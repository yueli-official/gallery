DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM authorization_instances
         WHERE instance_key LIKE 'gallery:%'
           AND (
               catalog_version <> 2
               OR catalog_digest <> 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b'
           )
    ) THEN
        RAISE EXCEPTION 'Gallery authorization catalog is not at the expected v2 digest';
    END IF;
END
$$;

DELETE FROM authorization_projection_rules
 WHERE instance_key LIKE 'gallery:%'
   AND capability_key = 'gallery.discovery.manage';

DELETE FROM authorization_policy_bindings
 WHERE instance_key LIKE 'gallery:%'
   AND target_kind = 'role'
   AND target_key = 'administrator'
   AND capability_key = 'gallery.discovery.manage';

UPDATE authorization_instances
   SET catalog_version = 1,
       catalog_digest = 'c7161d978168d00341536df2c6be34347a04ceddc544c03d4bcf65cf3a9a32f9',
       updated_at = NOW()
 WHERE instance_key LIKE 'gallery:%'
   AND catalog_version = 2
   AND catalog_digest = 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b';
