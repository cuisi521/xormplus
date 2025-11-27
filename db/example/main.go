package main

import (
	"fmt"
	"time"

	"github.com/cuisi521/zap-wrapper/logger"

	"xormplus/db"
)

func main() {
	_, err := initLogger()
	if err != nil {
		panic(err)
	}
	logger.Info("数据库初始化开始...")
	c := db.Config{
		Driver: "postgres", // or "mysql"
		Master: "postgres://postgres:clm@2023@localhost:5433/test?sslmode=disable",
		Slaves: []string{
			"postgres://postgres:clm@2023@localhost:5433/test?sslmode=disable",
		},
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ShowSQL:         true,
	}

	// 初始化管理器
	manager, err := db.Install(c)
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	logger.Info("数据库初始化成功")

	// 创建用户
	CreateUser()
	// 查询用户
	QueryUser()
}

type User struct {
	Id   int64
	Name string
}

func CreateUser() {
	engine := db.GetDB() // 获取 EngineGroup
	engine.Sync(new(User))
	user := &User{Name: "Cuisi"}
	affected, err := engine.Insert(user)
	// ...
	if err != nil {
		panic(err)
	}
	fmt.Printf("插入 %d 条记录\n", affected)
}

// 查询用户
func QueryUser() {
	engine := db.GetDB() // 获取 EngineGroup
	var users []User
	err := engine.Find(&users)
	if err != nil {
		panic(err)
	}
	fmt.Printf("查询到 %d 条用户记录\n", len(users))
}

func initLogger() (*logger.Logger, error) {
	// 创建日志器，同时会自动设置为全局日志器
	log, err := logger.New(
		// logger.WithAsyncMode(true), // 启用异步
		logger.WithLevel(logger.DebugLevel),
		logger.WithEncoding(logger.ConsoleEncoding),
		logger.WithBasePath("logs"),
		logger.WithConsoleOutput(true), // 启用控制台输出
		logger.WithCaller(true),        // 显示调用者信息
		logger.WithStacktrace(true),    // 显示堆栈跟踪

	)

	if err != nil {
		fmt.Printf("创建日志器失败: %v\n", err)
		return log, err
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Sync失败: %v\n", err)
		}
	}()

	return log, nil
}
