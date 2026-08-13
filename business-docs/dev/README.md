# 开发管理业务手册

本目录记录项目、模块、版本、需求、任务、缺陷、变更时间线和研发统计的当前行为。

## 阅读顺序

1. 先读 [开发管理领域词汇](./CONTEXT.md)。
2. 项目、模块和版本读 [项目与版本规则](./project-planning.md)。
3. 需求、任务、缺陷和变更时间线读 [研发工作项规则](./work-items.md)。
4. 仪表盘和工作台统计读 [开发统计规则](./statistics.md)。
5. 任务异步导出还要读 [下载中心规则](../system/download-center.md)。
6. 涉及页面时继续阅读前端 hive/business-docs/dev。

## 模块覆盖

| 模块 | 规则正文 | 后端入口 |
|---|---|---|
| 项目、模块、版本 | [project-planning.md](./project-planning.md) | dev_project、dev_module、dev_version |
| 需求、任务、缺陷、变更记录 | [work-items.md](./work-items.md) | dev_story、dev_task、dev_bug、dev_change_history |
| 开发统计 | [statistics.md](./statistics.md) | /api/statistics/dev |

## 规则编号

- DEV-PLAN-*：项目、模块和版本。
- DEV-ITEM-*：需求、任务、缺陷和变更记录。
- DEV-STAT-*：开发统计。
