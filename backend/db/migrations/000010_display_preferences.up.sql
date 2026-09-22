CREATE TABLE owner_display_preferences (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    locale text NOT NULL CHECK (locale IN ('en', 'fr')),
    theme text NOT NULL CHECK (theme IN ('system', 'light', 'dark')),
    version bigint NOT NULL CHECK (version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
