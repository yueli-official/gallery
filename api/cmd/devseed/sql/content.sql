-- The shared local test account is an E2E fixture, not personal development
-- data. Clear its mutable workspace before rebuilding the deterministic demo
-- so repeated browser runs cannot inherit favorites or uploads from an older run.
DELETE FROM gallery_cases
WHERE reporter_kind = 'user'
  AND reporter_id = 'TestA123';
DELETE FROM gallery_collections
WHERE kind = 'gallery.favorites' AND owner_kind = 'user'
  AND owner_id = 'TestA123';
DELETE FROM gallery_submissions
WHERE subject_kind = 'user'
  AND subject_id = 'TestA123';
DELETE FROM gallery_comments
WHERE user_key = 'TestA123'
  AND id::text NOT LIKE '019b5000-0000-7000-9000-%';

INSERT INTO gallery_tags (id, catalog_id, current_slug, current_name, status, created_at, updated_at)
SELECT
    ('019b2000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    '019817c8-0000-7000-8000-000000000002'::uuid,
    'demo-tag-' || i,
    (ARRAY['晨光','海岸','山野','城市','建筑','人物','静物','夜色','胶片','极简','蓝调','旅行'])[i],
    'active', TIMESTAMPTZ '2026-07-01 08:00:00+00', TIMESTAMPTZ '2026-07-01 08:00:00+00'
FROM generate_series(1, 12) AS i
ON CONFLICT (catalog_id, current_slug) DO UPDATE SET current_name = EXCLUDED.current_name, status = 'active', replacement_id = NULL;

INSERT INTO gallery_tag_lookup_entries (catalog_id, lookup_key, target_tag_id, kind, display_value)
SELECT
    '019817c8-0000-7000-8000-000000000002'::uuid,
    'demo-tag-' || i,
    ('019b2000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    'canonical',
    (ARRAY['晨光','海岸','山野','城市','建筑','人物','静物','夜色','胶片','极简','蓝调','旅行'])[i]
FROM generate_series(1, 12) AS i
ON CONFLICT (catalog_id, lookup_key) DO UPDATE SET target_tag_id = EXCLUDED.target_tag_id, display_value = EXCLUDED.display_value;

INSERT INTO gallery_images (
    id, asset_id, title, description, source_url, alt_text, width, height,
    dominant_color, processing_state, review_state, publication_state,
    safety_state, public_rendition_ready, published_at, hidden_at, deleted_at,
    created_at, updated_at
)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    ('019b0000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    (ARRAY['海岸晨光','山间薄雾','城市切片','安静建筑','夏日树影','夜色漫游','日常静物','远方公路'])[((i - 1) % 8) + 1] || ' ' || lpad(i::text, 3, '0'),
    '用于本地中型图库验收的确定性演示内容，第 ' || i || ' 张。',
    'https://example.com/gallery/demo/' || i,
    '演示图片：' || (ARRAY['海岸晨光','山间薄雾','城市切片','安静建筑','夏日树影','夜色漫游','日常静物','远方公路'])[((i - 1) % 8) + 1],
    960, 720,
    (ARRAY['#304f5c','#403b57','#355f50','#6b523f','#465a78','#764657','#4a6556','#5c4c6e'])[((i - 1) % 8) + 1],
    CASE WHEN i BETWEEN 105 AND 112 THEN 'failed' WHEN i BETWEEN 121 AND 124 THEN 'processing' ELSE 'ready' END,
    CASE WHEN i <= 96 OR i BETWEEN 113 AND 120 THEN 'approved' WHEN i >= 125 THEN 'rejected' ELSE 'pending' END,
    CASE WHEN i <= 84 THEN 'published' WHEN i <= 96 THEN 'hidden' WHEN i >= 125 THEN 'deleted' ELSE 'draft' END,
    CASE WHEN i <= 104 THEN 'safe' WHEN i <= 112 THEN 'unavailable' WHEN i <= 120 THEN 'uncertain' WHEN i <= 124 THEN 'pending' ELSE 'blocked' END,
    i <= 96,
    CASE WHEN i <= 84 THEN TIMESTAMPTZ '2026-07-10 08:00:00+00' - (i || ' hours')::interval END,
    CASE WHEN i BETWEEN 85 AND 96 THEN TIMESTAMPTZ '2026-07-12 08:00:00+00' END,
    CASE WHEN i >= 125 THEN TIMESTAMPTZ '2026-07-13 08:00:00+00' END,
    TIMESTAMPTZ '2026-07-01 08:00:00+00' + (i || ' minutes')::interval,
    TIMESTAMPTZ '2026-07-01 08:00:00+00' + (i || ' minutes')::interval
FROM generate_series(1, 128) AS i
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title, description = EXCLUDED.description, source_url = EXCLUDED.source_url,
    alt_text = EXCLUDED.alt_text, dominant_color = EXCLUDED.dominant_color,
    processing_state = EXCLUDED.processing_state, review_state = EXCLUDED.review_state,
    publication_state = EXCLUDED.publication_state, safety_state = EXCLUDED.safety_state,
    public_rendition_ready = EXCLUDED.public_rendition_ready, published_at = EXCLUDED.published_at,
    hidden_at = EXCLUDED.hidden_at, deleted_at = EXCLUDED.deleted_at, updated_at = EXCLUDED.updated_at;

INSERT INTO gallery_image_category_assignments (image_id, category_id)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    ('019817c8-0000-7000-8300-00000000000' || (((i - 1) % 3) + 1))::uuid
FROM generate_series(1, 128) AS i ON CONFLICT DO NOTHING;

INSERT INTO gallery_image_primary_categories (image_id, category_id)
SELECT image_id, category_id FROM gallery_image_category_assignments
WHERE image_id::text LIKE '019b1000-0000-7000-9000-%'
ON CONFLICT (image_id) DO UPDATE SET category_id = EXCLUDED.category_id;

INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    '019817c8-0000-7000-8200-000000000011'::uuid
FROM generate_series(1, 128) AS i ON CONFLICT DO NOTHING;

INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    (ARRAY[
        '019817c8-0000-7000-8200-000000000001'::uuid,
        '019817c8-0000-7000-8200-000000000003'::uuid,
        '019817c8-0000-7000-8200-000000000004'::uuid,
        '019817c8-0000-7000-8200-000000000005'::uuid,
        '019817c8-0000-7000-8200-000000000006'::uuid,
        '019817c8-0000-7000-8200-000000000007'::uuid
    ])[((i - 1) % 6) + 1]
FROM generate_series(1, 128) AS i ON CONFLICT DO NOTHING;

INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    (ARRAY[
        '019817c8-0000-7000-8200-000000000021'::uuid,
        '019817c8-0000-7000-8200-000000000022'::uuid
    ])[((i - 1) % 2) + 1]
FROM generate_series(1, 128) AS i ON CONFLICT DO NOTHING;

INSERT INTO gallery_image_tag_assignments (image_id, tag_id)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    ('019b2000-0000-7000-9000-' || lpad((((i - 1) % 12) + 1)::text, 12, '0'))::uuid
FROM generate_series(1, 128) AS i ON CONFLICT DO NOTHING;

INSERT INTO gallery_comments (
    id, image_id, parent_id, user_key, author_name, author_email,
    content, status, created_at, updated_at
)
VALUES
    ('019b5000-0000-7000-9000-000000000001', '019b1000-0000-7000-9000-000000000001', NULL,
     'TestA123', '', '', '这张海岸晨光的颜色很舒服。', 'approved',
     TIMESTAMPTZ '2026-07-11 08:00:00+00', TIMESTAMPTZ '2026-07-11 08:00:00+00'),
    ('019b5000-0000-7000-9000-000000000002', '019b1000-0000-7000-9000-000000000001',
     '019b5000-0000-7000-9000-000000000001', '', '路过的访客', '', '我也很喜欢这组蓝色。', 'approved',
     TIMESTAMPTZ '2026-07-11 08:05:00+00', TIMESTAMPTZ '2026-07-11 08:05:00+00'),
    ('019b5000-0000-7000-9000-000000000003', '019b1000-0000-7000-9000-000000000001', NULL,
     '', '清晨', '', '适合做一张安静的桌面背景。', 'approved',
     TIMESTAMPTZ '2026-07-11 08:10:00+00', TIMESTAMPTZ '2026-07-11 08:10:00+00'),
    ('019b5000-0000-7000-9000-000000000004', '019b1000-0000-7000-9000-000000000002', NULL,
     '', '等待审核的访客', 'pending@example.test', '这是一条等待运营审核的评论。', 'pending',
     TIMESTAMPTZ '2026-07-11 08:15:00+00', TIMESTAMPTZ '2026-07-11 08:15:00+00'),
    ('019b5000-0000-7000-9000-000000000005', '019b1000-0000-7000-9000-000000000002', NULL,
     '', '可疑访客', '', '这是一条垃圾评论验收样本。', 'spam',
     TIMESTAMPTZ '2026-07-11 08:20:00+00', TIMESTAMPTZ '2026-07-11 08:20:00+00'),
    ('019b5000-0000-7000-9000-000000000006', '019b1000-0000-7000-9000-000000000003', NULL,
     'TestA123', '', '', '这是一条回收站评论验收样本。', 'trash',
     TIMESTAMPTZ '2026-07-11 08:25:00+00', TIMESTAMPTZ '2026-07-11 08:25:00+00')
ON CONFLICT (id) DO UPDATE SET
    image_id = EXCLUDED.image_id, parent_id = EXCLUDED.parent_id,
    user_key = EXCLUDED.user_key, author_name = EXCLUDED.author_name,
    author_email = EXCLUDED.author_email, content = EXCLUDED.content,
    status = EXCLUDED.status, deleted_at = NULL, updated_at = EXCLUDED.updated_at;

INSERT INTO gallery_submissions (
    id, subject_kind, subject_id, asset_id, title, description, source_url,
    alt_text, processing_state, review_state, safety_state, outcome,
    width, height, dominant_color, public_rendition_ready, failure_code,
    review_note, created_at, updated_at
)
SELECT
    ('019b3000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    'user', 'TestA123',
    ('019b0000-0000-7000-9000-' || lpad((200 + i)::text, 12, '0'))::uuid,
    '批量投稿样本 ' || lpad(i::text, 2, '0'),
    '用于审核队列筛选、分页与局部失败恢复。',
    'https://example.com/submissions/' || i,
    '批量投稿演示图片 ' || i,
    CASE WHEN i BETWEEN 25 AND 32 THEN 'failed' WHEN i BETWEEN 41 AND 48 THEN 'processing' ELSE 'ready' END,
    CASE WHEN i BETWEEN 33 AND 40 THEN 'approved' ELSE 'pending' END,
    CASE WHEN i BETWEEN 17 AND 24 THEN 'uncertain' WHEN i BETWEEN 25 AND 32 THEN 'unavailable' ELSE 'safe' END,
    CASE WHEN i BETWEEN 33 AND 40 THEN 'published' WHEN i BETWEEN 25 AND 32 THEN 'failed' ELSE 'pending' END,
    960, 720, '#465a78', i BETWEEN 33 AND 40,
    CASE WHEN i BETWEEN 25 AND 32 THEN 'derive_failed' ELSE '' END,
    CASE WHEN i BETWEEN 17 AND 24 THEN '需要人工确认安全边界' ELSE '' END,
    TIMESTAMPTZ '2026-07-08 08:00:00+00' + (i || ' minutes')::interval,
    TIMESTAMPTZ '2026-07-08 08:00:00+00' + (i || ' minutes')::interval
FROM generate_series(1, 48) AS i
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title, processing_state = EXCLUDED.processing_state,
    review_state = EXCLUDED.review_state, safety_state = EXCLUDED.safety_state,
    outcome = EXCLUDED.outcome, failure_code = EXCLUDED.failure_code,
    review_note = EXCLUDED.review_note, updated_at = EXCLUDED.updated_at;

INSERT INTO gallery_tag_proposals (id, submission_id, input_value, lookup_key, status, created_at, updated_at)
SELECT
    ('019b4000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    ('019b3000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    (ARRAY['电影感','雨夜','留白','低饱和','通勤','窗边'])[i],
    'demo-proposal-' || i, 'pending',
    TIMESTAMPTZ '2026-07-09 08:00:00+00' + (i || ' minutes')::interval,
    TIMESTAMPTZ '2026-07-09 08:00:00+00' + (i || ' minutes')::interval
FROM generate_series(1, 6) AS i
ON CONFLICT (id) DO UPDATE SET input_value = EXCLUDED.input_value, status = 'pending', resolved_tag_id = NULL;

INSERT INTO gallery_cases (
    id, image_id, submission_id, kind, status, reporter_kind, reporter_id,
    reason, description, proposed_source_url, created_at, updated_at
)
SELECT
    ('019b5000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    CASE WHEN i <= 8 THEN ('019b3000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid END,
    (ARRAY['report','source_correction','safety_uncertain','near_duplicate','takedown'])[((i - 1) % 5) + 1],
    CASE WHEN i <= 10 THEN 'open' WHEN i <= 14 THEN 'reviewing' WHEN i <= 17 THEN 'resolved' ELSE 'dismissed' END,
    'user', 'TestA123', '演示工单原因 ' || i, '用于申诉队列的状态、类型与搜索验收。',
    CASE WHEN i % 5 = 2 THEN 'https://example.com/corrected-source/' || i END,
    TIMESTAMPTZ '2026-07-10 08:00:00+00' + (i || ' minutes')::interval,
    TIMESTAMPTZ '2026-07-10 08:00:00+00' + (i || ' minutes')::interval
FROM generate_series(1, 20) AS i
ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, reason = EXCLUDED.reason, updated_at = EXCLUDED.updated_at;

INSERT INTO gallery_collections (id, kind, resource_kind, owner_kind, owner_id, visibility, name, description, created_at, updated_at)
VALUES
('019b6000-0000-7000-9000-000000000001', 'gallery.editorial', 'gallery.image', 'site', 'gallery', 'public', '安静的蓝色时刻', '从海岸到夜色的蓝调图片精选。', '2026-07-05T08:00:00Z', '2026-07-12T08:00:00Z'),
('019b6000-0000-7000-9000-000000000002', 'gallery.editorial', 'gallery.image', 'site', 'gallery', 'public', '城市与几何', '建筑、街道与清晰构图。', '2026-07-05T08:00:00Z', '2026-07-11T08:00:00Z'),
('019b6000-0000-7000-9000-000000000003', 'gallery.editorial', 'gallery.image', 'site', 'gallery', 'public', '向自然出发', '山野、树影与远方公路。', '2026-07-05T08:00:00Z', '2026-07-10T08:00:00Z'),
('019b6000-0000-7000-9000-000000000004', 'gallery.editorial', 'gallery.image', 'site', 'gallery', 'private', '下期专题草稿', '尚未公开的运营编排。', '2026-07-05T08:00:00Z', '2026-07-09T08:00:00Z')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, visibility = EXCLUDED.visibility, updated_at = EXCLUDED.updated_at;

INSERT INTO gallery_collection_editorial (collection_id, slug, cover_image_id, seo_title, seo_description)
VALUES
('019b6000-0000-7000-9000-000000000001', 'blue-hour', '019b1000-0000-7000-9000-000000000001', '安静的蓝色时刻', '海岸、城市与夜色图片精选'),
('019b6000-0000-7000-9000-000000000002', 'city-geometry', '019b1000-0000-7000-9000-000000000003', '城市与几何', '建筑与街道摄影精选'),
('019b6000-0000-7000-9000-000000000003', 'into-nature', '019b1000-0000-7000-9000-000000000005', '向自然出发', '山野与自然图片精选'),
('019b6000-0000-7000-9000-000000000004', 'next-feature-draft', NULL, '', '')
ON CONFLICT (collection_id) DO UPDATE SET slug = EXCLUDED.slug, cover_image_id = EXCLUDED.cover_image_id, seo_title = EXCLUDED.seo_title, seo_description = EXCLUDED.seo_description;

INSERT INTO gallery_collection_members (collection_id, image_id, manual_position)
SELECT
    ('019b6000-0000-7000-9000-00000000000' || collection_no)::uuid,
    ('019b1000-0000-7000-9000-' || lpad((((collection_no - 1) * 20) + position)::text, 12, '0'))::uuid,
    position - 1
FROM generate_series(1, 4) AS collection_no
CROSS JOIN generate_series(1, 20) AS position
ON CONFLICT (collection_id, image_id) DO UPDATE SET manual_position = EXCLUDED.manual_position;

INSERT INTO gallery_image_metrics_daily (image_id, metric_date, exposures, qualified_views, unique_visitors, favorites, shares, reports)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    CURRENT_DATE - day_offset,
    28 + (i % 17) + (ARRAY[8,12,10,16,14,20,18,28,24,38,31,47,26,19])[((29 - day_offset) % 14) + 1],
    3 + (i % 11) + (ARRAY[2,4,3,7,5,9,6,16,11,24,18,31,13,8])[((29 - day_offset) % 14) + 1],
    2 + (i % 7) + (ARRAY[1,2,2,3,3,4,3,6,5,8,7,10,5,4])[((29 - day_offset) % 14) + 1],
    (i % 3) + (ARRAY[0,0,1,0,1,1,1,2,2,3,2,4,2,1])[((29 - day_offset) % 14) + 1],
    (i % 2) + (ARRAY[0,1,0,1,1,2,1,2,2,3,2,3,1,1])[((29 - day_offset) % 14) + 1],
    CASE WHEN (i + day_offset) % 53 = 0 THEN 1 ELSE 0 END
FROM generate_series(1, 84) AS i
CROSS JOIN generate_series(0, 29) AS day_offset
ON CONFLICT (image_id, metric_date) DO UPDATE SET
    exposures = EXCLUDED.exposures, qualified_views = EXCLUDED.qualified_views,
    unique_visitors = EXCLUDED.unique_visitors, favorites = EXCLUDED.favorites,
    shares = EXCLUDED.shares, reports = EXCLUDED.reports;;

-- Home spotlight and category covers are Gallery-owned presentation data.
INSERT INTO gallery_spotlight_characters (
    id, position, name, series, faction, image_url, thumb_url, accent, wash, enabled
) VALUES
    ('019b3000-0000-7000-9000-000000000001', 1, '砂狼白子', 'Blue Archive', '阿拜多斯', '/spotlight/shiroko.webp', '/spotlight/shiroko-thumb.webp', '#7fd8ff', 'radial-gradient(circle at 50% 34%, #17394a 0%, #090d12 68%)', TRUE),
    ('019b3000-0000-7000-9000-000000000002', 2, '小鸟游星野', 'Blue Archive', '阿拜多斯', '/spotlight/hoshino.webp', '/spotlight/hoshino-thumb.webp', '#ffd277', 'radial-gradient(circle at 50% 34%, #3d2a12 0%, #0d0b08 68%)', TRUE),
    ('019b3000-0000-7000-9000-000000000003', 3, '陆八魔阿露', 'Blue Archive', '格黑娜', '/spotlight/aru.webp', '/spotlight/aru-thumb.webp', '#ff7e95', 'radial-gradient(circle at 50% 34%, #3a1420 0%, #0e080a 68%)', TRUE),
    ('019b3000-0000-7000-9000-000000000004', 4, '早濑优香', 'Blue Archive', '千年', '/spotlight/yuuka.webp', '/spotlight/yuuka-thumb.webp', '#8dffd9', 'radial-gradient(circle at 50% 34%, #0e3a2e 0%, #060d0b 68%)', TRUE),
    ('019b3000-0000-7000-9000-000000000005', 5, '圣园弥香', 'Blue Archive', '崔尼蒂', '/spotlight/mika.webp', '/spotlight/mika-thumb.webp', '#ffb3e6', 'radial-gradient(circle at 50% 34%, #3a1a30 0%, #0e080d 68%)', TRUE)
ON CONFLICT (id) DO UPDATE SET
    position = EXCLUDED.position, name = EXCLUDED.name, series = EXCLUDED.series,
    faction = EXCLUDED.faction, image_url = EXCLUDED.image_url, thumb_url = EXCLUDED.thumb_url,
    accent = EXCLUDED.accent, wash = EXCLUDED.wash, enabled = EXCLUDED.enabled,
    updated_at = NOW();

INSERT INTO gallery_category_covers (category_id, image_url, thumb_url, accent)
VALUES
    ('019817c8-0000-7000-8300-000000000001', '/spotlight/hoshino.webp', '/spotlight/hoshino-thumb.webp', '#ffd277'),
    ('019817c8-0000-7000-8300-000000000002', '/spotlight/mika.webp', '/spotlight/mika-thumb.webp', '#ffb3e6'),
    ('019817c8-0000-7000-8300-000000000003', '/home-editorial-panorama.webp', '/home-editorial-panorama.webp', '#7fd8ff')
ON CONFLICT (category_id) DO UPDATE SET
    image_url = EXCLUDED.image_url, thumb_url = EXCLUDED.thumb_url,
    accent = EXCLUDED.accent, updated_at = NOW();

-- Local composition: if real uploaded assets exist, keep the public gallery on
-- those images and rebuild the three editorial collections from real works.
DO $$
DECLARE
    has_real BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM gallery_images
        WHERE asset_id::text LIKE '01a167%' AND publication_state = 'published'
    ) INTO has_real;

    IF NOT has_real THEN
        RETURN;
    END IF;

    UPDATE gallery_images
       SET publication_state = 'hidden', hidden_at = NOW(), updated_at = NOW()
     WHERE asset_id::text LIKE '019b%' AND publication_state = 'published';

    UPDATE gallery_collections
       SET name = '阿拜多斯 / 沙漠前线',
           description = '阿拜多斯对策委员会的日常与战斗。',
           updated_at = NOW()
     WHERE id = '019b6000-0000-7000-9000-000000000001';
    UPDATE gallery_collection_editorial
       SET slug = 'abydos-frontline',
           cover_image_id = (SELECT id FROM gallery_images WHERE title LIKE '%砂狼白子%' AND publication_state='published' LIMIT 1),
           seo_title = '阿拜多斯 / 沙漠前线',
           seo_description = '阿拜多斯对策委员会的日常与战斗。'
     WHERE collection_id = '019b6000-0000-7000-9000-000000000001';
    DELETE FROM gallery_collection_members WHERE collection_id = '019b6000-0000-7000-9000-000000000001';
    INSERT INTO gallery_collection_members (collection_id, image_id, manual_position, added_at)
    SELECT '019b6000-0000-7000-9000-000000000001', i.id, v.pos, NOW()
    FROM (VALUES ('砂狼白子', 0), ('小鸟游星野', 1), ('Ayane · Blue Archive', 2), ('Serika · Blue Archive', 3), ('Nonomi · Blue Archive', 4)) AS v(title, pos)
    JOIN gallery_images i ON i.title = v.title AND i.publication_state = 'published';

    UPDATE gallery_collections
       SET name = '格黑娜 / 风纪委员会',
           description = '格黑娜风纪委员会的秩序与火力。',
           updated_at = NOW()
     WHERE id = '019b6000-0000-7000-9000-000000000002';
    UPDATE gallery_collection_editorial
       SET slug = 'gehenna-prefect',
           cover_image_id = (SELECT id FROM gallery_images WHERE title LIKE '%天雨亚子%' AND publication_state='published' LIMIT 1),
           seo_title = '格黑娜 / 风纪委员会',
           seo_description = '格黑娜风纪委员会的秩序与火力。'
     WHERE collection_id = '019b6000-0000-7000-9000-000000000002';
    DELETE FROM gallery_collection_members WHERE collection_id = '019b6000-0000-7000-9000-000000000002';
    INSERT INTO gallery_collection_members (collection_id, image_id, manual_position, added_at)
    SELECT '019b6000-0000-7000-9000-000000000002', i.id, v.pos, NOW()
    FROM (VALUES
        ('天雨亚子', 0), ('陆八魔阿露', 1), ('Hina · Blue Archive', 2), ('Iori · Blue Archive', 3),
        ('Chinatsu · Blue Archive', 4), ('Kayoko · Blue Archive', 5), ('Haruna · Blue Archive', 6),
        ('Fuuka · Blue Archive', 7), ('Juri · Blue Archive', 8), ('Izumi · Blue Archive', 9), ('Cherino · Blue Archive', 10)
    ) AS v(title, pos)
    JOIN gallery_images i ON i.title = v.title AND i.publication_state = 'published';

    UPDATE gallery_collections
       SET name = '千年 / 游戏开发部',
           description = '千年游戏开发部的角色、装备与日常。',
           updated_at = NOW()
     WHERE id = '019b6000-0000-7000-9000-000000000003';
    UPDATE gallery_collection_editorial
       SET slug = 'millennium-game-dev',
           cover_image_id = (SELECT id FROM gallery_images WHERE title = 'Alice · Blue Archive' AND publication_state='published' LIMIT 1),
           seo_title = '千年 / 游戏开发部',
           seo_description = '千年游戏开发部的角色、装备与日常。'
     WHERE collection_id = '019b6000-0000-7000-9000-000000000003';
    DELETE FROM gallery_collection_members WHERE collection_id = '019b6000-0000-7000-9000-000000000003';
    INSERT INTO gallery_collection_members (collection_id, image_id, manual_position, added_at)
    SELECT '019b6000-0000-7000-9000-000000000003', i.id, v.pos, NOW()
    FROM (VALUES
        ('Alice · Blue Archive', 0), ('早濑优香', 1), ('Noa · Blue Archive', 2), ('Momoi · Blue Archive', 3),
        ('Midori · Blue Archive', 4), ('Neru · Blue Archive', 5), ('Toki · Blue Archive', 6),
        ('Rio · Blue Archive', 7), ('Maki · Blue Archive', 8), ('Chihiro · Blue Archive', 9),
        ('Kotama · Blue Archive', 10), ('一之濑明日奈', 11), ('角楯卡琳', 12)
    ) AS v(title, pos)
    JOIN gallery_images i ON i.title = v.title AND i.publication_state = 'published';

    DELETE FROM gallery_ranking_entries;
    DELETE FROM gallery_ranking_snapshots;
END $$;
