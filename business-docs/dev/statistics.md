# 开发统计业务规则

开发统计是对研发工作项的查询投影，不修改项目、版本、需求、任务或缺陷。

## 业务规则

### DEV-STAT-001 日统计按指定日期查询

getTaskFindDay 接收 YYYY/MM/DD 格式日期，格式错误或为空时拒绝查询。统计口径以当前服务中的任务日期字段和未删除数据为准。

### DEV-STAT-002 年统计要求合法年份

getTaskFindYear 按年份返回年度分布，非法年份不进入查询。

### DEV-STAT-003 工作台枚举是当前范围概览

getWorkspaceEnum 汇总当前用户数据范围内的未删除需求、任务和缺陷，并把状态 0 计入当前活动数量。需求按创建人或参与人、任务按创建人或执行人、缺陷按创建人或处理人统计；它不替代各模块列表。

### DEV-STAT-004 统计接口只读

当前统计路由需要登录，但没有注册独立原子权限码。若后续纳入细粒度权限，必须同步 Router、菜单按钮、前端展示和本文。

### DEV-STAT-005 统计和来源列表使用同一数据范围

日统计、年统计和工作台枚举都先应用当前角色数据范围再聚合，不能返回全局总数后由前端过滤。多角色并集、部门变化和无范围时的关闭访问语义与 [系统数据权限规则](../system/data-permission.md) 一致。

### DEV-STAT-006 版本详情统计

`GET /dev/versions/statistics?versionId=...` 聚合指定版本下的需求、任务、缺陷概览、完成趋势与多维分布。先按版本创建人校验版本可见，再对各工作项套用与对应列表一致的数据范围（需求按创建人/参与人、任务按创建人/执行人、缺陷按创建人/修复人），最后按 `version_id` 过滤。完成口径：需求已关闭=`story_status=99`、任务已完成=`task_status=99`、缺陷已修复=`bug_status=20`。分布项 `name` 直接返回可展示文案：字典维度取 `sys_dict` 中文 label，人员维度取 `real_name`，模块维度取 `module_title`，空值归为“未分配”；完成趋势按版本 `start_date~end_date` 逐日，无起止时取创建日到当天，最多 366 天。该路由与 `/dev/versions/:versionId` 共存，静态路由优先，注册时保持静态路径在参数路由之前。

## 代码入口

- Model/DTO：models/dev_statistics.go。
- Service：services/dev_statistics_service.go。
- Controller：controllers/dev_statistics_controller.go、dev_version_controller.go。
- Router：/api/statistics/dev/getTaskFindDay、getTaskFindYear、getWorkspaceEnum，以及 /api/dev/versions/statistics。
- 前端：dashboard、开发版本详情及 src/api/statistics/dev.ts、src/api/dev/versions.ts。
