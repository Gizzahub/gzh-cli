package gzhclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_UpdateConfig(t *testing.T) {
	tests := []struct {
		name   string
		update ClientConfig
	}{
		{
			name: "shorter timeout",
			update: func() ClientConfig {
				cfg := DefaultConfig()
				cfg.Timeout = 10 * time.Second
				return cfg
			}(),
		},
		{
			name: "change log level",
			update: func() ClientConfig {
				cfg := DefaultConfig()
				cfg.LogLevel = "debug"
				return cfg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(DefaultConfig())
			assert.NoError(t, err)

			err = c.UpdateConfig(tt.update)
			assert.NoError(t, err)
			assert.Equal(t, tt.update, c.GetConfig())
		})
	}
}
