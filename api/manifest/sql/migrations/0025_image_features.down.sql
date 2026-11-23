-- Feature assignments and editorial changes are user data.
DO $$ BEGIN
    RAISE EXCEPTION 'Restore the pre-feature database backup to roll back without losing feature assignments';
END $$;
