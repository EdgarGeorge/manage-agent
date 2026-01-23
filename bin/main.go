package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
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

//go:embed html
var htmlFiles embed.FS

var (
	configFile string
	logger     *logrus.Logger
	exeDir     string // exe 文件所在目录
)

// getExeDir 获取可执行文件所在目录
func getExeDir() string {
	if exeDir != "" {
		return exeDir
	}

	// Android 环境：使用环境变量指定的数据目录
	if androidDataDir := os.Getenv("ANDROID_DATA_DIR"); androidDataDir != "" {
		exeDir = androidDataDir
		return exeDir
	}

	exe, err := os.Executable()
	if err != nil {
		// 如果获取失败，使用当前工作目录
		exeDir, _ = os.Getwd()
		return exeDir
	}
	exeDir = filepath.Dir(exe)
	return exeDir
}

// showError 显示错误信息并写入日志文件
func showError(title, message string) {
	// 写入错误日志文件
	errorLogPath := filepath.Join(getExeDir(), "error.log")
	if f, err := os.OpenFile(errorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		fmt.Fprintf(f, "[%s] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), title, message)
		f.Close()
	}
	// 同时输出到标准错误（如果有控制台）
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), title, message)
}

func InitLogging() {
	// 使用相对于 exe 目录的日志路径
	logDir := filepath.Join(getExeDir(), "log")
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		showError("日志目录创建失败", err.Error())
		logrus.Fatalf("配置日志文件失败: %v", err)
	}

	logFile, err := rotatelogs.New(
		filepath.Join(logDir, "log-%Y%m%d.json"),
		rotatelogs.WithLinkName(filepath.Join(logDir, "latest.json")),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		showError("日志文件配置失败", err.Error())
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
	Database *server.DatabaseConfig `json:"database"`
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
		Database: server.NewDefaultDatabaseConfig(),
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

	// 获取 exe 目录
	exeDir = getExeDir()

	// 加载配置文件（使用相对于 exe 目录的路径）
	configPath := filepath.Join(exeDir, "config.json")
	if configFile != "" {
		configPath = configFile
	}
	config, err := LoadConfigFromFile(configPath)
	log.Printf("config:%v", config)
	if err != nil {
		errorMsg := fmt.Sprintf("加载配置文件失败: %s\n配置文件路径: %s", err.Error(), configPath)
		showError("配置文件加载失败", errorMsg)
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			log.Printf("配置文件不存在，使用默认配置")
			config = NewDefaultConfig()
		} else {
			panic("load config fail. err:" + err.Error())
		}
	}

	// 如果配置中的路径是相对路径，转换为相对于 exe 目录的绝对路径
	if config.Database != nil && config.Database.SQLite != nil && config.Database.SQLite.Path != "" {
		if !filepath.IsAbs(config.Database.SQLite.Path) {
			config.Database.SQLite.Path = filepath.Join(exeDir, config.Database.SQLite.Path)
		}
	}

	// 初始化数据库（本地默认使用 SQLite）
	if err := server.InitDatabase(config.Database); err != nil {
		errorMsg := fmt.Sprintf("初始化数据库失败: %s", err.Error())
		showError("数据库初始化失败", errorMsg)
		panic("init database fail. err:" + err.Error())
	}

	if err := InitDB(config); err != nil {
		errorMsg := fmt.Sprintf("数据库建表失败: %s", err.Error())
		showError("数据库建表失败", errorMsg)
		panic("init db fail. err:" + err.Error())
	}

	// 初始化日志
	InitLogging()

	// 启动Server
	r := gin.Default()

	// 中间件注册
	r.Use(server.RequestRecordMiddleware())
	r.Use(server.PanicRecoverMiddleware())

	// 从嵌入的文件系统加载前端资源
	htmlFS, err := fs.Sub(htmlFiles, "html")
	if err != nil {
		panic("加载前端资源失败: " + err.Error())
	}

	// 加载 HTML 模板
	r.LoadHTMLFS(http.FS(htmlFS), "index.html")

	// 静态资源：assets 目录
	assetsFS, err := fs.Sub(htmlFS, "assets")
	if err != nil {
		panic("加载 assets 资源失败: " + err.Error())
	}
	r.StaticFS("/assets", http.FS(assetsFS))

	// 静态资源：api 目录
	apiFS, err := fs.Sub(htmlFS, "api")
	if err != nil {
		panic("加载 api 资源失败: " + err.Error())
	}
	r.StaticFS("/static/api", http.FS(apiFS))

	// 路由注册
	api := r.Group("/api/v1")
	{
		server.RegisterFinanceRouter(api)
	}

	// 根路径：渲染 index.html
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

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
