package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"manage-agent/server"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

var (
	configFile string
	logger     *logrus.Logger
)

func InitLogging() {
	logFile, err := rotatelogs.New(
		"../log/log-%Y%m%d.json",
		rotatelogs.WithLinkName("../log/latest.json"),
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
	logger.Info("main init")

	server.SetLogger(logger)

}

func errorRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"msg":  "url don't exist",
		"code": http.StatusNotFound,
	})
}

func DBAutoMigrate() error {
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

type Config struct {
	Mysql *server.MySQLConfig `json:"mysql"`
}

func init() {
	// 绑定-c参数到configFile
	flag.StringVar(&configFile, "c", "", "config file")
}

func InitDB(config *Config) error {
	// DB建表
	err := DBAutoMigrate()
	if err != nil {
		return err
	}

	return nil
}

func NewDefaultConfig() *Config {
	return &Config{
		Mysql: server.NewDefaultMysqlConfig(),
	}
}

func LoadConfigFromFile(filename string) (*Config, error) {
	config := NewDefaultConfig()

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

func main() {

	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	// 解析启动参数, eg. .\main -c .\config.json
	flag.Parse()

	// 加载配置文件
	config, err := LoadConfigFromFile("./config.json")
	log.Printf("config:%v", config)
	if err != nil {
		panic("load config fail. err:" + err.Error())
	}

	// 初始化DB
	if err := server.DBConnect(*config.Mysql); err != nil {
		panic("connect db fail. err:" + err.Error())
	}

	if err := InitDB(config); err != nil {
		panic("init db fail. err:" + err.Error())
	}

	// 初始化日志
	InitLogging()

	// 启动Server
	r := gin.Default()

	// 中间件注册
	r.Use(server.RequestRecordMiddleware())
	r.Use(server.PanicRecoverMiddleware())

	// 加载 HTML 模板文件
	r.LoadHTMLGlob("html/index.html")
	r.Static("/assets", "./html/assets")
	// 前端 API 脚本改用独立前缀，避免与后端 /api 路由冲突
	r.Static("/static/api", "./html/api")

	// 路由注册
	api := r.Group("/api/v1")
	{
		server.RegisterFinanceRouter(api)

	}

	// 托管静态文件（index.html放在项目根目录文件夹下）
	r.StaticFile("/", "html/index.html")
	// 如果有其他静态资源（如js/css），需同步托管
	// r.Static("/static", "./static")

	// 程序启动后自动打开浏览器
	go func() {
		url := "http://localhost:8090"
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		case "darwin": // mac
			cmd = exec.Command("open", url)
		case "linux": // linux
			cmd = exec.Command("xdg-open", url)
		}
		if cmd != nil {
			_ = cmd.Start()
		}
	}()

	// 内置路由
	r.NoRoute(errorRouteHandler)

	// 启动Server
	server := &http.Server{
		Addr:           ":8090",
		Handler:        r,
		ReadTimeout:    120 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic("run server fail. err:" + err.Error())
	}
}
