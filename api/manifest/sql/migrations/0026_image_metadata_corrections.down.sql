DO $$ BEGIN
    RAISE EXCEPTION 'Image metadata suggestions contain editorial data. Assess and restore the verified database backup instead of automatically downgrading.';
END $$;
