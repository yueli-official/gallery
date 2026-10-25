ALTER TABLE gallery_tag_proposals
    ADD COLUMN proposer_kind TEXT NOT NULL DEFAULT 'user'
    CHECK (proposer_kind IN ('user', 'guest'));

DROP INDEX gallery_image_tag_proposal_contributor_idx;
CREATE UNIQUE INDEX gallery_image_tag_proposal_contributor_idx
    ON gallery_tag_proposals (image_id, proposer_kind, proposer_id, lookup_key)
    WHERE image_id IS NOT NULL;
