# 📊 Performance Profiling Guide

The `gz profile` command collects profiles of the `gz` process that runs it,
with Go's standard `runtime/pprof` and `net/http/pprof` packages. For
interactive exploration, open the saved files with `go tool pprof`.

> ⚠️ Every subcommand profiles only its own `gz` process, which waits or sleeps
> while it collects. It does not profile other `gz` commands or other
> programs. To profile code paths, use `go test -cpuprofile` /
> `-memprofile` and open the result with `go tool pprof`.

## 📋 Table of Contents

- [Commands](#commands)
- [Output Files](#output-files)
- [Analyzing Profiles with go tool pprof](#analyzing-profiles-with-go-tool-pprof)

## Commands

`gz profile` has five subcommands. The source of truth is `cmd/profile/`
(`gz profile --help`).

| Subcommand   | Purpose                                          | Flags                                                                                                     |
| ------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------------- |
| `server`     | Serve the `/debug/pprof/` endpoints until Ctrl+C | `--port` (default `6060`)                                                                                 |
| `cpu`        | Collect a CPU profile for a fixed duration       | `--duration` (default `30s`)                                                                              |
| `memory`     | Write the current heap profile                   | —                                                                                                         |
| `stats`      | Print runtime statistics                         | —                                                                                                         |
| `continuous` | Collect profiles at an interval                  | `--interval` (default `5m`), `--duration` (default `1h`), `--type cpu\|memory\|goroutine` (default `cpu`) |

### `server`

```bash
gz profile server
gz profile server --port 6061
```

The server listens on `127.0.0.1` only and exposes `/debug/pprof/` with the
standard `net/http/pprof` handlers (for example `profile`, `heap`, and
`goroutine`). The profiles describe the waiting `gz profile server` process
itself. It runs in the foreground and stops on Ctrl+C. Block and mutex
profiling rates are not enabled, so those profiles are empty.

### `cpu` and `memory`

```bash
gz profile cpu
gz profile cpu --duration 10s
gz profile memory
```

`cpu` records the `gz` process while it sleeps for `--duration`; `memory`
writes the heap of the `gz` process. Both print the saved file path together
with a `go tool pprof` hint.

### `stats`

```bash
gz profile stats
```

Prints the goroutine count, heap allocated/system/in-use, stack in use, the
number of GC runs, and the last GC time of the `gz` process.

### `continuous`

```bash
gz profile continuous --interval 1m --duration 30m
gz profile continuous --type memory --interval 5m --duration 2h
```

Collects one profile of its own process per interval until `--duration`
elapses or Ctrl+C is pressed; each CPU sample runs for 10 seconds.
Analyze the collected files with `go tool pprof` (see below).

## Output Files

Profiles are written to `tmp/profiles/` under the current working directory, as
`<type>_<YYYYMMDD_HHMMSS>_<sequence>.prof` with mode `0600`. If that directory
cannot be created, the files go to the current directory. There is no
configuration file setting for the directory, and old profiles are not removed
automatically.

## Analyzing Profiles with go tool pprof

`gz profile` does not analyze, compare, or render profiles itself. Use
`go tool pprof`: `-top` lists the hottest functions of one profile, and
`-diff_base=<base> <current>` shows what changed between two profiles.

```bash
go tool pprof -top tmp/profiles/<file>.prof
go tool pprof -list 'FunctionName' tmp/profiles/<file>.prof
go tool pprof -http=127.0.0.1:8080 tmp/profiles/<file>.prof
go tool pprof -diff_base=<before>.prof <after>.prof
```

`-http` opens the web UI with flame graph, graph, and source views.
