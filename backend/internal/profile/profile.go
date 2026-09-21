// Package profile owns the authenticated owner's minimum personal profile.
package profile

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

var (
	// ErrConflict means a replacement used a stale profile version.
	ErrConflict = errors.New("profile version conflict")
	// ErrInvalid means one or more bounded profile values are invalid.
	ErrInvalid = errors.New("invalid profile")
)

// Location is a migration-seeded coarse Canadian location available to profiles.
type Location struct {
	Key, CityEN, CityFR, ProvinceEN, ProvinceFR string
}

// Profile is the complete persisted personal profile for one owner.
type Profile struct {
	DisplayName        string
	AdultAttestedAt    time.Time
	LocationKey        string
	RelationshipIntent string
	Biography          string
	AvatarKey          string
	Locale             string
	Theme              string
	Version            int64
}

// Input contains the bounded values required to create or replace a profile.
type Input struct {
	DisplayName, LocationKey, RelationshipIntent, Biography, AvatarKey, Locale, Theme string
	AdultAttested                                                                     bool
	ExpectedVersion                                                                   int64
}

// Snapshot combines an owner's optional profile with the safe location catalogue.
type Snapshot struct {
	Profile   *Profile
	Locations []Location
}

// ValidationError names the profile fields that failed domain validation.
type ValidationError struct{ Fields []string }

func (e *ValidationError) Error() string { return ErrInvalid.Error() }

// Is allows callers to classify a ValidationError as ErrInvalid.
func (e *ValidationError) Is(target error) bool { return target == ErrInvalid }

// Repository provides actor-scoped profile persistence to the domain service.
type Repository interface {
	Get(context.Context, uuid.UUID) (*Profile, []Location, error)
	Save(context.Context, uuid.UUID, Input, time.Time) (Profile, error)
}

// Service validates and coordinates personal-profile reads and writes.
type Service struct {
	repository Repository
	clock      func() time.Time
}

// NewService creates a profile service from complete dependencies.
func NewService(repository Repository, clock func() time.Time) (*Service, error) {
	if repository == nil || clock == nil {
		return nil, errors.New("incomplete profile service configuration")
	}
	return &Service{repository: repository, clock: clock}, nil
}

// Get returns the authenticated actor's profile and safe location catalogue.
func (s *Service) Get(ctx context.Context, actor auth.Actor) (Snapshot, error) {
	profile, locations, err := s.repository.Get(ctx, actor.UserID())
	return Snapshot{Profile: profile, Locations: locations}, err
}

// Save validates and atomically creates or replaces the authenticated actor's profile.
func (s *Service) Save(ctx context.Context, actor auth.Actor, input Input) (Profile, error) {
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Biography = strings.TrimSpace(input.Biography)
	if fields := invalidFields(input); len(fields) > 0 {
		return Profile{}, &ValidationError{Fields: fields}
	}
	return s.repository.Save(ctx, actor.UserID(), input, s.clock().UTC())
}

func invalidFields(input Input) []string {
	var fields []string
	if count := utf8.RuneCountInString(input.DisplayName); count < 1 || count > 60 || strings.ContainsRune(input.DisplayName, '\x00') || displayEmpty(input.DisplayName) {
		fields = append(fields, "displayName")
	}
	if !input.AdultAttested {
		fields = append(fields, "adultAttested")
	}
	if !allowed(locationKeys, input.LocationKey) {
		fields = append(fields, "locationKey")
	}
	if !allowed(intents, input.RelationshipIntent) {
		fields = append(fields, "relationshipIntent")
	}
	if count := utf8.RuneCountInString(input.Biography); count < 1 || count > 500 || strings.ContainsRune(input.Biography, '\x00') || displayEmpty(input.Biography) {
		fields = append(fields, "biography")
	}
	if !allowed(avatars, input.AvatarKey) {
		fields = append(fields, "avatarKey")
	}
	if !allowed(locales, input.Locale) {
		fields = append(fields, "locale")
	}
	if !allowed(themes, input.Theme) {
		fields = append(fields, "theme")
	}
	if input.ExpectedVersion < 0 {
		fields = append(fields, "expectedVersion")
	}
	return fields
}

func allowed(values map[string]struct{}, value string) bool { _, ok := values[value]; return ok }

func displayEmpty(value string) bool {
	for _, character := range value {
		if !unicode.IsSpace(character) && !unicode.IsControl(character) && !unicode.Is(unicode.Cf, character) {
			return false
		}
	}
	return true
}

var locationKeys = set("halifax-ns", "montreal-qc", "ottawa-on", "toronto-on", "calgary-ab", "vancouver-bc")
var intents = set("long-term", "open-to-long-term", "figuring-it-out")
var avatars = set("aurora", "cedar", "ember", "harbour", "meadow", "solstice")
var locales = set("en", "fr")
var themes = set("system", "light", "dark")

func set(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
