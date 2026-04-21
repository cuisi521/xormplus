# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xormplus is a Go library wrapping [xorm](https://xorm.io/) to provide multi-database support, master-slave read-write separation, connection pool management, and transaction safety.

## Common Commands

```bash
# Build the project
go build ./...

# Run example
go run db/example/main.go

# Tidy dependencies
go mod tidy
```

## Architecture

The library has 4 main files in the `db/` package:

- **config.go** - `Config` struct defining database connection parameters (Driver, Master, Slaves, connection pool settings)
- **engine.go** - `DBManager` struct, `Install()` function, and instance management (global defaultManager and named managers map)
- **operations.go** - CRUD operations (`Get`, `Find`, `Count`, `Delete`), `WithTx` transaction wrapper, `BatchInsert`, health check logic
- **example/main.go** - Usage demonstration

### Key Concepts

- **EngineGroup**: xorm's built-in support for master-slave replication with round-robin policy for read operations
- **WithTx**: Closure-based transaction wrapper that handles Begin/Commit/Rollback automatically, including panic recovery
- **Health Check**: Background goroutine pings master database every minute to monitor connectivity
- **Multiple Instances**: Support for multiple named database configurations via `managers` map; `GetDB(name)` and `DBM(name)` retrieve by name

### Entry Points

- `db.Install(c Config, name ...string)` - Initialize database connection, returns `*DBManager`
- `db.GetDB(name ...string)` - Get `*xorm.EngineGroup` for CRUD operations
- `db.DBM(name ...string)` - Get `*DBManager` for advanced operations like transactions
- `manager.WithTx(ctx, fn)` - Execute function within a transaction

### Dependencies

- `xorm.io/xorm` - Core ORM
- `github.com/lib/pq` - PostgreSQL driver (included)
- `github.com/cuisi521/zap-wrapper` - Logging interface