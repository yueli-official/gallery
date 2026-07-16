DROP INDEX IF EXISTS gallery_ranking_snapshots_fresh_idx;
DROP INDEX IF EXISTS gallery_image_events_visitor_window_idx;
DROP INDEX IF EXISTS gallery_cases_admin_queue_idx;
DROP INDEX IF EXISTS gallery_submissions_admin_queue_idx;
DROP INDEX IF EXISTS gallery_images_admin_lifecycle_idx;
DROP INDEX IF EXISTS gallery_images_public_description_search_idx;
DROP INDEX IF EXISTS gallery_images_public_title_search_idx;
DROP INDEX IF EXISTS gallery_images_public_title_idx;

-- pg_trgm is database-level shared infrastructure and is intentionally retained.
