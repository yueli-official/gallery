DELETE FROM authorization_projection_rules
 WHERE instance_key LIKE 'gallery:%'
   AND capability_key IN (
       'gallery.comment.read',
       'gallery.comment.moderate',
       'gallery.comment.delete'
   );

DELETE FROM authorization_policy_bindings
 WHERE instance_key LIKE 'gallery:%'
   AND capability_key IN (
       'gallery.comment.read',
       'gallery.comment.moderate',
       'gallery.comment.delete'
   );

UPDATE authorization_instances
   SET catalog_version = 2,
       catalog_digest = 'a3e15c72caa9cfb861a4488716ec824bea9e4dc650fd0e028717eb653493943b',
       updated_at = NOW()
 WHERE instance_key LIKE 'gallery:%'
   AND catalog_version = 3
   AND catalog_digest = '5fa1a646f1ed1099141258578339993035e4db78d3931ab5ea702ac593315780';
