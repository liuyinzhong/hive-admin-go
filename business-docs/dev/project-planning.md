# 项目与版本业务规则

## 项目和模块

### DEV-PLAN-001 项目是研发数据的归属入口

项目标题在未删除项目中唯一。当前项目提供列表、创建、详情和更新，没有删除接口；模块、版本和工作项通过 projectId 归属项目。

### DEV-PLAN-002 模块属于单一项目

创建模块时，同一项目内模块标题不能重复。模块支持排序、详情、修改和批量逻辑删除。

当前实现存在一项需注意的差异：创建时按“同一项目”检查标题，更新时的重复检查没有带项目条件，可能把其它项目的同名模块也判定为冲突。修改该逻辑前必须先确认目标规则并同步本文。

## 版本

### DEV-PLAN-003 版本号在项目内唯一

版本关联一个项目，可记录版本类型、负责人、起止日期、发布状态和变更说明。同一项目下版本号不能重复，结束时间不能早于开始时间。

### DEV-PLAN-004 发布状态由研发字典表达

releaseStatus 以数字字符串在接口中传输，显示名称来自 RELEASE\_STATUS 字典。创建、整体编辑和“流转”接口都可以写入状态；当前后端未维护固定状态流转矩阵，不能仅依据接口名假设只能前进一个状态。

### DEV-PLAN-005 版本流转追加变更记录

新建、修改、流转和删除会按版本业务类型写入变更记录。流转可附带富文本说明。

### DEV-PLAN-006 只有初始状态版本可删除

当前只有 releaseStatus=0 的版本允许逻辑删除。最新版本接口按当前实现返回项目范围内的最新记录，调用方必须传递并核对项目条件。

### DEV-PLAN-007 项目维度全局共享，版本按创建人受限

项目和模块是研发公共规划主数据，不按创建人应用角色部门范围，仍由各自原子权限控制。版本列表、全量选项、最新版本、详情、修改、流转和删除按版本 `creator_id` 使用当前角色数据范围；跨范围版本不能通过 ID 直接读取或写入。

### DEV-PLAN-008 项目用户统一管理人员归属

项目用户表 `dev_project_user` 以 `project_id + user_id` 唯一约束管理项目成员。项目成员仅表示人员归属，不携带负责状态；需求的状态负责人在需求流转时指定并写入 `dev_story_user`。

项目用户管理接口采用全删全插模式：PUT 请求提交完整用户列表，后端先删除该项目所有行再插入新行。

项目用户不做记录级数据权限过滤：登录用户可见任意项目的成员列表；写入（全量保存）需要 `dev:project:user` 权限码。

需求参与人、缺陷修复人、任务执行人的候选范围均为项目用户：候选下拉只展示项目用户。需求流转指定的状态负责人必须是该项目成员。

## 权限和接口

* 项目：dev:project:list、create、detail、update。

* 项目用户：dev:project:user（全量保存）；查询登录即可。

* 模块：dev:module:list、create、detail、update、delete。

* 版本：dev:version:list、latest、create、advance、detail、update、delete。

* 路由：/api/dev/projects、/project-users、/modules、/versions。

## 代码入口

* Model：models/models.go 中 DevProject、DevProjectUser、DevModule、DevVersion。

* Service：services/dev\_project\_service.go、dev\_project\_user\_service.go、dev\_module\_service.go、dev\_version\_service.go。

* Controller：controllers/dev\_\*\_controller.go。

* 前端：hive/apps/web-antdv-next/src/views/dev/project、dev/versions、dev/base/baseSchema.ts。

