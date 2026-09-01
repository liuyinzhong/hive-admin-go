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

需求归属项目，可选关联版本和模块，并可关联文件、类型、级别和来源。支持单条与批量创建。创建和编辑不接收参与人，参与人完全由需求流转产生。

需求参与人表 `dev_story_user` 记录历次流转指定的状态负责人：每次流转（`PUT /dev/storys/{storyId}/next`）由操作人指定目标状态负责人（`userId` 须为该项目成员），写入一行（`story_id` + `user_id` + `story_status`=目标状态）；流转到 99（已关闭）时无需指定负责人，不写关联表、不推送通知。同一需求同一状态仅保留最新指定的负责人（先删后插）；"参与人"即历次各状态的负责人集合。

局部字段接口只允许修改 storyType、storyLevel、source。只有 storyStatus=0 的待评审需求可以删除。

需求流转未填写流转说明时，后端自动写入默认内容"流转至「目标状态名称」，请及时跟进"（状态名称取 STORY\_STATUS 字典），保证变更记录始终有正文；填写了说明则以填写内容为准。流转成功后仅向本次指定的状态负责人推送需求管理菜单未读消息。

需求列表与详情响应均返回 `thisUser` 字段：当前需求状态负责人（单个用户对象，参与人中 `story_status` 等于需求当前 `story_status` 的用户，结构与 `userList` 单项一致，无负责人时为 `null`），前端列表"当前负责人"列和详情"当前负责人"项复用 `UserAvatarGroup` 组件渲染头像+姓名。

## 任务

### DEV-ITEM-005 任务连接需求、执行人和工时

任务归属项目，可选关联需求、版本和模块，并维护执行人、任务类型、计划/实际工时与起止时间；结束时间不得早于开始时间。支持单条与批量创建。

局部字段接口只允许修改 userId、taskType、startDate、endDate。只有 taskStatus=0 的待执行任务可以删除。

### DEV-ITEM-006 任务导出进入下载中心

任务导出按照当前列表筛选和排序创建 devTask 来源的异步下载任务，不在请求中同步生成文件。创建权限为 dev:task:export；Worker 执行时重新解析创建用户当前数据范围，任务状态、下载和保留期遵循系统下载中心规则。

## 缺陷

### DEV-ITEM-007 缺陷同时维护处理状态和确认状态

缺陷可关联项目、版本、模块、需求和处理人，并维护级别、环境、来源和类型。支持单条与批量创建。

局部字段接口只允许修改 userId、bugLevel、bugEnv、bugType、bugSource。只有 bugStatus=0 的待确认缺陷可以删除。

### DEV-ITEM-008 缺陷确认是独立动作

确认状态写为 1 时，当前实现同时把缺陷状态改为 10；确认状态写为 2 时，缺陷状态改为 1。状态重新流转到 0 时会重置确认状态。调整这些数值含义必须同步 BUG\_STATUS、BUG\_CONFIRM\_STATUS 字典和前端动作。

### DEV-ITEM-011 缺陷打回次数随流转累计

缺陷维护打回次数 `return_num`，默认 0；流转接口 `/dev/bugs/{bugId}/next` 推进到 bugStatus=10（待修复）时在数据库中原子自增 1。打回次数是系统累计值，不在新建、编辑、局部字段和确认接口中修改；确认动作写状态 10 不计入打回。列表、全量选项和详情响应均返回 `returnNum`（int），前端缺陷列表以「打回次数」列只读展示。

## 数据权限

### DEV-ITEM-009 工作项范围贯穿全部入口

需求按创建人或参与人可见，参与人通过独立关联表 `dev_story_user` 的 EXISTS 子查询过滤；任务按创建人或执行人可见；缺陷按创建人或处理人可见。角色部门范围通过这些归属用户的当前启用部门展开。列表、全量选项、详情、整体编辑、局部编辑、状态动作、确认和删除使用同一规则，批量操作任一记录越界时整批失败。

变更记录继承对应需求、任务、缺陷或版本的访问范围，不按变更记录创建人单独放宽。

### DEV-ITEM-010 写入引用必须处于当前范围

新建或编辑需求时，附件和关联版本必须对操作者可见；新建或编辑任务、缺陷时，执行/处理用户、关联版本和关联需求必须可见。流转接口指定的状态负责人必须是需求所属项目的成员，不能通过直接提交 ID 建立越界引用。

## 权限

* 需求：dev:story:list、create、batchCreate、detail、update、fieldUpdate、advance、delete。

* 任务：dev:task:list、create、batchCreate、detail、update、fieldUpdate、advance、delete、export。

* 缺陷：dev:bug:list、create、batchCreate、detail、update、fieldUpdate、advance、confirm、delete。

* 变更记录：dev:changeHistory:list、create。

## 代码入口

* Model：models/models.go 中 DevStory、DevTask、DevBug、DevChangeHistory。

* Service：services/dev\_story\_service.go、dev\_task\_service.go、dev\_bug\_service.go、dev\_change\_history\_service.go。

* Router：/api/dev/storys、/tasks、/bugs、/changeHistory。storys 是当前既有接口拼写，修改需同步前后端。

* 前端：hive/apps/web-antdv-next/src/views/dev/story、task、bug。

