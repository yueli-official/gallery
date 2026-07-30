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
		('gallery', 'collections', TRUE, 1, '从专题进入', '沿着一个清晰主题，查看经过整理的图片集合。', '查看全部专题', 4),
		('gallery', 'latest', TRUE, 2, '最新入库', '最近完成处理和审核的公开图片。', '查看全部图片', 8),
		('gallery', 'trending', TRUE, 3, '正在被发现', '近期获得更多有效浏览的图片。', '查看排行', 8)
		ON CONFLICT (site_key, section_key) DO NOTHING;

INSERT INTO gallery_classification_catalogs (id, catalog_key, revision)
		VALUES ('019817c8-0000-7000-8000-000000000002', 'gallery', 1)
		ON CONFLICT (catalog_key) DO NOTHING;

INSERT INTO gallery_categories (
			id, catalog_id, slug, name, description, status, editorial_position, first_activated_at
		)
		VALUES
		('019817c8-0000-7000-8300-000000000001', '019817c8-0000-7000-8000-000000000002', 'wallpaper', '壁纸', '适合桌面、手机与其他屏幕展示的图片。', 'active', 10, NOW()),
		('019817c8-0000-7000-8300-000000000002', '019817c8-0000-7000-8000-000000000002', 'illustration', '插画', '数字或传统绘制的视觉作品。', 'active', 20, NOW()),
		('019817c8-0000-7000-8300-000000000003', '019817c8-0000-7000-8000-000000000002', 'photography', '摄影', '通过摄影媒介形成的图像。', 'active', 30, NOW())
		ON CONFLICT (catalog_id, slug) DO NOTHING;

INSERT INTO gallery_facets (
			id, catalog_id, slug, name, description, status, editorial_position, first_activated_at
		)
		VALUES
		('019817c8-0000-7000-8100-000000000001', '019817c8-0000-7000-8000-000000000002', 'scene', '场景', '图片呈现的主体与环境。', 'active', 10, NOW()),
		('019817c8-0000-7000-8100-000000000002', '019817c8-0000-7000-8000-000000000002', 'orientation', '方向', '图片的横竖与方形比例。', 'active', 20, NOW()),
		('019817c8-0000-7000-8100-000000000003', '019817c8-0000-7000-8000-000000000002', 'style', '风格', '运营维护的视觉风格。', 'active', 30, NOW())
		ON CONFLICT (catalog_id, slug) DO NOTHING;

INSERT INTO gallery_facet_values (
			id, catalog_id, facet_id, parent_id, slug, name, description, status, editorial_position, first_activated_at
		)
		VALUES
		('019817c8-0000-7000-8200-000000000001', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', NULL, 'people', '人物', '人物、角色与肖像。', 'active', 10, NOW()),
		('019817c8-0000-7000-8200-000000000002', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', NULL, 'nature', '自然', '自然环境与自然现象。', 'active', 20, NOW()),
		('019817c8-0000-7000-8200-000000000003', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', '019817c8-0000-7000-8200-000000000002', 'landscape', '风景', '山川、海洋、森林与天空。', 'active', 10, NOW()),
		('019817c8-0000-7000-8200-000000000004', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', NULL, 'animals', '动物', '动物与生物。', 'active', 30, NOW()),
		('019817c8-0000-7000-8200-000000000005', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', NULL, 'architecture', '建筑', '建筑、室内与空间。', 'active', 40, NOW()),
		('019817c8-0000-7000-8200-000000000006', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', NULL, 'objects', '静物', '物品、食物与日常细节。', 'active', 50, NOW()),
		('019817c8-0000-7000-8200-000000000007', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000001', NULL, 'abstract', '抽象', '非具象与实验视觉。', 'active', 60, NOW()),
		('019817c8-0000-7000-8200-000000000011', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000002', NULL, 'landscape', '横图', '宽度大于高度。', 'active', 10, NOW()),
		('019817c8-0000-7000-8200-000000000012', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000002', NULL, 'portrait', '竖图', '高度大于宽度。', 'active', 20, NOW()),
		('019817c8-0000-7000-8200-000000000013', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000002', NULL, 'square', '方图', '宽高接近。', 'active', 30, NOW()),
		('019817c8-0000-7000-8200-000000000021', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000003', NULL, 'minimal', '极简', '克制元素与清晰构成。', 'active', 10, NOW()),
		('019817c8-0000-7000-8200-000000000022', '019817c8-0000-7000-8000-000000000002', '019817c8-0000-7000-8100-000000000003', NULL, 'retro', '复古', '历史媒介与旧印刷语言。', 'active', 20, NOW())
		ON CONFLICT (catalog_id, facet_id, slug) DO NOTHING;

INSERT INTO gallery_classification_policy_profiles (
			catalog_id, policy_key, schema_version, policy_revision, document
		)
		VALUES (
			'019817c8-0000-7000-8000-000000000002',
			'gallery.image.public',
			1,
			1,
			'{
				"category": {
					"minAssignments": 1,
					"maxAssignments": 3,
					"requirePrimary": true,
					"leafOnly": false,
					"maxDepth": 0
				},
				"facets": [
					{"facetId": "019817c8-0000-7000-8100-000000000001", "minValues": 1, "maxValues": 2, "leafOnly": false, "maxDepth": 0},
					{"facetId": "019817c8-0000-7000-8100-000000000002", "minValues": 0, "maxValues": 1, "leafOnly": true, "maxDepth": 1},
					{"facetId": "019817c8-0000-7000-8100-000000000003", "minValues": 0, "maxValues": 3, "leafOnly": true, "maxDepth": 1}
				],
				"tags": {"unknown": "propose"},
				"discovery": {"defaultSort": "editorial"}
			}'::jsonb
		)
		ON CONFLICT (catalog_id, policy_key) DO NOTHING;
