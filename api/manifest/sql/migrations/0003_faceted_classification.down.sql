DROP INDEX gallery_artwork_facet_assignments_filter_idx;
ALTER INDEX gallery_artwork_facet_assignments_value_idx RENAME TO gallery_artwork_categories_category_idx;

ALTER TABLE gallery_artwork_facet_assignments
    DROP CONSTRAINT gallery_artwork_facet_assignments_value_fk,
    DROP COLUMN facet_id;
ALTER TABLE gallery_artwork_facet_assignments RENAME COLUMN value_id TO category_id;
ALTER TABLE gallery_artwork_facet_assignments RENAME CONSTRAINT gallery_artwork_facet_assignments_pkey TO gallery_artwork_categories_pkey;
ALTER TABLE gallery_artwork_facet_assignments
    ADD CONSTRAINT gallery_artwork_categories_category_id_fkey
        FOREIGN KEY (category_id) REFERENCES gallery_facet_values(id) ON DELETE CASCADE;
ALTER TABLE gallery_artwork_facet_assignments RENAME TO gallery_artwork_categories;

ALTER TABLE gallery_facet_values
    DROP CONSTRAINT gallery_facet_values_parent_same_facet_fk,
    DROP CONSTRAINT gallery_facet_values_facet_id_id_unique,
    DROP COLUMN updated_at,
    DROP COLUMN created_at,
    DROP COLUMN metadata,
    DROP COLUMN status;

ALTER INDEX gallery_facet_values_parent_idx RENAME TO gallery_categories_parent_idx;
ALTER INDEX gallery_facet_values_facet_slug_uidx RENAME TO gallery_categories_facet_slug_uidx;
ALTER TABLE gallery_facet_values
    ADD CONSTRAINT gallery_categories_parent_fk
        FOREIGN KEY (parent_id) REFERENCES gallery_facet_values(id) ON DELETE SET NULL;
ALTER TABLE gallery_facet_values RENAME CONSTRAINT gallery_facet_values_facet_fk TO gallery_categories_facet_fk;
ALTER TABLE gallery_facet_values RENAME CONSTRAINT gallery_facet_values_pkey TO gallery_categories_pkey;
ALTER TABLE gallery_facet_values RENAME TO gallery_categories;

ALTER TABLE gallery_facets
    DROP COLUMN updated_at,
    DROP COLUMN created_at,
    DROP COLUMN status,
    DROP COLUMN filterable,
    DROP COLUMN required_on_publish;
