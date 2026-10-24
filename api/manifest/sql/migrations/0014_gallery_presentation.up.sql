ALTER TABLE gallery_site_settings
    ADD COLUMN sidebar_mode TEXT NOT NULL DEFAULT 'plain' CHECK (sidebar_mode IN ('plain', 'image')),
    ADD COLUMN sidebar_image_id UUID REFERENCES gallery_images(id) ON DELETE RESTRICT,
    ADD COLUMN sidebar_overlay DOUBLE PRECISION NOT NULL DEFAULT 0.82 CHECK (sidebar_overlay BETWEEN 0.55 AND 0.95),
    ADD CONSTRAINT gallery_sidebar_image_check CHECK (sidebar_mode <> 'image' OR sidebar_image_id IS NOT NULL);

CREATE TABLE gallery_classification_presentation (
    kind TEXT NOT NULL CHECK (kind IN ('category', 'facet_value')),
    node_id UUID NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('icon', 'image')),
    icon TEXT NOT NULL DEFAULT 'photo',
    image_id UUID REFERENCES gallery_images(id) ON DELETE RESTRICT,
    PRIMARY KEY (kind, node_id),
    CHECK (mode <> 'image' OR image_id IS NOT NULL)
);

CREATE TABLE gallery_saved_filters (
    id UUID PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 60),
    query JSONB NOT NULL CHECK (jsonb_typeof(query) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, name)
);
CREATE INDEX gallery_saved_filters_owner_idx ON gallery_saved_filters (owner_id, created_at DESC, id);


-- Product-owned, optional dimensions. Values are configured by administrators;
-- existing images are not assigned fictional work/character relationships.
INSERT INTO gallery_facets(id,catalog_id,slug,name,status,editorial_position,first_activated_at)
SELECT v.id::uuid,c.id,v.slug,v.name,'active',v.position,NOW()
FROM gallery_classification_catalogs c CROSS JOIN (VALUES
('019c1000-0000-7000-8100-000000000001','work','作品',40),
('019c1000-0000-7000-8100-000000000002','character','角色',50),
('019c1000-0000-7000-8100-000000000003','purpose','用途',60)
) AS v(id,slug,name,position) WHERE c.catalog_key='gallery'
ON CONFLICT (catalog_id,slug) DO NOTHING;
UPDATE gallery_classification_policy_profiles p
SET document=jsonb_set(document,'{facets}',(p.document->'facets') ||
COALESCE((SELECT jsonb_agg(jsonb_build_object('facetId',f.id::text,'minValues',0,'maxValues',16,'leafOnly',false,'maxDepth',0))
FROM gallery_facets f WHERE f.catalog_id=p.catalog_id AND f.slug IN ('work','character','purpose')
AND NOT EXISTS(SELECT 1 FROM jsonb_array_elements(p.document->'facets') rule WHERE rule->>'facetId'=f.id::text)),'[]'::jsonb)),
policy_revision=policy_revision+1
WHERE p.policy_key='gallery.image.public'
AND EXISTS(SELECT 1 FROM gallery_facets f WHERE f.catalog_id=p.catalog_id AND f.slug IN ('work','character','purpose')
AND NOT EXISTS(SELECT 1 FROM jsonb_array_elements(p.document->'facets') rule WHERE rule->>'facetId'=f.id::text));
UPDATE gallery_classification_catalogs SET revision=revision+1,updated_at=NOW() WHERE catalog_key='gallery';
