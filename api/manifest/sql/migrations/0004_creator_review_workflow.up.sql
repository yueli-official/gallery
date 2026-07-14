ALTER TABLE gallery_creator_profiles
    DROP CONSTRAINT gallery_creator_profiles_status_check;

ALTER TABLE gallery_creator_profiles
    ADD CONSTRAINT gallery_creator_profiles_status_check
        CHECK (status IN ('pending', 'active', 'rejected', 'suspended')),
    ADD COLUMN application_note TEXT NOT NULL DEFAULT '',
    ADD COLUMN review_note TEXT NOT NULL DEFAULT '',
    ADD COLUMN reviewed_by TEXT NOT NULL DEFAULT '',
    ADD COLUMN reviewed_at TIMESTAMPTZ;

ALTER TABLE gallery_artworks
    ADD COLUMN review_note TEXT NOT NULL DEFAULT '',
    ADD COLUMN reviewed_by TEXT NOT NULL DEFAULT '',
    ADD COLUMN reviewed_at TIMESTAMPTZ;

CREATE INDEX gallery_creator_profiles_review_idx
    ON gallery_creator_profiles (status, created_at DESC);

CREATE INDEX gallery_artworks_review_idx
    ON gallery_artworks (status, updated_at DESC);
