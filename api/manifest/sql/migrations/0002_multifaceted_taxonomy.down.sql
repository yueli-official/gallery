ALTER TABLE gallery_artworks ADD COLUMN category_id TEXT REFERENCES gallery_categories(id);

UPDATE gallery_artworks artwork
SET category_id = relation.category_id
FROM (
    SELECT artwork_id, MIN(category_id) AS category_id
    FROM gallery_artwork_categories
    GROUP BY artwork_id
) relation
WHERE artwork.id = relation.artwork_id;

DROP INDEX gallery_artwork_categories_category_idx;
DROP TABLE gallery_artwork_categories;
DROP INDEX gallery_categories_parent_idx;
DROP INDEX gallery_categories_facet_slug_uidx;

ALTER TABLE gallery_categories
    DROP CONSTRAINT gallery_categories_parent_fk,
    DROP CONSTRAINT gallery_categories_facet_fk,
    DROP COLUMN cover_asset_id,
    DROP COLUMN parent_id,
    DROP COLUMN facet_id,
    ADD CONSTRAINT gallery_categories_slug_key UNIQUE (slug);

DROP TABLE gallery_facets;
