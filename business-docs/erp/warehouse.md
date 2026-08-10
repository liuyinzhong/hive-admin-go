# 仓库业务规则

本文记录仓库、库区和货位的当前业务边界。库存目前只核算到仓库，不核算到库区或货位。

## 结构

```text
仓库
└── 库区
    └── 货位
```

当前只支持固定两级内部结构，不支持仓库父子层级、多级库区或货位直接归属仓库。

## 业务规则

### ERP-WH-001 仓库编码由系统生成

仓库编码创建时由公共编码流水生成，用户不能录入或修改；数据库主键继续使用 UUID。

### ERP-WH-002 仓库名称全局唯一

未删除仓库的规范化名称不能重复。库区名称只要求在同一仓库内唯一，货位名称只要求在同一库区内唯一。

### ERP-WH-003 仓库状态控制新增业务选择

仓库状态为启用或停用。选项接口只返回启用且未删除的仓库；期初库存、采购单确认、采购入库和其它出库均要求目标仓库处于启用状态。停用不改变历史库存与历史单据。

### ERP-WH-004 储存类型和业务范围只是基础标签

储存类型为 `NORMAL`、`REFRIGERATED`、`FROZEN`、`COOL`、`HAZARDOUS`；业务范围为 `DRUG`、`CONSUMABLE`、`DEVICE`、`COMPREHENSIVE`。当前不据此执行温控、产品准入或出入库限制。

### ERP-WH-005 库区和货位没有启停状态

库区维护名称、类型和备注；货位维护名称和备注。两者当前都没有启停、容量、占用或库存数量属性。

### ERP-WH-006 修改和删除使用行版本

仓库、库区和货位的修改与删除通过 `expectedRowVersion` 进行乐观并发校验，冲突时要求刷新后重试。

### ERP-WH-007 当前删除为软删除且不检查业务引用

仓库、库区和货位删除只修改删除标记。当前不会检查库存余额、库存批次、采购单、入库单或出库单引用，也不会级联删除下级资料。

这是当前实现的高风险边界；引入严格数据治理前，不应把该行为扩展为物理删除或级联删除。

## 查询与入口

- 仓库列表支持关键字、储存类型、业务范围、状态和排序。
- 仓库关键字匹配仓库编码和名称。
- 库区、货位由仓库列表逐级进入，不建立独立顶级菜单。
- 仓库列表直接返回库区数量；库区列表直接返回货位数量，前端不逐行请求计数。
- 仓库、库区和货位选项接口需要登录，但当前不注册独立按钮权限。

## 权限

- 仓库：`erp:warehouse:list`、`erp:warehouse:detail`、`erp:warehouse:create`、`erp:warehouse:update`、`erp:warehouse:status`、`erp:warehouse:delete`。
- 库区：`erp:warehouseZone:list`、`erp:warehouseZone:detail`、`erp:warehouseZone:create`、`erp:warehouseZone:update`、`erp:warehouseZone:delete`。
- 货位：`erp:warehouseLocation:list`、`erp:warehouseLocation:detail`、`erp:warehouseLocation:create`、`erp:warehouseLocation:update`、`erp:warehouseLocation:delete`。

## 代码入口

- Model/DTO：`models/erp_warehouse.go`、`models/erp_warehouse_space.go`。
- Service：`services/erp_warehouse_service.go`、`services/erp_warehouse_space_service.go`。
- Controller：`controllers/erp_warehouse_controller.go`、`controllers/erp_warehouse_space_controller.go`。
- Router：`router/router.go` 中 `/api/erp/warehouses`。
- 前端：`hive/apps/web-antdv-next/src/views/erp/warehouse`。
