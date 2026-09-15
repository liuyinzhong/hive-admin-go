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
	Apifox   ApifoxConfig   `json:"apifox"`
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
}

type AuditLogConfig struct {
	RetentionDays int `json:"retentionDays"`
	CleanupHour   int `json:"cleanupHour"`
}

// ApifoxConfig 接口文档同步配置；token 或 projectUrl 为空时启动跳过同步，不向外部发送数据
type ApifoxConfig struct {
	Token      string `json:"token"`
	ProjectURL string `json:"projectUrl"`
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
