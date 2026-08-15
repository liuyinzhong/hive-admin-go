package services

import (
	"encoding/json"
	"fmt"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"github.com/xuri/excelize/v2"
)

type inventoryBalanceDownloadExporter struct{}

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
	service := NewErpInventoryService()
	query := permission.Apply(service.baseInventoryBalanceExportQuery(), "erp_inventory_balance.creator_id")
	query, err = service.applyInventoryBalanceFilters(query, req)
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
	service := NewErpInventoryService()
	query := permission.Apply(service.baseInventoryBalanceExportQuery(), "erp_inventory_balance.creator_id")
	query, err = service.applyInventoryBalanceFilters(query, req)
	if err != nil {
		return 0, err
	}
	order := buildInventoryBalanceOrder(req.Sorts)
	if order == "" {
		order = "erp_inventory_balance.update_date desc, erp_inventory_balance.create_date desc"
	}

	headers := []interface{}{
		"仓库编码", "仓库名称", "SKU编码", "商品名称", "规格", "生产企业", "包装规格", "批准文号",
		"批次号", "有效期", "单位成本", "包装单位库存", "最小单位库存", "库存金额", "更新时间",
	}
	var processed int64
	err = writeDownloadWorkbook(filePath, "库存余额", headers, func(writer *excelize.StreamWriter) error {
		rows, err := query.Select(erpInventoryBalanceExportSelectFields()).Order(order).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var row erpInventoryBalanceQueryRow
			if err := query.ScanRows(rows, &row); err != nil {
				return err
			}
			processed++
			if processed > totalRows {
				return ErrDownloadDataChanged
			}
			cell, err := excelize.CoordinatesToCellName(1, int(processed)+1)
			if err != nil {
				return err
			}
			values := []interface{}{
				row.WarehouseCode,
				row.WarehouseName,
				row.SkuCode,
				row.ProductName,
				row.SpecName,
				row.EnterpriseName,
				row.PackageSpecName,
				row.ApprovalNo,
				row.BatchNo,
				formatErpInventoryDate(row.ExpiryDate),
				row.UnitCost,
				fmt.Sprintf("%d%s", row.PackageUnitCount, row.PackageUnitName),
				fmt.Sprintf("%d%s", row.MinUnitCount, row.MinUnitName),
				multiplyErpInventoryAmount(row.UnitCost, row.PackageUnitCount),
				utils.TimeToString(row.UpdateDate),
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

func parseInventoryBalanceExportRequest(payload, creatorID string) (models.ErpInventoryBalanceListRequest, datapermission.Permission, error) {
	request, err := decodeInventoryBalanceExportRequest(payload)
	if err != nil {
		return models.ErpInventoryBalanceListRequest{}, datapermission.Permission{}, err
	}
	if creatorID == "" {
		return models.ErpInventoryBalanceListRequest{}, datapermission.Permission{}, fmt.Errorf("导出任务缺少创建用户")
	}
	permission, err := resolveDataPermission(creatorID)
	if err != nil {
		return models.ErpInventoryBalanceListRequest{}, datapermission.Permission{}, err
	}
	return models.ErpInventoryBalanceListRequest{
		WarehouseID:  request.WarehouseID,
		SkuCode:      request.SkuCode,
		BatchNo:      request.BatchNo,
		OnlyPositive: request.OnlyPositive,
		Sorts:        request.Sorts,
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
