# AGENTS.md - registry

## Coding conventions

- Use Go standard style and run `make fmt` before committing.
- Keep Cobra command implementations simple and avoid unnecessary abstractions.

## Testing and logging

- Run `go test ./cmd/registry -v` before submitting changes.
- Prefer the repository logger for output; use `t.Logf` for test logging.

## Setup and review

- Review existing CLI flags and documentation for registry before modifying.
- Update usage examples when command behavior changes.
