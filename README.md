# XormPlus 🚀

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![Xorm Version](https://img.shields.io/badge/Xorm-1.3.2-green.svg)](https://xorm.io)
[![License](https://img.shields.io/badge/License-MIT-brightgreen.svg)](LICENSE)

XormPlus 是一个基于 Xorm 的增强版 Go 语言数据库 ORM 封装，提供多实例管理、健康检查、连接池监控等企业级特性。

## ✨ 特性

- 🎯 **多数据库实例管理** - 支持多个主从数据库实例
- 🔄 **连接池管理** - 智能连接池配置和监控
- ❤️ **健康检查** - 自动健康检查和故障恢复
- 📊 **统计信息** - 详细的连接池统计信息
- 🛡️ **错误处理** - 完善的错误处理机制
- ⚡ **高性能** - 基于 Xorm 的高性能封装
- 🔧 **简单易用** - 简洁的 API 设计

## 🚀 快速开始

### 安装

```bash
go get github.com/cuisi521/xormplus
```

### 基本使用

```go
package main

import (
    "log"
    "time"

    "github.com/yourname/xormplus"
)

func main() {
    // 配置数据库
    config := xormplus.Config{
        Driver:          "mysql",
        Link:            []string{
            "user:pass@tcp(127.0.0.1:3306)/dbmaster?charset=utf8",
            "user:pass@tcp(127.0.0.1:3307)/dbslave?charset=utf8",
        },
        ShowSQL:         true,
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour * 2,
    }

    // 初始化默认数据库
    engine, err := xormplus.InitDefault(config)
    if err != nil {
        log.Fatal(err)
    }
    defer xormplus.CloseAll()

    // 执行查询
    results := make([]map[string]interface{}, 0)
    err = engine.Table("users").Find(&results)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 多实例管理

```go
// 初始化多个数据库实例
config1 := xormplus.Config{/* ... */}
config2 := xormplus.Config{/* ... */}

engine1, _ := xormplus.InitAndRegister("db1", config1)
engine2, _ := xormplus.InitAndRegister("db2", config2)

// 获取实例
db1, _ := xormplus.Get("db1")
db2, _ := xormplus.Get("db2")
```

### 健康检查

```go
// 单个实例健康检查
if err := engine.HealthCheck(); err != nil {
    log.Printf("Health check failed: %v", err)
}

// 所有实例健康检查
results := xormplus.HealthCheckAll()
for name, err := range results {
    if err != nil {
        log.Printf("Instance %s: %v", name, err)
    }
}
```

### 配置说明

```go
type Config struct {
    Driver          string        // 数据库驱动 (mysql, postgres, sqlite3)
    Link            []string      // 连接串，第一个为主库
    ShowSQL         bool          // 是否打印SQL
    LogLevel        int           // 日志级别
    ConnMaxLifetime time.Duration // 连接最大生命周期
    MaxIdleConns    int           // 最大空闲连接数
    MaxOpenConns    int           // 最大打开连接数
    ConnTimeout     time.Duration // 连接超时时间
}
```
### 🔧 支持的数据库
- MySQL
- PostgreSQL
- SQLite
- MSSQL
- 其他 Xorm 支持的数据库

### 📊 监控统计

```go
stats := engine.GetStats()
fmt.Printf("连接池统计: %+v\n", stats)
```

### 输出示例
```json
{
  "maxOpenConnections": 100,
  "openConnections": 5,
  "inUse": 2,
  "idle": 3,
  "waitCount": 0,
  "healthy": true
}
```

### 🤝 贡献
欢迎提交 Issue 和 Pull Request！

### 📄 许可证
本项目采用 MIT 许可证 - 查看 LICENSE 文件了解详情。

### 🙏 致谢
XORM - 优秀的 Go 语言 ORM 库


### LICENSE
```text
MIT License

Copyright (c) 2024 YourName

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
