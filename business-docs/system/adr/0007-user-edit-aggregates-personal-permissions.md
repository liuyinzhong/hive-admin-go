# 用户编辑抽屉聚合个人权限（删除独立 personalPermissions 路由）

个人权限此前通过独立路由 `GET/PUT /system/users/{userId}/personalPermissions` 维护，前端是列表操作列的独立抽屉入口；编辑一个用户的基础信息与个人权限要开两个入口、走两套请求。决定把个人权限聚合进用户编辑抽屉对应的接口：`GET /system/users/{userId}` 在用户基础信息之上同时返回个人额外授权与个人禁止的菜单ID集合；`POST /system/users` 与 `PUT /system/users/{userId}` 都接受可选的 `grantMenuIds`/`denyMenuIds` 字段（nil 表示不修改，非 nil 按完整替换语义处理），新建与编辑共用同一抽屉表单、同等支持个人权限。可勾选的启用菜单树由独立查询 `GET /system/users/personalPermissionMenuTree` 提供，作为新建与编辑的唯一树来源（新建尚无 userId，无法从详情接口取树；拆出后详情响应也不必携带全量菜单树）。原 personalPermissions 两条路由删除，不留兼容层。

考虑过的备选：其一，前端并行调用两个既有接口（后端零改动），但"一个编辑动作、两个接口拼装"的契约分裂会保留在 API 层；其二，菜单树并入用户详情响应，但新建抽屉同样需要树，会出现详情内嵌与独立接口两个树来源，故拒绝。聚合后需要处理权限码边界：编辑/创建路由按 SYS-ACL-031 只能硬编码挂 `system:user:update` / `system:user:create`，因此请求体携带个人权限字段时必须在服务层校验操作者持有 `system:user:personalPermission`；防提升校验（授予与禁止都须在操作者生效权限内，全部数据范围操作者豁免）、系统内置用户禁配个人权限、两集合不得交叉等约束原样保留。个人权限维护入口并入用户编辑抽屉后，`system:user:personalPermission` 权限码语义不变：无该权限码的操作者在新建与编辑抽屉中均不展示也不提交个人权限字段。

已知代价（有意识接受）：编辑抽屉打开时需两次查询（详情 + 菜单树）；用户列表另行为每行返回权限摘要计数（角色授权节点数求和不去重、额外授权数、禁止数），沿用列表既有的逐行关联查询模式。
