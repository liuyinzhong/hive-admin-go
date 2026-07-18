package config

import (
	"encoding/json"
	"os"
)

var AppConfig *Config

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	AuditLog AuditLogConfig `json:"auditLog"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Mode string `json:"mode"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
	Charset  string `json:"charset"`
}

type JWTConfig struct {
	Secret string `json:"secret"`
	Expire int    `json:"expire"`
}

type AuditLogConfig struct {
	RetentionDays int `json:"retentionDays"`
	CleanupHour   int `json:"cleanupHour"`
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	AppConfig = &Config{}
	if err := json.Unmarshal(data, AppConfig); err != nil {
		return err
	}
	if AppConfig.AuditLog.RetentionDays <= 0 {
		AppConfig.AuditLog.RetentionDays = 180
	}
	if AppConfig.AuditLog.CleanupHour <= 0 || AppConfig.AuditLog.CleanupHour > 23 {
		AppConfig.AuditLog.CleanupHour = 3
	}
	return nil
}
