# 菜单消息推送业务规则

当前消息能力以“某个用户在某个叶子页面菜单下有多少条未读消息”为核心。系统持久化消息记录，通过 SSE 推送完整未读汇总或瞬时任务变化事件。

## 业务规则

### SYS-MSG-001 菜单消息不是通用通知中心

菜单消息归属于一个用户和一个菜单，当前页面只消费未读数量，不提供逐条消息列表。消息标题和内容会持久化，但尚未展示在右上角通知组件或独立消息列表中。

### SYS-MSG-002 未读数量按当前用户和菜单聚合

未读汇总只统计当前登录用户 `readAt` 为空的记录，并按菜单 ID 聚合。菜单路径优先使用菜单外链地址，否则使用菜单路径；没有未读消息时返回空数组。

### SYS-MSG-003 选择叶子菜单会整组标记已读

当前交互不是逐条已读。用户选择一个带未读角标的叶子菜单时，会把该用户在该菜单下的全部未读记录一次性写入相同已读时间，并重新推送完整汇总。

后端已读接口只使用登录身份限定用户，不允许调用方指定其他用户。

### SYS-MSG-004 菜单角标向上汇总

叶子菜单显示自身未读数量，目录菜单显示所有子孙菜单未读数量之和。数量超过 99 时展示 `99+`；归零后恢复菜单原有角标配置，而不是永久清空菜单元数据。

### SYS-MSG-005 SSE 先发送全量状态再发送变化

建立连接后，服务端立即发送一次 `unreadSummary`，之后发送未读汇总变化和 `downloadTaskChanged` 事件，并每 30 秒发送心跳。前端连接失败后先重新请求完整未读汇总，等待 2 秒再连接。

实时事件只是刷新提示，不是权威数据：菜单角标使用完整汇总覆盖本地状态；下载中心收到任务变化后重新查询列表。

### SYS-MSG-006 推送只在当前后端进程内广播

订阅关系保存在进程内内存中，每个用户连接使用有限缓冲并在拥堵时保留较新的事件。当前没有 Redis、消息队列或数据库事件总线，多实例之间不会相互转发 SSE 事件。

### SYS-MSG-007 Demo 消息受目标限制

Demo 接口每次可为一个或多个目标用户各创建 1 至 1000 条固定标题和内容的测试消息。目标菜单必须启用、未删除、属于页面类型且没有非按钮子菜单；目标用户必须启用、未删除且不能是系统用户。重复用户 ID 会去重，整批创建在一个事务中完成。

### SYS-MSG-008 下载终态生成持久化菜单消息

下载任务成功、失败或因服务重启中断时，系统按菜单名称 `downloadCenter` 创建一条归属于任务创建者的菜单消息。创建消息失败不会回滚已经完成的下载任务状态。

### SYS-MSG-009 菜单消息当前没有生命周期清理

当前没有消息删除、过期或定期归档逻辑；已读记录仍保留在 `sys_menu_message`。增加清理规则前必须明确审计需要、保留期和角标统计影响。

## 接口与权限

| 接口 | 用途 | 权限边界 |
|---|---|---|
| `GET /api/system/messages/unreadSummary` | 当前用户未读汇总 | 登录即可，只返回本人 |
| `GET /api/system/messages/stream` | 当前用户 SSE | 登录即可，只订阅本人 |
| `POST /api/system/messages/read` | 当前用户某菜单全部已读 | 登录即可，只更新本人 |
| `POST /api/system/messages/demo` | 批量创建测试消息 | `system:messageDemo:create` |

`CreateMenuMessageForMenuName` 是后端内部能力，不是外部 HTTP 接口。

## 代码入口

- Model/DTO：`models/menu_message.go`。
- Service：`services/menu_message_service.go`。
- Controller：`controllers/system_menu_message_controller.go`。
- Router：`router/router.go` 中 `/api/system/messages`。
- 前端 Store：`hive/apps/web-antdv-next/src/store/menu-message.ts`。
- 前端 Demo：`hive/apps/web-antdv-next/src/views/system/message/demo.vue`。

