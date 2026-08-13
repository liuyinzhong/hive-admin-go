# 打印业务手册

本目录记录打印模板设计、发布、预览以及业务单据打印数据的当前规则。

## 阅读顺序

1. 先读 [打印领域词汇](./CONTEXT.md)。
2. 模板设计和发布读 [打印模板规则](./template.md)。
3. 真实单据打印读 [打印文档规则](./document.md)，并继续阅读来源领域文档。
4. 涉及页面时读前端 [打印 UI 手册](../../../hive/business-docs/print/README.md)。

## 模块覆盖

| 模块 | 规则正文 | 后端入口 |
|---|---|---|
| 模板、元数据、预览、发布 | [template.md](./template.md) | print_template_controller.go、print_template_service.go |
| 单据数据和正式打印 | [document.md](./document.md) | print_document_controller.go、print_document_service.go、print_document_registry.go |

## 规则编号

- PRINT-TPL-*：模板设计、校验和发布。
- PRINT-DOC-*：数据来源和正式打印。
