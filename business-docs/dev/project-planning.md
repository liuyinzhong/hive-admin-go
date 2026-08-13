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

releaseStatus 以数字字符串在接口中传输，显示名称来自 RELEASE_STATUS 字典。创建、整体编辑和“流转”接口都可以写入状态；当前后端未维护固定状态流转矩阵，不能仅依据接口名假设只能前进一个状态。

### DEV-PLAN-005 版本流转追加变更记录

新建、修改、流转和删除会按版本业务类型写入变更记录。流转可附带富文本说明。

### DEV-PLAN-006 只有初始状态版本可删除

当前只有 releaseStatus=0 的版本允许逻辑删除。最新版本接口按当前实现返回项目范围内的最新记录，调用方必须传递并核对项目条件。

## 权限和接口

- 项目：dev:project:list、create、detail、update。
- 模块：dev:module:list、create、detail、update、delete。
- 版本：dev:version:list、latest、create、advance、detail、update、delete。
- 路由：/api/dev/projects、/modules、/versions。

## 代码入口

- Model：models/models.go 中 DevProject、DevModule、DevVersion。
- Service：services/dev_project_service.go、dev_module_service.go、dev_version_service.go。
- Controller：controllers/dev_*_controller.go。
- 前端：hive/apps/web-antdv-next/src/views/dev/project、dev/versions。
