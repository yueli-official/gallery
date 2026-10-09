# 2026-10-09 正式发布准备

用户明确纳入 首次发布 范围，Distribution 排除。本文件记录准备证据，不表示已经上线。

域名候选：gallery.yuelili.com（Nav/YTV 保留原入口；其余新域名等待用户确认）。服务器正式产品目录拟为 `/projects/yuelili.com/gallery`。当前 checkout `03410417859d80e8758ae437b438d059c02a6670`；应用制品来源以下方独立 SHA 为准，后续文档或迁移工具提交不会自动改变既有应用制品。

后端源 `c7b2c2763e92ec5b43b9c6165fd8b64063f09092`：GOWORK=off 的 go test ./...、Linux/amd64/CGO=0 固定构建通过；命令 bindingcheck, bootstrap, gallery, healthcheck。迁移文件 24 项 SHA 记录在 manifest.json，数据库集成覆盖以测试配置/日志为准，不能把未设置数据库的单测当生产验证。

前端：固定源 03410417859d80e8758ae437b438d059c02a6670；冻结安装、前端测试、类型、生产构建、已安装内容 verified 均通过。Web tar SHA-256 2186e7d31c3ecbf2d08bdd9df9005535035dae8607426020e8d6c13edd00ecfa。UI 为 .6 固定 SHA 06407c95454b215c61b5cf3ef432476de6677d3a78542b07fe706a7c52106c0a，sourceOverrides=[]。

迁移入口：api/manifest/sql/migrations + cmd/bootstrap。只在新建/明确备份的目标数据库预检并应用，不运行 devseed，不复制本地账号/业务数据，不修改旧 checksum。现有生产 Provider 按精确能力绑定，禁止直接套用 standalone compose 重建中央服务。

所需注册：Identity + Asset gallery-submission/profile 与 service-token 授权。

管理员/所有权：现有启动入口仍要求 Bootstrap subject，必须补齐共享首位管理员认领。

网络与密钥：Web/API 仅发布宿主机回环端口。所有 browser issuer/callback 使用正式 HTTPS Origin；Seal Secret、服务客户端 secret  由私有部署配置独立生成并备份，不使用开发默认值。证书需覆盖所选域名，现有证书不能作为新域名证明。



本次响应式修复完成 CLI Playwright 36 个图片／评论／投稿场景和 8 个专题编辑器场景，Axe、横向溢出与页面错误检查通过；隔离组合已停止，固定依赖已恢复。证据 gallery-followup-result.json、gallery-editor-result.json。

完成门禁：固定源码/依赖与制品可追溯 → 域名/TLS → Provider 注册及绑定检查 → 数据库迁移预检/初始化 → 候选健康/API → 精确切换 → 正式域名 CLI Playwright 手机/桌面、明暗主题、实际交互与登录门禁。初始化认领由用户显式完成；线上管理写入不以匿名测试代替。

现有部署资料：deploy\asset.consumer.json, deploy\compose.gallery.yaml, deploy\contracts\requirements.json, deploy\deployment.lock.json, deploy\includes\asset\compose.managed-identity.yaml, deploy\includes\asset\compose.yaml, deploy\includes\asset\config.yaml, deploy\includes\identity\compose.yaml, deploy\topologies\gallery.hybrid.yaml, deploy\topologies\gallery.standalone.yaml。

本机证据 `C:/Users/yl/.codex/task-artifacts/sites-ui-upgrade/20261009-3/gallery` 与 `gallery-api`；原始依赖对照在相邻 20261009-2/first-release-original-dependencies.json。源码快照/依赖为可再生过程产物，收尾删除；固定归档、二进制、Web 包、哈希、日志和迁移准备保留用于后续发布。
