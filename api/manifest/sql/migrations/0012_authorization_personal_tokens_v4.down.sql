DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM authorization_instances
         WHERE instance_key LIKE 'gallery:%'
           AND (
               catalog_version <> 4
               OR catalog_digest <> '425cad0b712884350e16a98c42fc3d0b161dd62f6f8b4cacc8ff4e4042a10d5a'
           )
    ) THEN
        RAISE EXCEPTION 'Gallery authorization catalog is not at the expected v4 digest';
    END IF;
END
$$;

DELETE FROM authorization_projection_rules
 WHERE instance_key LIKE 'gallery:%'
   AND capability_key IN (
       'gallery.submission.create',
       'gallery.submission.read_own',
       'gallery.submission.withdraw',
       'gallery.favorite.read',
       'gallery.favorite.manage',
       'gallery.comment.create'
   );

DELETE FROM authorization_policy_bindings
 WHERE instance_key LIKE 'gallery:%'
   AND target_kind = 'access_layer'
   AND target_key = 'authenticated'
   AND capability_key IN (
       'gallery.submission.create',
       'gallery.submission.read_own',
       'gallery.submission.withdraw',
       'gallery.favorite.read',
       'gallery.favorite.manage',
       'gallery.comment.create'
   );

UPDATE authorization_instances
   SET catalog_version = 3,
       catalog_digest = '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780',
       updated_at = NOW()
 WHERE instance_key LIKE 'gallery:%'
   AND catalog_version = 4
   AND catalog_digest = '425cad0b712884350e16a98c42fc3d0b161dd62f6f8b4cacc8ff4e4042a10d5a';
