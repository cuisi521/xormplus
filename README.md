
```markdown
# Xorm Store Adapter

这是一个基于 [xorm](https://xorm.io/) 的数据库操作封装库。旨在提供开箱即用的多数据库支持、主从读写分离、健壮的连接池管理以及事务处理能力。

## ✨ 特性

- **多数据库支持**: 兼容 MySQL, PostgreSQL, SQLite, SQL Server 等主流数据库。
- **读写分离**: 内置主从架构支持，自动将读操作路由到从库，写操作路由到主库。
- **智能连接池**: 预配置的连接池参数，支持最大空闲、最大打开连接及生命周期管理。
- **事务安全**: `WithTx` 闭包式事务封装，自动处理 Commit/Rollback，防止 panic 导致的死锁。
- **健康监控**: 后台协程自动监控数据库健康状态。
- **高扩展性**: 支持注入自定义 Logger (如 zap)。

## 📦 安装

```bash
go get github.com/yourusername/project/pkg/db
```

## 🚀 快速开始

### 1. 初始化配置

```go
package main

import (
    "time"
    "github.com/yourusername/project/pkg/db"
    "github.com/zeromicro/go-zero/core/logx"
)

func main() {
    c := db.Config{
        Driver:          "postgres", // or "mysql"
        Master:          "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
        Slaves:          []string{
            "postgres://user:pass@localhost:5433/mydb?sslmode=disable",
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
    
    logx.Info("数据库初始化成功")
}
```

### 2. 基础 CRUD 操作

```go
type User struct {
    Id   int64
    Name string
}

func CreateUser() {
    engine := db.GetDB() // 获取 EngineGroup
    
    user := &User{Name: "Cuisi"}
    affected, err := engine.Insert(user)
    // ...
}
```

### 3. 使用事务 (WithTx)

```go
func TransferMoney(ctx context.Context) error {
    manager := db.defaultManager // 或通过依赖注入获取

    return manager.WithTx(ctx, func(session *xorm.Session) error {
        // 1. 扣款
        if _, err := session.Exec("UPDATE account SET balance = balance - 100 WHERE id = 1"); err != nil {
            return err // 自动 Rollback
        }

        // 2. 加款
        if _, err := session.Exec("UPDATE account SET balance = balance + 100 WHERE id = 2"); err != nil {
            return err // 自动 Rollback
        }

        return nil // 自动 Commit
    })
}
```

### 4. 集成 Zap 日志 (可选)

如果您使用 `github.com/cuisi521/zap-wrapper`，可以实现 `xorm.io/xorm/log.ContextLogger` 接口并注入：

```go
// 伪代码示例
zapLogger := NewXormZapAdapter(zap.L())
manager.SetLogger(zapLogger)
```

## ⚙️ 配置说明

| 字段 | 类型 | 说明 |
|Data Type|Description|
|---|---|---|
| `Driver` | string | 数据库驱动名称 (postgres, mysql, sqlite3, mssql) |
| `Master` | string | 主库 DSN 连接字符串 |
| `Slaves` | []string | 从库 DSN 列表 |
| `MaxIdleConns` | int | 连接池最大空闲连接数 |
| `MaxOpenConns` | int | 连接池最大打开连接数 |
| `ConnMaxLifetime` | duration | 连接最大存活时间 |
| `DisableCache` | bool | 是否禁用 xorm 内置缓存 (建议为 true) |

## ⚠️ 注意事项

1. **驱动引入**: 本库默认引入了 `lib/pq`。如果使用 MySQL，请在您的 main 文件或此包中取消注释 `_ "github.com/go-sql-driver/mysql"`。
2. **缓存策略**: 默认建议在业务层（如 Redis）处理缓存，因此配置中提供了 `DisableCache` 选项。

## 📄 License

MIT
```

### 关键改动解释

1.  **移除全局 Map (`dbEngine`)**: 旧代码使用全局 map 存储，这在测试和多实例场景下很难维护。新代码通过 `Install` 返回一个 `DBManager` 实例，同时也保留了一个可选的 `defaultManager` 以兼容旧的使用习惯。
2.  **增强的 `Install`**:
    *   明确区分了 `Master` 和 `Slaves` 的配置，而不是将所有链接混在一起通过逻辑判断。
    *   增加了 `Ping` 检查，确保服务启动时数据库是可用的。
3.  **`WithTx` 事务封装**: 这是一个非常实用的模式。它利用 Go 的闭包特性，消除了到处写 `session.Begin()`, `session.Commit()`, `defer session.Close()` 的样板代码，并且安全地处理了 panic。
4.  **健康检查**: 增加了一个后台 goroutine 定期 ping 数据库。虽然 xorm 内部有保活机制，但应用层的健康检查对于对接 Prometheus 或 K8s 探针非常有用。
5.  **日志脱敏**: 增加了 `maskDSN` 函数，防止在日志中明文打印数据库密码。