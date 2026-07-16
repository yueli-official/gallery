-- Gallery scale baseline: keep medium-sized catalog search and operational queues
-- inside PostgreSQL until the documented external-search triggers are reached.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX gallery_images_public_title_idx
    ON gallery_images (LOWER(title), id)
    WHERE processing_state = 'ready'
      AND review_state IN ('not_required', 'approved')
      AND publication_state = 'published'
      AND safety_state = 'safe'
      AND public_rendition_ready;

CREATE INDEX gallery_images_public_title_search_idx
    ON gallery_images USING GIN (title gin_trgm_ops)
    WHERE processing_state = 'ready'
      AND review_state IN ('not_required', 'approved')
      AND publication_state = 'published'
      AND safety_state = 'safe'
      AND public_rendition_ready;

CREATE INDEX gallery_images_public_description_search_idx
    ON gallery_images USING GIN (description gin_trgm_ops)
    WHERE processing_state = 'ready'
      AND review_state IN ('not_required', 'approved')
      AND publication_state = 'published'
      AND safety_state = 'safe'
      AND public_rendition_ready;

CREATE INDEX gallery_images_admin_lifecycle_idx
    ON gallery_images (
        publication_state,
        processing_state,
        review_state,
        safety_state,
        updated_at DESC,
        id DESC
    );

CREATE INDEX gallery_submissions_admin_queue_idx
    ON gallery_submissions (
        outcome,
        processing_state,
        review_state,
        safety_state,
        updated_at DESC,
        id DESC
    );

CREATE INDEX gallery_cases_admin_queue_idx
    ON gallery_cases (status, kind, updated_at DESC, id DESC);

CREATE INDEX gallery_image_events_visitor_window_idx
    ON gallery_image_events (image_id, event_type, occurred_at DESC, subject_key, session_key);

CREATE INDEX gallery_ranking_snapshots_fresh_idx
    ON gallery_ranking_snapshots (ranking_kind, window_key, generated_at DESC)
    INCLUDE (expires_at);
