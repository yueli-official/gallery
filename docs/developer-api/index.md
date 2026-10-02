# 开发者 API 生成入口

- [能力目录](developer-api.json)：可申请权限、允许路由、媒体 Profile 与 scope 组合，供脚本和助手读取。
- [开发者 OpenAPI](developer-openapi.json)：只包含 PAT 白名单及公开兼容操作。
- [接口说明](developer-api.md)：自动生成的人类可读接口目录。
- [使用指南](../developer-tokens.md)：投稿、上传和后台治理的使用边界。

声明直接来自 Gallery 运行时权限目录、路由 allowlist、公开读取路由和媒体授权 Profile，字段来自产品 OpenAPI。

~~~powershell
./scripts/generate-developer-api.ps1 -GeneratorPath ../foundation/tools/developer-api/generate.py
./scripts/generate-developer-api.ps1 -GeneratorPath ../foundation/tools/developer-api/generate.py -Check
~~~

-Check 会重新编译导出声明和 OpenAPI，并比对声明、完整 OpenAPI 与三份生成资料。生成成功只表示当前源码支持，不代表生产部署状态。
