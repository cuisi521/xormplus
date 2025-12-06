// package db
// @Author cuisi
// @Date 2024/11/25
// @Desc 数据库操作和高级功能

package db

import (
	"context"
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/cuisi521/zap-wrapper/logger"
)

// WithTx 事务包装器
// 自动处理 Commit 和 Rollback，panic 时自动 Rollback
func (m *DBManager) WithTx(ctx context.Context, fn func(session *xorm.Session) error) (err error) {
	session := m.engineGroup.NewSession()
	if ctx != nil {
		session.Context(ctx)
	}
	defer session.Close()

	if err = session.Begin(); err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = session.Rollback()
			panic(p) // 重新抛出 panic
		} else if err != nil {
			_ = session.Rollback()
		} else {
			err = session.Commit()
		}
	}()

	err = fn(session)
	return err
}

// BatchInsert 批量插入工具
func (m *DBManager) BatchInsert(bean interface{}, batchSize int) (int64, error) {
	// xorm 自带 Insert 支持切片，但为了控制大批量数据的内存，这里可以做分批处理的封装
	// 此处直接利用 xorm 的能力，预留接口供扩展
	return m.engineGroup.Insert(bean)
}

// startHealthCheck 简单的健康检查与日志告警
func (m *DBManager) startHealthCheck(ctx context.Context) {
	defer m.healthCheckWg.Done()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.check()
		case <-ctx.Done():
			// context 被取消，退出循环
			logger.Info("Health check goroutine stopped.")
			return
		}
	}

}

func (m *DBManager) check() {

	master := m.GetMaster()
	if master != nil {
		err := master.Ping()
		if err != nil {
			fmt.Printf("Master DB health check failed: %v\n", err)
			// 这里可以调用 m.logger.Error(...)
		} else {
			// fmt.Println("Master DB is healthy.")
		}
	} else {
		fmt.Println("Master DB engine is nil during health check.")
	}

	// 检查从库 (如果需要更细粒度的控制，需要遍历 Slaves)
	// xorm 的 EngineGroup Ping 会根据策略检查，这里简单检查 Master 即可代表连通性
}

// Get 获取单条记录
func (m *DBManager) Get(bean interface{}, conditions ...interface{}) (bool, error) {
	if len(conditions) == 0 {
		return m.engineGroup.Get(bean)
	}
	return m.engineGroup.Where(conditions[0], conditions[1:]...).Get(bean)
}

// Find 获取多条记录
func (m *DBManager) Find(beans interface{}, conditions ...interface{}) error {
	if len(conditions) == 0 {
		return m.engineGroup.Find(beans)
	}
	return m.engineGroup.Where(conditions[0], conditions[1:]...).Find(beans)
}

// Count 统计记录数量
func (m *DBManager) Count(bean interface{}, conditions ...interface{}) (int64, error) {
	if len(conditions) == 0 {
		return m.engineGroup.Count(bean)
	}
	return m.engineGroup.Where(conditions[0], conditions[1:]...).Count(bean)
}

// Iterate 迭代查询大量数据
func (m *DBManager) Iterate(bean interface{}, conditions interface{}, iterator func(idx int, bean interface{}) error) error {
	return m.engineGroup.Where(conditions).Iterate(bean, iterator)
}

// Delete 删除记录
func (m *DBManager) Delete(bean interface{}, conditions ...interface{}) (int64, error) {
	if len(conditions) == 0 {
		return m.engineGroup.Delete(bean)
	}
	return m.engineGroup.Where(conditions[0], conditions[1:]...).Delete(bean)
}

// DeleteByID 根据ID删除记录
func (m *DBManager) DeleteByID(bean interface{}, id interface{}) (int64, error) {
	return m.engineGroup.ID(id).Delete(bean)
}

// GetDB 获取全局默认 DB 实例，支持指定数据库名称
func GetDB(name ...string) *xorm.EngineGroup {
	mu.RLock()
	defer mu.RUnlock()

	var mgr *DBManager
	if len(name) > 0 && name[0] != "" {
		mgr = managers[name[0]]
	} else {
		mgr = defaultManager
	}

	if mgr == nil {
		logger.Error("Database manager not initialized, call Install() first")
		return nil
	}
	return mgr.GetEngineGroup()
}

// GetManager 获取全局默认 DBManager 实例，支持指定数据库名称
func DBM(name ...string) *DBManager {
	mu.RLock()
	defer mu.RUnlock()

	var mgr *DBManager
	if len(name) > 0 && name[0] != "" {
		mgr = managers[name[0]]
	} else {
		mgr = defaultManager
	}

	if mgr == nil {
		logger.Error("Database manager not initialized, call Install() first")
		return nil
	}
	return mgr
}
