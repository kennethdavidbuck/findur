package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/profile"
)

func TestProfileRepositoryCreatesReplacesAndIsolatesOwners(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner, other := uuid.New(), uuid.New()
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO users (id,origin) VALUES ($1,'oauth'),($2,'oauth')`, owner, other); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewProfileRepository(fixture.pool)
	input := profile.Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system"}
	created, err := repository.Save(fixture.ctx, owner, input, fixture.now)
	if err != nil || created.Version != 1 || !created.AdultAttestedAt.Equal(fixture.now) {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	otherProfile, locations, err := repository.Get(fixture.ctx, other)
	if err != nil || otherProfile != nil || len(locations) != 6 {
		t.Fatalf("other=%+v locations=%d err=%v", otherProfile, len(locations), err)
	}
	input.DisplayName = "Alexis"
	input.LocationKey = "montreal-qc"
	input.RelationshipIntent = "figuring-it-out"
	input.Biography = "Updated biography"
	input.AvatarKey = "cedar"
	input.Locale = "fr"
	input.Theme = "dark"
	input.ExpectedVersion = 1
	updated, err := repository.Save(fixture.ctx, owner, input, fixture.now.Add(1))
	if err != nil || updated.Version != 2 || !updated.AdultAttestedAt.Equal(fixture.now) || !matchesMutableProfile(updated, input) {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if _, err := repository.Save(fixture.ctx, owner, input, fixture.now.Add(2)); !errors.Is(err, profile.ErrConflict) {
		t.Fatalf("stale save error=%v", err)
	}
	loaded, _, err := repository.Get(fixture.ctx, owner)
	if err != nil || loaded == nil || loaded.Version != 2 || !loaded.AdultAttestedAt.Equal(fixture.now) || !matchesMutableProfile(*loaded, input) {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
}

func matchesMutableProfile(value profile.Profile, input profile.Input) bool {
	return value.DisplayName == input.DisplayName &&
		value.LocationKey == input.LocationKey &&
		value.RelationshipIntent == input.RelationshipIntent &&
		value.Biography == input.Biography &&
		value.AvatarKey == input.AvatarKey &&
		value.Locale == input.Locale &&
		value.Theme == input.Theme
}

func TestProfileRepositoryRejectsACompleteInvalidCreateAtomically(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := uuid.New()
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO users (id,origin) VALUES ($1,'oauth')`, owner); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewProfileRepository(fixture.pool)
	input := profile.Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "not-seeded", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system"}
	if _, err := repository.Save(fixture.ctx, owner, input, fixture.now); err == nil {
		t.Fatal("invalid location was persisted")
	}
	loaded, _, err := repository.Get(fixture.ctx, owner)
	if err != nil || loaded != nil {
		t.Fatalf("partial profile=%+v err=%v", loaded, err)
	}
}
