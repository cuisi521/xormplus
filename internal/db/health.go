package db

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type healthChecker struct {
	manager     *DBManager
	interval    time.Duration
	consecutive int
	threshold   int
	stopped     atomic.Bool
}

func newHealthChecker(mgr *DBManager, interval time.Duration) *healthChecker {
	if interval <= 0 {
		interval = time.Minute
	}
	return &healthChecker{
		manager:   mgr,
		interval:  interval,
		threshold: 3,
	}
}

func (hc *healthChecker) start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.check()
		case <-ctx.Done():
			hc.stopped.Store(true)
			return
		}
	}
}

func (hc *healthChecker) check() {
	master := hc.manager.GetMaster()
	if master == nil {
		hc.consecutive++
		hc.logFailure("master engine is nil")
		return
	}

	if err := master.Ping(); err != nil {
		hc.consecutive++
		hc.logFailure(err.Error())
	} else {
		hc.consecutive = 0
	}
}

func (hc *healthChecker) logFailure(msg string) {
	if hc.consecutive >= hc.threshold {
		fmt.Printf("[WARN] Database health check failed %d consecutive times: %s\n", hc.consecutive, msg)
	}
}

func (m *DBManager) startHealthCheck(ctx context.Context) {
	defer m.healthCheckWg.Done()

	interval := m.config.HealthCheckInterval
	if interval <= 0 {
		interval = time.Minute
	}

	hc := newHealthChecker(m, interval)
	hc.start(ctx)
}