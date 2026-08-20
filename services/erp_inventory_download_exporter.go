package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"github.com/xuri/excelize/v2"
)

type inventoryBalanceDownloadExporter struct{}

type inventoryBalanceExportRequest struct {
	models.ErpInventoryBalanceListRequest
	Filename  string
	SheetName string
	Columns   []models.DownloadExportColumn
	IsHeader  *bool
	IsTitle   *bool
	Original  *bool
}

type inventoryBalanceExportColumn struct {
	Field string
	Title string
	Width int
}

var inventoryBalanceExportColumnTitles = map[string]string{
	"warehouseCode":    "仓库编码",
	"warehouseName":    "仓库名称",
	"skuCode":          "SKU编码",
	"productName":      "商品名称",
	"specName":         "规格",
	"enterpriseName":   "生产企业",
	"packageSpecName":  "包装规格",
	"approvalNo":       "批准文号",
	"batchNo":          "批次号",
	"expiryDate":       "有效期",
	"unitCost":         "单位成本",
	"packageUnitCount": "包装单位库存",
	"minUnitCount":     "最小单位库存",
	"inventoryAmount":  "库存金额",
	"updateDate":       "更新时间",
}

type InventoryBalanceExportPayload struct {
	Request models.InventoryBalanceExportRequest `json:"request"`
}

func NewInventoryBalanceExportPayload(request models.InventoryBalanceExportRequest) InventoryBalanceExportPayload {
	return InventoryBalanceExportPayload{Request: request}
}

func (e *inventoryBalanceDownloadExporter) Count(payload, creatorID string) (int64, error) {
	req, permission, err := parseInventoryBalanceExportRequest(payload, creatorID)
	if err != nil {
		return 0, err
	}
	if _, err := resolveInventoryBalanceExportColumns(req); err != nil {
		return 0, err
	}
	service := NewErpInventoryService()
	query := permission.Apply(service.baseInventoryBalanceExportQuery(), "erp_inventory_balance.creator_id")
	query, err = service.applyInventoryBalanceFilters(query, req.ErpInventoryBalanceListRequest)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (e *inventoryBalanceDownloadExporter) Export(payload, creatorID, filePath string, totalRows int64, onProgress func(int64)) (int64, error) {
	req, permission, err := parseInventoryBalanceExportRequest(payload, creatorID)
	if err != nil {
		return 0, err
	}
	columns, err := resolveInventoryBalanceExportColumns(req)
	if err != nil {
		return 0, err
	}
	service := NewErpInventoryService()
	query := permission.Apply(service.baseInventoryBalanceExportQuery(), "erp_inventory_balance.creator_id")
	query, err = service.applyInventoryBalanceFilters(query, req.ErpInventoryBalanceListRequest)
	if err != nil {
		return 0, err
	}
	order := buildInventoryBalanceOrder(req.Sorts)
	if order == "" {
		order = "erp_inventory_balance.update_date desc, erp_inventory_balance.create_date desc"
	}

	includeHeader := exportBoolValue(req.IsHeader, true)
	useTitle := exportBoolValue(req.IsTitle, true)
	original := exportBoolValue(req.Original, false)
	headers := make([]interface{}, 0, len(columns))
	widths := make([]int, 0, len(columns))
	for _, column := range columns {
		if useTitle {
			headers = append(headers, column.Title)
		} else {
			headers = append(headers, column.Field)
		}
		widths = append(widths, column.Width)
	}
	sheetName := normalizeDownloadSheetName(req.SheetName)
	var processed int64
	err = writeDownloadWorkbookWithWidths(filePath, sheetName, headers, includeHeader, widths, func(writer *excelize.StreamWriter) error {
		rows, err := query.Select(erpInventoryBalanceExportSelectFields()).Order(order).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		rowOffset := 0
		if includeHeader {
			rowOffset = 1
		}
		for rows.Next() {
			var row erpInventoryBalanceQueryRow
			if err := query.ScanRows(rows, &row); err != nil {
				return err
			}
			processed++
			if processed > totalRows {
				return ErrDownloadDataChanged
			}
			cell, err := excelize.CoordinatesToCellName(1, int(processed)+rowOffset)
			if err != nil {
				return err
			}
			values := make([]interface{}, 0, len(columns))
			for _, column := range columns {
				values = append(values, inventoryBalanceExportColumnValue(column.Field, row, original))
			}
			if err := writer.SetRow(cell, values); err != nil {
				return err
			}
			onProgress(processed)
		}
		return rows.Err()
	})
	return processed, err
}

func parseInventoryBalanceExportRequest(payload, creatorID string) (inventoryBalanceExportRequest, datapermission.Permission, error) {
	request, err := decodeInventoryBalanceExportRequest(payload)
	if err != nil {
		return inventoryBalanceExportRequest{}, datapermission.Permission{}, err
	}
	if creatorID == "" {
		return inventoryBalanceExportRequest{}, datapermission.Permission{}, fmt.Errorf("导出任务缺少创建用户")
	}
	permission, err := resolveDataPermission(creatorID)
	if err != nil {
		return inventoryBalanceExportRequest{}, datapermission.Permission{}, err
	}
	return inventoryBalanceExportRequest{
		ErpInventoryBalanceListRequest: models.ErpInventoryBalanceListRequest{
			WarehouseID:  request.WarehouseID,
			SkuCode:      request.SkuCode,
			BatchNo:      request.BatchNo,
			OnlyPositive: request.OnlyPositive,
			Sorts:        request.Sorts,
		},
		Filename:  request.Filename,
		SheetName: request.SheetName,
		Columns:   request.Columns,
		IsHeader:  request.IsHeader,
		IsTitle:   request.IsTitle,
		Original:  request.Original,
	}, permission, nil
}

func decodeInventoryBalanceExportRequest(payload string) (models.InventoryBalanceExportRequest, error) {
	var envelope struct {
		Request json.RawMessage `json:"request"`
	}
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return models.InventoryBalanceExportRequest{}, err
	}
	requestPayload := []byte(payload)
	if len(envelope.Request) > 0 && string(envelope.Request) != "null" {
		requestPayload = envelope.Request
	}
	var request models.InventoryBalanceExportRequest
	if err := json.Unmarshal(requestPayload, &request); err != nil {
		return models.InventoryBalanceExportRequest{}, err
	}
	return request, nil
}

func resolveInventoryBalanceExportColumns(request inventoryBalanceExportRequest) ([]inventoryBalanceExportColumn, error) {
	if len(request.Columns) == 0 {
		return nil, fmt.Errorf("导出列不能为空")
	}
	columns := make([]inventoryBalanceExportColumn, 0, len(request.Columns))
	seen := make(map[string]struct{}, len(request.Columns))
	for _, selectedColumn := range request.Columns {
		field := strings.TrimSpace(selectedColumn.Field)
		title, ok := inventoryBalanceExportColumnTitles[field]
		if !ok {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		if value := strings.TrimSpace(selectedColumn.Title); value != "" {
			titleRunes := []rune(value)
			if len(titleRunes) > 255 {
				value = string(titleRunes[:255])
			}
			title = value
		}
		columns = append(columns, inventoryBalanceExportColumn{
			Field: field,
			Title: title,
			Width: selectedColumn.Width,
		})
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("没有可导出的有效列")
	}
	return columns, nil
}

func inventoryBalanceExportColumnValue(field string, row erpInventoryBalanceQueryRow, original bool) interface{} {
	switch field {
	case "warehouseCode":
		return row.WarehouseCode
	case "warehouseName":
		return row.WarehouseName
	case "skuCode":
		return row.SkuCode
	case "productName":
		return row.ProductName
	case "specName":
		return row.SpecName
	case "enterpriseName":
		return row.EnterpriseName
	case "packageSpecName":
		return row.PackageSpecName
	case "approvalNo":
		return row.ApprovalNo
	case "batchNo":
		return row.BatchNo
	case "expiryDate":
		if original {
			return row.ExpiryDate
		}
		return formatErpInventoryDate(row.ExpiryDate)
	case "unitCost":
		return row.UnitCost
	case "packageUnitCount":
		if original {
			return row.PackageUnitCount
		}
		return fmt.Sprintf("%d%s", row.PackageUnitCount, row.PackageUnitName)
	case "minUnitCount":
		if original {
			return row.MinUnitCount
		}
		return fmt.Sprintf("%d%s", row.MinUnitCount, row.MinUnitName)
	case "inventoryAmount":
		return multiplyErpInventoryAmount(row.UnitCost, row.PackageUnitCount)
	case "updateDate":
		return utils.TimeToString(row.UpdateDate)
	default:
		return ""
	}
}

func (e *inventoryBalanceDownloadExporter) ResolveFileName(payload string, _ time.Time) (string, error) {
	request, err := decodeInventoryBalanceExportRequest(payload)
	if err != nil {
		return "", err
	}
	return normalizeDownloadFileName(request.Filename), nil
}
