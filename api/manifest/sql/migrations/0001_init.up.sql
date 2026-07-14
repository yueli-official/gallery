CREATE TABLE gallery_site_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    search_placeholder TEXT NOT NULL,
    footer_tagline TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_creator_profiles (
    id TEXT PRIMARY KEY,
    account_sub TEXT NOT NULL UNIQUE,
    handle TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    cover_asset_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_categories (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0)
);

CREATE TABLE gallery_artworks (
    id TEXT PRIMARY KEY,
    legacy_id BIGINT,
    creator_id TEXT NOT NULL REFERENCES gallery_creator_profiles(id),
    category_id TEXT REFERENCES gallery_categories(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cover_asset_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'published', 'rejected', 'restricted', 'archived')),
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'unlisted', 'private')),
    content_rating TEXT NOT NULL DEFAULT 'general' CHECK (content_rating IN ('general', 'sensitive', 'adult')),
    ai_usage TEXT NOT NULL DEFAULT 'none' CHECK (ai_usage IN ('none', 'assistive', 'mostly_generated')),
    ai_training_permission TEXT NOT NULL DEFAULT 'unspecified' CHECK (ai_training_permission IN ('unspecified', 'allow', 'disallow')),
    rights_basis TEXT NOT NULL DEFAULT 'original' CHECK (rights_basis IN ('original', 'authorized_repost', 'public_domain', 'licensed_material')),
    license TEXT NOT NULL DEFAULT 'all_rights_reserved',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (legacy_id)
);

CREATE INDEX gallery_artworks_public_idx ON gallery_artworks (published_at DESC, id DESC)
    WHERE status = 'published' AND visibility = 'public';

CREATE TABLE gallery_artwork_assets (
    id TEXT PRIMARY KEY,
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    asset_id TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    width INTEGER NOT NULL DEFAULT 0 CHECK (width >= 0),
    height INTEGER NOT NULL DEFAULT 0 CHECK (height >= 0),
    format TEXT NOT NULL DEFAULT '',
    alt_text TEXT NOT NULL DEFAULT '',
    placeholder_url TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    card_url TEXT NOT NULL DEFAULT '',
    detail_url TEXT NOT NULL DEFAULT '',
    original_url TEXT NOT NULL DEFAULT '',
    content_rating TEXT NOT NULL DEFAULT 'general' CHECK (content_rating IN ('general', 'sensitive', 'adult')),
    ai_usage TEXT NOT NULL DEFAULT 'none' CHECK (ai_usage IN ('none', 'assistive', 'mostly_generated')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (artwork_id, sort_order),
    UNIQUE (artwork_id, asset_id)
);

CREATE TABLE gallery_tags (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_tag_aliases (
    alias TEXT PRIMARY KEY,
    tag_id TEXT NOT NULL REFERENCES gallery_tags(id) ON DELETE CASCADE
);

CREATE TABLE gallery_artwork_tags (
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    tag_id TEXT NOT NULL REFERENCES gallery_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (artwork_id, tag_id)
);

CREATE TABLE gallery_series (
    id TEXT PRIMARY KEY,
    creator_id TEXT NOT NULL REFERENCES gallery_creator_profiles(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'unlisted', 'private')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_series_items (
    series_id TEXT NOT NULL REFERENCES gallery_series(id) ON DELETE CASCADE,
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    PRIMARY KEY (series_id, artwork_id),
    UNIQUE (series_id, sort_order)
);

CREATE TABLE gallery_follows (
    account_sub TEXT NOT NULL,
    creator_id TEXT NOT NULL REFERENCES gallery_creator_profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_sub, creator_id)
);

CREATE TABLE gallery_bookmarks (
    account_sub TEXT NOT NULL,
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    visibility TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('public', 'private')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_sub, artwork_id)
);

CREATE TABLE gallery_collections (
    id TEXT PRIMARY KEY,
    account_sub TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('public', 'private')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_collection_items (
    collection_id TEXT NOT NULL REFERENCES gallery_collections(id) ON DELETE CASCADE,
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, artwork_id)
);

CREATE TABLE gallery_view_history (
    account_sub TEXT NOT NULL,
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    last_viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_sub, artwork_id)
);

CREATE TABLE gallery_interaction_events (
    id TEXT PRIMARY KEY,
    artwork_id TEXT REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    account_sub TEXT,
    session_key TEXT,
    event_type TEXT NOT NULL CHECK (event_type IN ('impression', 'qualified_view', 'bookmark', 'follow_from_artwork', 'share')),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX gallery_interaction_events_window_idx ON gallery_interaction_events (event_type, occurred_at DESC);

CREATE TABLE gallery_editorial_features (
    id TEXT PRIMARY KEY,
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    placement TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ,
    operator_sub TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_reports (
    id TEXT PRIMARY KEY,
    artwork_id TEXT REFERENCES gallery_artworks(id),
    reporter_sub TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'reviewing', 'resolved', 'dismissed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_moderation_cases (
    id TEXT PRIMARY KEY,
    artwork_id TEXT REFERENCES gallery_artworks(id),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'approved', 'rejected', 'restricted', 'appealed', 'closed')),
    reason TEXT NOT NULL DEFAULT '',
    operator_sub TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

