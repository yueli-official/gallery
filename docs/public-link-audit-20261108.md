# 其他站点公开地址检查

2026-11-08，只读检查相邻 checkout 的公开内容路由与相关标识生成。没有修改或发布其他站点；本表不代表生产版本审计，也不把所有 API 资源 ID、业务编号、媒体键或授权令牌统一为 8 位。

| 产品 | 源码现状与证据入口 | 后续建议 |
| --- | --- | --- |
| Resource | `resource/api/internal/dao/discovery.go` 直接用资源 UUID 生成 `/resources/{id}`；主键由 `identifier.MustNew` 创建，网页 `pages/resources/[id].vue` 同样用 ID | 已有唯一 slug 可优先评估复用，否则增加独立公开键；页面、Discovery/SEO 和分享入口同步调整 |
| BVideo | `bvideo/api/v1/types.go` 的视频详情 ID 明确校验 UUID；`web/app/pages/videos/[id].vue` 按该值读取；视频表已有唯一 `bvid`，部分写入仍直接调用 `uuid.NewString` | 明确用自身短键或既有唯一 BVID 做视频业务地址；内部生成入口同时对齐 Foundation，不额外复制 BV 算法 |
| Forum | `forum/internal/forum/api.go` 为主题生成 UUID 并直接写 `/t/{id}`；`web/app/pages/t/[id]/index.vue` 按该值查询 | 主题增加独立公开短键；回复锚点、通知、分享和引用地址一起核对 |
| Nav | `nav/api/internal/catalog/service.go` 为分类、分组生成 UUID；公开 `/categories/[categoryId]` 与 `/topics/[groupId]` 使用这些 ID | 分类/专题入口可改稳定 slug 或公开短键；外部导航目标链接不需要改 |
| Paste | `paste/internal/paste/service.go` 使用 `identifier.Allocate(CompactURLV1)`；公开 `/p/[code]` | 已符合 8 位短码方向，不需重复迁移 |
| Shortlink | `shortlink/internal/link/service.go` 使用 `Allocate(ShortLocatorV1)` | 已有共享 6 位短链语义，保留该产品格式 |
| Identity | `identity/internal/user/identifiers.go` 已有共享 8 位稳定公开用户键 | 本次无需改为另一套格式；handle 与用户键各有用途 |
| Blog / Docs | Blog 的文章、分类、系列主路由使用 slug；Docs 使用 collection slug 和文档路径 | 主内容地址不因内部 UUID 改写；作者页等非主内容入口另按实际公开键核对 |
| WWW / Distribution / Shop | WWW 的内容与应用、Distribution 软件、Shop 商品页面使用 slug | 已有可读业务地址，无需换成随机短键 |
| YTV | 标题与播放路径带 `source`、外部内容 `id`、集数 | 属于来源组合定位，不按 UUID 长度问题直接替换 |

本轮未检查全部后台记录、订单和支付标识、激活/访问令牌、Asset 媒体协议及所有历史 URL。它们需要按领域合同判断，不能由本次公开图片短键迁移推断调整。
