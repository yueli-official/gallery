-- Gallery authorization catalog v4 adds authenticated user capabilities used
-- by personal access tokens. Existing grants and policy revisions are kept.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM authorization_instances
         WHERE instance_key LIKE 'gallery:%'
           AND (
               catalog_version <> 3
               OR catalog_digest <> '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780'
           )
    ) THEN
        RAISE EXCEPTION 'Gallery authorization catalog is not at the expected v3 digest';
    END IF;
END
$$;

INSERT INTO authorization_policy_bindings (
    instance_key, revision, target_kind, target_key, capability_key
)
SELECT revisions.instance_key,
       revisions.revision,
       'access_layer',
       'authenticated',
       capabilities.capability_key
  FROM authorization_policy_revisions AS revisions
  JOIN authorization_instances AS instances
    ON instances.instance_key = revisions.instance_key
 CROSS JOIN (VALUES
       ('gallery.submission.create'),
       ('gallery.submission.read_own'),
       ('gallery.submission.withdraw'),
       ('gallery.favorite.read'),
       ('gallery.favorite.manage'),
       ('gallery.comment.create')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'gallery:%'
   AND instances.catalog_version = 3
   AND instances.catalog_digest = '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780'
ON CONFLICT DO NOTHING;

INSERT INTO authorization_projection_rules (
    instance_key, policy_revision, rule_kind, subject_key, role_key,
    capability_key, scope_id, provenance
)
SELECT instances.instance_key,
       instances.active_policy_revision,
       'access_layer',
       'authenticated',
       'authenticated',
       capabilities.capability_key,
       instances.root_scope_id,
       '{}'::jsonb
  FROM authorization_instances AS instances
 CROSS JOIN (VALUES
       ('gallery.submission.create'),
       ('gallery.submission.read_own'),
       ('gallery.submission.withdraw'),
       ('gallery.favorite.read'),
       ('gallery.favorite.manage'),
       ('gallery.comment.create')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'gallery:%'
   AND instances.catalog_version = 3
   AND instances.catalog_digest = '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780'
ON CONFLICT DO NOTHING;

UPDATE authorization_instances
   SET catalog_version = 4,
       catalog_digest = '425cad0b712884350e16a98c42fc3d0b161dd62f6f8b4cacc8ff4e4042a10d5a',
       updated_at = NOW()
 WHERE instance_key LIKE 'gallery:%'
   AND catalog_version = 3
   AND catalog_digest = '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780';
