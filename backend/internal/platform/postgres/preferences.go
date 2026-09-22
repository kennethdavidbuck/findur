package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/profile"
)

// PreferenceRepository persists display preferences independently of personal-profile completion.
type PreferenceRepository struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewPreferenceRepository(pool *pgxpool.Pool, now func() time.Time) *PreferenceRepository {
	return &PreferenceRepository{pool: pool, now: now}
}

func (r *PreferenceRepository) GetDisplayPreferences(ctx context.Context, owner uuid.UUID) (*profile.DisplayPreferences, error) {
	var value profile.DisplayPreferences
	err := r.pool.QueryRow(ctx, `SELECT locale,theme,version FROM owner_display_preferences WHERE user_id=$1`, owner).Scan(&value.Locale, &value.Theme, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *PreferenceRepository) SaveDisplayPreferences(ctx context.Context, owner uuid.UUID, input profile.DisplayPreferencesInput) (profile.DisplayPreferences, error) {
	now := r.now().UTC()
	var value profile.DisplayPreferences
	if input.ExpectedVersion == 0 {
		err := r.pool.QueryRow(ctx, `INSERT INTO owner_display_preferences (user_id,locale,theme,version,created_at,updated_at) VALUES ($1,$2,$3,1,$4,$4) ON CONFLICT (user_id) DO NOTHING RETURNING locale,theme,version`, owner, input.Locale, input.Theme, now).Scan(&value.Locale, &value.Theme, &value.Version)
		if errors.Is(err, pgx.ErrNoRows) {
			return profile.DisplayPreferences{}, profile.ErrConflict
		}
		return value, err
	}
	err := r.pool.QueryRow(ctx, `UPDATE owner_display_preferences SET locale=$3,theme=$4,version=version+1,updated_at=$5 WHERE user_id=$1 AND version=$2 RETURNING locale,theme,version`, owner, input.ExpectedVersion, input.Locale, input.Theme, now).Scan(&value.Locale, &value.Theme, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return profile.DisplayPreferences{}, profile.ErrConflict
	}
	return value, err
}
