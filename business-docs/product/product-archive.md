# 产品档案业务规则

产品档案采用 `SPU → RP → MP → SKU` 四级结构。ERP 直接引用最末级 SKU，不直接以 SPU、RP 或 MP 作为库存对象。

## 业务规则

### PROD-ARC-001 层级归属不可跳级

RP 必须归属于一个 SPU，MP 必须归属于一个 RP，SKU 必须归属于一个 MP。详情接口将四级结构展开为表格行，但展开结果不改变真实归属关系。

### PROD-ARC-002 产品编码由后端生成

SPU、RP 和 SKU 的业务编码由公共编码流水在创建事务中生成，不能由前端填写或复用。数据库主键仍使用 UUID，业务编码不承担主键职责。

### PROD-ARC-003 同层级名称或包装组合必须唯一

- SPU：同一产品类型下，标准化后的通用名称不能重复。
- RP：同一 SPU 下，标准化后的规格名称不能重复。
- MP：同一 RP 下，生产企业和标准化批准文号的组合不能重复。
- SKU：同一 MP 下，包装换算、最小单位、包装单位、整箱换算和整箱单位的组合不能重复。

### PROD-ARC-004 SKU 是业务选择边界

SKU 保存包装链、条码或 GTIN、UDI、是否拆零和追溯模式。包装规格和完整包装链名称由后端根据包装字段派生，调用方不能把展示名称当成独立主数据维护。

追溯模式当前仅允许：

- `NONE`：无需追溯码。
- `REQUIRED`：相关库存动作必须按现有 ERP 规则采集追溯码。

### PROD-ARC-005 启停不级联修改下级

SPU、RP、MP、SKU 均独立维护 `0=停用、1=启用`。停用上级不会批量改写下级状态，也不会删除历史引用。

用于新业务选择的 SKU 选项要求 SPU、RP、MP、SKU 整条层级均启用；RP 选项要求 SPU 和 RP 启用。因此“当前节点启用”不等于“当前层级可供新业务选择”。

### PROD-ARC-006 核心档案使用乐观并发控制

SPU、RP、MP、SKU 的修改和状态切换必须提交 `expectedRowVersion`。版本不一致时拒绝覆盖并要求刷新；成功写入后版本递增。

### PROD-ARC-007 核心层级当前不提供删除

SPU、RP、MP、SKU 当前只有创建、详情、修改和启停接口，没有删除接口。停用用于阻止其进入新的业务选择，历史单据和库存引用继续保留。

## 页面动作与权限

| 资源 | 动作权限 |
|---|---|
| SPU | `product:spu:list`、`create`、`detail`、`update`、`status` |
| RP | `product:rp:list`、`create`、`detail`、`update`、`status` |
| MP | `product:mp:list`、`create`、`detail`、`update`、`status` |
| SKU | `product:sku:list`、`create`、`detail`、`update`、`status` |

完整权限码均为 `product:资源:动作`。SPU、RP、SKU 另有供下拉选择使用的 options 接口，其中部分接口当前未配置按钮权限，修改时必须重新评估数据暴露边界。

## 代码入口

- Model/DTO：`models/product_spu.go`、`product_rp.go`、`product_mp.go`、`product_sku.go`。
- Service：`services/product_spu_service.go`、`product_rp_service.go`、`product_mp_service.go`、`product_sku_service.go`。
- Router：`router/router.go` 中 `/api/product/spus`、`rps`、`mps`、`skus`。
- 前端：`hive/apps/web-antdv-next/src/views/product/spu`。

