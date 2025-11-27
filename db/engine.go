// package db
// @Author cuisi
// @Date 2024/11/25
// @Desc 数据库引擎创建和工具函数
package db

import (
	"context"
	"fmt"
	"time"

	"xorm.io/xorm"
)

// createEngine 创建单个数据库引擎实例
func createEngine(c Config, dsn string) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine(c.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// 连接池配置
	engine.SetMaxIdleConns(c.MaxIdleConns)
	engine.SetMaxOpenConns(c.MaxOpenConns)
	engine.SetConnMaxLifetime(c.ConnMaxLifetime)

	// 禁用默认缓存 (通常我们在 Service 层做缓存，ORM 层缓存容易导致一致性问题)
	if c.DisableCache {
		engine.SetDefaultCacher(nil)
	}

	// 初始 Ping 测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.PingContext(ctx); err != nil {
		// 这里即使 Ping 失败也返回 engine，以便后续自动重连机制生效？
		// 还是严格模式直接报错？这里选择严格模式，启动时必须连通。
		engine.Close()
		return nil, fmt.Errorf("ping database failed: %w", err)
	}

	return engine, nil
}

// maskDSN 简单的脱敏处理，用于日志
func maskDSN(dsn string) string {
	if len(dsn) < 8 {
		return "******"
	}
	return dsn[:8] + "******"
}
