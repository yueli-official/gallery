-- Reviewed contributions are durable user data; an automatic rollback would lose them.
DO $$ BEGIN RAISE EXCEPTION 'Image tag contributions require an explicit data migration before rollback'; END $$;
