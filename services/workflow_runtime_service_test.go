package services

import (
	"strings"
	"testing"
	"time"

	"hive-admin-go/models"
)

func TestEvaluateWorkflowRule(t *testing.T) {
	variables := map[string]interface{}{
		"amount":    120.5,
		"applicant": map[string]interface{}{"dept": "研发部"},
		"tags":      []interface{}{"urgent", "purchase"},
		"remark":    "",
	}
	tests := []struct {
		name string
		rule workflowConditionRule
		want bool
	}{
		{name: "nested field", rule: workflowConditionRule{Field: "applicant.dept", Operator: "equal", Value: "研发部"}, want: true},
		{name: "number comparison", rule: workflowConditionRule{Field: "amount", Operator: "greaterThan", Value: "100"}, want: true},
		{name: "slice contains", rule: workflowConditionRule{Field: "tags", Operator: "contains", Value: "urgent"}, want: true},
		{name: "empty value", rule: workflowConditionRule{Field: "remark", Operator: "empty"}, want: true},
		{name: "missing value", rule: workflowConditionRule{Field: "missing", Operator: "equal", Value: "x"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateWorkflowRule(tt.rule, variables)
			if err != nil {
				t.Fatalf("evaluateWorkflowRule() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("evaluateWorkflowRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkflowCategoryPrefix(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{category: "general", want: "TY"},
		{category: "finance", want: "CW"},
		{category: "hr", want: "RS"},
		{category: "administration", want: "XZ"},
		{category: "procurement", want: "CG"},
		{category: "development", want: "DEV"},
		{category: "system", want: "SYS"},
		{category: "other", want: "QT"},
	}
	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got, err := workflowCategoryPrefix(&tt.category)
			if err != nil {
				t.Fatalf("workflowCategoryPrefix() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("workflowCategoryPrefix() = %s, want %s", got, tt.want)
			}
		})
	}

	if _, err := workflowCategoryPrefix(nil); err == nil {
		t.Fatal("workflowCategoryPrefix() accepted nil category")
	}
	unknown := "unknown"
	if _, err := workflowCategoryPrefix(&unknown); err == nil {
		t.Fatal("workflowCategoryPrefix() accepted unknown category")
	}
}

func TestFormatWorkflowInstanceNo(t *testing.T) {
	now := time.Date(2026, time.July, 15, 15, 49, 19, 0, time.Local)
	got, err := formatWorkflowInstanceNo("CW", now, 1)
	if err != nil {
		t.Fatalf("formatWorkflowInstanceNo() error = %v", err)
	}
	if got != "CW20260715000001" {
		t.Fatalf("formatWorkflowInstanceNo() = %s", got)
	}

	if _, err = formatWorkflowInstanceNo("CW", now, 0); err == nil {
		t.Fatal("formatWorkflowInstanceNo() accepted zero sequence")
	}
	if _, err = formatWorkflowInstanceNo("CW", now, 1_000_000); err == nil {
		t.Fatal("formatWorkflowInstanceNo() accepted overflowing sequence")
	}
}

func TestWorkflowInstanceTitle(t *testing.T) {
	got := workflowInstanceTitle("报销流程", "李四员工")
	if got != "报销流程-李四员工" {
		t.Fatalf("workflowInstanceTitle() = %s", got)
	}
}

func TestValidateWorkflowTaskVariableChanges(t *testing.T) {
	permissions := map[string]string{"amount": "editable", "remark": "readonly"}
	if err := validateWorkflowTaskVariableChanges(permissions, map[string]interface{}{"amount": 100}); err != nil {
		t.Fatalf("validateWorkflowTaskVariableChanges() error = %v", err)
	}
	if err := validateWorkflowTaskVariableChanges(permissions, map[string]interface{}{"remark": "changed"}); err == nil || !strings.Contains(err.Error(), "不可编辑") {
		t.Fatalf("validateWorkflowTaskVariableChanges() readonly error = %v", err)
	}
}

func TestWorkflowNodeFieldPermissions(t *testing.T) {
	nodeInstance := &models.WfProcessNodeInstance{
		NodeName:         "部门审批",
		FieldPermissions: `{"amount":"editable","remark":"readonly"}`,
	}
	permissions, err := workflowNodeFieldPermissions(nodeInstance)
	if err != nil {
		t.Fatalf("workflowNodeFieldPermissions() error = %v", err)
	}
	if permissions["amount"] != "editable" || permissions["remark"] != "readonly" {
		t.Fatalf("workflowNodeFieldPermissions() = %#v", permissions)
	}

	nodeInstance.FieldPermissions = "null"
	if _, err = workflowNodeFieldPermissions(nodeInstance); err == nil || !strings.Contains(err.Error(), "快照无效") {
		t.Fatalf("workflowNodeFieldPermissions() null error = %v", err)
	}
}

func TestWorkflowNodeDisplayAction(t *testing.T) {
	tests := []struct {
		name    string
		node    models.WfProcessNodeInstance
		records []models.WorkflowRecordResponse
		want    string
	}{
		{name: "planned", node: models.WfProcessNodeInstance{Status: models.WorkflowNodeStatusPlanned}, want: "planned"},
		{name: "active", node: models.WfProcessNodeInstance{Status: models.WorkflowNodeStatusActive}, want: "pending"},
		{name: "record action", node: models.WfProcessNodeInstance{Status: models.WorkflowNodeStatusTerminated}, records: []models.WorkflowRecordResponse{{Action: "return"}}, want: "return"},
		{name: "completed end", node: models.WfProcessNodeInstance{Status: models.WorkflowNodeStatusCompleted, NodeType: "end"}, want: "complete"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := workflowNodeDisplayAction(tt.node, tt.records); got != tt.want {
				t.Fatalf("workflowNodeDisplayAction() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestWorkflowApprovalDurationSeconds(t *testing.T) {
	startTime := time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local)
	endTime := startTime.Add(75 * time.Minute)
	responseTime := startTime.Add(90 * time.Minute)

	tests := []struct {
		name string
		node models.WfProcessNodeInstance
		want *int64
	}{
		{
			name: "active approval",
			node: models.WfProcessNodeInstance{NodeName: "主管审批", NodeType: "approve", Status: models.WorkflowNodeStatusActive, StartDate: &startTime},
			want: int64Ptr(5400),
		},
		{
			name: "completed approval",
			node: models.WfProcessNodeInstance{NodeName: "主管审批", NodeType: "approve", Status: models.WorkflowNodeStatusCompleted, StartDate: &startTime, EndDate: &endTime},
			want: int64Ptr(4500),
		},
		{
			name: "terminated approval",
			node: models.WfProcessNodeInstance{NodeName: "主管审批", NodeType: "approve", Status: models.WorkflowNodeStatusTerminated, StartDate: &startTime, EndDate: &endTime},
			want: int64Ptr(4500),
		},
		{
			name: "planned approval",
			node: models.WfProcessNodeInstance{NodeName: "主管审批", NodeType: "approve", Status: models.WorkflowNodeStatusPlanned},
		},
		{
			name: "automatic node",
			node: models.WfProcessNodeInstance{NodeName: "条件分支", NodeType: "condition", Status: models.WorkflowNodeStatusCompleted, StartDate: &startTime, EndDate: &endTime},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workflowApprovalDurationSeconds(tt.node, responseTime)
			if err != nil {
				t.Fatalf("workflowApprovalDurationSeconds() error = %v", err)
			}
			if tt.want == nil {
				if got != nil {
					t.Fatalf("workflowApprovalDurationSeconds() = %v, want nil", *got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Fatalf("workflowApprovalDurationSeconds() = %v, want %v", got, *tt.want)
			}
		})
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}

func TestBuildWorkflowReturnTargets(t *testing.T) {
	firstID := "approve-first"
	secondID := "approve-second"
	currentID := "approve-current"
	nodeInstances := []models.WfProcessNodeInstance{
		{NodeID: secondID, NodeName: "二级审批"},
		{NodeID: currentID, NodeName: "当前审批"},
		{NodeID: secondID, NodeName: "二级审批"},
		{NodeID: firstID, NodeName: "一级审批"},
	}

	targets := buildWorkflowReturnTargets(nodeInstances, currentID)
	if len(targets) != 2 || targets[0].NodeID != secondID || targets[1].NodeID != firstID {
		t.Fatalf("buildWorkflowReturnTargets() = %#v", targets)
	}
}

func TestSelectWorkflowConditionEdge(t *testing.T) {
	graph := &workflowGraph{Edges: []workflowEdge{
		{ID: "default", SourceNodeID: "condition", TargetNodeID: "end-default", Properties: workflowEdgeProperties{IsDefaultBranch: true}},
		{ID: "low", SourceNodeID: "condition", TargetNodeID: "end-low", Properties: workflowEdgeProperties{Priority: 2, ConditionLogic: "and", ConditionRules: []workflowConditionRule{{Field: "amount", Operator: "greaterThan", Value: "10"}}}},
		{ID: "high", SourceNodeID: "condition", TargetNodeID: "end-high", Properties: workflowEdgeProperties{Priority: 1, ConditionLogic: "and", ConditionRules: []workflowConditionRule{{Field: "amount", Operator: "greaterThan", Value: "100"}}}},
	}}

	edge, err := selectWorkflowConditionEdge(graph, "condition", map[string]interface{}{"amount": 200})
	if err != nil {
		t.Fatalf("selectWorkflowConditionEdge() error = %v", err)
	}
	if edge.ID != "high" {
		t.Fatalf("selected edge = %s, want high", edge.ID)
	}

	edge, err = selectWorkflowConditionEdge(graph, "condition", map[string]interface{}{"amount": 1})
	if err != nil {
		t.Fatalf("selectWorkflowConditionEdge() default error = %v", err)
	}
	if edge.ID != "default" {
		t.Fatalf("selected edge = %s, want default", edge.ID)
	}
}

func TestParseWorkflowGraphRejectsInvalidRuntimeConfiguration(t *testing.T) {
	base := `{"nodes":[{"id":"start","properties":{"nodeType":"start"}},{"id":"approve","text":{"value":"审批"},"properties":{"nodeType":"approve","assigneeType":"user","assigneeIds":["u1"],"approvalMode":"%s"}},{"id":"end","properties":{"nodeType":"end"}}],"edges":[{"id":"e1","sourceNodeId":"start","targetNodeId":"approve"},{"id":"e2","sourceNodeId":"approve","targetNodeId":"end"}]}`
	_, err := parseWorkflowGraph(strings.Replace(base, "%s", "invalid", 1))
	if err == nil || !strings.Contains(err.Error(), "审批方式无效") {
		t.Fatalf("parseWorkflowGraph() error = %v, want invalid approval mode", err)
	}

	if _, err = parseWorkflowGraph(strings.Replace(base, "%s", "all", 1)); err != nil {
		t.Fatalf("parseWorkflowGraph() valid graph error = %v", err)
	}

	if _, err = parseWorkflowGraph(strings.Replace(base, `"approvalMode":"%s"`, `"approvalMode":""`, 1)); err == nil {
		t.Fatal("parseWorkflowGraph() accepted missing approval mode")
	}

	stringText := strings.Replace(base, `"text":{"value":"审批"}`, `"text":"审批"`, 1)
	stringText = strings.Replace(stringText, "%s", "all", 1)
	if _, err = parseWorkflowGraph(stringText); err == nil {
		t.Fatal("parseWorkflowGraph() accepted legacy string node text")
	}

	missingLogic := `{"nodes":[{"id":"start","properties":{"nodeType":"start"}},{"id":"condition","text":{"value":"条件"},"properties":{"nodeType":"condition"}},{"id":"approve","text":{"value":"审批"},"properties":{"nodeType":"approve","assigneeType":"user","assigneeIds":["u1"],"approvalMode":"any"}},{"id":"end","properties":{"nodeType":"end"}}],"edges":[{"id":"e1","sourceNodeId":"start","targetNodeId":"condition"},{"id":"branch","sourceNodeId":"condition","targetNodeId":"approve","properties":{"priority":1,"conditionRules":[{"field":"amount","operator":"greaterThan","value":"100"}]}},{"id":"default","sourceNodeId":"condition","targetNodeId":"end","properties":{"isDefaultBranch":true}},{"id":"e2","sourceNodeId":"approve","targetNodeId":"end"}]}`
	if _, err = parseWorkflowGraph(missingLogic); err == nil || !strings.Contains(err.Error(), "规则关系") {
		t.Fatalf("parseWorkflowGraph() missing condition logic error = %v", err)
	}
}
