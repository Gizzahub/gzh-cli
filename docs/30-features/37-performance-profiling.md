# 📊 Performance Profiling Guide

The `gz profile` command collects and inspects profiles of the running `gz`
process with Go's standard `runtime/pprof` and `net/http/pprof` packages. For
interactive exploration, open the saved files with `go tool pprof`.

## 📋 Table of Contents

- [Commands](#commands)
- [Output Files](#output-files)
- [Workflows](#workflows)
- [Analyzing Profiles with go tool pprof](#analyzing-profiles-with-go-tool-pprof)

## Commands

`gz profile` has seven subcommands. The source of truth is `cmd/profile/`
(`gz profile --help`).

| Subcommand                               | Purpose                                          | Flags                                                                                                                       |
| ---------------------------------------- | ------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `server`                                 | Serve the `/debug/pprof/` endpoints until Ctrl+C | `--port` (default `6060`)                                                                                                   |
| `cpu`                                    | Collect a CPU profile for a fixed duration       | `--duration` (default `30s`)                                                                                                |
| `memory`                                 | Write the current heap profile                   | —                                                                                                                           |
| `stats`                                  | Print runtime statistics                         | —                                                                                                                           |
| `compare <baseline.prof> <current.prof>` | Simulated comparison (see below)                 | `--format text\|json` (default `text`), `--threshold` percent (default `5.0`)                                               |
| `continuous`                             | Collect profiles at an interval                  | `--interval` (default `5m`), `--duration` (default `1h`), `--type cpu\|memory\|goroutine` (default `cpu`), `--auto-analyze` |
| `analyze <profile.prof>`                 | Simulated analysis (see below)                   | `--threshold` percent (default `5.0`), `--format text\|json` (default `text`), `--auto-suggest` (default `true`)            |

> ⚠️ `compare` and `analyze` do not parse the profile data yet. Their output is
> a fixed sample chosen from the file names (for example `cpu`, `memory`,
> `baseline`, `current`), so it does not describe your program. Use
> `go tool pprof` for real comparison and analysis.

### `server`

```bash
gz profile server
gz profile server --port 6061
```

The server listens on `127.0.0.1` only and exposes `/debug/pprof/`,
`/debug/pprof/profile`, `/debug/pprof/heap`, `/debug/pprof/goroutine`,
`/debug/pprof/block`, and `/debug/pprof/mutex`. It runs in the foreground and
stops on Ctrl+C.

### `cpu` and `memory`

```bash
gz profile cpu
gz profile cpu --duration 10s
gz profile memory
```

Both commands profile the `gz` process that runs them and print the saved file
path together with a `go tool pprof` hint.

### `stats`

```bash
gz profile stats
```

Prints the goroutine count, heap allocated/system/in-use, stack in use, the
number of GC runs, and the last GC time of the `gz` process.

### `compare`

```bash
gz profile compare baseline.prof current.prof
gz profile compare --threshold 10 baseline.prof current.prof
gz profile compare --format json baseline.prof current.prof
```

The output is simulated; see the warning above.

### `continuous`

```bash
gz profile continuous --interval 1m --duration 30m
gz profile continuous --type memory --interval 5m --duration 2h --auto-analyze
```

Collects one profile per interval until `--duration` elapses or Ctrl+C is
pressed; each CPU sample runs for 10 seconds. `--auto-analyze` passes each
profile to the simulated `analyze` step.

### `analyze`

```bash
gz profile analyze cpu.prof
gz profile analyze --threshold 10 memory.prof
gz profile analyze --format json profile.prof
```

The output is simulated; see the warning above. Sample issues below
`--threshold` percent impact are dropped, and suggestions are included unless
`--auto-suggest=false` is given.

## Output Files

Profiles are written to `tmp/profiles/` under the current working directory, as
`<type>_<YYYYMMDD_HHMMSS>_<sequence>.prof` with mode `0600`. There is no
configuration file setting for the directory, and old profiles are not removed
automatically.

## Workflows

### Compare before and after a change

```bash
gz profile cpu --duration 30s
# make the change, rebuild gz
gz profile cpu --duration 30s
go tool pprof -diff_base=tmp/profiles/<before>.prof tmp/profiles/<after>.prof
```

### Profile a long-running command

Start `gz profile server` in one terminal, run the workload, and fetch profiles
from the endpoints in another:

```bash
go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=30
go tool pprof http://127.0.0.1:6060/debug/pprof/heap
```

## Analyzing Profiles with go tool pprof

`gz profile` does not render reports or graphs itself. Use `go tool pprof`:

```bash
go tool pprof -top tmp/profiles/<file>.prof
go tool pprof -list 'FunctionName' tmp/profiles/<file>.prof
go tool pprof -http=127.0.0.1:8080 tmp/profiles/<file>.prof
go tool pprof -diff_base=<before>.prof <after>.prof
```

`-http` opens the web UI with flame graph, graph, and source views.
