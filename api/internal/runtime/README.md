# Gallery API 运行时装配

本目录只负责 Gallery API 的进程边界装配，包括鉴权、HTTP 中间件、健康检查、数据库、OpenAPI、遥测和 Webhook。

领域规则仍归 `internal/gallery` 等业务包所有；跨产品可复用的稳定原语由 Foundation 提供。这样 Gallery 独立迁出后，不需要依赖 Platform 仓库里的 `gokit`。

新增能力时先判断归属：

- Gallery 特有的启动策略和错误映射放在这里。
- 能被多个产品稳定复用的协议原语放入 Foundation。
- 领域决策不要放入本目录。
