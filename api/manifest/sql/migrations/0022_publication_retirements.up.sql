-- Replacing an imported image's original must durably revoke its old public
-- capability. This product effect shares the image's reference delivery job.
CREATE TABLE gallery_publication_retirements (
 image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE RESTRICT,
 asset_id UUID NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(image_id,asset_id)
);
