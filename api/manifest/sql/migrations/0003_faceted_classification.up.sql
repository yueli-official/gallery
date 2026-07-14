ALTER TABLE gallery_facets
    ADD COLUMN required_on_publish BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN filterable BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE gallery_facets
SET required_on_publish = TRUE
WHERE slug = 'medium';

ALTER TABLE gallery_categories RENAME TO gallery_facet_values;
ALTER TABLE gallery_facet_values RENAME CONSTRAINT gallery_categories_pkey TO gallery_facet_values_pkey;
ALTER TABLE gallery_facet_values RENAME CONSTRAINT gallery_categories_facet_fk TO gallery_facet_values_facet_fk;
ALTER TABLE gallery_facet_values DROP CONSTRAINT gallery_categories_parent_fk;
ALTER INDEX gallery_categories_facet_slug_uidx RENAME TO gallery_facet_values_facet_slug_uidx;
ALTER INDEX gallery_categories_parent_idx RENAME TO gallery_facet_values_parent_idx;

ALTER TABLE gallery_facet_values
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD CONSTRAINT gallery_facet_values_facet_id_id_unique UNIQUE (facet_id, id),
    ADD CONSTRAINT gallery_facet_values_parent_same_facet_fk
        FOREIGN KEY (facet_id, parent_id)
        REFERENCES gallery_facet_values (facet_id, id);

ALTER TABLE gallery_artwork_categories RENAME TO gallery_artwork_facet_assignments;
ALTER TABLE gallery_artwork_facet_assignments RENAME CONSTRAINT gallery_artwork_categories_pkey TO gallery_artwork_facet_assignments_pkey;
ALTER TABLE gallery_artwork_facet_assignments
    DROP CONSTRAINT gallery_artwork_categories_category_id_fkey;
ALTER TABLE gallery_artwork_facet_assignments RENAME COLUMN category_id TO value_id;
ALTER TABLE gallery_artwork_facet_assignments ADD COLUMN facet_id TEXT;

UPDATE gallery_artwork_facet_assignments assignment
SET facet_id = value.facet_id
FROM gallery_facet_values value
WHERE value.id = assignment.value_id;

ALTER TABLE gallery_artwork_facet_assignments
    ALTER COLUMN facet_id SET NOT NULL,
    ADD CONSTRAINT gallery_artwork_facet_assignments_value_fk
        FOREIGN KEY (facet_id, value_id)
        REFERENCES gallery_facet_values (facet_id, id)
        ON DELETE CASCADE;

ALTER INDEX gallery_artwork_categories_category_idx RENAME TO gallery_artwork_facet_assignments_value_idx;
CREATE INDEX gallery_artwork_facet_assignments_filter_idx
    ON gallery_artwork_facet_assignments (facet_id, value_id, artwork_id);
