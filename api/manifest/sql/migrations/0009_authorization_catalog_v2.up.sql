-- Gallery authorization catalog v2 adds the protected
-- gallery.discovery.manage capability. Catalog changes are applied explicitly
-- so the PostgreSQL adapter can continue to fail closed on unknown drift.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM authorization_instances
         WHERE instance_key LIKE 'gallery:%'
           AND (
               catalog_version <> 1
               OR catalog_digest <> 'c7161d978168d00341536df2c6be34347a04ceddc544c03d4bcf65cf3a9a32f9'
           )
    ) THEN
        RAISE EXCEPTION 'Gallery authorization catalog is not at the expected v1 digest';
    END IF;
END
$$;

INSERT INTO authorization_policy_bindings (
    instance_key,
    revision,
    target_kind,
    target_key,
    capability_key
)
SELECT revisions.instance_key,
       revisions.revision,
       'role',
       'administrator',
       'gallery.discovery.manage'
  FROM authorization_policy_revisions AS revisions
  JOIN authorization_instances AS instances
    ON instances.instance_key = revisions.instance_key
 WHERE instances.instance_key LIKE 'gallery:%'
   AND instances.catalog_version = 1
   AND instances.catalog_digest = 'c7161d978168d00341536df2c6be34347a04ceddc544c03d4bcf65cf3a9a32f9'
ON CONFLICT DO NOTHING;

INSERT INTO authorization_projection_rules (
    instance_key,
    policy_revision,
    rule_kind,
    subject_key,
    role_key,
    capability_key,
    scope_id,
    provenance
)
SELECT instances.instance_key,
       instances.active_policy_revision,
       'permission',
       '',
       roles.id,
       'gallery.discovery.manage',
       instances.root_scope_id,
       jsonb_build_object('role_key', roles.role_key)
  FROM authorization_instances AS instances
  JOIN authorization_role_definitions AS roles
    ON roles.instance_key = instances.instance_key
   AND roles.role_key = 'administrator'
   AND roles.protected = TRUE
 WHERE instances.instance_key LIKE 'gallery:%'
   AND instances.catalog_version = 1
   AND instances.catalog_digest = 'c7161d978168d00341536df2c6be34347a04ceddc544c03d4bcf65cf3a9a32f9'
ON CONFLICT DO NOTHING;

UPDATE authorization_instances
   SET catalog_version = 2,
       catalog_digest = 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b',
       updated_at = NOW()
 WHERE instance_key LIKE 'gallery:%'
   AND catalog_version = 1
   AND catalog_digest = 'c7161d978168d00341536df2c6be34347a04ceddc544c03d4bcf65cf3a9a32f9';
