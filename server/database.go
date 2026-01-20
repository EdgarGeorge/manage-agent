package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
)

var (
	Mysql *gorm.DB
)

type MySQLConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Schema   string `json:"schema"`
}

func NewDefaultMysqlConfig() *MySQLConfig {
	return &MySQLConfig{
		Username: DefaultMysqlUsername,
		Password: DefaultMysqlPassword,
		Host:     DefaultHost,
		Port:     DefaultMySQLPort,
		Schema:   DefaultSchema,
	}
}

func DBConnect(db MySQLConfig) error {
	if err := MysqlConnect(db); err != nil {
		return err
	}
	return nil
}

func MysqlConnect(db MySQLConfig) error {
	//初始化mysql连接
	dsn := strings.Join([]string{db.Username, ":", db.Password, "@tcp(", db.Host, ":", db.Port, ")/", db.Schema, "?charset=utf8&parseTime=True&loc=Asia%2FShanghai"}, "")
	err := NewMySQLPool(dsn)
	if err != nil {
		return err
	}
	return nil
}

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
	// //设置连接的最长生命周期
	// db.SetConnMaxLifetime(time.Minute * 10)
	// //设置数据库最大闲置连接数
	// db.SetMaxIdleConns(100)
	// //设置最大打开的连接数
	// db.SetMaxOpenConns(100)

	// err = db.Ping()
	// if err != nil {
	// 	panic(err)
	// }
	Mysql = db
	return nil
}
