package services

import "hive-admin-go/models"

func printDocumentTypes() []models.PrintDocumentTypeDefinition {
	return []models.PrintDocumentTypeDefinition{
		{
			Code: models.PrintDocumentTypePurchaseInbound,
			Name: "采购入库单",
		},
	}
}

func printFieldGroups() []models.PrintFieldGroup {
	return []models.PrintFieldGroup{
		{
			Code: "header",
			Name: "单据头",
			Fields: []models.PrintFieldDefinition{
				{Path: "header.inboundNo", Label: "入库单号", DataType: "string", Scope: "header", Example: "PIN00000001"},
				{Path: "header.inboundDate", Label: "入库日期", DataType: "date", Scope: "header", Example: "2026-08-03"},
				{Path: "header.supplierName", Label: "供应商", DataType: "string", Scope: "header", Example: "示例供应商"},
				{Path: "header.warehouseName", Label: "入库仓库", DataType: "string", Scope: "header", Example: "中心库"},
				{Path: "header.remark", Label: "单据备注", DataType: "string", Scope: "header", Example: "采购到货入库"},
				{Path: "header.createDate", Label: "创建时间", DataType: "datetime", Scope: "header", Example: "2026-08-03 10:00:00"},
			},
		},
		{
			Code: "item",
			Name: "明细行",
			Fields: []models.PrintFieldDefinition{
				{Path: "items.lineNo", Label: "行号", DataType: "number", Scope: "item", Example: "1"},
				{Path: "items.skuCode", Label: "SKU编码", DataType: "string", Scope: "item", Example: "SKU000001"},
				{Path: "items.productName", Label: "产品名称", DataType: "string", Scope: "item", Example: "阿莫西林胶囊"},
				{Path: "items.specName", Label: "规格", DataType: "string", Scope: "item", Example: "0.25g"},
				{Path: "items.enterpriseName", Label: "生产企业", DataType: "string", Scope: "item", Example: "示例药业"},
				{Path: "items.packageSpecName", Label: "包装规格", DataType: "string", Scope: "item", Example: "10粒/盒"},
				{Path: "items.packageUnitName", Label: "包装单位", DataType: "string", Scope: "item", Example: "盒"},
				{Path: "items.minUnitName", Label: "最小单位", DataType: "string", Scope: "item", Example: "粒"},
				{Path: "items.batchNo", Label: "批号", DataType: "string", Scope: "item", Example: "B20260803001"},
				{Path: "items.expiryDate", Label: "有效期至", DataType: "date", Scope: "item", Example: "2028-12-31"},
				{Path: "items.quantity", Label: "数量", DataType: "number", Scope: "item", Example: "20"},
				{Path: "items.unitCost", Label: "单价", DataType: "currency", Scope: "item", Example: "12.0000"},
				{Path: "items.amount", Label: "金额", DataType: "currency", Scope: "item", Example: "240.0000"},
				{Path: "items.remark", Label: "明细备注", DataType: "string", Scope: "item", Example: "首批到货"},
			},
		},
		{
			Code: "summary",
			Name: "汇总",
			Fields: []models.PrintFieldDefinition{
				{Path: "summary.lineCount", Label: "明细行数", DataType: "number", Scope: "summary", Example: "2"},
				{Path: "summary.totalAmount", Label: "合计金额", DataType: "currency", Scope: "summary", Example: "240.0000"},
			},
		},
		{
			Code: "system",
			Name: "系统字段",
			Fields: []models.PrintFieldDefinition{
				{Path: "system.pageNumber", Label: "当前页码", DataType: "number", Scope: "system", Example: "1"},
				{Path: "system.totalPages", Label: "总页数", DataType: "number", Scope: "system", Example: "2"},
				{Path: "system.printTime", Label: "打印时间", DataType: "datetime", Scope: "system", Example: "2026-08-03 10:00:00"},
			},
		},
	}
}

func printFieldDefinitionMap() map[string]models.PrintFieldDefinition {
	result := make(map[string]models.PrintFieldDefinition)
	for _, group := range printFieldGroups() {
		for _, field := range group.Fields {
			result[field.Path] = field
		}
	}
	return result
}
