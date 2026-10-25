-- Run once per catalog. Operator edits, deletions and inactive tags survive restarts.
DO $$
DECLARE catalog UUID; entry RECORD; target UUID;
BEGIN
    SELECT id INTO catalog FROM gallery_classification_catalogs
    WHERE catalog_key = 'gallery' AND tag_presets_seeded_at IS NULL FOR UPDATE;
    IF catalog IS NULL THEN RETURN; END IF;
    FOR entry IN SELECT * FROM (VALUES
        (1, '白发'), (2, '黑发'), (3, '金发'), (4, '银发'), (5, '蓝发'), (6, '粉发'),
        (7, '长发'), (8, '短发'), (9, '双马尾'), (10, '单马尾'),
        (11, '白色丝袜'), (12, '黑色丝袜'), (13, '过膝袜'), (14, '连裤袜'),
        (15, '制服'), (16, '连衣裙'), (17, '和服'), (18, '女仆装'),
        (19, '眼镜'), (20, '发饰'), (21, '帽子'), (22, '耳机'),
        (23, '微笑'), (24, '背影'), (25, '海边'), (26, '夜景'), (27, '樱花'), (28, '室内')
    ) AS seed(position, name) LOOP
        SELECT target_tag_id INTO target FROM gallery_tag_lookup_entries
        WHERE catalog_id = catalog AND lookup_key = entry.name;
        IF target IS NULL THEN
            target := ('019c6000-0000-7000-a100-' || lpad(entry.position::text, 12, '0'))::uuid;
            INSERT INTO gallery_tags (id, catalog_id, current_name, current_slug)
            VALUES (target, catalog, entry.name, entry.name);
            INSERT INTO gallery_tag_lookup_entries (catalog_id, lookup_key, target_tag_id, kind, display_value)
            VALUES (catalog, entry.name, target, 'canonical', entry.name);
        END IF;
    END LOOP;
    INSERT INTO gallery_tag_lookup_entries (catalog_id, lookup_key, target_tag_id, kind, display_value)
    SELECT catalog, alias.name, canonical.target_tag_id, 'alias', alias.name
    FROM (VALUES ('白丝','白色丝袜'), ('黑丝','黑色丝袜'), ('白毛','白发'), ('黑毛','黑发'), ('双马尾辫','双马尾')) AS alias(name, target)
    JOIN gallery_tag_lookup_entries canonical ON canonical.catalog_id = catalog AND canonical.lookup_key = alias.target
    ON CONFLICT DO NOTHING;
    UPDATE gallery_classification_catalogs SET tag_presets_seeded_at = NOW(), revision = revision + 1, updated_at = NOW()
    WHERE id = catalog;
END $$;
