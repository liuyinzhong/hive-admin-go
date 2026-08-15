# Hive 后端业务知识库

本目录是 Hive 业务人员、开发、测试和 AI 共用的后端业务规则库。进入已有业务模块时，先用这里的文档理解目标、术语、状态、边界和运行逻辑，再阅读代码核对当前实现；不要从逐文件钻研代码开始反推业务。

## 固定阅读顺序

1. 在本页按源码入口定位领域。
2. 阅读仓库根 [CONTEXT-MAP.md](../CONTEXT-MAP.md)，确认跨领域依赖。
3. 阅读领域 CONTEXT.md，只建立统一词汇。
4. 阅读领域 README.md 和当前模块规则正文。
5. 涉及页面时阅读前端 hive/business-docs 中对应 UI 文档。
6. 最后阅读 Router、Controller、Service、Model/DTO 与前端 API，验证文档和代码是否一致。

文档与代码不一致时，先记录差异、实际行为和影响；不得静默选一方覆盖另一方。

## 业务域总览

| 业务域 | 当前覆盖的后端能力 | 入口 |
|---|---|---|
| 基础资料 | 本机构、企业主体、分类体系与分类节点 | [base](./base/README.md) |
| 产品档案 | SPU、RP、MP、SKU、SKU 价格与阶梯价格 | [product](./product/README.md) |
| 医疗 | 科室、医生、患者、诊断、挂号、挂号费、排班、候诊、接诊、处方与审核 | [medical](./medical/README.md) |
| ERP | 仓库、库区货位、采购、入库、库存、追溯码与其它出库 | [erp](./erp/README.md) |
| 打印 | 打印模板、字段注册表、采购入库单打印数据与打印文档 | [print](./print/README.md) |
| 开发管理 | 项目、模块、版本、需求、任务、缺陷、变更记录与统计 | [dev](./dev/README.md) |
| 表单 | 可复用表单 Schema、设计约束和服务端提交校验 | [form](./form/README.md) |
| 工作流 | 流程定义、画布、表单绑定、实例、待办、抄送和审批操作 | [workflow](./workflow/README.md) |
| 系统管理 | 登录授权、角色数据范围、用户角色部门菜单、字典参数、文件、日志、外部页面、支付渠道、消息与下载 | [system](./system/README.md) |

## 后端源码覆盖表

| 源码或路由范围 | 文档归属 |
|---|---|
| controllers/base_*、classification_controller.go；services/base_*、classification_* | [基础资料](./base/README.md) |
| controllers/product_*；services/product_*；models/product_* | [产品档案](./product/README.md) |
| controllers/medical_*；services/medical_*；models/medical* | [医疗](./medical/README.md) |
| controllers/erp_*；services/erp_*；models/erp_* | [ERP](./erp/README.md) |
| controllers/print_*；services/print_*；models/print.go | [打印](./print/README.md) |
| controllers/dev_*；services/dev_*；models/models.go 中 Dev*；statistics/dev | [开发管理](./dev/README.md) |
| form_schema_controller.go、form_schema_service.go、models/form_schema.go | [表单](./form/README.md) |
| workflow_*_controller.go、workflow_*_service.go、models/workflow* | [工作流](./workflow/README.md) |
| auth、system、user、role、dept、menu、permission、datapermission、dict、param、file、audit、external_page、pay_channel、menu_message、download_task | [系统管理](./system/README.md) |
| router/router.go | 按路由所属资源阅读上表对应领域；跨域关系再读 CONTEXT-MAP.md |

## 不属于业务规则正文的目录

- docs 是 Swagger 自动生成目录，不手工维护业务规则。
- database、middleware、utils 和统一 response 属于基础设施；若它们改变用户可见业务语义，文档仍回写到受影响领域。
- 测试文件用于验证实现，不替代业务规则。
- 根 CONTEXT.md 是历史综合材料，尚未机械删除；新任务以本目录和 [CONTEXT-MAP.md](../CONTEXT-MAP.md) 为入口。

## 覆盖维护

新增 Controller、Service、Model、受保护路由或业务状态时，必须在同一次修改中：

1. 把源码入口登记到本页或现有领域 README。
2. 补充领域词汇、模块规则和跨领域关系。
3. 同步前端 UI 文档与就近 AGENTS.md。
4. 核对规则编号唯一、相对链接有效、文件为 UTF-8，且业务文档没有进入 docs。
