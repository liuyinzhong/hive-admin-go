# 本机构业务规则

本机构页面维护当前系统唯一医疗机构的基础、展示、资质、联系、地址、银行和开票资料。

## 业务规则

### BASE-INS-001 系统只保存一个本机构聚合

本机构使用固定单例键识别。首次保存时创建，后续保存更新同一聚合；未初始化时详情查询可以返回空数据。本模块没有机构列表、启停或删除动作。

### BASE-INS-002 保存按完整聚合处理

机构主资料、资质、联系人、地址和银行账户在同一事务中保存。四类子资料都必须传完整数组，空数组表示清空该类资料；保存会整体替换子资料，不保留子行历史。

### BASE-INS-003 机构类型和属性使用固定枚举

当前机构类型只支持 HOSPITAL。机构性质、医院类别和医院等级必须使用模型中已支持的枚举，不接收任意文本。

### BASE-INS-004 身份和日期必须有效

机构名称必填；统一社会信用代码只允许字母和数字。资质名称和证书编号必填，资质有效期不能早于发证日期。

### BASE-INS-005 主要资料保持唯一

联系人最多设置一个主要联系人，地址最多设置一个主要地址，银行账户最多设置一个默认账户。银行账户名称、开户行和账号必填。

### BASE-INS-006 权限边界

查看使用 base:institution:detail，保存使用 base:institution:update。页面隐藏保存按钮不能替代后端权限校验。

## 代码入口

- Model/DTO：models/base_institution.go。
- Service：services/base_institution_service.go。
- Controller：controllers/base_institution_controller.go。
- Router：/api/base/institution。
- 前端：hive/apps/web-antdv-next/src/views/base/institution。
