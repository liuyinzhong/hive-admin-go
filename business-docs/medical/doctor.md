# 医生管理业务规则

医生档案管理医生个人和执业信息，并通过医生科室关系确定其可配置排班的临床科室。

## 业务规则

### MED-DOC-001 医生编号唯一

医生编号和姓名必填。医生编号在所有未删除医生中唯一；删除后的编号当前可再次使用。

### MED-DOC-002 系统用户绑定可选且一对一

医生可不绑定系统用户。绑定时，用户必须存在、启用且未删除，同一系统用户只能绑定一个未删除医生档案。

### MED-DOC-003 至少一个出诊科室且主科室必须在其中

创建和修改医生时必须提交至少一个不重复的出诊科室，并指定其中一个为主科室。所有所选科室必须存在、启用且未删除。

保存科室集合采用整体同步：选中关系被创建或重新启用，默认开放预约；不再选中的关系被停用但当前不设置删除标记。该操作与医生档案写入处于同一事务。

### MED-DOC-004 字典字段必须使用启用值

专业职称使用 `MED_DOCTOR_TITLE`，用工类型使用 `MED_EMPLOYMENT_TYPE`，性别使用 `GENDER`（与患者性别字典一致）。必填字典值为空、值不存在或已停用时拒绝保存。

### MED-DOC-005 服务能力字段有明确范围

- 默认接诊时长传 `0` 时按 15 分钟保存，允许范围为 5 至 240 分钟。
- 线上问诊、允许预约、公开展示只能为 `0` 或 `1`。
- 离职日期不能早于入职日期。
- 工作邮箱必须是合法邮箱且不超过 128 字符；电话不超过 20 字符。

### MED-DOC-006 医生启停与科室关系独立

医生状态为 `0=停用、1=启用`。状态接口只修改医生档案，不级联修改医生科室关系；排班创建、生成和发布会要求医生仍为启用状态。

医生档案上的“允许预约”与医生科室关系上的预约开关是两个字段。当前排班维度直接要求对应医生科室关系启用且开放预约；不要把两个开关误当成同一个字段。

### MED-DOC-007 删除为关联软删除

批量删除会在同一事务中停用并软删除医生档案及其医生科室关系。当前实现不先检查排班或挂号引用，也不删除相关排班。

## 页面动作与权限

| 动作 | 权限 |
|---|---|
| 列表 | `medical:doctor:list` |
| 新增 | `medical:doctor:create` |
| 详情 | `medical:doctor:detail` |
| 修改 | `medical:doctor:update` |
| 启停 | `medical:doctor:status` |
| 删除 | `medical:doctor:delete` |

`GET /api/medical/doctors/all` 只返回启用医生，当前作为下拉数据源使用。

## 代码入口

- Model/DTO：`models/medical.go` 的 `MedDoctor`、`MedDoctorDepartment` 和医生请求响应结构。
- Service：`services/medical_doctor_service.go`。
- Controller：`controllers/medical_controller.go`。
- Router：`router/medical.go` 中 `/api/medical/doctors`。
- 前端：`hive/apps/web-antdv-next/src/views/medical/doctor`。

