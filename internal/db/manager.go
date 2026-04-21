package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var (
	ErrInvalidConfig   = errors.New("invalid database config")
	ErrManagerNotFound = errors.New("database manager not initialized, call Install() first")
)

type EngineGroup = xorm.EngineGroup

// DBManager 数据库管理器
type DBManager struct {
	mu            sync.RWMutex
	config        Config
	engineGroup   *xorm.EngineGroup
	logger        log.ContextLogger
	cancelFunc    context.CancelFunc
	healthCheckWg sync.WaitGroup
	closed        bool
}

// 全局实例管理
var (
	defaultManager *DBManager
	managers       = make(map[string]*DBManager)
	mu             sync.RWMutex
)

// Install 初始化数据库连接
func Install(cfg Config, name ...string) (*DBManager, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	mgr := &DBManager{config: cfg}

	masterEngine, err := createEngine(cfg, cfg.Master)
	if err != nil {
		return nil, fmt.Errorf("create master engine: %w", err)
	}

	slaveEngines := make([]*xorm.Engine, 0, len(cfg.Slaves))
	for _, link := range cfg.Slaves {
		if link == "" {
			continue
		}
		slave, err := createEngine(cfg, link)
		if err != nil {
			continue
		}
		slaveEngines = append(slaveEngines, slave)
	}

	eg, err := xorm.NewEngineGroup(masterEngine, slaveEngines, xorm.RoundRobinPolicy())
	if err != nil {
		return nil, fmt.Errorf("create engine group: %w", err)
	}

	mgr.engineGroup = eg

	if cfg.ShowSQL {
		eg.ShowSQL(true)
		eg.Logger().SetLevel(log.LOG_INFO)
	}

	ctx, cancel := context.WithCancel(context.Background())
	mgr.cancelFunc = cancel
	mgr.healthCheckWg.Add(1)
	go mgr.startHealthCheck(ctx)

	mu.Lock()
	defer mu.Unlock()
	if len(name) > 0 && name[0] != "" {
		managers[name[0]] = mgr
	}
	if defaultManager == nil {
		defaultManager = mgr
	}

	return mgr, nil
}

func validateConfig(cfg Config) error {
	if cfg.Driver == "" {
		return errors.New("driver is required")
	}
	if cfg.Master == "" {
		return errors.New("master is required")
	}
	if cfg.MaxIdleConns < 0 {
		return errors.New("max_idle_conns must be non-negative")
	}
	if cfg.MaxOpenConns < cfg.MaxIdleConns {
		return errors.New("max_open_conns must be >= max_idle_conns")
	}
	return nil
}

func createEngine(cfg Config, dsn string) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine(cfg.Driver, dsn)
	if err != nil {
		return nil, err
	}

	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetMaxOpenConns(cfg.MaxOpenConns)
	engine.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	timeout := cfg.ConnectTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := engine.PingContext(ctx); err != nil {
		_ = engine.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return engine, nil
}

// GetEngineGroup 获取 xorm EngineGroup 实例
func (m *DBManager) GetEngineGroup() *xorm.EngineGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engineGroup
}

// GetMaster 获取主库实例
func (m *DBManager) GetMaster() *xorm.Engine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.engineGroup == nil {
		return nil
	}
	return m.engineGroup.Master()
}

// Close 关闭所有连接
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	if m.cancelFunc != nil {
		m.cancelFunc()
		m.healthCheckWg.Wait()
	}

	if m.engineGroup != nil {
		err := m.engineGroup.Close()
		m.closed = true
		return err
	}

	m.closed = true
	return nil
}

// SetLogger 设置日志记录器
func (m *DBManager) SetLogger(logger log.ContextLogger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
	if m.engineGroup != nil {
		m.engineGroup.SetLogger(logger)
	}
}

// IsClosed 检查管理器是否已关闭
func (m *DBManager) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

func maskDSN(dsn string) string {
	if len(dsn) < 8 {
		return "******"
	}
	return dsn[:8] + "******"
}