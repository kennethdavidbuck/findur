CREATE TABLE profile_locations (
    key text PRIMARY KEY,
    city_en text NOT NULL,
    city_fr text NOT NULL,
    province_en text NOT NULL,
    province_fr text NOT NULL,
    latitude double precision NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude double precision NOT NULL CHECK (longitude BETWEEN -180 AND 180)
);

INSERT INTO profile_locations (key, city_en, city_fr, province_en, province_fr, latitude, longitude) VALUES
    ('halifax-ns', 'Halifax', 'Halifax', 'Nova Scotia', 'Nouvelle-Écosse', 44.6488, -63.5752),
    ('montreal-qc', 'Montréal', 'Montréal', 'Quebec', 'Québec', 45.5019, -73.5674),
    ('ottawa-on', 'Ottawa', 'Ottawa', 'Ontario', 'Ontario', 45.4215, -75.6972),
    ('toronto-on', 'Toronto', 'Toronto', 'Ontario', 'Ontario', 43.6532, -79.3832),
    ('calgary-ab', 'Calgary', 'Calgary', 'Alberta', 'Alberta', 51.0447, -114.0719),
    ('vancouver-bc', 'Vancouver', 'Vancouver', 'British Columbia', 'Colombie-Britannique', 49.2827, -123.1207);

CREATE TABLE personal_profiles (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name text NOT NULL CHECK (char_length(display_name) BETWEEN 1 AND 60),
    adult_attested_at timestamptz NOT NULL,
    location_key text NOT NULL REFERENCES profile_locations(key),
    relationship_intent text NOT NULL CHECK (relationship_intent IN ('long-term', 'open-to-long-term', 'figuring-it-out')),
    biography text NOT NULL CHECK (char_length(biography) BETWEEN 1 AND 500),
    avatar_key text NOT NULL CHECK (avatar_key IN ('aurora', 'cedar', 'ember', 'harbour', 'meadow', 'solstice')),
    locale text NOT NULL CHECK (locale IN ('en', 'fr')),
    theme text NOT NULL CHECK (theme IN ('system', 'light', 'dark')),
    version bigint NOT NULL CHECK (version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
