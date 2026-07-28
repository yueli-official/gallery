---
version: 1
slug: "web-app-pages-index-vue"
primary_target: "web/app/pages/index.vue"
related_targets: ["web/app/components/GalleryHeader.vue","web/app/components/GalleryGlobalSearch.vue","web/app/components/GalleryMasonry.vue"]
---

# 首页

## 范围与模式

- 目标：`web/app/pages/index.vue`
- 模式：Experience
- 相关界面：`GalleryHeader.vue`、`GalleryGlobalSearch.vue`、`GalleryMasonry.vue`

## 用户、任务与动作

- 用户在手机或桌面寻找下一张可用图片。
- 首要任务是连续发现；明确查询时可按 Category、Facet 和 Tag 精炼。
- 首要动作是打开图片；次要动作是收藏、换一批、进入专题或投稿。

## 内容与约束

- 只使用真实 Gallery 图片、分类、Tag、Facet 和专题数据。
- 首页只有一个搜索框；顶栏不再重复第二个大型搜索。
- 图片先于产品介绍，搜索能力不得弱于现有目录页。
- 保留路由、数据合同、登录、投稿和颜色模式能力。
- 手机端保持完整搜索与收藏能力，满足键盘、触摸和减少动态效果要求。

## 已选方向

以 Pinterest 式连续图片发现为首要体验，以 Pexels 式查询、相关词和筛选补足搜索。视觉采用成熟图库范式，
但以矿物蓝索引线、轻微切角筛选标签和嵌入图片流的内容块建立月离辨识度。选定构图：
`.impeccable/mocks/home-image-first.webp`。

记忆点：首屏标题不是 Hero，而是瀑布流的第一块；离开这一块后，图片成为页面唯一主角。

## 实现清单

| 可见元素 | 构图承诺 | 实现媒介 |
| --- | --- | --- |
| 顶栏 | 64–72px，品牌左、唯一宽搜索居中、投稿/主题/账户右 | 语义 HTML、Nuxt UI、现有图标 |
| 搜索增强带 | Category 常驻，Facet 与 Tag 可展开；吸顶但不遮挡顶栏 | HTML/CSS、真实 API 数据 |
| 标题块 | 瀑布流第一块，桌面接近竖图比例，移动端回到自然宽度 | HTML/CSS |
| 图片流 | 无边框、高密度、比例保真；桌面首张图形成视觉锚点 | 现有 Asset rendition 与 CSS columns |
| 收藏操作 | hover/focus 显示，触摸端常驻小型按钮 | 现有收藏 API、语义按钮 |
| 图片文案 | 标题在图片底部渐隐层；不展示投稿者 | HTML/CSS |
| 专题与排行 | 作为图片流后的节奏变化，不使用仪表盘卡片 | 真实专题与排行数据 |
| 动效 | 图片轻微抬升/压暗，筛选带吸顶；减少动态时关闭 | CSS |

无需新增品牌摄影或生成式图片资产；真实开发夹具就是视觉内容。

## 未决

- 当前后端 discovery 响应是否已经携带足够 Tag/Facet 用于首页增强带；缺失时只链接到目录查询，不虚构计数。
