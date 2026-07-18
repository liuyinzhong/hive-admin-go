package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	projectDocs "hive-admin-go/docs"
	"os"
	"path/filepath"
	"strings"
)

const swaggerJSONPath = "docs/swagger.json"

var httpMethods = map[string]struct{}{
	"delete":  {},
	"get":     {},
	"head":    {},
	"options": {},
	"patch":   {},
	"post":    {},
	"put":     {},
	"trace":   {},
}

func loadLatestSwaggerDoc() (string, error) {
	data, err := os.ReadFile(filepath.Clean(swaggerJSONPath))
	if err == nil {
		return string(data), nil
	}

	doc := projectDocs.SwaggerInfo.ReadDoc()
	if strings.TrimSpace(doc) == "" {
		return "", fmt.Errorf("swagger doc is empty and %s cannot be read: %w", swaggerJSONPath, err)
	}

	return doc, nil
}

func buildApifoxSwaggerDoc(swaggerDoc string) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(swaggerDoc))
	decoder.UseNumber()

	var doc map[string]interface{}
	if err := decoder.Decode(&doc); err != nil {
		return "", fmt.Errorf("parse swagger doc: %w", err)
	}

	paths, ok := doc["paths"].(map[string]interface{})
	if !ok {
		return marshalSwaggerDoc(doc)
	}

	for _, pathValue := range paths {
		pathItem, ok := pathValue.(map[string]interface{})
		if !ok {
			continue
		}

		for method, operationValue := range pathItem {
			if _, ok := httpMethods[strings.ToLower(method)]; !ok {
				continue
			}

			operation, ok := operationValue.(map[string]interface{})
			if !ok {
				continue
			}

			tag := firstOperationTag(operation)
			folder := apifoxFolderFromTag(tag)
			if folder == "" {
				continue
			}

			operation["x-apifox-folder"] = folder
		}
	}

	return marshalSwaggerDoc(doc)
}

func firstOperationTag(operation map[string]interface{}) string {
	tags, ok := operation["tags"].([]interface{})
	if !ok || len(tags) == 0 {
		return ""
	}

	tag, ok := tags[0].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(tag)
}

func apifoxFolderFromTag(tag string) string {
	if tag == "" {
		return ""
	}

	if strings.Contains(tag, "/") {
		return tag
	}

	parts := strings.SplitN(tag, "-", 2)
	if len(parts) != 2 {
		return tag
	}

	parent := strings.TrimSpace(parts[0])
	child := strings.TrimSpace(parts[1])
	if parent == "" || child == "" {
		return tag
	}

	return parent + "/" + child
}

func marshalSwaggerDoc(doc map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(doc); err != nil {
		return "", fmt.Errorf("marshal swagger doc: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}
