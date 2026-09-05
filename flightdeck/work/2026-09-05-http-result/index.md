# Gallery 基础合同升级

## Goal
升级Foundation/Identity/Asset消费与HTTP Result合同，完成真实图库发现、投稿、收藏、管理和失败反馈验收并提交。

## Status
Finished

## Current
已完成Gallery本地基础升级：70个HTTP operation、声明式错误目录、OpenAPI/Go/TS/i18n生成物、创建201/Location、无正文204、统一分页、批量Problem与安全前端/BFF失败反馈。Vue/Router显式对齐，删除旧Envelope解码与产品重认证逻辑。现有UI成果独立保存在bbd4ae8，未接入底部导航。

## Next
None。本地迁移已交付；底部导航须等待用户逐站验收后明确启动。正式制品组合、远端CI和发布由Workspace后续门禁拥有。

## References
- [Context](context.md)

## Verification
- GOWORK=off完整Go测试与最终vet通过；新增批量cause保密、分页wire、validation保密测试和trace修正的定向Go测试通过。
- Web79项单测、最终Nuxt类型检查与独立构建目录生产build通过。
- Project -check验证70个operation；已发布v0.4.1 CLI的Go/TS/i18n生成检查通过。
- CLI Playwright累计47个有效用例通过：Journeys13、Management25、HTTP Result4、Responsive5；初次失败修复后只复跑受影响项，无未解决失败。
- 真实上传经同源Asset代理完成并返回投稿201；验证了专题201/Location、评论201/删除204、事件204、批量逐项Problem及一致traceId、收藏新增/页面显示/恢复、后台3次SSR刷新。
- 390/1440px图片编辑的字段错误、未知字段摘要、技术详情、保留草稿、隐藏raw detail和Axe通过；320/1920px及200%有效缩放检查通过。
- 修复Vue重复运行时导致的SSR/评论水合错误；旧测试的标签游标夹具改为当前DTO，专题测试选择有成员的夹具，工作区宽度按共享壳边界验证。
- 证据：web/test-results/e2e/http-result-*、http-result-fields-final/、http-result-favorites-final/、http-result-initial-evidence/和http-result-build.log。

## Runtime boundary
Gallery使用隔离Identity/Asset组合；原共享Asset源码/prepare指纹不匹配，原共享Provider保持运行。浏览器入口http://192.168.5.7:3007，后台/manage。当前Session为20260905T123000Z-43296，后端仅回环8391/8481/8482，Account3400。正式发布/镜像/远端CI不在本轮已完成范围。
