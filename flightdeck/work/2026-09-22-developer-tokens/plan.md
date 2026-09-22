# 实施与验收计划

- [x] 产品认证与授权：接入 Foundation PAT verifier、可信权限目录、用户 capability、现有后台 capability 与显式路由 allowlist。
- [x] 媒体链路：为 `media.upload` 提供在线授权回查，只签发 Gallery 投稿 Profile。
- [x] 持久目录：授权定义升级到 v4，迁移现有实例且保留已有角色、授权与策略修订。
- [x] 产品合同与工具：更新 OpenAPI/HTTP Result 合同、开发者令牌指南和可重复本地入口。
- [x] Workspace 接线：登记 Identity PAT application、Gallery verifier 与 Asset authority，不影响其他 Target。
- [x] 验证：相关 Go 测试、合同 freshness、必要 Web 检查，以及真实 Workspace 组合中的桌面/手机 Account 权限目录、Asset 上传与 Gallery API 闭环。
