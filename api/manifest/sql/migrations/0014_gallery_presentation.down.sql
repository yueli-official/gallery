DROP TABLE gallery_saved_filters;
DROP TABLE gallery_classification_presentation;
ALTER TABLE gallery_site_settings DROP CONSTRAINT gallery_sidebar_image_check,
    DROP COLUMN sidebar_mode, DROP COLUMN sidebar_image_id, DROP COLUMN sidebar_overlay;
