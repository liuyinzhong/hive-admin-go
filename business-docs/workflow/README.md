# 工作流业务手册

本目录记录流程定义设计、发布和运行时审批的当前规则。

## 阅读顺序

1. 先读 [工作流领域词汇](./CONTEXT.md)。
2. 设计、表单绑定和发布读 [流程定义规则](./definition.md)。
3. 发起、待办、抄送和审批动作读 [流程运行规则](./runtime.md)。
4. 表单结构继续读 [表单业务手册](../form/README.md)。
5. 权限分层继续读 [系统数据权限规则](../system/data-permission.md)。
6. 涉及页面时读前端 [工作流 UI 手册](../../../hive/business-docs/workflow/README.md)。

## 模块覆盖

| 模块 | 规则正文 | 后端入口 |
|---|---|---|
| 定义、画布、表单绑定、发布 | [definition.md](./definition.md) | workflow_definition_controller.go、workflow_definition_service.go |
| 实例、待办、抄送、审批操作 | [runtime.md](./runtime.md) | workflow_runtime_controller.go、workflow_runtime_service.go |

## 规则编号

- WF-DEF-*：流程定义、画布和发布。
- WF-RUN-*：实例、任务、抄送和运行操作。
