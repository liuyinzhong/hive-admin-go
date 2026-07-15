package services

import (
	"strings"
	"testing"

	"hive-admin-go/models"
)

// TestValidateWorkflowFormSchema 验证表单字段标识、类型和选项约束。
func TestValidateWorkflowFormSchema(t *testing.T) {
	schema := workflowFormTestSchema()
	if err := validateWorkflowFormSchema(schema); err != nil {
		t.Fatalf("validateWorkflowFormSchema() error = %v", err)
	}

	schema.Fields[1].Key = "amount"
	if err := validateWorkflowFormSchema(schema); err == nil || !strings.Contains(err.Error(), "不能重复") {
		t.Fatalf("validateWorkflowFormSchema() duplicate key error = %v", err)
	}
}

// TestValidateWorkflowFormVariables 验证动态申请数据的必填、类型和未知字段约束。
func TestValidateWorkflowFormVariables(t *testing.T) {
	schema := workflowFormTestSchema()
	valid := map[string]interface{}{"amount": float64(100), "level": "high"}
	if err := validateWorkflowFormVariables(schema, valid); err != nil {
		t.Fatalf("validateWorkflowFormVariables() error = %v", err)
	}

	missing := map[string]interface{}{"level": "high"}
	if err := validateWorkflowFormVariables(schema, missing); err == nil || !strings.Contains(err.Error(), "申请金额") {
		t.Fatalf("validateWorkflowFormVariables() required error = %v", err)
	}

	unknown := map[string]interface{}{"amount": 100, "extra": "x"}
	if err := validateWorkflowFormVariables(schema, unknown); err == nil || !strings.Contains(err.Error(), "未知字段") {
		t.Fatalf("validateWorkflowFormVariables() unknown field error = %v", err)
	}
}

// TestValidateWorkflowFormGrid 验证栅格列宽及列内字段可以参与表单校验。
func TestValidateWorkflowFormGrid(t *testing.T) {
	schema := &models.WorkflowFormSchema{
		Version: 1,
		Fields: []models.WorkflowFormField{{
			ID:   "grid_1",
			Type: "grid",
			Columns: []models.WorkflowFormGridColumn{
				{ID: "column_1", Span: 8, Fields: []models.WorkflowFormField{{ID: "amount", Key: "amount", Label: "申请金额", Type: "number", Required: true}}},
				{ID: "column_2", Span: 16, Fields: []models.WorkflowFormField{{ID: "level", Key: "level", Label: "等级", Type: "select", Options: []models.WorkflowFormOption{{Label: "高", Value: "high"}}}}},
			},
		}},
	}
	if err := validateWorkflowFormSchema(schema); err != nil {
		t.Fatalf("validateWorkflowFormSchema() grid error = %v", err)
	}
	valid := map[string]interface{}{"amount": float64(100), "level": "high"}
	if err := validateWorkflowFormVariables(schema, valid); err != nil {
		t.Fatalf("validateWorkflowFormVariables() grid error = %v", err)
	}

	schema.Fields[0].Columns[1].Span = 17
	if err := validateWorkflowFormSchema(schema); err == nil || !strings.Contains(err.Error(), "不能超过24") {
		t.Fatalf("validateWorkflowFormSchema() grid span error = %v", err)
	}
}

// TestValidateWorkflowConditionFields 验证条件分支只能引用存在且类型兼容的表单字段。
func TestValidateWorkflowConditionFields(t *testing.T) {
	schema := workflowFormTestSchema()
	graph := &workflowGraph{Edges: []workflowEdge{{
		Properties: workflowEdgeProperties{ConditionRules: []workflowConditionRule{{
			Field: "amount", Operator: "greaterThan", Value: "10",
		}}},
	}}}
	if err := validateWorkflowConditionFields(graph, schema); err != nil {
		t.Fatalf("validateWorkflowConditionFields() error = %v", err)
	}

	graph.Edges[0].Properties.ConditionRules[0].Field = "level"
	if err := validateWorkflowConditionFields(graph, schema); err == nil || !strings.Contains(err.Error(), "不是数字") {
		t.Fatalf("validateWorkflowConditionFields() type error = %v", err)
	}
}

// TestValidateWorkflowFieldPermissions 验证审批节点字段权限只能引用表单字段和受支持的权限值。
func TestValidateWorkflowFieldPermissions(t *testing.T) {
	schema := workflowFormTestSchema()
	graph := &workflowGraph{Nodes: []workflowNode{{
		ID: "approve",
		Properties: workflowNodeProperties{
			NodeType:         "approve",
			FieldPermissions: map[string]string{"amount": "editable", "level": "hidden"},
		},
	}}}
	if err := validateWorkflowConditionFields(graph, schema); err != nil {
		t.Fatalf("validateWorkflowConditionFields() permission error = %v", err)
	}

	graph.Nodes[0].Properties.FieldPermissions["missing"] = "readonly"
	if err := validateWorkflowConditionFields(graph, schema); err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Fatalf("validateWorkflowConditionFields() missing field error = %v", err)
	}
	delete(graph.Nodes[0].Properties.FieldPermissions, "missing")
	graph.Nodes[0].Properties.FieldPermissions["amount"] = "invalid"
	if err := validateWorkflowConditionFields(graph, schema); err == nil || !strings.Contains(err.Error(), "权限无效") {
		t.Fatalf("validateWorkflowConditionFields() invalid permission error = %v", err)
	}
}

// workflowFormTestSchema 返回流程表单校验测试使用的固定结构。
func workflowFormTestSchema() *models.WorkflowFormSchema {
	return &models.WorkflowFormSchema{
		Version: 1,
		Fields: []models.WorkflowFormField{
			{ID: "amount", Key: "amount", Label: "申请金额", Type: "number", Required: true},
			{ID: "level", Key: "level", Label: "等级", Type: "select", Options: []models.WorkflowFormOption{{Label: "高", Value: "high"}}},
		},
	}
}
