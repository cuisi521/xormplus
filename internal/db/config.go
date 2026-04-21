// Package db 提供数据库连接池管理、主从读写分离、事务封装等核心功能。
package db

import (
	"time"

	_ "github.com/lib/pq"
)

// Config 数据库配置结构体
type Config struct {
	Driver          string        `json:"driver" yaml:"driver"`
	Master          string        `json:"master" yaml:"master"`
	Slaves          []string      `json:"slaves" yaml:"slaves"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ShowSQL         bool          `json:"show_sql" yaml:"show_sql"`
	// 健康检查配置
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	// 连接超时配置
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxIdleConns:        10,
		MaxOpenConns:        100,
		ConnMaxLifetime:     time.Hour,
		HealthCheckInterval: time.Minute,
		ConnectTimeout:      5 * time.Second,
		ShowSQL:             false,
	}
}