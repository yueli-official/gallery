ALTER TABLE gallery_classification_presentation
    ADD COLUMN asset_id UUID;

UPDATE gallery_classification_presentation p
SET asset_id = i.asset_id FROM gallery_images i WHERE i.id = p.image_id;

ALTER TABLE gallery_classification_presentation
    DROP CONSTRAINT gallery_classification_presentation_kind_check,
    DROP CONSTRAINT gallery_classification_presentation_check,
    ADD CONSTRAINT gallery_classification_presentation_kind_check CHECK (kind IN ('category', 'facet', 'facet_value', 'tag')),
    ADD CONSTRAINT gallery_classification_presentation_cover_check CHECK (mode <> 'image' OR asset_id IS NOT NULL);
