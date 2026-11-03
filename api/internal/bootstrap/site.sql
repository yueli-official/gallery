INSERT INTO gallery_site_settings (
			id, site_key, name, title, description, search_placeholder, footer_tagline,
			random_batch_size, random_candidate_size
		)
		VALUES (
			'019817c8-0000-7000-8000-000000000001', 'gallery', '月离图库', '月离图库',
			'公开浏览、收藏和投稿值得反复观看的图片。', '搜索图片、分类或标签',
			'月离图库，安静地收藏互联网中的好图片。', 24, 240
		)
		ON CONFLICT (site_key) DO NOTHING;

INSERT INTO gallery_home_sections (
			site_key, section_key, enabled, position, title, description, action_label, item_limit
		)
		VALUES
		('gallery', 'random', TRUE, 0, '随机看看', '', '换一批', 24),
		('gallery', 'collections', TRUE, 1, '从专题进入', '', '查看全部专题', 4),
		('gallery', 'latest', TRUE, 2, '最新入库', '', '查看全部图片', 8),
		('gallery', 'trending', TRUE, 3, '正在被发现', '', '查看排行', 8)
		ON CONFLICT (site_key, section_key) DO NOTHING;

INSERT INTO gallery_classification_catalogs (id, catalog_key, revision)
		VALUES ('019817c8-0000-7000-8000-000000000002', 'gallery', 1)
		ON CONFLICT (catalog_key) DO NOTHING;

-- Content subject is independent from format, medium, and pixel dimensions.
INSERT INTO gallery_categories
    (id,catalog_id,slug,name,description,status,editorial_position,first_activated_at)
VALUES
('019d25a0-0000-7000-8300-000000000001','019817c8-0000-7000-8000-000000000002','game','游戏','游戏作品、角色与相关创作。','active',10,NOW()),
('019d25a0-0000-7000-8300-000000000002','019817c8-0000-7000-8000-000000000002','anime','动漫','动画、漫画作品及其角色创作。','active',20,NOW()),
('019d25a0-0000-7000-8300-000000000003','019817c8-0000-7000-8000-000000000002','original','原创','原创角色与不依附现有作品的创作。','active',30,NOW())
ON CONFLICT (catalog_id,slug) DO NOTHING;

-- Work roots own their character children. File orientation
-- and resolution are computed from media facts, never duplicated as facets.
INSERT INTO gallery_facets
    (id,catalog_id,slug,name,description,status,editorial_position,first_activated_at)
VALUES
('019c1000-0000-7000-8100-000000000001','019817c8-0000-7000-8000-000000000002','work','作品与角色','按主要所属作品或系列组织角色；联动不改变归属。','active',10,NOW())
ON CONFLICT (catalog_id,slug) DO NOTHING;

INSERT INTO gallery_facet_values
    (id,catalog_id,facet_id,slug,name,status,editorial_position,first_activated_at)
SELECT v.id::uuid,f.catalog_id,f.id,v.slug,v.name,'active',v.position,NOW()
FROM gallery_facets f CROSS JOIN (VALUES
('019e2300-0000-7000-8200-000000000001','independent-characters','独立角色',10000),
('019e2300-0000-7000-8200-000000000002','unassigned-characters','待归属角色',10001)
) AS v(id,slug,name,position)
WHERE f.catalog_id='019817c8-0000-7000-8000-000000000002' AND f.slug='work'
ON CONFLICT (catalog_id,facet_id,slug) DO NOTHING;

INSERT INTO gallery_classification_policy_profiles
    (catalog_id,policy_key,schema_version,policy_revision,document)
VALUES (
'019817c8-0000-7000-8000-000000000002','gallery.image.public',1,1,
'{
  "category":{"minAssignments":1,"maxAssignments":3,"requirePrimary":true,"leafOnly":false,"maxDepth":0},
  "facets":[
    {"facetId":"019c1000-0000-7000-8100-000000000001","minValues":0,"maxValues":32,"leafOnly":false,"maxDepth":0}
  ],
  "tags":{"unknown":"propose"},
  "discovery":{"defaultSort":"editorial"}
}'::jsonb)
ON CONFLICT (catalog_id,policy_key) DO NOTHING;
