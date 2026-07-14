package services

import (
	"strings"
	"testing"
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

func TestSelectWorkflowConditionEdge(t *testing.T) {
	graph := &workflowGraph{Edges: []workflowEdge{
		{ID: "default", SourceNodeID: "condition", TargetNodeID: "end-default", Properties: workflowEdgeProperties{IsDefaultBranch: true}},
		{ID: "low", SourceNodeID: "condition", TargetNodeID: "end-low", Properties: workflowEdgeProperties{Priority: 2, ConditionRules: []workflowConditionRule{{Field: "amount", Operator: "greaterThan", Value: "10"}}}},
		{ID: "high", SourceNodeID: "condition", TargetNodeID: "end-high", Properties: workflowEdgeProperties{Priority: 1, ConditionRules: []workflowConditionRule{{Field: "amount", Operator: "greaterThan", Value: "100"}}}},
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
}
