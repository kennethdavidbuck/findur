package profile

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// DisplayPreferences are the independently persisted, non-sensitive display settings for one owner.
type DisplayPreferences struct {
	Locale  string
	Theme   string
	Version int64
}

// DisplayPreferencesInput supports optimistic replacement. Version zero creates an absent record.
type DisplayPreferencesInput struct {
	Locale, Theme   string
	ExpectedVersion int64
}

// PreferenceRepository is deliberately separate from complete personal profile persistence.
type PreferenceRepository interface {
	GetDisplayPreferences(context.Context, uuid.UUID) (*DisplayPreferences, error)
	SaveDisplayPreferences(context.Context, uuid.UUID, DisplayPreferencesInput) (DisplayPreferences, error)
}

// PreferenceService validates owner-scoped display preferences independently of profile completion.
type PreferenceService struct{ repository PreferenceRepository }

func NewPreferenceService(repository PreferenceRepository) (*PreferenceService, error) {
	if repository == nil {
		return nil, errors.New("incomplete preference service configuration")
	}
	return &PreferenceService{repository: repository}, nil
}

func (s *PreferenceService) Get(ctx context.Context, actor auth.Actor) (*DisplayPreferences, error) {
	return s.repository.GetDisplayPreferences(ctx, actor.UserID())
}

func (s *PreferenceService) Save(ctx context.Context, actor auth.Actor, input DisplayPreferencesInput) (DisplayPreferences, error) {
	if !allowed(locales, input.Locale) || !allowed(themes, input.Theme) || input.ExpectedVersion < 0 {
		return DisplayPreferences{}, ErrInvalid
	}
	return s.repository.SaveDisplayPreferences(ctx, actor.UserID(), input)
}
