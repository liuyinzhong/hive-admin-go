package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hive-admin-go/config"
	"hive-admin-go/database"
	projectDocs "hive-admin-go/docs"
	"hive-admin-go/router"
	"hive-admin-go/utils"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Hive Admin API
// @version 1.0
// @description Hive Admin Go 后端 API 接口文档
// @host localhost:9191
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

//go:generate go install github.com/swaggo/swag/cmd/swag@latest

func main() {
	// 自动生成 Swagger 文档（如果需要）
	autoGenerateSwagger()

	// 启动服务
	startServer()
}

func startServer() {
	if err := config.LoadConfig("config.json"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	utils.InitJWT()

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置 GIN 为 release 模式，减少日志输出
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard // 禁用 GIN 默认日志输出

	r := router.SetupRouter()

	// 添加 Swagger 路由
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/doc.json" {
			doc := projectDocs.SwaggerInfo.ReadDoc()

			var buf bytes.Buffer
			if err := json.Compact(&buf, []byte(doc)); err != nil {
				c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(doc))
				return
			}

			c.Data(http.StatusOK, "application/json; charset=utf-8", buf.Bytes())
			return
		}

		swaggerHandler(c)
	})

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("✅ 服务启动成功，端口: %d", config.AppConfig.Server.Port)
	log.Printf("📖 Swagger UI: http://localhost:%d/swagger/index.html", config.AppConfig.Server.Port)
	log.Printf("📡 API Base URL: http://localhost:%d/api", config.AppConfig.Server.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// syncToApify 同步接口文档到 Apifox
func syncToApify() {
	swaggerDoc := projectDocs.SwaggerInfo.ReadDoc()

	options := map[string]interface{}{
		"endpointOverwriteBehavior":     "OVERWRITE_EXISTING",
		"schemaOverwriteBehavior":       "OVERWRITE_EXISTING",
		"updateFolderOfChangedEndpoint": true,
		"prependBasePath":               false,
		"deleteUnmatchedResources":      true,
	}

	payloadMap := map[string]interface{}{
		"input":   swaggerDoc,
		"options": options,
	}

	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		log.Printf("❌ 构建 Apifox 请求体失败: %v", err)
		return
	}

	url := "https://api.apifox.com/v1/projects/8280529/import-openapi?locale=zh-CN"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Printf("❌ 创建 Apifox 请求失败: %v", err)
		return
	}

	req.Header.Add("X-Apifox-Api-Version", "2024-03-28")
	req.Header.Add("Authorization", "Bearer afxp_1bf9abbkG6NB0mOwxgkxQRKoYiFvMPpbnC9A")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Printf("❌ 调用 Apifox API 失败: %v", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("❌ 读取 Apifox 响应失败: %v", err)
		return
	}

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		log.Println("✅ 接口文档已成功同步到 Apifox!")
	} else {
		log.Printf("⚠️ 同步到 Apifox 失败，状态码: %d, 响应: %s", res.StatusCode, string(body))
	}
}

// autoGenerateSwagger 自动生成 Swagger 文档
func autoGenerateSwagger() {
	swaggerDir := "docs"
	docsFile := filepath.Join(swaggerDir, "docs.go")

	_, err := os.Stat(docsFile)
	if os.IsNotExist(err) {
		log.Println("🔄 Swagger 文档不存在，正在自动生成...")
		if success := generateSwagger(); success {
			log.Println("✅ Swagger 文档自动生成成功!")
		}
		return
	}

	controllersModified, err := getLatestModifiedTime("controllers")
	if err == nil {
		docsModified, err := os.Stat(docsFile)
		if err == nil {
			if controllersModified.After(docsModified.ModTime()) {
				log.Println("🔄 检测到 controllers 有更新，正在重新生成 Swagger 文档...")
				if success := generateSwagger(); success {
					log.Println("✅ Swagger 文档已自动更新!")
				}
				return
			}
		}
	}
	// 在启动服务之前，先尝试同步到 Apifox
	go syncToApify()
	log.Println("✅ Swagger 文档已是最新版本，跳过生成")
}

func generateSwagger() bool {
	commands := [][]string{
		{"swag", "init"},
		{"C:\\Users\\Admin\\go\\bin\\swag.exe", "init"},
	}

	var lastErr error
	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = "."
		output, err := c.CombinedOutput()
		if err != nil {
			lastErr = err
			continue
		}
		log.Println(strings.TrimSpace(string(output)))
		return true
	}

	log.Printf("❌ 自动生成 Swagger 文档失败: %v", lastErr)
	log.Println("💡 提示: 请手动运行 'swag init' 或 '& \"C:\\Users\\Admin\\go\\bin\\swag.exe\" init'")
	return false
}

func getLatestModifiedTime(dir string) (time.Time, error) {
	var latest time.Time
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}
		return nil
	})
	return latest, err
}
