ALTER TABLE gallery_images ADD COLUMN public_id TEXT;
ALTER TABLE gallery_images ADD CONSTRAINT gallery_images_public_id_shape
    CHECK (public_id ~ '^[1-9A-HJ-NP-Za-km-z]{8}$');
ALTER TABLE gallery_images ADD CONSTRAINT gallery_images_public_id_unique UNIQUE (public_id);
-- Run cmd/imagekeys after this schema migration and before starting API/Web.
-- The shared Go Identifier module owns allocation and the atomic NOT NULL seal.
