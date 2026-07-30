# 边界与已知状态

## 已有用户改动

开始本工作时 Gallery 工作树已有 README、公共图片详情/Viewer、目录状态、主题样式、E2E 契约和测试改动，
并已删除 `doctor.yaml`。这些改动属于用户，必须保留。

## 已有开发控制面

Workspace 的 `environments/gallery-local/environment.yaml` 已把 Identity、Account、Asset、Gallery API/Web
按能力依赖组合，并提供 `workspace dev prepare/up/status/logs/down`；Gallery 仓不再拥有本地多仓进程清单。

## 已关闭的部署缺口

- 根目录已有三个标准 Compose 入口、对应严格环境样例和精确部署锁。
- API 镜像已有 runtime、migration、bootstrap、bindingcheck 与 healthcheck 目标。
- `GALLERY_DATABASE_URL` 会转为 GoFrame 嵌套数据库环境，Web 在启动时校验 URL、Cookie 与 seal secret。
- Web 暴露独立 `/healthz`，Compose 不再用会触发 SSR 和下游调用的首页作为健康门禁；Nuxt 4 冻结
  runtime config 只做一致性校验，Cookie 安全派生值由 Compose 显式传入。
- Identity 的默认 `/api/v1` BFF 只属于 Identity；Gallery Web 必须通过具名 `gallery` 客户端进入
  `/api/gallery`。该 BFF 使用统一运行库限制方法、Header 和请求体，并从 Identity 密封会话或 Guest
  会话解析服务端 Bearer，浏览器不持有访问令牌。
- Gallery BFF 只为 `/api/gallery/me/*` actor 范围的首次 GET 建立 Guest 会话；普通公开 GET 保持无状态，
  避免浏览目录时无条件创建持久 Guest 记录。Gallery E2E 本地地址从 Workspace 的 `LOCAL_*` 变量派生，
  确保站点、Identity、Account 与 OIDC cookie 使用同一 LAN origin 和端口。
- README 只保留 Workspace 本地开发和 Compose 部署入口，不再把 Doctor 当作接口。
- 完整独立拓扑已覆盖 PostgreSQL 18 + pgvector、Identity OIDC client、Asset profile/rendition 和 migration 顺序。

## 不能声称的验证

当前主机没有 Docker Engine，因此不能执行 `docker compose up` 或真实容器健康链路。已使用官方
Compose 5.1.4 独立 CLI 完成三种模型解析，源码测试、Linux 构建和 CI 门禁也已通过；真实三拓扑验收必须在
Docker 主机补做。

独立 Asset 的扫描移除目前仍是未发布工作树；Gallery 自持 Compose 已删除 ClamAV，但部署锁只能在 Asset
Asset 扫描移除已发布为 `v0.3.0 / 7007ec9503219635b769a99c68bdd1b27ccb4776`；Gallery 产品部署锁和
默认构建 revision 已同步，旧 `v0.1.2` 不再是部署输入。
