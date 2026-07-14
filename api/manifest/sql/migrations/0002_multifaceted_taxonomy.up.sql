CREATE TABLE gallery_facets (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    selection_mode TEXT NOT NULL DEFAULT 'multiple' CHECK (selection_mode IN ('single', 'multiple')),
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0)
);

INSERT INTO gallery_facets (id, slug, name, description, selection_mode, sort_order)
VALUES
    ('facet-medium', 'medium', '媒介', '作品采用的主要视觉媒介或内容形态。', 'multiple', 10),
    ('facet-subject', 'subject', '题材', '作品描绘的对象、场景或主题。', 'multiple', 20),
    ('facet-style', 'style', '风格', '作品呈现的视觉语言与创作风格。', 'multiple', 30),
    ('facet-purpose', 'purpose', '用途', '作品适合的使用与浏览场景。', 'multiple', 40);

ALTER TABLE gallery_categories
    ADD COLUMN facet_id TEXT,
    ADD COLUMN parent_id TEXT,
    ADD COLUMN cover_asset_id TEXT NOT NULL DEFAULT '';

UPDATE gallery_categories SET facet_id = 'facet-subject' WHERE facet_id IS NULL;

ALTER TABLE gallery_categories
    ALTER COLUMN facet_id SET NOT NULL,
    ADD CONSTRAINT gallery_categories_facet_fk FOREIGN KEY (facet_id) REFERENCES gallery_facets(id),
    ADD CONSTRAINT gallery_categories_parent_fk FOREIGN KEY (parent_id) REFERENCES gallery_categories(id) ON DELETE SET NULL,
    DROP CONSTRAINT gallery_categories_slug_key;

CREATE UNIQUE INDEX gallery_categories_facet_slug_uidx ON gallery_categories (facet_id, slug);
CREATE INDEX gallery_categories_parent_idx ON gallery_categories (parent_id, sort_order);

CREATE TABLE gallery_artwork_categories (
    artwork_id TEXT NOT NULL REFERENCES gallery_artworks(id) ON DELETE CASCADE,
    category_id TEXT NOT NULL REFERENCES gallery_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (artwork_id, category_id)
);

INSERT INTO gallery_artwork_categories (artwork_id, category_id)
SELECT id, category_id
FROM gallery_artworks
WHERE category_id IS NOT NULL;

ALTER TABLE gallery_artworks DROP COLUMN category_id;

CREATE INDEX gallery_artwork_categories_category_idx ON gallery_artwork_categories (category_id, artwork_id);
