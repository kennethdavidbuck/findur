package profile

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

type repositoryStub struct {
	getOwner  uuid.UUID
	saveOwner uuid.UUID
	saved     Input
	saves     int
	err       error
}

func (r *repositoryStub) Get(_ context.Context, owner uuid.UUID) (*Profile, []Location, error) {
	r.getOwner = owner
	return nil, nil, nil
}
func (r *repositoryStub) Save(_ context.Context, owner uuid.UUID, input Input, _ time.Time) (Profile, error) {
	r.saveOwner = owner
	r.saved = input
	r.saves++
	return Profile{Version: 1}, r.err
}

type actorSessionRepository struct{ owner uuid.UUID }

func (r actorSessionRepository) FindActive(context.Context, []byte, time.Time) (auth.Session, error) {
	return auth.Session{UserID: r.owner}, nil
}
func (r actorSessionRepository) AuthenticateAndTouch(context.Context, []byte, time.Time, time.Time) (auth.Session, error) {
	return auth.Session{UserID: r.owner}, nil
}
func (actorSessionRepository) RevokeCurrent(context.Context, []byte, time.Time) (bool, error) {
	return true, nil
}
func (actorSessionRepository) CleanupSessions(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestSaveRejectsIncompleteProfileWithoutPersistence(t *testing.T) {
	repository := &repositoryStub{}
	service, _ := NewService(repository, time.Now)
	_, err := service.Save(context.Background(), auth.Actor{}, Input{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Save() error = %v", err)
	}
	if repository.saves != 0 {
		t.Fatal("invalid input was persisted")
	}
}

func TestSaveRejectsEveryUnsupportedCatalogueValue(t *testing.T) {
	valid := Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system"}
	tests := map[string]func(*Input){
		"location": func(input *Input) { input.LocationKey = "precise-coordinates" },
		"intent":   func(input *Input) { input.RelationshipIntent = "unsupported" },
		"avatar":   func(input *Input) { input.AvatarKey = "https://example.com/avatar.png" },
		"locale":   func(input *Input) { input.Locale = "es" },
		"theme":    func(input *Input) { input.Theme = "neon" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repository := &repositoryStub{}
			service, _ := NewService(repository, time.Now)
			input := valid
			mutate(&input)
			if _, err := service.Save(context.Background(), auth.Actor{}, input); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Save() error = %v", err)
			}
			if repository.saves != 0 {
				t.Fatal("unsupported value reached persistence")
			}
		})
	}
}

func TestSaveRejectsNULAndDisplayEmptyText(t *testing.T) {
	valid := Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system"}
	tests := []struct {
		name, field string
		mutate      func(*Input)
	}{
		{name: "display name NUL", field: "displayName", mutate: func(input *Input) { input.DisplayName = "Alex\x00" }},
		{name: "biography NUL", field: "biography", mutate: func(input *Input) { input.Biography = "Hello\x00" }},
		{name: "format-only display name", field: "displayName", mutate: func(input *Input) { input.DisplayName = "\u200b\u2060" }},
		{name: "control-only biography", field: "biography", mutate: func(input *Input) { input.Biography = "\u0001\u0002" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &repositoryStub{}
			service, _ := NewService(repository, time.Now)
			input := valid
			test.mutate(&input)
			_, err := service.Save(context.Background(), auth.Actor{}, input)
			var invalid *ValidationError
			if !errors.As(err, &invalid) || !contains(invalid.Fields, test.field) {
				t.Fatalf("Save() error = %#v", err)
			}
			if repository.saves != 0 {
				t.Fatal("invalid input was persisted")
			}
		})
	}
}

func TestActorIdentityReachesRepositoryReadsAndWrites(t *testing.T) {
	owner := uuid.New()
	sessions, err := auth.NewSessionService(auth.SessionConfig{HashKey: bytes.Repeat([]byte{7}, 32), Clock: time.Now}, actorSessionRepository{owner: owner})
	if err != nil {
		t.Fatal(err)
	}
	actor, err := sessions.Authenticate(context.Background(), "opaque-session")
	if err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{}
	service, _ := NewService(repository, time.Now)
	if _, err := service.Get(context.Background(), actor); err != nil {
		t.Fatal(err)
	}
	input := Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system"}
	if _, err := service.Save(context.Background(), actor, input); err != nil {
		t.Fatal(err)
	}
	if repository.getOwner != owner || repository.saveOwner != owner {
		t.Fatalf("owners: get=%s save=%s want=%s", repository.getOwner, repository.saveOwner, owner)
	}
}

func TestSavePreservesConflictSemantics(t *testing.T) {
	repository := &repositoryStub{err: ErrConflict}
	service, _ := NewService(repository, time.Now)
	input := Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system", ExpectedVersion: 2}
	if _, err := service.Save(context.Background(), auth.Actor{}, input); !errors.Is(err, ErrConflict) {
		t.Fatalf("Save() error = %v", err)
	}
}

func TestSaveAcceptsCompleteAllowlistedProfile(t *testing.T) {
	repository := &repositoryStub{}
	service, _ := NewService(repository, time.Now)
	input := Input{DisplayName: " Alex ", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: " Hello ", AvatarKey: "aurora", Locale: "en", Theme: "system"}
	if _, err := service.Save(context.Background(), auth.Actor{}, input); err != nil {
		t.Fatal(err)
	}
	if repository.saved.DisplayName != "Alex" || repository.saved.Biography != "Hello" {
		t.Fatalf("input not normalized: %#v", repository.saved)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
