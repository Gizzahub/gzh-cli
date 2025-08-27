# Architecture

## Dependency Injection

The CLI uses a lightweight dependency injection pattern based on `AppContext`.
The context provides shared services like structured logging and global configuration.

```go
// internal/app/context.go
type AppContext struct {
    Logger *logger.StructuredLogger
    Config *config.GlobalConfig
}
```

The root command creates the context and passes it to command constructors:

```go
cfg, _ := config.LoadGlobalConfig()
log := logger.NewStructuredLogger("gzh-cli", logger.LevelInfo)
appCtx := &app.AppContext{Logger: log, Config: cfg}

cmd.AddCommand(synclone.NewSyncCloneCmd(ctx, appCtx))
```

Commands and services then access shared dependencies through the context:

```go
logger := appCtx.Logger.WithContext("component", "synclone")
cfg := appCtx.Config
```

This approach centralizes configuration and logging while keeping
command implementations simple and testable.
