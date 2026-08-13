# 分类体系业务规则

分类体系提供独立命名空间下的层级分类树，供产品或其它业务模块通过稳定体系编码选择分类节点。

## 业务规则

### BASE-CLS-001 体系编码全局唯一

体系编码和名称必填；体系编码在未删除分类体系中全局唯一。修改分类体系使用 expectedRowVersion 做乐观并发控制。

### BASE-CLS-002 有节点的体系不能删除

删除分类体系前必须确认其没有未删除节点。删除采用逻辑删除并递增版本，不级联删除分类节点。

### BASE-CLS-003 节点编码在所属体系内唯一

节点必须归属于一个有效分类体系，节点编码和名称必填；同一体系内节点编码唯一，不同体系可以使用相同节点编码。

### BASE-CLS-004 分类树不能跨体系或形成环

父节点必须存在且属于同一体系。节点不能选择自身、不能跨体系移动，也不能移动到自己的任一后代下面。

### BASE-CLS-005 节点启停和修改使用行版本

节点状态只允许 0 或 1。编辑节点或单独切换状态都必须提交 expectedRowVersion，成功后版本递增。

### BASE-CLS-006 有子节点的节点不能删除

删除前只要存在未删除直接子节点就拒绝操作；删除会停用并逻辑删除当前节点，不级联处理后代。

### BASE-CLS-007 管理树和业务选项职责不同

管理树可以查看停用节点，关键词或状态筛选时保留匹配节点的祖先链。业务选项按体系编码仅返回启用且未删除节点树。

### BASE-CLS-008 权限边界

分类体系使用 base:classificationSystem:list、create、detail、update、delete；分类节点使用 base:classificationNode:list、create、detail、update、status、delete。选项接口用于业务选择，不替代来源模块权限和有效性校验。

## 代码入口

- Model/DTO：models/classification.go。
- Service：services/classification_system_service.go、classification_node_service.go。
- Controller：controllers/classification_controller.go。
- Router：/api/base/classificationSystems。
- 前端：hive/apps/web-antdv-next/src/views/base/classification。
