DROP INDEX gallery_artworks_review_idx;
DROP INDEX gallery_creator_profiles_review_idx;

ALTER TABLE gallery_artworks
    DROP COLUMN reviewed_at,
    DROP COLUMN reviewed_by,
    DROP COLUMN review_note;

ALTER TABLE gallery_creator_profiles
    DROP COLUMN reviewed_at,
    DROP COLUMN reviewed_by,
    DROP COLUMN review_note,
    DROP COLUMN application_note,
    DROP CONSTRAINT gallery_creator_profiles_status_check;

UPDATE gallery_creator_profiles
SET status = 'suspended'
WHERE status = 'rejected';

ALTER TABLE gallery_creator_profiles
    ADD CONSTRAINT gallery_creator_profiles_status_check
        CHECK (status IN ('pending', 'active', 'suspended'));
