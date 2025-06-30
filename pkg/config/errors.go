package config

import "errors"

// Configuration errors
var (
	ErrMissingVersion    = errors.New("missing required field: version")
	ErrMissingToken      = errors.New("missing required field: token")
	ErrMissingName       = errors.New("missing required field: name")
	ErrInvalidVisibility = errors.New("invalid visibility: must be 'public', 'private', or 'all'")
	ErrInvalidStrategy   = errors.New("invalid strategy: must be 'reset', 'pull', or 'fetch'")
	ErrInvalidRegex      = errors.New("invalid regex pattern")
	ErrFileNotFound      = errors.New("configuration file not found")
	ErrInvalidYAML       = errors.New("invalid YAML format")
	ErrInvalidCloneDir   = errors.New("invalid clone directory")
	ErrUnsafePath        = errors.New("unsafe path: contains '..'")
)
