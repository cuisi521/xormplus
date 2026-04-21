package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cuisi521/xormplus/internal/db"
	"xorm.io/xorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := db.DefaultConfig()
	cfg.Driver = "postgres"
	cfg.Master = "postgres://postgres:clm@2023@localhost:5433/test?sslmode=disable"
	cfg.Slaves = []string{
		"postgres://postgres:clm@2023@localhost:5433/test?sslmode=disable",
	}
	cfg.MaxIdleConns = 10
	cfg.MaxOpenConns = 100
	cfg.ConnMaxLifetime = time.Hour
	cfg.ShowSQL = true
	cfg.HealthCheckInterval = time.Minute

	mgr, err := db.Install(cfg)
	if err != nil {
		return fmt.Errorf("install database: %w", err)
	}
	defer func() {
		if err := mgr.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := createUser(); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	if err := queryUsers(); err != nil {
		return fmt.Errorf("query users: %w", err)
	}

	if err := testTransaction(); err != nil {
		return fmt.Errorf("test transaction: %w", err)
	}

	time.Sleep(time.Second)
	return nil
}

type User struct {
	Id   int64  `xorm:"pk autoincr"`
	Name string `xorm:"varchar(255)"`
}

func createUser() error {
	engine := db.GetDB()
	if engine == nil {
		return db.ErrManagerNotFound
	}

	_ = engine.Sync(new(User))

	user := &User{Name: "Cuisi"}
	affected, err := engine.Insert(user)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	fmt.Printf("插入 %d 条记录, ID: %d\n", affected, user.Id)
	return nil
}

func queryUsers() error {
	engine := db.GetDB()
	if engine == nil {
		return db.ErrManagerNotFound
	}

	var users []User
	if err := engine.Find(&users); err != nil {
		return fmt.Errorf("find: %w", err)
	}
	fmt.Printf("查询到 %d 条用户记录\n", len(users))
	return nil
}

func testTransaction() error {
	mgr, err := db.GetManager()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return mgr.WithTx(ctx, func(session *xorm.Session) error {
		user := &User{Name: "TxUser"}
		_, err := session.Insert(user)
		if err != nil {
			return fmt.Errorf("insert in tx: %w", err)
		}
		fmt.Printf("事务中插入用户, ID: %d\n", user.Id)
		return nil
	})
}