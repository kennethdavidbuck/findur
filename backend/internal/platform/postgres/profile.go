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

// ProfileRepository persists personal profiles and reads seeded profile catalogues.
type ProfileRepository struct{ pool *pgxpool.Pool }

// NewProfileRepository creates a PostgreSQL-backed profile repository.
func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

// Get returns one owner's optional profile and the safe location catalogue.
func (r *ProfileRepository) Get(ctx context.Context, owner uuid.UUID) (*profile.Profile, []profile.Location, error) {
	locations, err := r.locations(ctx)
	if err != nil {
		return nil, nil, err
	}
	var result profile.Profile
	err = r.pool.QueryRow(ctx, `SELECT display_name,adult_attested_at,location_key,relationship_intent,biography,avatar_key,locale,theme,version FROM personal_profiles WHERE user_id=$1`, owner).
		Scan(&result.DisplayName, &result.AdultAttestedAt, &result.LocationKey, &result.RelationshipIntent, &result.Biography, &result.AvatarKey, &result.Locale, &result.Theme, &result.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, locations, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &result, locations, nil
}

func (r *ProfileRepository) locations(ctx context.Context) ([]profile.Location, error) {
	rows, err := r.pool.Query(ctx, `SELECT key,city_en,city_fr,province_en,province_fr FROM profile_locations ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []profile.Location
	for rows.Next() {
		var location profile.Location
		if err := rows.Scan(&location.Key, &location.CityEN, &location.CityFR, &location.ProvinceEN, &location.ProvinceFR); err != nil {
			return nil, err
		}
		result = append(result, location)
	}
	return result, rows.Err()
}

// Save atomically creates or replaces one active owner's complete profile.
func (r *ProfileRepository) Save(ctx context.Context, owner uuid.UUID, input profile.Input, now time.Time) (profile.Profile, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return profile.Profile{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var active bool
	if err := tx.QueryRow(ctx, `SELECT active FROM users WHERE id=$1 FOR UPDATE`, owner).Scan(&active); err != nil || !active {
		if errors.Is(err, pgx.ErrNoRows) || !active {
			return profile.Profile{}, profile.ErrInvalid
		}
		return profile.Profile{}, err
	}
	var result profile.Profile
	if input.ExpectedVersion == 0 {
		err = tx.QueryRow(ctx, `INSERT INTO personal_profiles (user_id,display_name,adult_attested_at,location_key,relationship_intent,biography,avatar_key,locale,theme,version,created_at,updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,1,$3,$3)
			ON CONFLICT (user_id) DO NOTHING
			RETURNING display_name,adult_attested_at,location_key,relationship_intent,biography,avatar_key,locale,theme,version`, owner, input.DisplayName, now, input.LocationKey, input.RelationshipIntent, input.Biography, input.AvatarKey, input.Locale, input.Theme).
			Scan(&result.DisplayName, &result.AdultAttestedAt, &result.LocationKey, &result.RelationshipIntent, &result.Biography, &result.AvatarKey, &result.Locale, &result.Theme, &result.Version)
	} else {
		err = tx.QueryRow(ctx, `UPDATE personal_profiles SET display_name=$3,location_key=$4,relationship_intent=$5,biography=$6,avatar_key=$7,locale=$8,theme=$9,version=version+1,updated_at=$10
			WHERE user_id=$1 AND version=$2
			RETURNING display_name,adult_attested_at,location_key,relationship_intent,biography,avatar_key,locale,theme,version`, owner, input.ExpectedVersion, input.DisplayName, input.LocationKey, input.RelationshipIntent, input.Biography, input.AvatarKey, input.Locale, input.Theme, now).
			Scan(&result.DisplayName, &result.AdultAttestedAt, &result.LocationKey, &result.RelationshipIntent, &result.Biography, &result.AvatarKey, &result.Locale, &result.Theme, &result.Version)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return profile.Profile{}, profile.ErrConflict
	}
	if err != nil {
		return profile.Profile{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return profile.Profile{}, err
	}
	return result, nil
}
