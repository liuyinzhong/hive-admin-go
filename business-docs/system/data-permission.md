# 数据权限规则与接口分类

数据权限是认证和原子接口权限之后的第三层访问控制，用于限制一个已获准调用接口的用户能够读取、修改、导出或统计哪些业务记录。它不替代菜单/按钮权限，也不替代医生、流程参与者等领域归属校验。

## 生效顺序

```text
Bearer Token 认证
  -> 路由原子权限
  -> 解析当前角色数据范围
  -> Service 查询/写入/统计/导出或领域归属校验
```

- **SYS-DATA-001** 认证失败返回未认证；缺少原子权限返回禁止访问；通过两层门禁后，Service 仍必须执行本页定义的数据边界。
- **SYS-DATA-002** 数据权限解析在每个受认证业务请求进入分组时执行一次。用户不存在、已停用或已删除时请求终止；解析数据库失败时关闭访问，不以全量权限兜底。
- **SYS-DATA-003** 当前用户没有有效角色、角色范围未知或所有角色均为 `none` 时，角色数据范围结果为空；不能因配置缺失自动放宽。

## 角色数据范围

| 值 | 中文含义 | 贡献的可见范围 |
|---|---|---|
| `all` | 全部数据 | 所有记录 |
| `customDepartment` | 自定义部门 | 角色配置的启用部门中的归属用户记录 |
| `department` | 本部门 | 当前用户所属的启用部门中的归属用户记录 |
| `departmentAndChildren` | 本部门及下级 | 当前用户所属启用部门及全部启用后代部门中的归属用户记录 |
| `self` | 仅本人 | 当前用户作为记录归属人或业务参与人的记录 |
| `none` | 无数据 | 不贡献可见记录 |

- **SYS-DATA-010** 一个用户拥有多个有效角色时取范围并集；任一角色为 `all` 时结果为全部数据。`none` 只是不贡献范围，不抵消其它角色。
- **SYS-DATA-011** `is_sys=1` 的启用系统内置用户直接取得全部数据范围。普通用户不能通过用户名、角色名称或前端标记获得该特例。
- **SYS-DATA-012** 部门范围使用当前有效的 `sys_user_dept`、`sys_dept`、`sys_user_role`、`sys_role` 和 `sys_role_dept` 关系实时计算，不把部门快照复制到每张业务表。用户调岗、角色停用或自定义部门调整会影响后续请求。
- **SYS-DATA-013** 业务记录以 `creator_id`、`user_id`、`operator_id` 等明确归属用户字段建立范围；多归属字段取 OR。需求参与人沿用现有逗号分隔字段并参与可见性判断。
- **SYS-DATA-014** 归属字段为空且无法可靠推断的历史记录只对 `all` 可见。禁止把空归属解释为公共数据。
- **SYS-DATA-015** `none` 只控制记录范围，不撤销已经授予的创建动作权限。若角色拥有创建权限但范围为 `none`，其新建记录仍按当前用户落归属，但该用户之后不能仅凭 `none` 查询该记录；角色配置应同时核对动作权限和数据范围。

## 查询、写入、统计和导出

- **SYS-DATA-020** 同一资源的分页列表、全量选项、详情、批量读取、状态动作、修改和删除必须使用同一范围语义，不能只过滤列表而留下按 ID 越权入口。
- **SYS-DATA-021** 批量修改或删除要求提交的每个有效 ID 都存在且可操作；发现任一越界或不存在记录时整批失败，不对可见子集静默成功。
- **SYS-DATA-022** 新建记录统一写入当前用户归属。写入引用的用户、附件、版本、需求、采购单、库存余额、患者或排班等资源在本页矩阵中被列为角色数据范围时，必须先验证引用也在当前范围内；全局主数据和领域归属资源继续按各自规则校验。
- **SYS-DATA-023** 用户管理的部门写入必须完全落在操作者部门范围内。非全部范围操作者只能分配自己持有的有效角色，且被分配角色的部门展开结果不能大于操作者范围。
- **SYS-DATA-024** 一个用户同时属于多个启用部门时，只与其中一个部门相交即可出现在查询结果；但非全部范围操作者只有在该用户的全部启用部门都可管理时才能修改、启停或删除，避免跨部门关联被局部管理员覆盖。
- **SYS-DATA-025** 统计必须先应用与来源列表相同的范围再聚合。页面不得把全局统计与受限列表混用。
- **SYS-DATA-026** 异步导出以下载任务记录的 `creator_id` 为权限主体，不信任请求负载中的用户字段，也不保存一份长期有效的权限快照。Worker 在计数和生成文件时重新解析该用户当时的有效角色和部门；用户失效或权限收紧时任务失败或只导出收紧后的结果。
- **SYS-DATA-027** 下载任务、菜单消息、流程待办等天然个人资源继续按创建者、接收者或参与者控制，不套用角色部门范围扩大为“同部门可见”。
- **SYS-DATA-028** 一条已获准查看的业务记录可以携带其页面展示必需的关联编号、名称和附件信息，例如任务中的需求标题、入库单中的采购单号；这只是父子响应投影，不授予关联资源的独立列表、详情或写入权限。

## 全局主数据与领域归属

- **SYS-DATA-030** 全局主数据是全系统共同引用的一套配置或档案，不按创建人切分；是否可读写由原子接口权限和业务状态校验决定。拥有 `none` 数据范围但拥有相应原子权限的用户仍可操作全局主数据。
- **SYS-DATA-031** 领域归属资源使用比通用角色范围更严格或更贴合业务的参与者规则，例如流程发起人/办理人/抄送人、当前排班医生、处方开具人和审核人。角色的 `all` 不自动绕过这些业务身份校验。
- **SYS-DATA-032** 打印文档、变更记录、候诊队列等从属资源继承来源业务对象的范围，不另以自身创建人形成不一致边界。

## 全路由分类矩阵

下表覆盖 `router/` 目录当前注册的全部 HTTP 资源。`/**` 表示该资源下已注册的列表、选项、详情和动作路由均按同一分类处理；动作仍需各自原子权限。

| 路由范围 | 分类 | 范围或原因 |
|---|---|---|
| `/uploads/**` | 公开静态资源 | Gin 静态目录，不经过认证、接口权限或数据权限；不得用于需要私有下载授权的文件 |
| `/api/public/externalPages/:name` | 公开接口 | 按已启用外部页面名称公开解析 |
| `/api/public/downloads/preview/:token` | 公开接口 | 通过 JWT 签名校验确保 token 由本服务签发且未过期，再复用创建者和文件可用性校验；kkFileView 等外部预览服务无登录态时通过此接口取文件 |
| `POST /api/auth/login` | 公开接口 | 建立登录身份 |
| `/api/auth/profile`（GET 读取、PUT 更新头像和邮箱）、`menus`、`codes`、`logout` | 当前用户上下文 | 只处理当前 Token 对应用户，不使用角色行范围 |
| `/api/system/downloads/**` | 当前用户归属 | 任务列表、文件下载和预览链接生成只允许创建者 |
| `/api/system/messages/unreadSummary`、`stream`、`read` | 当前用户归属 | 只查询、订阅或更新当前接收者 |
| `POST /api/system/messages/demo` | 角色数据范围 | 除接口权限外，每个目标用户还必须在当前用户数据范围内；整批越界则失败 |
| `POST /api/system/upload` | 创建归属 | 元数据创建人为当前用户；实际 `/uploads/**` 访问仍是公开边界 |
| `GET /api/system/files` | 角色数据范围 | 按文件 `creator_id` |
| `/api/system/users/**` | 角色数据范围 | 按用户当前有效部门；本人范围可读取本人，管理写入另受 SYS-DATA-023/024 约束 |
| `/api/system/operationLogs/**`、`loginLogs/**` | 角色数据范围 | 按日志 `user_id`；未认证或空用户日志仅 `all` 可见 |
| `/api/system/roles/**` | 全局授权配置 | 列表/详情用于授权配置；角色创建、更新、启停、删除额外要求操作者为 `all` |
| `/api/system/menus/**`、`externalPages/**`、`depts/**`、`dicts/**`、`params/**`、`payChannels/**` | 全局主数据 | 全系统共享配置，由各自原子权限和业务校验控制 |
| `/api/dev/projects/**`、`modules/**` | 全局主数据 | 研发公共规划维度，不按创建人拆分 |
| `/api/dev/versions/**` | 角色数据范围 | 按版本 `creator_id`，包括 latest、all、详情和写动作 |
| `/api/dev/storys/**` | 角色数据范围 | 创建人或参与人；参与人使用现有 `user_ids` |
| `/api/dev/tasks/**` | 角色数据范围 | 创建人或执行人；异步导出重新解析创建者当前范围 |
| `/api/dev/bugs/**` | 角色数据范围 | 创建人或处理人 |
| `/api/dev/changeHistory` | 来源继承 | 按需求、任务、缺陷或版本父对象校验 |
| `/api/statistics/dev/**` | 角色数据范围 | 各指标先应用需求、任务、缺陷或版本同源范围再聚合 |
| `/api/form/schemas/**` | 全局主数据 | 工作流共同引用的结构定义 |
| `/api/base/institution/**`、`enterprises/**`、`classificationSystems/**` | 全局主数据 | 机构、企业主体和分类体系是共享基础资料 |
| `/api/erp/warehouses/**`（含库区、货位） | 全局主数据 | 仓储结构由全系统业务单据共同引用 |
| `/api/erp/inventory/balances/**` | 角色数据范围 | 按库存余额 `creator_id`；导出使用同一范围 |
| `/api/erp/inventory/balances/:id/movements`、`movements` | 角色数据范围 | 按库存流水 `operator_id`；余额入口还先校验余额范围 |
| `/api/erp/inventory/traceCodes/**` | 角色数据范围 | 追溯码按 `creator_id`，单码流水同时受流水操作者范围 |
| `POST /api/erp/inventory/initialStocks` | 创建归属/角色数据范围 | 新余额、流水和追溯码归属当前操作者；累加已有余额前校验其范围；仓库和 SKU 是全局主数据 |
| `/api/erp/purchaseOrders/**` | 角色数据范围 | 按采购单 `creator_id`，详情、修改、状态动作和日志一致 |
| `/api/erp/purchaseInbounds/**` | 角色数据范围 | 按入库单 `creator_id`；创建前必须能访问来源采购单及命中的已有库存余额 |
| `/api/erp/otherOutbounds/**` | 角色数据范围 | 按出库单 `creator_id`；创建前必须能访问每个库存余额 |
| `/api/printTemplates/**` | 全局主数据 | 全系统共享模板、元数据和发布状态 |
| `/api/printDocuments/purchaseInbound/**` | 来源继承 | 继承采购入库单数据范围，预览和正式打印一致 |
| `/api/product/spus/**`、`rps/**`、`mps/**`、`skus/**`（含价格、阶梯价） | 全局主数据 | 产品档案和价格是共享业务主数据 |
| `/api/medical/departments/**`、`doctors/**`、`diagnoses/**`、`registrationFeeRules/**` | 全局主数据 | 医疗基础档案和计费配置 |
| `/api/medical/patients/**` | 角色数据范围 | 按患者档案 `creator_id`；敏感字段权限另行叠加 |
| `/api/medical/registrations/**` | 角色数据范围 | 按挂号单 `creator_id`；创建时患者和实际排班都必须可见 |
| `/api/medical/scheduleTemplates/**`、`scheduleTasks/**` | 全局主数据/运维记录 | 周期模板和自动任务由排班管理权限控制，不按创建人过滤 |
| `/api/medical/schedules/**` | 角色数据范围 | 按实际排班 `creator_id`；生成、发布、编辑、停诊、结束和删除遵循同一范围 |
| `/api/medical/schedules/:id/visitQueues` | 来源继承 | 先校验实际排班范围，队列患者信息始终脱敏 |
| `/api/medical/doctorWorkbench/**`、`outpatientRecords/**`、`prescriptions/**` | 领域归属 | 当前用户必须绑定当前排班医生；病历、处方和患者历史按接诊关系控制 |
| `/api/medical/prescriptionReviews/**` | 领域归属 | 审核权限、处方状态及“不得审核本人处方”共同控制 |
| `/api/workflow/definitions/**` | 全局主数据 | 流程定义和表单绑定是共享配置 |
| `/api/workflow/instances/**` | 领域归属 | 发起人或流程参与者；撤销只允许发起人 |
| `/api/workflow/tasks/**` | 领域归属 | 当前办理人及审批组规则 |
| `/api/workflow/copies/**` | 当前用户归属 | 当前抄送接收者，阅读只更新本人记录 |

## 文件公开边界

- **SYS-DATA-040** `sys_file` 元数据列表受角色数据范围限制，但 `/uploads/**` 当前仍是公开静态路径；知道 URL 的访问者不需要登录。
- **SYS-DATA-041** 需求附件等现有业务可继续引用公开上传文件，但身份证件、处方原件、合同原件或其它需要逐次授权的敏感文件不得继续放入该目录。此类需求必须先建设私有文件下载 Controller、归属校验和不可猜测/短期访问方案。
- **SYS-DATA-042** 下载中心使用私有文件接口并校验创建者，不受 `/uploads/**` 公开边界影响。

## 迁移、兼容和回滚

2. 为保持升级兼容，迁移把存量未删除角色设为 `all`；新建角色未传范围时默认 `self`。发布后应由管理员逐角色收紧并验证页面结果。
3. 03 脚本只回填可以从自动生成批次和模板可靠推断的历史排班。其它空归属历史数据仍仅 `all` 可见。
4. 回滚应用前先恢复不读取 `data_scope` 的旧后端，再按 `03 → 02 → 01` 执行 down 脚本。03 的 down 会把符合条件的新旧自动排班归属一并清空，执行前必须由实施人员自行备份相关记录。
5. 本功能不自动执行迁移、不创建数据库备份对象，也不自动修改角色配置。
6. 发布时无需清理遗留等待导出任务：Worker 兼容旧版裸筛选 JSON 和新版请求封装，并始终以任务表创建者重新解析权限。

## 代码入口

- 范围模型与 SQL 过滤：`datapermission/permission.go`、`datapermission/gorm_store.go`。
- 请求解析：`middleware/data_permission.go`、`controllers/data_permission.go`。
- 业务校验：各领域 Service，以及 `services/data_permission_validation.go`。
- 异步重新解析：`services/data_permission_service.go`、两个下载导出器。
- 角色配置：`models/models.go`、`models/response.go`、`services/role_service.go`。
- 设计决定：[ADR-0001](./adr/0001-role-department-data-permission.md)。
