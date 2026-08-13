# 表单业务手册

本目录记录可复用表单 Schema 的设计、维护、提交校验以及与工作流的关系。

## 阅读顺序

1. 先读 [表单领域词汇](./CONTEXT.md)。
2. 再读 [表单 Schema 规则](./form-schema.md)。
3. 涉及流程时继续读 [工作流业务手册](../workflow/README.md)。
4. 涉及页面时读前端 [表单 UI 手册](../../../hive/business-docs/form/README.md)。
5. 最后核对 form_schema Controller、Service、Model 和前端 API。

## 模块覆盖

| 模块 | 规则正文 | 后端入口 |
|---|---|---|
| Schema 列表、设计、预览、启停、删除和提交校验 | [form-schema.md](./form-schema.md) | form_schema_controller.go、form_schema_service.go、models/form_schema.go |

## 规则编号

- FORM-SCHEMA-*：Schema、字段结构和提交校验。
- FORM-LINK-*：与工作流定义的联动。

已有编号不得复用为其它含义。
