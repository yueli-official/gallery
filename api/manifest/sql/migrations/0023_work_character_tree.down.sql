-- Newly created work and character values share one identity space after this
-- migration. Splitting it automatically would invent ownership or lose edits.
DO $$ BEGIN
    RAISE EXCEPTION 'Work/character migration requires restoring the pre-migration database backup to roll back';
END $$;
