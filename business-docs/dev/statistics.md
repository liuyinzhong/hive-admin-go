# 开发统计业务规则

开发统计是对研发工作项的查询投影，不修改项目、版本、需求、任务或缺陷。

## 业务规则

### DEV-STAT-001 日统计按指定日期查询

getTaskFindDay 接收 YYYY/MM/DD 格式日期，格式错误或为空时拒绝查询。统计口径以当前服务中的任务日期字段和未删除数据为准。

### DEV-STAT-002 年统计要求合法年份

getTaskFindYear 按年份返回年度分布，非法年份不进入查询。

### DEV-STAT-003 工作台枚举是全局概览

getWorkspaceEnum 汇总未删除需求、任务和缺陷总数，并把状态 0 计入当前活动数量。它不是权限数据，也不替代各模块列表。

### DEV-STAT-004 统计接口只读

当前统计路由需要登录，但没有注册独立原子权限码。若后续纳入细粒度权限，必须同步 Router、菜单按钮、前端展示和本文。

## 代码入口

- Model/DTO：models/dev_statistics.go。
- Service：services/dev_statistics_service.go。
- Controller：controllers/dev_statistics_controller.go。
- Router：/api/statistics/dev/getTaskFindDay、getTaskFindYear、getWorkspaceEnum。
- 前端：dashboard、开发版本详情及 src/api/statistics/dev.ts。
