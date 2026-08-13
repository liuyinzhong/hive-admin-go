# 企业主体业务规则

企业主体统一承载生产企业、上市许可持有人、注册人、备案人、进口代理、供应商、经销商、经营商和客户等业务身份。

## 业务规则

### BASE-ENT-001 一个主体可以承担多个企业角色

企业类型支持 ENTERPRISE、MEDICAL_ORG、INDIVIDUAL、PUBLIC_INSTITUTION 和 OTHER。保存时至少选择一个受支持的企业角色；角色集合采用全量保存语义。

### BASE-ENT-002 企业编码由系统生成

创建企业主体时由基础编号序列生成稳定企业编码，调用方不能自行指定或复用编码。

### BASE-ENT-003 名称和信用代码保持唯一

企业名称按规范化值在未删除主体中唯一；非空统一社会信用代码同样唯一且只允许字母和数字。

### BASE-ENT-004 修改使用乐观并发

更新主体和切换状态必须提交 expectedRowVersion。版本不匹配时拒绝保存并要求刷新，成功后行版本递增。

### BASE-ENT-005 停用不删除历史主体

状态只允许 0 或 1。当前没有企业删除接口；停用主体保留档案和既有业务引用，但不进入新业务的企业选项。

### BASE-ENT-006 业务选项只返回可用主体

企业选项接口只查询启用且未删除主体，可按关键词和单个企业角色过滤。具体采购、产品或其它来源模块仍需再次校验其所需角色。

### BASE-ENT-007 权限边界

列表、创建、详情、更新和状态分别使用 base:enterprise:list、create、detail、update、status。

## 代码入口

- Model/DTO：models/base_enterprise.go。
- Service：services/base_enterprise_service.go、base_code_sequence_service.go。
- Controller：controllers/base_enterprise_controller.go。
- Router：/api/base/enterprises。
- 前端：hive/apps/web-antdv-next/src/views/base/enterprise。
