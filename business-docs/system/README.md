# 系统管理业务手册

本目录记录 Hive 系统级业务能力。系统模块提供登录授权和共享基础能力，不替代医疗、ERP 等领域自身的业务规则。

## 阅读顺序

1. 先读 [系统管理领域词汇](./CONTEXT.md)。
2. 按当前任务选择下表中的规则正文。
3. 消息或下载任务还要阅读其来源模块文档。
4. 涉及页面时读前端 [系统管理 UI 手册](../../../hive/business-docs/system/README.md)。
5. 最后核对 Router、Controller、Service、Model 和前端 API。

## 模块覆盖

| 模块 | 规则正文 | 主要后端入口 |
|---|---|---|
| 登录、用户、角色、部门、菜单和接口权限 | [access-control.md](./access-control.md) | auth、user、role、dept、menu、permission |
| 字典和系统参数 | [dictionary-parameter.md](./dictionary-parameter.md) | dict、param |
| 上传与文件列表 | [file-management.md](./file-management.md) | file_controller.go、file_service.go |
| 操作日志和登录日志 | [audit-log.md](./audit-log.md) | operation_log、login_log、audit middleware |
| 外部页面 | [external-page.md](./external-page.md) | external_page_controller.go、external_page_service.go |
| 支付渠道 | [payment-channel.md](./payment-channel.md) | pay_channel_controller.go、pay_channel_service.go |
| 菜单消息与 SSE | [message-push.md](./message-push.md) | menu_message、system_menu_message |
| 异步导出与文件下载 | [download-center.md](./download-center.md) | download_task、download exporter |

## 规则编号

- SYS-ACL-*：登录、用户、角色、部门、菜单和权限。
- SYS-CFG-*：字典和参数。
- SYS-FILE-*：上传文件。
- SYS-AUD-*：审计日志。
- SYS-EXT-*：外部页面。
- SYS-PAY-*：支付渠道。
- SYS-MSG-*：菜单消息和实时事件。
- SYS-DL-*：下载任务和导出文件。

已有编号不得分配给新的含义。

## 跨模块关系

- 系统用户可以与医生档案关联，但用户授权与医生业务身份是两层校验。
- 字典只提供可配置选项，不能代替业务文档中定义的状态流转。
- 来源模块创建下载任务；下载中心只负责异步生成、交付和清理。
- 外部页面公开解析和后台配置是不同访问边界。
- 支付渠道只维护接入配置，当前不管理支付订单或退款业务。

