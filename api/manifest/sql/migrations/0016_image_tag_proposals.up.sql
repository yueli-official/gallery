ALTER TABLE gallery_tag_proposals
    ALTER COLUMN submission_id DROP NOT NULL,
    ADD COLUMN image_id UUID REFERENCES gallery_images(id) ON DELETE CASCADE,
    ADD COLUMN proposer_id TEXT NOT NULL DEFAULT '',
    ADD CONSTRAINT gallery_tag_proposal_target_check CHECK (
        (submission_id IS NOT NULL AND image_id IS NULL) OR
        (submission_id IS NULL AND image_id IS NOT NULL AND proposer_id <> '')
    );

CREATE UNIQUE INDEX gallery_image_tag_proposal_contributor_idx
    ON gallery_tag_proposals (image_id, proposer_id, lookup_key)
    WHERE image_id IS NOT NULL;
CREATE INDEX gallery_tag_proposal_contributor_created_idx
    ON gallery_tag_proposals (proposer_id, created_at)
    WHERE image_id IS NOT NULL;

ALTER TABLE gallery_classification_catalogs ADD COLUMN tag_presets_seeded_at TIMESTAMPTZ;
