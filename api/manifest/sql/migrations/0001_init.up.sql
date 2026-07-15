-- Gallery is greenfield and seed-owned. PostgreSQL 18 provides uuidv7().
-- Runtime identifiers are UUIDv7; public 22-character IDs are only an API encoding.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE gallery_site_settings (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    site_key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    search_placeholder TEXT NOT NULL,
    footer_tagline TEXT NOT NULL,
    random_batch_size INTEGER NOT NULL DEFAULT 30 CHECK (random_batch_size BETWEEN 12 AND 80),
    random_candidate_size INTEGER NOT NULL DEFAULT 240 CHECK (random_candidate_size BETWEEN 40 AND 2000),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_classification_catalogs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    catalog_key TEXT NOT NULL UNIQUE,
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gallery_classification_policy_profiles (
    catalog_id UUID NOT NULL REFERENCES gallery_classification_catalogs(id) ON DELETE RESTRICT,
    policy_key TEXT NOT NULL,
    schema_version INTEGER NOT NULL CHECK (schema_version > 0),
    policy_revision BIGINT NOT NULL CHECK (policy_revision > 0),
    document JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (catalog_id, policy_key),
    CONSTRAINT gallery_classification_policy_document_object_check
        CHECK (jsonb_typeof(document) = 'object')
);

CREATE TABLE gallery_categories (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    catalog_id UUID NOT NULL REFERENCES gallery_classification_catalogs(id) ON DELETE RESTRICT,
    parent_id UUID,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'inactive', 'replaced')),
    editorial_position INTEGER CHECK (editorial_position IS NULL OR editorial_position >= 0),
    replacement_id UUID,
    first_activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (catalog_id, id),
    UNIQUE (catalog_id, slug),
    FOREIGN KEY (catalog_id, parent_id) REFERENCES gallery_categories(catalog_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (catalog_id, replacement_id) REFERENCES gallery_categories(catalog_id, id) ON DELETE RESTRICT,
    CONSTRAINT gallery_category_replacement_state_check CHECK (
        (status = 'replaced' AND replacement_id IS NOT NULL) OR
        (status <> 'replaced' AND replacement_id IS NULL)
    ),
    CONSTRAINT gallery_category_self_replacement_check CHECK (replacement_id IS NULL OR replacement_id <> id)
);

CREATE INDEX gallery_categories_parent_idx
    ON gallery_categories (catalog_id, parent_id, editorial_position, id);
CREATE INDEX gallery_categories_replacement_idx
    ON gallery_categories (catalog_id, replacement_id)
    WHERE replacement_id IS NOT NULL;

CREATE TABLE gallery_facets (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    catalog_id UUID NOT NULL REFERENCES gallery_classification_catalogs(id) ON DELETE RESTRICT,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'inactive', 'replaced')),
    editorial_position INTEGER CHECK (editorial_position IS NULL OR editorial_position >= 0),
    replacement_id UUID,
    first_activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (catalog_id, id),
    UNIQUE (catalog_id, slug),
    FOREIGN KEY (catalog_id, replacement_id) REFERENCES gallery_facets(catalog_id, id) ON DELETE RESTRICT,
    CONSTRAINT gallery_facet_replacement_state_check CHECK (
        (status = 'replaced' AND replacement_id IS NOT NULL) OR
        (status <> 'replaced' AND replacement_id IS NULL)
    ),
    CONSTRAINT gallery_facet_self_replacement_check CHECK (replacement_id IS NULL OR replacement_id <> id)
);

CREATE TABLE gallery_facet_values (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    catalog_id UUID NOT NULL,
    facet_id UUID NOT NULL,
    parent_id UUID,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'inactive', 'replaced')),
    editorial_position INTEGER CHECK (editorial_position IS NULL OR editorial_position >= 0),
    replacement_id UUID,
    first_activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (catalog_id, facet_id, id),
    UNIQUE (catalog_id, facet_id, slug),
    FOREIGN KEY (catalog_id, facet_id) REFERENCES gallery_facets(catalog_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (catalog_id, facet_id, parent_id)
        REFERENCES gallery_facet_values(catalog_id, facet_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (catalog_id, facet_id, replacement_id)
        REFERENCES gallery_facet_values(catalog_id, facet_id, id) ON DELETE RESTRICT,
    CONSTRAINT gallery_facet_value_replacement_state_check CHECK (
        (status = 'replaced' AND replacement_id IS NOT NULL) OR
        (status <> 'replaced' AND replacement_id IS NULL)
    ),
    CONSTRAINT gallery_facet_value_self_replacement_check CHECK (replacement_id IS NULL OR replacement_id <> id)
);

CREATE INDEX gallery_facet_values_parent_idx
    ON gallery_facet_values (catalog_id, facet_id, parent_id, editorial_position, id);
CREATE INDEX gallery_facets_replacement_idx
    ON gallery_facets (catalog_id, replacement_id)
    WHERE replacement_id IS NOT NULL;
CREATE INDEX gallery_facet_values_replacement_idx
    ON gallery_facet_values (catalog_id, facet_id, replacement_id)
    WHERE replacement_id IS NOT NULL;

CREATE TABLE gallery_tags (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    catalog_id UUID NOT NULL REFERENCES gallery_classification_catalogs(id) ON DELETE RESTRICT,
    current_name TEXT NOT NULL,
    current_slug TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'replaced')),
    replacement_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (catalog_id, id),
    UNIQUE (catalog_id, current_slug),
    FOREIGN KEY (catalog_id, replacement_id) REFERENCES gallery_tags(catalog_id, id) ON DELETE RESTRICT,
    CONSTRAINT gallery_tag_replacement_state_check CHECK (
        (status = 'replaced' AND replacement_id IS NOT NULL) OR
        (status <> 'replaced' AND replacement_id IS NULL)
    ),
    CONSTRAINT gallery_tag_self_replacement_check CHECK (replacement_id IS NULL OR replacement_id <> id)
);

CREATE TABLE gallery_tag_lookup_entries (
    catalog_id UUID NOT NULL,
    lookup_key TEXT NOT NULL,
    target_tag_id UUID NOT NULL,
    source_tag_id UUID,
    kind TEXT NOT NULL CHECK (kind IN ('canonical', 'alias', 'replacement')),
    display_value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (catalog_id, lookup_key),
    FOREIGN KEY (catalog_id, target_tag_id) REFERENCES gallery_tags(catalog_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (catalog_id, source_tag_id) REFERENCES gallery_tags(catalog_id, id) ON DELETE RESTRICT,
    CONSTRAINT gallery_tag_lookup_replacement_source_check CHECK (
        (kind = 'replacement' AND source_tag_id IS NOT NULL) OR
        (kind <> 'replacement' AND source_tag_id IS NULL)
    )
);

CREATE INDEX gallery_tag_lookup_target_idx
    ON gallery_tag_lookup_entries (catalog_id, target_tag_id, lookup_key);
CREATE INDEX gallery_tag_lookup_source_idx
    ON gallery_tag_lookup_entries (catalog_id, source_tag_id)
    WHERE source_tag_id IS NOT NULL;
CREATE INDEX gallery_tags_replacement_idx
    ON gallery_tags (catalog_id, replacement_id)
    WHERE replacement_id IS NOT NULL;
CREATE INDEX gallery_tags_name_cursor_idx
    ON gallery_tags (LOWER(current_name), id);

CREATE TABLE gallery_classification_outbox (
    event_id UUID PRIMARY KEY DEFAULT uuidv7(),
    catalog_id UUID NOT NULL REFERENCES gallery_classification_catalogs(id) ON DELETE RESTRICT,
    revision BIGINT NOT NULL CHECK (revision > 0),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    UNIQUE (catalog_id, revision, event_id)
);

CREATE INDEX gallery_classification_outbox_unpublished_idx
    ON gallery_classification_outbox (occurred_at, event_id)
    WHERE published_at IS NULL;

CREATE FUNCTION gallery_notify_classification_catalog_changed()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM pg_notify(
        'classification_catalog_changed',
        NEW.catalog_key || ':' || NEW.revision::text
    );
    RETURN NEW;
END;
$$;

CREATE TRIGGER gallery_classification_catalog_changed
AFTER UPDATE OF revision ON gallery_classification_catalogs
FOR EACH ROW
WHEN (OLD.revision IS DISTINCT FROM NEW.revision)
EXECUTE FUNCTION gallery_notify_classification_catalog_changed();

CREATE TABLE gallery_submissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    subject_kind TEXT NOT NULL CHECK (subject_kind IN ('user', 'guest', 'operator')),
    subject_id TEXT NOT NULL,
    asset_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT,
    alt_text TEXT NOT NULL,
    processing_state TEXT NOT NULL DEFAULT 'queued' CHECK (processing_state IN ('queued', 'processing', 'ready', 'failed')),
    review_state TEXT NOT NULL DEFAULT 'pending' CHECK (review_state IN ('not_required', 'pending', 'approved', 'rejected')),
    safety_state TEXT NOT NULL DEFAULT 'pending' CHECK (safety_state IN ('pending', 'safe', 'uncertain', 'blocked', 'unavailable')),
    outcome TEXT NOT NULL DEFAULT 'pending' CHECK (outcome IN ('pending', 'published', 'duplicate', 'rejected', 'withdrawn', 'failed')),
    image_id UUID,
    exact_sha256 BYTEA,
    pdq_hash BIT(256),
	width INTEGER NOT NULL DEFAULT 0 CHECK (width >= 0),
	height INTEGER NOT NULL DEFAULT 0 CHECK (height >= 0),
	dominant_color TEXT NOT NULL DEFAULT '',
	public_rendition_ready BOOLEAN NOT NULL DEFAULT FALSE,
    failure_code TEXT NOT NULL DEFAULT '',
    review_note TEXT NOT NULL DEFAULT '',
    reviewed_by TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    claimed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT gallery_submission_source_url_check CHECK (source_url IS NULL OR source_url ~* '^https?://')
);

CREATE INDEX gallery_submissions_owner_idx ON gallery_submissions (subject_kind, subject_id, created_at DESC);
CREATE INDEX gallery_submissions_review_idx ON gallery_submissions (review_state, created_at ASC) WHERE outcome = 'pending';

CREATE TABLE gallery_submission_category_assignments (
    submission_id UUID NOT NULL REFERENCES gallery_submissions(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES gallery_categories(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (submission_id, category_id)
);

CREATE INDEX gallery_submission_category_filter_idx
    ON gallery_submission_category_assignments (category_id, submission_id);

CREATE TABLE gallery_submission_primary_categories (
    submission_id UUID PRIMARY KEY REFERENCES gallery_submissions(id) ON DELETE CASCADE,
    category_id UUID NOT NULL,
    FOREIGN KEY (submission_id, category_id)
        REFERENCES gallery_submission_category_assignments(submission_id, category_id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE gallery_submission_facet_assignments (
    submission_id UUID NOT NULL REFERENCES gallery_submissions(id) ON DELETE CASCADE,
    facet_value_id UUID NOT NULL REFERENCES gallery_facet_values(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (submission_id, facet_value_id)
);

CREATE INDEX gallery_submission_facet_filter_idx
    ON gallery_submission_facet_assignments (facet_value_id, submission_id);

CREATE TABLE gallery_submission_tag_assignments (
    submission_id UUID NOT NULL REFERENCES gallery_submissions(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES gallery_tags(id) ON DELETE RESTRICT,
    PRIMARY KEY (submission_id, tag_id)
);

CREATE INDEX gallery_submission_tag_filter_idx
    ON gallery_submission_tag_assignments (tag_id, submission_id);

CREATE TABLE gallery_tag_proposals (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    submission_id UUID NOT NULL REFERENCES gallery_submissions(id) ON DELETE CASCADE,
    input_value TEXT NOT NULL,
    lookup_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    resolved_tag_id UUID REFERENCES gallery_tags(id) ON DELETE RESTRICT,
    reviewed_by TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (submission_id, lookup_key),
    CONSTRAINT gallery_tag_proposal_resolution_check CHECK (
        (status = 'approved' AND resolved_tag_id IS NOT NULL) OR
        (status <> 'approved' AND resolved_tag_id IS NULL)
    )
);

CREATE INDEX gallery_tag_proposals_queue_idx
    ON gallery_tag_proposals (status, created_at, id);
CREATE INDEX gallery_tag_proposals_resolved_idx
    ON gallery_tag_proposals (resolved_tag_id)
    WHERE resolved_tag_id IS NOT NULL;

CREATE TABLE gallery_images (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    asset_id UUID NOT NULL UNIQUE,
    origin_submission_id UUID UNIQUE REFERENCES gallery_submissions(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT,
    alt_text TEXT NOT NULL,
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    dominant_color TEXT NOT NULL DEFAULT '',
    focus_x NUMERIC(5,4) NOT NULL DEFAULT 0.5 CHECK (focus_x BETWEEN 0 AND 1),
    focus_y NUMERIC(5,4) NOT NULL DEFAULT 0.5 CHECK (focus_y BETWEEN 0 AND 1),
    processing_state TEXT NOT NULL DEFAULT 'queued' CHECK (processing_state IN ('queued', 'processing', 'ready', 'failed')),
    review_state TEXT NOT NULL DEFAULT 'pending' CHECK (review_state IN ('not_required', 'pending', 'approved', 'rejected')),
    publication_state TEXT NOT NULL DEFAULT 'draft' CHECK (publication_state IN ('draft', 'published', 'hidden', 'deleted')),
    safety_state TEXT NOT NULL DEFAULT 'pending' CHECK (safety_state IN ('pending', 'safe', 'uncertain', 'blocked', 'unavailable')),
    public_rendition_ready BOOLEAN NOT NULL DEFAULT FALSE,
    exact_sha256 BYTEA,
    pdq_hash BIT(256),
    published_at TIMESTAMPTZ,
    hidden_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT gallery_image_source_url_check CHECK (source_url IS NULL OR source_url ~* '^https?://'),
    CONSTRAINT gallery_image_publication_time_check CHECK (publication_state <> 'published' OR published_at IS NOT NULL)
);

ALTER TABLE gallery_submissions
    ADD CONSTRAINT gallery_submissions_image_fk FOREIGN KEY (image_id) REFERENCES gallery_images(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX gallery_images_exact_sha_uidx ON gallery_images (exact_sha256) WHERE exact_sha256 IS NOT NULL;
CREATE INDEX gallery_images_pdq_hnsw_idx ON gallery_images USING hnsw (pdq_hash bit_hamming_ops)
    WHERE pdq_hash IS NOT NULL;
CREATE INDEX gallery_images_public_idx ON gallery_images (published_at DESC, id DESC)
    WHERE processing_state = 'ready'
      AND review_state IN ('not_required', 'approved')
      AND publication_state = 'published'
      AND safety_state = 'safe'
      AND public_rendition_ready;

CREATE TABLE gallery_image_category_assignments (
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES gallery_categories(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (image_id, category_id)
);

CREATE INDEX gallery_image_category_filter_idx
    ON gallery_image_category_assignments (category_id, image_id);

CREATE TABLE gallery_image_primary_categories (
    image_id UUID PRIMARY KEY REFERENCES gallery_images(id) ON DELETE CASCADE,
    category_id UUID NOT NULL,
    FOREIGN KEY (image_id, category_id)
        REFERENCES gallery_image_category_assignments(image_id, category_id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE gallery_image_facet_assignments (
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    facet_value_id UUID NOT NULL REFERENCES gallery_facet_values(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (image_id, facet_value_id)
);

CREATE INDEX gallery_image_facet_filter_idx
    ON gallery_image_facet_assignments (facet_value_id, image_id);

CREATE TABLE gallery_image_tag_assignments (
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES gallery_tags(id) ON DELETE RESTRICT,
    PRIMARY KEY (image_id, tag_id)
);

CREATE INDEX gallery_image_tag_filter_idx
    ON gallery_image_tag_assignments (tag_id, image_id);

CREATE TABLE gallery_collections (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    kind TEXT NOT NULL CHECK (kind IN ('gallery.editorial', 'gallery.favorites')),
    resource_kind TEXT NOT NULL DEFAULT 'gallery.image' CHECK (resource_kind = 'gallery.image'),
    owner_kind TEXT NOT NULL CHECK (owner_kind IN ('site', 'user')),
    owner_id TEXT NOT NULL,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'public')),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT gallery_collection_kind_policy_check CHECK (
        (kind = 'gallery.favorites' AND owner_kind = 'user' AND visibility = 'private') OR
        (kind = 'gallery.editorial' AND owner_kind = 'site')
    )
);

CREATE UNIQUE INDEX gallery_favorites_singleton_uidx ON gallery_collections (owner_id)
    WHERE kind = 'gallery.favorites';

CREATE TABLE gallery_collection_editorial (
    collection_id UUID PRIMARY KEY REFERENCES gallery_collections(id) ON DELETE CASCADE,
    slug TEXT NOT NULL UNIQUE,
    cover_image_id UUID REFERENCES gallery_images(id) ON DELETE SET NULL,
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE gallery_collection_members (
    collection_id UUID NOT NULL REFERENCES gallery_collections(id) ON DELETE CASCADE,
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    manual_position INTEGER CHECK (manual_position IS NULL OR manual_position >= 0),
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, image_id)
);

CREATE INDEX gallery_collection_members_manual_idx ON gallery_collection_members (collection_id, manual_position, added_at DESC);

CREATE TABLE gallery_cases (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    image_id UUID REFERENCES gallery_images(id) ON DELETE SET NULL,
    submission_id UUID REFERENCES gallery_submissions(id) ON DELETE SET NULL,
    kind TEXT NOT NULL CHECK (kind IN ('report', 'source_correction', 'safety_uncertain', 'near_duplicate', 'takedown')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'reviewing', 'resolved', 'dismissed')),
    reporter_kind TEXT CHECK (reporter_kind IS NULL OR reporter_kind IN ('user', 'guest', 'operator')),
    reporter_id TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    proposed_source_url TEXT,
    operator_sub TEXT NOT NULL DEFAULT '',
    resolution_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    CONSTRAINT gallery_case_source_url_check CHECK (proposed_source_url IS NULL OR proposed_source_url ~* '^https?://')
);

CREATE INDEX gallery_cases_queue_idx ON gallery_cases (status, created_at ASC);

CREATE TABLE gallery_image_events (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    image_id UUID REFERENCES gallery_images(id) ON DELETE SET NULL,
    subject_key TEXT NOT NULL DEFAULT '',
    session_key TEXT NOT NULL DEFAULT '',
    event_type TEXT NOT NULL CHECK (event_type IN ('grid_exposure', 'qualified_view', 'favorite', 'unfavorite', 'share', 'report')),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX gallery_image_events_window_idx ON gallery_image_events (event_type, occurred_at DESC);

CREATE TABLE gallery_image_metrics_daily (
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    exposures BIGINT NOT NULL DEFAULT 0,
    qualified_views BIGINT NOT NULL DEFAULT 0,
    unique_visitors BIGINT NOT NULL DEFAULT 0,
    favorites BIGINT NOT NULL DEFAULT 0,
    shares BIGINT NOT NULL DEFAULT 0,
    reports BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (image_id, metric_date)
);

CREATE TABLE gallery_ranking_snapshots (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    ranking_kind TEXT NOT NULL CHECK (ranking_kind IN ('trending', 'most_viewed', 'most_favorited')),
    window_key TEXT NOT NULL CHECK (window_key IN ('24h', '7d', '30d', 'all')),
    generated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE (ranking_kind, window_key, generated_at)
);

CREATE TABLE gallery_ranking_entries (
    snapshot_id UUID NOT NULL REFERENCES gallery_ranking_snapshots(id) ON DELETE CASCADE,
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    rank INTEGER NOT NULL CHECK (rank > 0),
    score DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (snapshot_id, image_id),
    UNIQUE (snapshot_id, rank)
);

CREATE TABLE gallery_image_tombstones (
    image_id UUID PRIMARY KEY,
    reason TEXT NOT NULL DEFAULT '',
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
