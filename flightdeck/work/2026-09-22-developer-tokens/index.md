# Gallery 开发者令牌

## Goal

让用户可以在 Account 创建绑定 `gallery-main-web` 的细粒度开发者令牌，通过 Gallery 与 Asset 的现有 API 上传图片、创建和管理自己的投稿、管理收藏、发表评论；当前 Gallery 运营者可额外选择现有图片、投稿、专题、分类、处理单、发现策略和评论治理能力。每次请求都重新验证 PAT，并取令牌 scope 与账号当前 Gallery 权限的交集。

完成时应满足：普通用户只能操作自己的投稿和收藏；媒体上传只获得 `gallery-submission` Profile；运营能力不会授予普通用户；首位管理员认领、授权控制台和未声明路由不向 PAT 开放；显式 PAT 失败不回退浏览器 Cookie；开发者指南、OpenAPI、运行配置与真实本地浏览器/API 验收一致。

## Status

待查收

## Current

2026-09-22 本地实现与真实组合验收已完成。Gallery 已升级到 Foundation Go `v0.5.0`，接入 PersonalTokenVerifier、可信权限目录、显式路由 allowlist 和 Asset 媒体授权回查。Account 当前管理员目录展示 25 项 Gallery 权限：普通账号可委派上传、投稿、自己的投稿/收藏与评论；运营者/管理员可按当前账号权限额外委派后台读取、审核、编排、分类治理、处理单处置、站点/发现设置与评论治理。Asset 平台策略、授权控制台、角色申请/授予、Guest Claim 和未声明路由保持拒绝。

授权目录从 v3 升级到 v4，新增六项 authenticated access-layer capability 与 `0012_authorization_personal_tokens_v4` 迁移；Workspace 对已有 Gallery 数据库实际执行并再次校验迁移 checksum/digest，五进程正常启动，现有角色、授权和数据保留。

验证通过：`GOWORK=off go test -timeout 60s ./...`、全量 `go vet ./...`、72 项 HTTP Result/OpenAPI freshness，以及 PAT 权限交集、动态撤权、可信目录、媒体 Profile 和路由边界测试。最终隔离 Session `20260922T062602Z-31280` ready；CLI Playwright 1/1 通过，覆盖 Account UI 创建 25 权限令牌、桌面/390px 手机目录、Asset upload-init/上传/finalize、Gallery 投稿/评论/后台读取、敏感端点 403、撤回、撤销后 Gallery 401 与 Asset 拒绝。截图已目检且无横向溢出，临时令牌/评论已清理、投稿已撤回，五进程验收后已停止。

本 Work 只负责 Gallery，本地实现与验收，不部署生产。Paste 已在独立 Work 完成本地交付并等待用户确认；Nav 与 Distribution 后续各自建立单站 Work。

## Next

等待用户查收 Gallery 开发者令牌交付；收到反馈后处理对应问题。2026-09-22 用户已授权本地提交，尚未推送或生产部署。

## References

- [上下文与权限边界](context.md)
- [实施与验收计划](plan.md)
- [Foundation 个人令牌合同](../../../../foundation/flightdeck/knowledge/authorization/personal-access-tokens.md)
