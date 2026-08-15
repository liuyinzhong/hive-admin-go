package services

import (
	"fmt"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
)

// PrintDocumentService 负责把各业务单据适配成统一的打印数据协议。
// 协议统一，业务查询仍由各自领域 Service 负责，避免打印接口复制业务查询逻辑。
type PrintDocumentService struct {
	purchaseInboundService *ErpPurchaseInboundService
}

func NewPrintDocumentService() *PrintDocumentService {
	return &PrintDocumentService{
		purchaseInboundService: NewErpPurchaseInboundService(),
	}
}

func (s *PrintDocumentService) GetPrintDocument(documentType, documentID string, permission datapermission.Permission) (*models.PrintDocumentResponse, error) {
	switch documentType {
	case models.PrintDocumentTypePurchaseInbound:
		return s.getPurchaseInboundPrintDocument(documentID, permission)
	default:
		return nil, fmt.Errorf("%w: 不支持的打印单据类型", ErrPrintTemplateInvalidInput)
	}
}

func (s *PrintDocumentService) getPurchaseInboundPrintDocument(inboundID string, permission datapermission.Permission) (*models.PrintDocumentResponse, error) {
	detail, err := s.purchaseInboundService.GetPurchaseInboundDetail(inboundID, permission)
	if err != nil {
		return nil, err
	}

	header := map[string]interface{}{
		"inboundId":     detail.InboundID,
		"inboundNo":     detail.InboundNo,
		"inboundDate":   detail.InboundDate,
		"supplierName":  detail.SupplierName,
		"warehouseName": detail.WarehouseName,
		"remark":        detail.Remark,
		"creatorId":     detail.CreatorID,
		"createDate":    detail.CreateDate,
	}
	items := make([]map[string]interface{}, 0, len(detail.Items))
	for _, item := range detail.Items {
		items = append(items, map[string]interface{}{
			"inboundItemId":   item.InboundItemID,
			"lineNo":          item.LineNo,
			"skuId":           item.SkuID,
			"skuCode":         item.SkuCode,
			"productName":     item.ProductName,
			"specName":        item.SpecName,
			"enterpriseName":  item.EnterpriseName,
			"packageSpecName": item.PackageSpecName,
			"packageUnitName": item.PackageUnitName,
			"minUnitName":     item.MinUnitName,
			"batchNo":         item.BatchNo,
			"expiryDate":      item.ExpiryDate,
			"unitCost":        item.UnitCost,
			"quantity":        item.Quantity,
			"amount":          item.Amount,
			"remark":          item.Remark,
		})
	}

	return &models.PrintDocumentResponse{
		DocumentType:  models.PrintDocumentTypePurchaseInbound,
		SchemaVersion: 1,
		Header:        header,
		Items:         items,
		Summary: map[string]interface{}{
			"lineCount":   detail.LineCount,
			"totalAmount": detail.TotalAmount,
		},
	}, nil
}
