# Contributing

## Error handling

Use the helpers in [`internal/errors`](../internal/errors) for consistent error
management.

```go
import (
    "errors"
    "fmt"

    gerrors "github.com/Gizzahub/gzh-cli/internal/errors"
)

func load() error {
    if err := readConfig(); err != nil {
        return gerrors.Wrap(err, gerrors.ErrConfigNotFound)
    }
    return nil
}

func handle(err error) {
    if errors.Is(err, gerrors.ErrConfigNotFound) {
        // configuration is missing
    }

    var serr *gerrors.StandardError
    if errors.As(err, &serr) {
        fmt.Println("code:", serr.Code)
    }
}
```

These patterns allow callers to detect specific failure scenarios and extract
structured error information.
