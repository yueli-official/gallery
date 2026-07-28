CREATE TABLE gallery_home_sections (
    site_key TEXT NOT NULL REFERENCES gallery_site_settings(site_key) ON DELETE CASCADE,
    section_key TEXT NOT NULL CHECK (section_key IN ('random', 'collections', 'latest', 'trending')),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 20),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    action_label TEXT NOT NULL DEFAULT '',
    item_limit INTEGER NOT NULL CHECK (item_limit BETWEEN 1 AND 60),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (site_key, section_key),
    UNIQUE (site_key, position)
);

INSERT INTO gallery_home_sections (
    site_key,
    section_key,
    enabled,
    position,
    title,
    description,
    action_label,
    item_limit
)
SELECT
    site.site_key,
    defaults.section_key,
    defaults.enabled,
    defaults.position,
    defaults.title,
    defaults.description,
    defaults.action_label,
    defaults.item_limit
FROM gallery_site_settings AS site
CROSS JOIN (
    VALUES
        ('random', TRUE, 0, '随机看看', '', '换一批', 24),
        ('collections', TRUE, 1, '从专题进入', '沿着一个清晰主题，查看经过整理的图片集合。', '查看全部专题', 4),
        ('latest', TRUE, 2, '最新入库', '最近完成处理和审核的公开图片。', '查看全部图片', 8),
        ('trending', TRUE, 3, '正在被发现', '近期获得更多有效浏览的图片。', '查看排行', 8)
) AS defaults(section_key, enabled, position, title, description, action_label, item_limit)
WHERE site.site_key = 'gallery';

UPDATE gallery_site_settings
SET random_batch_size = 24,
    title = CASE WHEN title = '随机看看' THEN name ELSE title END,
    updated_at = NOW()
WHERE site_key = 'gallery';
