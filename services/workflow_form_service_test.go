package services

import (
	"encoding/json"
	"strings"
	"testing"

	"hive-admin-go/models"
)

func TestParseAndValidateFormSchema(t *testing.T) {
	raw := json.RawMessage(`[
		{"component":"InputNumber","fieldName":"amount","label":"申请金额","rules":[{"type":"required"},{"type":"number","min":10,"max":100}]},
		{"component":"Select","fieldName":"level","label":"等级","componentProps":{"options":[{"label":"高","value":"high"}]}}
	]`)
	fields, stored, err := parseAndValidateFormSchema(raw)
	if err != nil {
		t.Fatalf("parseAndValidateFormSchema() error = %v", err)
	}
	if len(fields) != 2 || !json.Valid([]byte(stored)) {
		t.Fatalf("parseAndValidateFormSchema() fields = %d, stored = %s", len(fields), stored)
	}

	duplicate := json.RawMessage(`[
		{"component":"Input","fieldName":"name"},
		{"component":"Input","fieldName":"name"}
	]`)
	if _, _, err := parseAndValidateFormSchema(duplicate); err == nil || !strings.Contains(err.Error(), "不能重复") {
		t.Fatalf("parseAndValidateFormSchema() duplicate error = %v", err)
	}

	parentConflict := json.RawMessage(`[
		{"component":"Input","fieldName":"applicant"},
		{"component":"Input","fieldName":"applicant.name"}
	]`)
	if _, _, err := parseAndValidateFormSchema(parentConflict); err == nil || !strings.Contains(err.Error(), "互为父子") {
		t.Fatalf("parseAndValidateFormSchema() parent conflict error = %v", err)
	}
}

func TestValidateNestedFormSchemaValues(t *testing.T) {
	fields := []models.FormSchemaField{
		{Component: "Input", FieldName: "applicant.name", Label: "申请人", Rules: []models.FormSchemaRule{{Type: "required"}}},
		{Component: "Input", FieldName: "applicant.dept", Label: "部门"},
	}
	values := map[string]interface{}{
		"applicant": map[string]interface{}{"name": "李四", "dept": "研发部"},
	}
	if err := validateFormSchemaValues(fields, values); err != nil {
		t.Fatalf("validateFormSchemaValues() nested error = %v", err)
	}
	values["applicant"].(map[string]interface{})["extra"] = true
	if err := validateFormSchemaValues(fields, values); err == nil || !strings.Contains(err.Error(), "applicant.extra") {
		t.Fatalf("validateFormSchemaValues() nested unknown error = %v", err)
	}
}

func TestValidateFormSchemaValues(t *testing.T) {
	fields := []models.FormSchemaField{
		{Component: "InputNumber", FieldName: "amount", Label: "申请金额", Rules: []models.FormSchemaRule{{Type: "required"}, {Type: "number", Min: formFloat64Ptr(10), Max: formFloat64Ptr(100)}}},
		{Component: "Input", FieldName: "name", Label: "名称", Rules: []models.FormSchemaRule{{Type: "string", Min: formFloat64Ptr(2), Max: formFloat64Ptr(5)}}},
	}
	if err := validateFormSchemaValues(fields, map[string]interface{}{"amount": float64(50), "name": "测试"}); err != nil {
		t.Fatalf("validateFormSchemaValues() error = %v", err)
	}
	tests := []struct {
		name   string
		values map[string]interface{}
		want   string
	}{
		{name: "required", values: map[string]interface{}{"name": "测试"}, want: "申请金额"},
		{name: "minimum", values: map[string]interface{}{"amount": float64(5), "name": "测试"}, want: "不能小于"},
		{name: "unknown", values: map[string]interface{}{"amount": float64(50), "name": "测试", "extra": true}, want: "未知字段"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateFormSchemaValues(fields, tt.values); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateFormSchemaValues() error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestValidateWorkflowConditionFields(t *testing.T) {
	fields := []models.FormSchemaField{
		{Component: "InputNumber", FieldName: "amount", Label: "申请金额"},
		{Component: "Select", FieldName: "level", Label: "等级"},
	}
	graph := &workflowGraph{Edges: []workflowEdge{{
		Properties: workflowEdgeProperties{ConditionRules: []workflowConditionRule{{
			Field: "amount", Operator: "greaterThan", Value: "10",
		}}},
	}}}
	if err := validateWorkflowConditionFields(graph, fields); err != nil {
		t.Fatalf("validateWorkflowConditionFields() error = %v", err)
	}
	graph.Edges[0].Properties.ConditionRules[0].Field = "level"
	if err := validateWorkflowConditionFields(graph, fields); err == nil || !strings.Contains(err.Error(), "不是数字") {
		t.Fatalf("validateWorkflowConditionFields() type error = %v", err)
	}
}

func TestValidateWorkflowFieldPermissions(t *testing.T) {
	fields := []models.FormSchemaField{{Component: "InputNumber", FieldName: "amount", Label: "申请金额"}}
	graph := &workflowGraph{Nodes: []workflowNode{{
		ID: "approve",
		Properties: workflowNodeProperties{
			NodeType: "approve", FieldPermissions: map[string]string{"amount": "editable"},
		},
	}}}
	if err := validateWorkflowConditionFields(graph, fields); err != nil {
		t.Fatalf("validateWorkflowConditionFields() permission error = %v", err)
	}
	graph.Nodes[0].Properties.FieldPermissions["missing"] = "readonly"
	if err := validateWorkflowConditionFields(graph, fields); err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Fatalf("validateWorkflowConditionFields() missing error = %v", err)
	}
}

func TestNormalizeFormSchemaRequestLayout(t *testing.T) {
	req := &models.UpsertFormSchemaRequest{
		SchemaKey:  "expense_apply",
		SchemaName: "报销申请表",
		Layout:     models.FormSchemaLayoutDouble,
		Schema:     json.RawMessage(`[{"component":"Input","fieldName":"name"}]`),
	}
	_, _, layout, _, _, err := normalizeFormSchemaRequest(req)
	if err != nil || layout != models.FormSchemaLayoutDouble {
		t.Fatalf("normalizeFormSchemaRequest() layout = %s, error = %v", layout, err)
	}
	req.Layout = "invalid"
	if _, _, _, _, _, err := normalizeFormSchemaRequest(req); err == nil {
		t.Fatal("normalizeFormSchemaRequest() accepted invalid layout")
	}
}

func formFloat64Ptr(value float64) *float64 { return &value }
