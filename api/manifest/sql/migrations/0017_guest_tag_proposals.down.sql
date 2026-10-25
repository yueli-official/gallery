DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM gallery_tag_proposals WHERE image_id IS NOT NULL AND proposer_kind = 'guest') THEN
        RAISE EXCEPTION 'Guest contributions require an explicit data migration before rollback';
    END IF;
END $$;
DROP INDEX gallery_image_tag_proposal_contributor_idx;
ALTER TABLE gallery_tag_proposals DROP COLUMN proposer_kind;
CREATE UNIQUE INDEX gallery_image_tag_proposal_contributor_idx
    ON gallery_tag_proposals (image_id, proposer_id, lookup_key)
    WHERE image_id IS NOT NULL;
