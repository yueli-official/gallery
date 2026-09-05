# Gallery HTTP Result

从 api/ 目录运行：

```sh
go run github.com/yueli-official/foundation/go/httpcontract/cmd/httpcontract -project contracts/http-result/project.json
go run github.com/yueli-official/foundation/go/httpcontract/cmd/httpcontract -project contracts/http-result/project.json -check
```

Project编排使用显式本地Foundation。CI使用已发布v0.4.1分步检查OpenAPI、Go/TS/i18n生成物和前后兼容性。原cmd/errorcatalog已移除，contracts/errors/catalog.json现在是声明式事实源，catalog_gen.go是生成物。

70个operation按端点声明成功状态与公开错误；BFF特有错误单独声明为未被后端operation使用的目录项。创建投稿、评论、处理单、专题和权限资源返回201；专题存在稳定读取路径，返回可由BFF改写的Location。删除图片/评论、无正文下架与事件记录返回204。删除收藏成员与撤销权限仍返回200资源DTO，因为调用方使用更新后的集合/授权状态。

图片、投稿、处理单、标签提案使用items/page/size/total，游标标签使用items/nextCursor。专题详情保留collection元数据，其成员分页使用items/page/size/total。页数由消费者根据total/size推导，不再发送pageSize/totalPages。内部Go字段命名不构成公开兼容协议。不存在code/data/message成功Envelope。

批量结果保留每项标识和success，失败通过failure: Problem投影，traceId来自HTTP中间件响应头；原始cause不会序列化。未知cause安全降级为common.internal。Validation只发稳定code和pointer，不发数据库/提供方原始detail；公开参数受有界预算约束。

前端删除旧Envelope解码和产品自建重认证流程，复用Foundation request/failure resolver与Identity会话能力。图片编辑提供字段inline错误、未知字段摘要与技术详情，失败保留编辑内容。上传XHR错误也进入结构化failure，不显示提供方原始响应。

运行基线：Nuxt4.5.1、Vue3.5.39、Vue Router5.1.0。直接声明Vue依赖，确保共享组件与消费者只使用同一运行时；DevTools默认关闭，可由NUXT_DEVTOOLS=true启用。

本地使用Workspace声明的隔离组合：Gallery3007/API8391、Account3400、Identity8481、Asset8482。原共享Asset运行指纹已过期，故未停止原共享Provider。浏览器登录和测试通过Account Origin，后端继续只监听回环。实际部署与正式制品组合验收另行执行；本Work未发布。
