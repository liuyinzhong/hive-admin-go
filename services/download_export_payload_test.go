package services

import (
	"strings"
	"testing"

	"hive-admin-go/models"
)

func TestDecodeDevTaskExportRequestSupportsLegacyAndEnvelopePayloads(t *testing.T) {
	payloads := []string{
		`{"projectId":"project-1","taskTitle":"任务一","taskStatus":[1,2],"sorts":"createDate,desc"}`,
		`{"request":{"projectId":"project-1","taskTitle":"任务一","taskStatus":[1,2],"sorts":"createDate,desc"},"userId":"legacy-user"}`,
	}
	for _, payload := range payloads {
		request, err := decodeDevTaskExportRequest(payload)
		if err != nil {
			t.Fatalf("decode payload %s: %v", payload, err)
		}
		if request.ProjectID != "project-1" || request.TaskTitle != "任务一" || len(request.TaskStatuses) != 2 {
			t.Fatalf("unexpected request decoded from %s: %+v", payload, request)
		}
	}
}

func TestDecodeInventoryBalanceExportRequestSupportsLegacyAndEnvelopePayloads(t *testing.T) {
	payloads := []string{
		`{"warehouseId":"warehouse-1","skuCode":"SKU-1","onlyPositive":true}`,
		`{"request":{"warehouseId":"warehouse-1","skuCode":"SKU-1","onlyPositive":true},"userId":"legacy-user"}`,
	}
	for _, payload := range payloads {
		request, err := decodeInventoryBalanceExportRequest(payload)
		if err != nil {
			t.Fatalf("decode payload %s: %v", payload, err)
		}
		if request.WarehouseID != "warehouse-1" || request.SkuCode != "SKU-1" || !request.OnlyPositive {
			t.Fatalf("unexpected request decoded from %s: %+v", payload, request)
		}
	}
}

func TestResolveDevTaskExportColumnsUsesWhitelistAndRequestOrder(t *testing.T) {
	columns, err := resolveDevTaskExportColumns(models.DevTaskExportRequest{
		Columns: []models.DownloadExportColumn{
			{Field: "taskStatus", Title: "状态值"},
			{Field: "operation", Title: "操作"},
			{Field: "taskStatus", Title: "重复列"},
			{Field: "taskTitle"},
		},
	})
	if err != nil {
		t.Fatalf("resolve columns: %v", err)
	}
	if len(columns) != 2 || columns[0].Field != "taskStatus" || columns[0].Title != "状态值" || columns[1].Field != "taskTitle" {
		t.Fatalf("unexpected columns: %+v", columns)
	}
}

func TestNormalizeDevTaskExportOptions(t *testing.T) {
	if got := normalizeDevTaskFileName(`任务:/列表`); got != "任务__列表.xlsx" {
		t.Fatalf("unexpected file name: %s", got)
	}
	if got := normalizeDevTaskSheetName("任务/管理" + strings.Repeat("名", 40)); len([]rune(got)) != 31 {
		t.Fatalf("unexpected sheet name length: %d", len([]rune(got)))
	}
	if !exportBoolValue(nil, true) || exportBoolValue(func() *bool { value := false; return &value }(), true) {
		t.Fatal("unexpected boolean fallback")
	}
}
