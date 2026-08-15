package services

import "testing"

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
