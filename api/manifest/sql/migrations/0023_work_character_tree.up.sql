-- Keep value identities and their image/submission assignments while moving
-- characters into the work facet. No relationship is inferred from co-occurrence.
DO $$
DECLARE
    catalog UUID;
    work UUID;
    characters UUID;
    pending UUID;
    work_ref TEXT;
    character_ref TEXT;
BEGIN
    SELECT id INTO catalog FROM gallery_classification_catalogs WHERE catalog_key = 'gallery';
    IF catalog IS NULL THEN RETURN; END IF;
    SELECT id INTO work FROM gallery_facets WHERE catalog_id = catalog AND slug = 'work';
    SELECT id INTO characters FROM gallery_facets WHERE catalog_id = catalog AND slug = 'character';
    IF work IS NULL THEN
        RAISE EXCEPTION 'Create the Gallery work facet before applying the work/character migration';
    END IF;
    -- Both UUID and compact IDs are valid in the existing saved-filter API.
    work_ref := '(work|' || work::text || '|' || rtrim(translate(encode(decode(replace(work::text, '-', ''), 'hex'), 'base64'), '+/', '-_'), '=') || ')';
    character_ref := '(character';
    IF characters IS NOT NULL THEN
        character_ref := character_ref || '|' || characters::text || '|' || rtrim(translate(encode(decode(replace(characters::text, '-', ''), 'hex'), 'base64'), '+/', '-_'), '=');
    END IF;
    character_ref := character_ref || ')';
    IF EXISTS (
        SELECT 1 FROM gallery_facet_values old JOIN gallery_facet_values target
          ON target.catalog_id = old.catalog_id AND target.slug = old.slug
        WHERE old.facet_id = characters AND target.facet_id = work
    ) THEN
        RAISE EXCEPTION 'Work/character slugs overlap; resolve the identities before migrating';
    END IF;
    -- Merging two OR groups into one changes their conjunction. Singleton
    -- groups can retain their meaning with matchAll; more complex saved queries
    -- require an explicit editorial choice rather than silently broadening them.
    IF EXISTS (
        SELECT 1 FROM gallery_saved_filters s
        WHERE s.query->>'facets' ~ ('(^|,)\s*' || work_ref || '\s*:')
          AND s.query->>'facets' ~ ('(^|,)\s*' || character_ref || '\s*:')
          AND COALESCE((s.query->>'matchAll')::boolean, false) = false
          AND EXISTS (
            SELECT 1 FROM regexp_split_to_table(s.query->>'facets', ',') selection
            GROUP BY CASE
                WHEN trim(selection) ~ ('^' || work_ref || '\s*:') THEN 'work'
                WHEN trim(selection) ~ ('^' || character_ref || '\s*:') THEN 'character'
                ELSE trim(split_part(selection, ':', 1)) END
            HAVING count(*) > 1
          )
    ) THEN
        RAISE EXCEPTION 'Review saved filters combining multiple work/character alternatives before migrating';
    END IF;
    -- A facet cover cannot be silently discarded or replace another cover.
    IF EXISTS (SELECT 1 FROM gallery_classification_presentation WHERE kind = 'facet' AND node_id = characters)
       OR EXISTS (SELECT 1 FROM gallery_facets WHERE replacement_id = characters) THEN
        RAISE EXCEPTION 'Resolve the character facet cover/replacements before migrating';
    END IF;

    INSERT INTO gallery_facet_values
        (id, catalog_id, facet_id, slug, name, status, editorial_position, first_activated_at)
    VALUES
        ('019e2300-0000-7000-8200-000000000001', catalog, work, 'independent-characters', '独立角色', 'active', 10000, NOW()),
        ('019e2300-0000-7000-8200-000000000002', catalog, work, 'unassigned-characters', '待归属角色', 'active', 10001, NOW())
    ON CONFLICT (catalog_id, facet_id, slug) DO NOTHING;
    SELECT id INTO pending FROM gallery_facet_values
      WHERE facet_id = work AND slug = 'unassigned-characters' AND parent_id IS NULL AND status = 'active';
    IF pending IS NULL OR NOT EXISTS (
        SELECT 1 FROM gallery_facet_values
        WHERE facet_id = work AND slug = 'independent-characters' AND parent_id IS NULL AND status = 'active'
    ) THEN
        RAISE EXCEPTION 'The independent/unassigned character groups must be active work roots';
    END IF;

    -- The existing PLN import has an explicit, verified ownership relationship.
    -- Other roots stay visible in 待归属角色 until an operator chooses their work.
    UPDATE gallery_facet_values v SET
        facet_id = work,
        parent_id = CASE
            WHEN v.parent_id IS NOT NULL THEN v.parent_id
            WHEN v.slug = 'plana-blue-archive' THEN COALESCE(
                (SELECT id FROM gallery_facet_values WHERE facet_id = work AND slug = 'blue-archive' AND parent_id IS NULL), pending)
            ELSE pending
        END,
        updated_at = NOW()
    WHERE v.catalog_id = catalog AND v.facet_id = characters;

    UPDATE gallery_facets SET name = '作品与角色',
        description = '按主要所属作品或系列组织角色；联动不改变归属。', updated_at = NOW()
    WHERE id = work;

    UPDATE gallery_classification_policy_profiles p SET
        document = jsonb_set(document, '{facets}', (
            SELECT COALESCE(jsonb_agg(
                CASE WHEN rule->>'facetId' = work::text THEN
                    rule || '{"maxValues":32,"leafOnly":false}'::jsonb ELSE rule END ORDER BY ordinal
            ), '[]'::jsonb)
            FROM jsonb_array_elements(p.document->'facets') WITH ORDINALITY AS rules(rule, ordinal)
            WHERE rule->>'facetId' IS DISTINCT FROM characters::text
        )), policy_revision = policy_revision + 1
    WHERE p.catalog_id = catalog;

    -- Saved filters are persisted data, migrated once rather than supported by
    -- a second runtime parser. Bare value IDs retain their original identities.
    UPDATE gallery_saved_filters SET query = jsonb_set(query, '{matchAll}', 'true'::jsonb)
    WHERE query->>'facets' ~ ('(^|,)\s*' || work_ref || '\s*:') AND query->>'facets' ~ ('(^|,)\s*' || character_ref || '\s*:');
    UPDATE gallery_saved_filters SET query = jsonb_set(query, '{facets}', to_jsonb(
        regexp_replace(regexp_replace(query->>'facets', '(^|,)\s*' || work_ref || '\s*:\s*', '\1work:', 'g'),
            '(^|,)\s*' || character_ref || '\s*:\s*', '\1work:', 'g')
    )) WHERE query->>'facets' ~ ('(^|,)\s*(' || work_ref || '|' || character_ref || ')\s*:');

    DELETE FROM gallery_facets WHERE id = characters;
    UPDATE gallery_classification_catalogs SET revision = revision + 1, updated_at = NOW() WHERE id = catalog;
END;
$$;
