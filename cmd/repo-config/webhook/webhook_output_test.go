// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package webhook

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookMachineOutputRedactsSecret(t *testing.T) {
	secret := "never-print-this-webhook-secret"
	hook := &github.Hook{
		ID:      github.Int64(42),
		URL:     github.String("https://api.github.test/hooks/42"),
		Type:    github.String("Repository"),
		Name:    github.String("web"),
		TestURL: github.String("https://api.github.test/hooks/42/test"),
		PingURL: github.String("https://api.github.test/hooks/42/pings"),
		LastResponse: map[string]interface{}{
			"code":    float64(200),
			"message": "ok",
		},
		Config: &github.HookConfig{
			URL:         github.String("https://example.test/webhook"),
			ContentType: github.String("json"),
			InsecureSSL: github.String("0"),
			Secret:      &secret,
		},
		Events: []string{"push", "issues"},
		Active: github.Bool(true),
	}

	tests := []struct {
		name       string
		format     string
		list       bool
		fullSchema bool
	}{
		{name: "JSON list", format: "json", list: true, fullSchema: true},
		{name: "JSON get", format: "json", fullSchema: true},
		{name: "yaml list", format: "yaml", list: true},
		{name: "yaml get", format: "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureWebhookOutput(t, func() error {
				if tt.list {
					return displayWebhooks([]*github.Hook{hook}, tt.format)
				}

				return displayWebhook(hook, tt.format)
			})

			assert.NotContains(t, string(output), secret)
			object := decodeWebhookOutput(t, output, tt.list)
			config := object
			if tt.fullSchema {
				config = requireJSONMap(t, object, "config")
			}
			assert.NotContains(t, config, "secret")
			assert.Equal(t, "https://example.test/webhook", config["url"])
			contentTypeKey := "contentType"
			if tt.fullSchema {
				contentTypeKey = "content_type"
			}
			assert.Equal(t, "json", config[contentTypeKey])

			if tt.fullSchema {
				assert.Equal(t, float64(42), object["id"])
				assert.Equal(t, "Repository", object["type"])
				assert.Equal(t, "https://api.github.test/hooks/42/test", object["test_url"])
				assert.Equal(t, "0", config["insecure_ssl"])
				assert.Equal(t, "ok", requireJSONMap(t, object, "last_response")["message"])
			}
		})
	}

	// 표시 과정이 생성·수정 요청에 재사용될 수 있는 원본 값을 바꾸면 안 된다.
	require.NotNil(t, hook.Config.Secret)
	assert.Equal(t, secret, *hook.Config.Secret)
}

func captureWebhookOutput(t *testing.T, fn func() error) []byte {
	t.Helper()

	outputFile, err := os.CreateTemp(t.TempDir(), "webhook-output-*.json")
	require.NoError(t, err)

	originalStdout := os.Stdout
	os.Stdout = outputFile
	defer func() {
		os.Stdout = originalStdout
	}()

	require.NoError(t, fn())
	require.NoError(t, outputFile.Close())

	output, err := os.ReadFile(outputFile.Name())
	require.NoError(t, err)

	return output
}

func decodeWebhookOutput(t *testing.T, output []byte, list bool) map[string]interface{} {
	t.Helper()

	if list {
		var objects []map[string]interface{}
		require.NoError(t, json.Unmarshal(output, &objects))
		require.Len(t, objects, 1)

		return objects[0]
	}

	var object map[string]interface{}
	require.NoError(t, json.Unmarshal(output, &object))

	return object
}

func requireJSONMap(t *testing.T, object map[string]interface{}, key string) map[string]interface{} {
	t.Helper()

	value, ok := object[key].(map[string]interface{})
	require.True(t, ok, "%s must be a JSON object", key)

	return value
}
