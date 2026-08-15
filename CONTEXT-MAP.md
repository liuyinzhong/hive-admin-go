# Hive 领域上下文地图

本文件帮助业务人员、开发、测试和 AI 快速定位术语、规则和跨领域关系。进入已有模块时先阅读文档建立业务理解，再阅读代码核实现状；领域 CONTEXT.md 只定义术语，运行规则位于各模块文档。

## 领域入口

总索引见 [后端业务知识库](./business-docs/README.md)，前端页面说明见 [前端业务文档总索引](../hive/business-docs/README.md)。

| 领域 | 词汇 | 业务手册 |
|---|---|---|
| 基础资料 | [CONTEXT](./business-docs/base/CONTEXT.md) | [base](./business-docs/base/README.md) |
| 产品档案 | [CONTEXT](./business-docs/product/CONTEXT.md) | [product](./business-docs/product/README.md) |
| 医疗 | [CONTEXT](./business-docs/medical/CONTEXT.md) | [medical](./business-docs/medical/README.md) |
| ERP | [CONTEXT](./business-docs/erp/CONTEXT.md) | [erp](./business-docs/erp/README.md) |
| 打印 | [CONTEXT](./business-docs/print/CONTEXT.md) | [print](./business-docs/print/README.md) |
| 开发管理 | [CONTEXT](./business-docs/dev/CONTEXT.md) | [dev](./business-docs/dev/README.md) |
| 表单 | [CONTEXT](./business-docs/form/CONTEXT.md) | [form](./business-docs/form/README.md) |
| 工作流 | [CONTEXT](./business-docs/workflow/CONTEXT.md) | [workflow](./business-docs/workflow/README.md) |
| 系统管理 | [CONTEXT](./business-docs/system/CONTEXT.md) | [system](./business-docs/system/README.md) |

根目录 [CONTEXT.md](./CONTEXT.md) 是历史综合材料，不再作为新任务入口。内容尚未机械删除，以便后续逐条核对迁移历史。

## 核心关系

- 基础资料的企业主体被产品生产企业、ERP 供应商等模块引用；分类体系为可复用层级主数据。
- 产品档案向 ERP 和医疗处方提供具体 SKU、包装换算、追溯模式和药品属性。
- 临床科室、医生科室关系和挂号费规则共同约束排班；排班发布固化费用并向挂号提供号源。
- 患者与已发布排班形成挂号；签到创建候诊记录；医生接诊创建门诊病历，完成时同步挂号和候诊终态。
- 疾病诊断档案向门诊病历提供标准诊断；处方提交固化诊断、药品和临床信息快照，审核通过后才可进入未来药房履约。
- 企业主体向采购提供供应商；采购入库和其它出库分别增加或减少 ERP 库存。
- 打印模块从采购入库等来源模块读取数据；模板只改变展示，不改变来源单据或库存。
- 表单 Schema 被工作流定义引用；结构变化使定义退回草稿。工作流实例保存定义、流程和表单快照。
- 开发任务和库存等来源模块创建异步下载任务；下载中心负责生成文件并通过菜单消息和 SSE 提示变化。
- 系统用户授权与医生身份、流程参与者、数据归属等业务校验叠加生效，不能互相替代。
- 角色数据范围把业务记录归属人映射到系统组织部门，并贯穿研发工作项、ERP 单据与库存、患者挂号排班、文件、日志、统计和异步导出；全局主数据及流程/医生领域归属按 [数据权限分类](./business-docs/system/data-permission.md) 处理。

## 跨域修改检查

改变任一关系的字段、状态、权限、事务或副作用时，必须同时更新关系两端的业务正文、前端 UI 文档、本地图和源码入口索引。若仅一侧实现发生变化，交付前必须说明另一侧现状和兼容影响。
