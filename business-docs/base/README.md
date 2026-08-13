# 基础资料业务手册

本目录记录本机构、企业主体和分类体系的业务规则。

## 阅读顺序

1. 先读 [基础资料领域词汇](./CONTEXT.md)。
2. 本机构资料读 [本机构规则](./institution.md)。
3. 生产企业、供应商、客户等主体读 [企业主体规则](./enterprise.md)。
4. 通用层级分类读 [分类体系规则](./classification.md)。
5. 涉及页面时继续阅读前端 hive/business-docs/base。

## 模块

| 模块 | 规则正文 | 主要源码 |
|---|---|---|
| 本机构 | [institution.md](./institution.md) | base_institution_* |
| 企业主体 | [enterprise.md](./enterprise.md) | base_enterprise_*、base_code_sequence_service.go |
| 分类体系 | [classification.md](./classification.md) | classification_* |

## 规则编号

- BASE-INS-*：本机构聚合。
- BASE-ENT-*：企业主体和企业角色。
- BASE-CLS-*：分类体系和分类节点。

已有编号不得复用为其它含义。
