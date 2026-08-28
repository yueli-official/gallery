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

INSERT INTO authorization_policy_bindings (
    instance_key, revision, target_kind, target_key, capability_key
)
SELECT revisions.instance_key,
       revisions.revision,
       'role',
       roles.role_key,
       capabilities.capability_key
  FROM authorization_policy_revisions AS revisions
  JOIN authorization_instances AS instances
    ON instances.instance_key = revisions.instance_key
 CROSS JOIN (VALUES ('administrator'), ('content_operator')) AS roles(role_key)
 CROSS JOIN (VALUES
       ('gallery.comment.read'),
       ('gallery.comment.moderate'),
       ('gallery.comment.delete')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'gallery:%'
   AND instances.catalog_version = 2
   AND instances.catalog_digest = 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b'
ON CONFLICT DO NOTHING;

INSERT INTO authorization_projection_rules (
    instance_key, policy_revision, rule_kind, subject_key, role_key,
    capability_key, scope_id, provenance
)
SELECT instances.instance_key,
       instances.active_policy_revision,
       'permission',
       '',
       roles.id,
       capabilities.capability_key,
       instances.root_scope_id,
       jsonb_build_object('role_key', roles.role_key)
  FROM authorization_instances AS instances
  JOIN authorization_role_definitions AS roles
    ON roles.instance_key = instances.instance_key
   AND roles.role_key IN ('administrator', 'content_operator')
 CROSS JOIN (VALUES
       ('gallery.comment.read'),
       ('gallery.comment.moderate'),
       ('gallery.comment.delete')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'gallery:%'
   AND instances.catalog_version = 2
   AND instances.catalog_digest = 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b'
ON CONFLICT DO NOTHING;

UPDATE authorization_instances
   SET catalog_version = 3,
       catalog_digest = '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780',
       updated_at = NOW()
 WHERE instance_key LIKE 'gallery:%'
   AND catalog_version = 2
   AND catalog_digest = 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b';
