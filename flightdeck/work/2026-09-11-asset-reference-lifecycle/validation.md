# 本地验证与接入

## 范围
投稿与图片事实源；撤稿撤销投稿，隐藏图片保留，删除后重试撤销公开派生图授权。

## 已通过
- 当前产品业务回归和 API main 编译；使用明确本地 Asset checkout 及已记录 Foundation 源码快照的临时 modfile。
- `internal/assetreferences` 真实 PostgreSQL 临时表测试：重复/共享素材、替换、删除或撤销、恢复；Shop 另验证组合交付、图集与合集，Gallery 另验证撤稿与公开授权失败重试，Commerce 另验证多文件及跨 site 隔离。
- SDK 真实 PostgreSQL + HTTP 夹具验证：读取失败不清空、503 后恢复、不让某类未解析素材阻断其他类型对账；MediaKey 节点解析及禁止重定向泄露机器凭据。
- 消费者 JSON Schema 校验通过。无新数据库迁移。

## 运行约束
Asset Go v0.3.0 尚无 referencesync；本轮是本地源码覆盖验证，未发布 SDK。SDK 依赖的新间接版本和校验和已同步到产品 go.mod/go.sum，无硬编码本地 replace。
共用环境变量：ASSET_BASE_URL、ASSET_REFERENCE_TOKEN_URL、ASSET_REFERENCE_CLIENT_ID、ASSET_REFERENCE_CLIENT_SECRET、ASSET_PUBLIC_ORIGIN。机器身份需要 asset:sign 及此消费者的 Registration binding；缺失配置拒绝启动，不能默默跳过登记。
Commerce 另需 ASSET_REFERENCE_SITE_KEY 指定订单业务 site（默认 shop）；本仓库演示目录显式 COMMERCE_REFERENCE_DEMO_CATALOG=.data/consumer-catalog.json，不能将其他消费者目录交给 Commerce 扫描。Gallery/Licensing 复用原 Asset 机器配置。

## 实际运行验收
2026-09-11 真实运行补登 125 图片/48 投稿/6 头像/6 封面，既有 12 隐藏图片全部保留、4 删除图片无公开授权残留。新测试投稿审核发布、隐藏保留引用、删除全部解引用与素材可删除均通过；另验证待审投稿撤回解引用。
CLI Playwright 桌面/移动端引用弹窗、删除保护和无溢出/脚本异常通过。图库独立组合定向健康检查全部 ready；修复脚本错误检查全 Workspace 导致假失败。
证据 E:/tmp/yueli-consumer-references-20260911/gallery-lifecycle-report.json、gallery-references-{desktop,mobile}.png。测试素材已删除，撤回/删除的业务审计行按产品规则保留。未部署或提交。
