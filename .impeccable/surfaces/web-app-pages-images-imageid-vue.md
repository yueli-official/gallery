---
version: 1
slug: "web-app-pages-images-imageid-vue"
primary_target: "web/app/pages/images/[imageId].vue"
related_targets: ["web/app/components/GalleryViewer.vue","web/app/components/GalleryImageGrid.vue","web/app/components/GalleryMasonry.vue","web/app/components/GalleryFavoriteGrid.vue"]
---

# 图片详情页

- 范围：`/images/:imageId` 公共图片详情及所有公共列表到详情页的进入方式。
- 模式：Experience。图片本身先于界面，详情信息帮助判断、保存和继续探索。
- 用户任务：完整观看单张图片，轻量确认标题、简介、来源、分类和标签；收藏或分享；按需查看尺寸；继续浏览相关图片。
- 主要动作：观看图片。收藏与分享是紧邻图片的图标动作；查看尺寸、举报和补充来源进一步降级。
- 内容与证据：使用真实 Gallery Image、Asset rendition、浏览与收藏指标、来源 URL、分类、标签和相关图片，不增加投稿者或社交信息。
- 约束：取消 Quick View/Modal；桌面和手机必须共享独立路由；保留键盘前后导航、缩放、全屏、滑动与错误状态；顶部全局导航保持不变。
- 方向：公共图片索引中的“单张挂展”。首屏以大幅原图和安静的信息栏构成，相关推荐从侧栏移到下方成为完整的开放存储区。
- 记忆点：图片像被放到一张宽阔检视台上，界面只在边缘提供精确动作；向下滚动立即进入同类图片阵列。
- 已选构图：`.impeccable/mocks/viewer-comp-b-bottom-dock.png`。全宽图片占据首屏主要视觉重量，信息在图片下方形成低位信息坞；不采用固定右侧栏、常驻缩略图索引或四项数据分组，非核心信息通过省略号渐进披露。
- 未决项：无。

## 实现清单

| 可见部分 | 实现媒介 | 构图承诺 |
| --- | --- | --- |
| 全局顶栏 | 既有 `GalleryHeader` | 保持单行、搜索与现有导航，不在详情页另造导航系统 |
| 图片展台 | 语义化 HTML/CSS 与真实 Asset rendition | 桌面接近满宽，图片视觉重量高于标题和数据；移动端无侧栏 |
| 缩放、全屏、前后浏览 | 既有 Nuxt UI 与 Tabler 图标 | 控件贴边、低对比，hover/focus 时增强，不遮挡主体 |
| 低位信息坞 | HTML/CSS Grid、语义链接与既有按钮 | 标题 1.05-1.25rem、任意中文可断行；收藏和分享仅保留带无障碍名称的图标；浏览和收藏是弱化的小图标数字 |
| 简介、标签与更多信息 | 语义段落、链接、原生 `details` 和按钮 | 图片数据中的简介直接位于标题下方且不加重复标签；标签与来源保持低对比；尺寸、举报、补充来源按需展开，点击外部、Esc 或切图自动关闭 |
| 相关图片 | 真实关联数据与现有 rendition | 位于信息坞之后，延续公开图片索引，不进入固定侧栏 |
| 构图探针 | 仅作设计证据的 PNG | 不进入产品运行包，不栅格化正文、控件或图片内容 |
