# SKU 价格业务规则

SKU 价格独立归属于一个 SKU，并按价格类型、适用范围和生效期共同识别使用场景。

## 业务规则

### PROD-PRICE-001 当前币种仅支持人民币

币种为空时归一化为 `CNY`，其它币种拒绝保存。基础价格和阶梯价格必须大于零，最多四位小数；是否含税必须明确为 `0` 或 `1`。

### PROD-PRICE-002 全局与非全局范围的对象规则不同

`GLOBAL` 价格不保存 `scopeId`；非全局范围必须提供 `scopeId`。当前前端提供 `GLOBAL`、`CUSTOMER`、`CHANNEL` 三种范围，并为非全局范围使用企业主体选项。

### PROD-PRICE-003 同维度生效期不能重叠

同一 SKU、价格类型、范围类型和范围对象下，任意两条价格的生效期不能重叠。创建、修改和重新启用时都会检查；停用记录仍参与期间重叠检查。

生效结束必须晚于生效开始，结束为空表示长期有效。新建价格的生效开始不能早于当前时间；修改时只有在改变生效开始的情况下才禁止改到过去。

### PROD-PRICE-004 价格变更使用乐观并发控制

修改、启停、删除价格以及保存阶梯价格均校验价格的行版本。任何成功变更都会递增价格版本，旧页面提交会被拒绝。

### PROD-PRICE-005 阶梯区间为闭区间且不能重叠

起始数量必须为正整数，结束数量为空表示无上限，否则必须大于等于起始数量。区间按起始数量排序后不能重叠，起始数量不能重复；当前不要求相邻区间连续。

### PROD-PRICE-006 阶梯保存是全量替换语义

保存时，提交数组中的已有 ID 被更新、新行被创建、未提交旧行被软删除。提交空数组可清空全部阶梯；字段缺失形成的 `nil` 数组会被拒绝。

### PROD-PRICE-007 删除价格同时删除阶梯

删除是软删除，并在同一事务中软删除该价格的所有阶梯。核心产品层级不会因为价格删除而变化。

## 权限

- 价格：`product:skuPrice:list`、`create`、`update`、`status`、`delete`。
- 阶梯：`product:skuPriceTier:list`、`save`。

完整权限前缀分别为 `product:skuPrice:` 和 `product:skuPriceTier:`。

## 代码入口

- Model/DTO：`models/product_sku_price.go`。
- Service：`services/product_sku_price_service.go`。
- Controller：`controllers/product_sku_price_controller.go`。
- Router：`router/product.go` 中 `/api/product/skus/:skuId/prices`。
- 前端：`hive/apps/web-antdv-next/src/views/product/spu/components/sku-price`。

