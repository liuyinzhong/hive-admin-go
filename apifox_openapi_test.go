package main

import (
	"encoding/json"
	"testing"
)

func TestBuildApifoxSwaggerDocAddsNestedFolderFromTag(t *testing.T) {
	swaggerDoc := `{
		"swagger": "2.0",
		"paths": {
			"/system/logs": {
				"get": {
					"tags": ["系统管理-日志管理"],
					"summary": "获取日志列表"
				}
			}
		}
	}`

	got, err := buildApifoxSwaggerDoc(swaggerDoc)
	if err != nil {
		t.Fatalf("buildApifoxSwaggerDoc() error = %v", err)
	}

	operation := mustOperation(t, got, "/system/logs", "get")
	if operation["x-apifox-folder"] != "系统管理/日志管理" {
		t.Fatalf("x-apifox-folder = %v, want 系统管理/日志管理", operation["x-apifox-folder"])
	}
}

func TestBuildApifoxSwaggerDocKeepsSingleLevelTag(t *testing.T) {
	swaggerDoc := `{
		"swagger": "2.0",
		"paths": {
			"/form/schemas": {
				"post": {
					"tags": ["表单管理"],
					"summary": "创建表单"
				}
			}
		}
	}`

	got, err := buildApifoxSwaggerDoc(swaggerDoc)
	if err != nil {
		t.Fatalf("buildApifoxSwaggerDoc() error = %v", err)
	}

	operation := mustOperation(t, got, "/form/schemas", "post")
	if operation["x-apifox-folder"] != "表单管理" {
		t.Fatalf("x-apifox-folder = %v, want 表单管理", operation["x-apifox-folder"])
	}
}

func TestBuildApifoxSwaggerDocSkipsPathMetadata(t *testing.T) {
	swaggerDoc := `{
		"swagger": "2.0",
		"paths": {
			"/system/logs": {
				"parameters": [
					{"name": "tenantId", "in": "header"}
				],
				"get": {
					"tags": ["系统管理-日志管理"],
					"summary": "获取日志列表"
				}
			}
		}
	}`

	got, err := buildApifoxSwaggerDoc(swaggerDoc)
	if err != nil {
		t.Fatalf("buildApifoxSwaggerDoc() error = %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(got), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	path := doc["paths"].(map[string]interface{})["/system/logs"].(map[string]interface{})
	if _, ok := path["parameters"].(map[string]interface{}); ok {
		t.Fatal("path-level parameters should not be treated as an operation")
	}
}

func TestBuildApifoxSwaggerDocRejectsInvalidJSON(t *testing.T) {
	if _, err := buildApifoxSwaggerDoc(`{"swagger":`); err == nil {
		t.Fatal("buildApifoxSwaggerDoc() error = nil, want error")
	}
}

func mustOperation(t *testing.T, swaggerDoc string, pathName string, method string) map[string]interface{} {
	t.Helper()

	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(swaggerDoc), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	paths, ok := doc["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("paths is missing")
	}
	path, ok := paths[pathName].(map[string]interface{})
	if !ok {
		t.Fatalf("path %s is missing", pathName)
	}
	operation, ok := path[method].(map[string]interface{})
	if !ok {
		t.Fatalf("operation %s %s is missing", method, pathName)
	}

	return operation
}
