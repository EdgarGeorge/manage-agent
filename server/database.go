package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	mlogger "gorm.io/gorm/logger"
)

const (
	DefaultHost          = "127.0.0.1"
	DefaultMysqlUsername = "root"
	DefaultMysqlPassword = "root"
	DefaultMySQLPort     = "3306"

	DefaultSchema = "manage-agent"

	DefaultSQLitePath = "./data/manage-agent.db"
)

var (
	Mysql *gorm.DB
)

// DatabaseType 表示数据库类型
type DatabaseType string

const (
	DBTypeMySQL  DatabaseType = "mysql"
	DBTypeSQLite DatabaseType = "sqlite"
)

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Schema   string `json:"schema"`
}

// SQLiteConfig SQLite 配置
type SQLiteConfig struct {
	Path string `json:"path"`
}

// DatabaseConfig 通用数据库配置
type DatabaseConfig struct {
	Type   DatabaseType  `json:"type"`
	MySQL  *MySQLConfig  `json:"mysql,omitempty"`
	SQLite *SQLiteConfig `json:"sqlite,omitempty"`
}

// NewDefaultMysqlConfig 保留给云端或 MySQL 场景使用
func NewDefaultMysqlConfig() *MySQLConfig {
	return &MySQLConfig{
		Username: DefaultMysqlUsername,
		Password: DefaultMysqlPassword,
		Host:     DefaultHost,
		Port:     DefaultMySQLPort,
		Schema:   DefaultSchema,
	}
}

// NewDefaultDatabaseConfig 默认使用本地 SQLite
func NewDefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type: DBTypeSQLite,
		SQLite: &SQLiteConfig{
			Path: DefaultSQLitePath,
		},
	}
}

// InitDatabase 根据配置初始化数据库连接
func InitDatabase(cfg *DatabaseConfig) error {
	if cfg == nil {
		cfg = NewDefaultDatabaseConfig()
	}

	switch cfg.Type {
	case DBTypeMySQL:
		if cfg.MySQL == nil {
			cfg.MySQL = NewDefaultMysqlConfig()
		}
		return MysqlConnect(*cfg.MySQL)
	case DBTypeSQLite, "":
		// 默认使用 SQLite
		return InitSQLite(cfg.SQLite)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}
}

// InitSQLite 初始化 SQLite 数据库（本地文件）
func InitSQLite(cfg *SQLiteConfig) error {
	path := DefaultSQLitePath
	if cfg != nil && cfg.Path != "" {
		path = cfg.Path
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	newLogger := mlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		mlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  mlogger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(sqlite.Open(path),
		&gorm.Config{
			Logger: newLogger,
		})
	if err != nil {
		return fmt.Errorf("打开 SQLite 数据库失败: %w", err)
	}

	Mysql = db
	return nil
}

// DBConnect 兼容旧接口，仍然连接 MySQL
func DBConnect(db MySQLConfig) error {
	if err := MysqlConnect(db); err != nil {
		return err
	}
	return nil
}

// MysqlConnect 初始化 MySQL 连接
func MysqlConnect(db MySQLConfig) error {
	// 初始化 mysql 连接
	dsn := strings.Join([]string{db.Username, ":", db.Password, "@tcp(", db.Host, ":", db.Port, ")/", db.Schema, "?charset=utf8&parseTime=True&loc=Asia%2FShanghai"}, "")
	err := NewMySQLPool(dsn)
	if err != nil {
		return err
	}
	return nil
}

// NewMySQLPool 创建 MySQL 连接池
func NewMySQLPool(dsn string) error {
	fmt.Println("Mysql DSN: ", dsn)

	newLogger := mlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		mlogger.Config{
			SlowThreshold:             time.Second,    // 慢 SQL 阈值
			LogLevel:                  mlogger.Silent, // 日志级别, Silent
			IgnoreRecordNotFoundError: true,           // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,           // 彩色打印
		},
	)
	db, err := gorm.Open(mysql.Open(dsn),
		&gorm.Config{
			Logger: newLogger,
		})

	if err != nil {
		return err
	}

	Mysql = db
	return nil
}
