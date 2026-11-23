ALTER TABLE gallery_cases DROP CONSTRAINT gallery_cases_kind_check;
ALTER TABLE gallery_cases ADD CONSTRAINT gallery_cases_kind_check
    CHECK (kind IN ('report', 'source_correction', 'metadata_correction', 'safety_uncertain', 'near_duplicate', 'takedown'));
ALTER TABLE gallery_cases ADD COLUMN proposed_metadata JSONB;
ALTER TABLE gallery_cases ADD CONSTRAINT gallery_case_metadata_check CHECK (
    (kind = 'metadata_correction' AND proposed_metadata IS NOT NULL AND jsonb_typeof(proposed_metadata) = 'object')
    OR (kind <> 'metadata_correction' AND proposed_metadata IS NULL)
);
CREATE INDEX gallery_cases_metadata_contributor_idx
    ON gallery_cases (reporter_kind, reporter_id, created_at DESC)
    WHERE kind = 'metadata_correction';
