package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"manage-agent/server"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var (
	configFile string
	logger     *logrus.Logger
)

// CloudConfig 云端服务配置
type CloudConfig struct {
	Database *server.DatabaseConfig `json:"database"`
	Server   struct {
		Addr string `json:"addr"` // 监听地址，例如 ":8080"
	} `json:"server"`
}

func init() {
	// 绑定 -c 参数到 configFile
	flag.StringVar(&configFile, "c", "", "config file")
}

// NewDefaultCloudConfig 默认云端配置（MySQL + 8080 端口）
func NewDefaultCloudConfig() *CloudConfig {
	cfg := &CloudConfig{
		Database: server.NewDefaultDatabaseConfig(),
	}
	// 云端默认使用 MySQL
	cfg.Database.Type = server.DBTypeMySQL
	cfg.Server.Addr = ":8080"
	return cfg
}

func LoadCloudConfigFromFile(filename string) (*CloudConfig, error) {
	config := NewDefaultCloudConfig()

	if filename != "" {
		file, err := os.Open(filepath.Clean(filename))
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

// InitLogging 初始化云端日志到 log 目录
func InitLogging() {
	logFile, err := rotatelogs.New(
		"../log/cloud-log-%Y%m%d.json",
		rotatelogs.WithLinkName("../log/cloud-latest.json"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		logrus.Fatalf("配置日志文件失败: %v", err)
	}

	logger = logrus.New()
	logger.SetOutput(logFile)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	logger.Info("cloud main init")

	server.SetLogger(logger)
}

// DBAutoMigrateCloud 执行数据表迁移
func DBAutoMigrateCloud() error {
	err := server.Mysql.AutoMigrate(
		&server.RequestRecordModel{},
		&server.WorthModel{},
		&server.FlowRecordModel{},
	)
	if err != nil {
		logrus.Fatalf("DB AutoMigrate fail. err:%v", err)
		return err
	}
	return nil
}

func errorRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"msg":  "url don't exist",
		"code": http.StatusNotFound,
	})
}

func main() {
	// 解析启动参数，例如: ./cloud -c ./cloud-config.json
	flag.Parse()

	// 加载配置文件
	config, err := LoadCloudConfigFromFile("./cloud-config.json")
	log.Printf("cloud config:%v", config)
	if err != nil {
		panic("load cloud config fail. err:" + err.Error())
	}

	// 初始化数据库（云端通常使用 MySQL）
	if err := server.InitDatabase(config.Database); err != nil {
		panic("init cloud database fail. err:" + err.Error())
	}

	// 建表
	if err := DBAutoMigrateCloud(); err != nil {
		panic("init cloud db fail. err:" + err.Error())
	}

	// 初始化日志
	InitLogging()

	// 启动 Gin（云端不需要加载本地 HTML 静态页面，仅提供 API）
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 中间件注册
	r.Use(server.RequestRecordMiddleware())
	r.Use(server.PanicRecoverMiddleware())

	// API 路由注册（与本地保持一致）
	api := r.Group("/api/v1")
	{
		server.RegisterFinanceRouter(api)
	}

	// 404 处理
	r.NoRoute(errorRouteHandler)

	// 启动 HTTP 服务器
	addr := config.Server.Addr
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    120 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic("run cloud server fail. err:" + err.Error())
	}
}
