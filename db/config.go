// package db
// @Author cuisi
// @Date 2024/11/25
// @Desc 数据库配置结构定义

package db

import (
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	// _ "github.com/go-sql-driver/mysql" // MySQL driver (uncomment if needed)
)

// ============================================================================
// Configuration
// ============================================================================

// Config 数据库配置结构体
type Config struct {
	Driver          string        `json:"driver" yaml:"driver"`                       // 驱动类型: postgres, mysql, sqlite3, mssql
	Master          string        `json:"master" yaml:"master"`                       // 主库连接字符串
	Slaves          []string      `json:"slaves" yaml:"slaves"`                       // 从库连接字符串列表
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`       // 最大空闲连接数
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`       // 最大打开连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"` // 连接最大存活时间
	ShowSQL         bool          `json:"show_sql" yaml:"show_sql"`                   // 是否打印SQL
	LogLevel        string        `json:"log_level" yaml:"log_level"`                 // 日志级别: debug, info, warn, error
	DisableCache    bool          `json:"disable_cache" yaml:"disable_cache"`         // 是否禁用xorm缓存
}

// ============================================================================
// Core Logic
// ============================================================================
