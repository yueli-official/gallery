CREATE TABLE gallery_spotlight_characters (
    id UUID PRIMARY KEY,
    position INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL,
    series TEXT NOT NULL DEFAULT '',
    faction TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL,
    thumb_url TEXT NOT NULL DEFAULT '',
    accent TEXT NOT NULL DEFAULT '',
    wash TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX gallery_spotlight_characters_position_idx
    ON gallery_spotlight_characters (enabled, position, id);

CREATE TABLE gallery_category_covers (
    category_id UUID PRIMARY KEY REFERENCES gallery_categories(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    thumb_url TEXT NOT NULL DEFAULT '',
    accent TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
