-- Reverting would discard direct cover references and dimension/tag visuals.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM gallery_classification_presentation WHERE kind IN ('facet','tag') OR (asset_id IS NOT NULL AND image_id IS NULL)) THEN
        RAISE EXCEPTION 'Remove direct covers and dimension/tag visuals before reverting migration 15';
    END IF;
END $$;
ALTER TABLE gallery_classification_presentation
    DROP CONSTRAINT gallery_classification_presentation_kind_check,
    DROP CONSTRAINT gallery_classification_presentation_cover_check,
    DROP COLUMN asset_id,
    ADD CONSTRAINT gallery_classification_presentation_kind_check CHECK (kind IN ('category','facet_value')),
    ADD CONSTRAINT gallery_classification_presentation_check CHECK (mode <> 'image' OR image_id IS NOT NULL);
