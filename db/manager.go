// package db
// @Author cuisi
// @Date 2024/11/25
// @Desc 数据库管理器核心结构和方法
package db

import (
	"context"
	"fmt"
	"sync"

	"xorm.io/xorm"
	"xorm.io/xorm/log"

	"github.com/cuisi521/zap-wrapper/logger"
)

// ============================================================================
// Manager & Interfaces
// ============================================================================

// DBManager 数据库管理器
type DBManager struct {
	engineGroup   *xorm.EngineGroup
	config        Config
	mu            sync.RWMutex
	logger        log.ContextLogger  // 预留给 zap-wrapper 的接口
	cancelFunc    context.CancelFunc // 用于取消 startHealthCheck goroutine
	healthCheckWg sync.WaitGroup     // 用于等待 health check goroutine 退出
}

// 全局实例（可选，推荐使用依赖注入）
var (
	defaultManager *DBManager
	managers       = make(map[string]*DBManager)
	mu             sync.RWMutex
)

// Install 初始化数据库连接
func Install(c Config, name ...string) (*DBManager, error) {
	mgr := &DBManager{
		config: c,
	}

	// 1. 构建主库
	masterEngine, err := createEngine(c, c.Master)
	if err != nil {
		return nil, fmt.Errorf("create master engine failed: %w", err)
	}

	// 2. 构建从库
	var slaveEngines []*xorm.Engine
	for _, link := range c.Slaves {
		if link == "" {
			continue
		}
		slave, err := createEngine(c, link)
		if err != nil {
			// 从库连接失败记录日志，但不阻断主库启动，除非这是必须的
			// logx.Errorf("create slave engine failed (link: %s): %v", maskDSN(link), err)
			continue
		}
		slaveEngines = append(slaveEngines, slave)
	}

	// 3. 创建 EngineGroup (读写分离策略：轮询)
	eg, err := xorm.NewEngineGroup(masterEngine, slaveEngines, xorm.RoundRobinPolicy())
	if err != nil {
		return nil, fmt.Errorf("create engine group failed: %w", err)
	}

	mgr.engineGroup = eg

	// 4. 设置日志 (这里可以集成 zap-wrapper)
	// mgr.SetLogger(NewZapLogger()) // 如果有 zap 实现
	if c.ShowSQL {
		eg.ShowSQL(true)
		eg.Logger().SetLevel(log.LOG_INFO)
	}

	ctx, cancel := context.WithCancel(context.Background())
	mgr.cancelFunc = cancel
	mgr.healthCheckWg.Add(1)

	// 5. 启动健康检查协程 (可选)
	go mgr.startHealthCheck(ctx)

	// 6. 存储实例
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

// GetEngineGroup 获取 xorm EngineGroup 实例
func (m *DBManager) GetEngineGroup() *xorm.EngineGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engineGroup
}

// GetMaster 获取主库实例
func (m *DBManager) GetMaster() *xorm.Engine {
	return m.engineGroup.Master()
}

// Close 关闭所有连接
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancelFunc != nil {
		m.cancelFunc()                                 // 调用 cancelFunc 停止 goroutine
		m.healthCheckWg.Wait()                         // 等待 goroutine 真正退出
		logger.Info("Health check goroutine stopped.") // 示例日志
	}
	if m.engineGroup != nil {
		logger.Info("Engine group closed.") // 示例日志
		return m.engineGroup.Close()
	}
	return nil
}

// SetLogger 允许注入自定义 Logger (如 zap-wrapper)
func (m *DBManager) SetLogger(logger log.ContextLogger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
	if m.engineGroup != nil {
		m.engineGroup.SetLogger(logger)
	}
}
