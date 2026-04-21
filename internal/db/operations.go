package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"xorm.io/xorm"
)

var (
	ErrNilSession     = errors.New("session is nil")
	ErrNilEngineGroup = errors.New("engine group is nil")
)

// WithTx 事务包装器
// 自动处理 Commit 和 Rollback，panic 时自动 Rollback
func (m *DBManager) WithTx(ctx context.Context, fn func(session *xorm.Session) error) error {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return ErrNilEngineGroup
	}

	session := eg.NewSession()
	if session == nil {
		return ErrNilSession
	}
	defer session.Close()

	if ctx != nil {
		_ = session.Context(ctx)
	}

	if err := session.Begin(); err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	var txErr error
	func() {
		defer func() {
			if p := recover(); p != nil {
				_ = session.Rollback()
				panic(p)
			}
			if txErr != nil {
				_ = session.Rollback()
			} else {
				txErr = session.Commit()
			}
		}()
		txErr = fn(session)
	}()

	return txErr
}

// BatchInsert 批量插入
func (m *DBManager) BatchInsert(bean any, batchSize int) (int64, error) {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return 0, ErrNilEngineGroup
	}
	return eg.Insert(bean)
}

// Get 获取单条记录
func (m *DBManager) Get(bean any, conditions ...any) (bool, error) {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return false, ErrNilEngineGroup
	}

	if len(conditions) == 0 {
		return eg.Get(bean)
	}
	return eg.Where(conditions[0], conditions[1:]...).Get(bean)
}

// Find 获取多条记录
func (m *DBManager) Find(beans any, conditions ...any) error {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return ErrNilEngineGroup
	}

	if len(conditions) == 0 {
		return eg.Find(beans)
	}
	return eg.Where(conditions[0], conditions[1:]...).Find(beans)
}

// Count 统计记录数量
func (m *DBManager) Count(bean any, conditions ...any) (int64, error) {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return 0, ErrNilEngineGroup
	}

	if len(conditions) == 0 {
		return eg.Count(bean)
	}
	return eg.Where(conditions[0], conditions[1:]...).Count(bean)
}

// Iterate 迭代查询大量数据
func (m *DBManager) Iterate(bean any, conditions any, iterFunc func(int, any) error) error {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return ErrNilEngineGroup
	}
	return eg.Where(conditions).Iterate(bean, iterFunc)
}

// Delete 删除记录
func (m *DBManager) Delete(bean any, conditions ...any) (int64, error) {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return 0, ErrNilEngineGroup
	}

	if len(conditions) == 0 {
		return eg.Delete(bean)
	}
	return eg.Where(conditions[0], conditions[1:]...).Delete(bean)
}

// DeleteByID 根据ID删除记录
func (m *DBManager) DeleteByID(bean any, id any) (int64, error) {
	m.mu.RLock()
	eg := m.engineGroup
	m.mu.RUnlock()

	if eg == nil {
		return 0, ErrNilEngineGroup
	}
	return eg.ID(id).Delete(bean)
}

// DB 返回 EngineGroup
func (m *DBManager) DB() *xorm.EngineGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engineGroup
}

// GetDB 获取全局默认 DB 实例
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
		return nil
	}
	return mgr.GetEngineGroup()
}

// DBM 获取全局默认 DBManager 实例
func DBM(name ...string) *DBManager {
	mu.RLock()
	defer mu.RUnlock()

	var mgr *DBManager
	if len(name) > 0 && name[0] != "" {
		mgr = managers[name[0]]
	} else {
		mgr = defaultManager
	}

	return mgr
}

// GetManager 获取 DBManager，带错误返回
func GetManager(name ...string) (*DBManager, error) {
	mu.RLock()
	defer mu.RUnlock()

	var mgr *DBManager
	if len(name) > 0 && name[0] != "" {
		mgr = managers[name[0]]
	} else {
		mgr = defaultManager
	}

	if mgr == nil {
		return nil, ErrManagerNotFound
	}
	return mgr, nil
}

// CloseAll 关闭所有数据库管理器
func CloseAll() error {
	mu.Lock()
	defer mu.Unlock()

	var errs []string

	if defaultManager != nil {
		if err := defaultManager.Close(); err != nil {
			errs = append(errs, err.Error())
		}
		defaultManager = nil
	}

	for name, mgr := range managers {
		if err := mgr.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
		delete(managers, name)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}