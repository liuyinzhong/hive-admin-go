# 研发工作项业务规则

需求、任务和缺陷共享项目、版本、模块等研发上下文，但拥有各自编号、状态和动作。

## 共用规则

### DEV-ITEM-001 工作项编号由数据库生成

需求、任务和缺陷分别使用自增业务编号。详情接口按业务编号查询，更新和动作接口按 UUID 操作，两者不可混用。

### DEV-ITEM-002 状态展示由字典驱动

storyStatus、taskStatus、bugStatus 和 bugConfirmStatus 以数字字符串在接口中表达，页面名称来自对应系统字典。当前后端的 advance 动作会解析目标值后直接更新，没有统一的合法流转矩阵；字典值、页面步骤和后端可接受状态必须一起核对。

### DEV-ITEM-003 变更记录形成只读时间线

创建、整体修改、局部字段修改、状态流转、确认和删除会追加变更记录。查询按 businessId 返回时间线；变更记录没有修改或删除接口。独立创建接口只用于明确的补充记录场景。

## 需求

### DEV-ITEM-004 需求连接版本、模块和参与人

需求归属项目，可选关联版本和模块，并可关联多名参与用户、文件、类型、级别和来源。支持单条与批量创建。

局部字段接口只允许修改 userIds、storyType、storyLevel、source。只有 storyStatus=0 的待评审需求可以删除。

## 任务

### DEV-ITEM-005 任务连接需求、执行人和工时

任务归属项目，可选关联需求、版本和模块，并维护执行人、任务类型、计划/实际工时与起止时间；结束时间不得早于开始时间。支持单条与批量创建。

局部字段接口只允许修改 userId、taskType、startDate、endDate。只有 taskStatus=0 的待执行任务可以删除。

### DEV-ITEM-006 任务导出进入下载中心

任务导出按照当前列表筛选和排序创建 devTask 来源的异步下载任务，不在请求中同步生成文件。创建权限为 dev:task:export，任务状态、下载和保留期遵循系统下载中心规则。

## 缺陷

### DEV-ITEM-007 缺陷同时维护处理状态和确认状态

缺陷可关联项目、版本、模块、需求和处理人，并维护级别、环境、来源和类型。支持单条与批量创建。

局部字段接口只允许修改 userId、bugLevel、bugEnv、bugType、bugSource。只有 bugStatus=0 的待确认缺陷可以删除。

### DEV-ITEM-008 缺陷确认是独立动作

确认状态写为 1 时，当前实现同时把缺陷状态改为 10；确认状态写为 2 时，缺陷状态改为 1。状态重新流转到 0 时会重置确认状态。调整这些数值含义必须同步 BUG_STATUS、BUG_CONFIRM_STATUS 字典和前端动作。

## 权限

- 需求：dev:story:list、create、batchCreate、detail、update、fieldUpdate、advance、delete。
- 任务：dev:task:list、create、batchCreate、detail、update、fieldUpdate、advance、delete、export。
- 缺陷：dev:bug:list、create、batchCreate、detail、update、fieldUpdate、advance、confirm、delete。
- 变更记录：dev:changeHistory:list、create。

## 代码入口

- Model：models/models.go 中 DevStory、DevTask、DevBug、DevChangeHistory。
- Service：services/dev_story_service.go、dev_task_service.go、dev_bug_service.go、dev_change_history_service.go。
- Router：/api/dev/storys、/tasks、/bugs、/changeHistory。storys 是当前既有接口拼写，修改需同步前后端。
- 前端：hive/apps/web-antdv-next/src/views/dev/story、task、bug。
