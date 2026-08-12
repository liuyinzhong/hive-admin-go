# 挂号候诊队列规则

本文定义挂号单签到后生成候诊记录和签到序号的首期规则。挂号主流程、状态矩阵和权限仍以根目录 `CONTEXT.md` 为准。

## 业务规则

### MED-QUE-001 候诊队列边界

一条实际排班对应一个候诊队列，不按号源档位、预约方式或挂号时间拆分。候诊记录必须保存 `schedule_id`，用于定位队列和建立高频查询索引。

### MED-QUE-002 签到排号

只有状态为 `10（已支付）` 的挂号单可以签到。签到时按所属实际排班当前最大的 `queue_sequence` 加一生成签到序号；没有候诊记录时从 `1` 开始。签到序号生成后不可调整。

挂号单状态更新为 `30（已签到）`、候诊记录创建和挂号单日志写入必须处于同一数据库事务，任一步失败均整体回滚。

### MED-QUE-003 并发与唯一性

生成签到序号前锁定所属实际排班记录，使同一实际排班下的并发签到串行分配序号。数据库同时保证：

- 一个挂号单最多一条候诊记录；
- 同一实际排班下的签到序号不得重复；
- 候诊队列按 `schedule_id + queue_status + queue_sequence` 建立查询索引。

### MED-QUE-004 候诊状态

当前定义 `0（候诊中）`、`10（已叫号）`、`15（已过号）`、`20（接诊中）` 和 `30（已完成）`。候诊记录创建时 `queue_status=0`、`call_count=0`。

叫号、过号、重新叫号、开始接诊和完成接诊规则见 [医生接诊与处方规则](./outpatient-workbench-draft.md)。每次首次或重新叫号使 `call_count + 1`。

### MED-QUE-005 表字段

`med_visit_queue` 只保存候诊生命周期自身需要的数据：

| 字段 | 含义 |
|---|---|
| `queue_id` | 候诊记录 UUID 主键 |
| `registration_id` | 挂号单 ID |
| `schedule_id` | 实际排班 ID，即队列边界 |
| `queue_sequence` | 同一实际排班内的签到序号 |
| `queue_status` | 候诊状态：`0` 候诊中、`10` 已叫号、`15` 已过号、`20` 接诊中、`30` 已完成 |
| `call_count` | 累计叫号次数，初始为 `0` |
| `create_date` | 候诊记录创建时间，即签到排号时间 |
| `creator_id` | 执行签到的系统用户 ID |

患者、医生、科室、号源档位及其名称不在候诊表重复保存，需要时通过挂号单取得。

### MED-QUE-006 接口响应、查询与权限

挂号详情和签到动作响应通过 `queueInfo` 返回候诊序号、候诊状态、叫号次数、创建时间和创建人。未签到详情中的 `queueInfo` 为空或省略；其它动作响应和挂号列表首期不加载候诊信息，列表也不增加候诊序号列。

排班列表响应通过 `queueCount` 返回该排班下候诊记录总数，统计全部候诊状态。候诊队列通过 `GET /api/medical/schedules/{scheduleId}/visitQueues` 一次性返回，不分页，始终按 `queue_sequence ASC` 排序，并由独占权限码 `medical:visitQueue:list` 保护。

队列项中的患者编号、挂号单号和号源时段来自挂号单快照。患者姓名和手机号必须由服务端无条件脱敏，即使调用方拥有 `medical:patient:viewSensitive` 或系统管理员权限也不返回完整值。候诊队列接口只读，不提供详情跳转或任何状态操作。

## 本期不包含

- 候诊序号调整、插队和优先级；
- 独立叫号日志和每次叫号时间；
- 过号后自动重新排号。

## 迁移、兼容性与回滚

- 本次只新增 `med_visit_queue`，不修改已有挂号、排班和挂号日志表。
- 发布时必须先执行 `migrations/20260810_create_med_visit_queue.sql`，确认建表成功后再发布后端代码；否则挂号详情和签到动作查询候诊表时会失败。
- 不提供历史候诊记录补录或修复脚本，系统按新建候诊表后的规则运行。
- 回滚时先停止新签到流量并恢复不读取候诊表的旧版后端。只有确认新增候诊数据可以丢弃后，才由实施人员手动执行 `DROP TABLE med_visit_queue`；删除表会永久丢失迁移后的签到序号和叫号次数。
- 完成接诊只允许通过正式门诊病历完成接口执行，要求候诊记录处于 `20（接诊中）`。

## 实现入口

- 数据库脚本：`migrations/20260810_create_med_visit_queue.sql`。
- Model：`models/medical_registration.go`。
- Service：`services/medical_registration_service.go`。
- 前端 API：`hive/apps/web-antdv-next/src/api/medical/registration.ts`、`schedule.ts`。
- 前端页面：`hive/apps/web-antdv-next/src/views/medical/registration`、`schedule/calendar`。
