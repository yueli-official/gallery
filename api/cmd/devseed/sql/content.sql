-- The shared local test account is an E2E fixture, not personal development
-- data. Clear its mutable workspace before rebuilding the deterministic demo
-- so repeated browser runs cannot inherit favorites or uploads from an older run.
DELETE FROM gallery_cases
WHERE reporter_kind = 'user' AND reporter_id = 'ac73d232-ce55-487d-bb39-fd336f1a9806';
DELETE FROM gallery_collections
WHERE kind = 'gallery.favorites' AND owner_kind = 'user'
  AND owner_id = 'ac73d232-ce55-487d-bb39-fd336f1a9806';
DELETE FROM gallery_submissions
WHERE subject_kind = 'user' AND subject_id = 'ac73d232-ce55-487d-bb39-fd336f1a9806';

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

INSERT INTO gallery_image_tag_assignments (image_id, tag_id)
SELECT
    ('019b1000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    ('019b2000-0000-7000-9000-' || lpad((((i - 1) % 12) + 1)::text, 12, '0'))::uuid
FROM generate_series(1, 128) AS i ON CONFLICT DO NOTHING;

INSERT INTO gallery_submissions (
    id, subject_kind, subject_id, asset_id, title, description, source_url,
    alt_text, processing_state, review_state, safety_state, outcome,
    width, height, dominant_color, public_rendition_ready, failure_code,
    review_note, created_at, updated_at
)
SELECT
    ('019b3000-0000-7000-9000-' || lpad(i::text, 12, '0'))::uuid,
    'user', 'ac73d232-ce55-487d-bb39-fd336f1a9806',
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
    'user', 'ac73d232-ce55-487d-bb39-fd336f1a9806', '演示工单原因 ' || i, '用于申诉队列的状态、类型与搜索验收。',
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
    DATE '2026-07-14', i * 31, i * 17, i * 11, i * 3, i, CASE WHEN i % 19 = 0 THEN 1 ELSE 0 END
FROM generate_series(1, 84) AS i
ON CONFLICT (image_id, metric_date) DO UPDATE SET
    exposures = EXCLUDED.exposures, qualified_views = EXCLUDED.qualified_views,
    unique_visitors = EXCLUDED.unique_visitors, favorites = EXCLUDED.favorites,
    shares = EXCLUDED.shares, reports = EXCLUDED.reports;;
