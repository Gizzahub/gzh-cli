// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package webhook

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v66/github"
)

// displayWebhooks displays a list of webhooks.
func displayWebhooks(webhooks []*github.Hook, format string) error {
	switch format {
	case "json":
		redacted := make([]*webhookDisplay, len(webhooks))
		for i, hook := range webhooks {
			redacted[i] = webhookForDisplay(hook)
		}

		return json.NewEncoder(os.Stdout).Encode(redacted)
	case "yaml":
		// 기존 YAML 선택 경로의 JSON 형식을 유지하면서 secret은 제외한다.
		configs := make([]webhookConfigDisplay, len(webhooks))
		for i, hook := range webhooks {
			configs[i] = webhookConfigForDisplay(hook)
		}

		data, err := json.MarshalIndent(configs, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))

		return nil
	default:
		// Table format
		fmt.Printf("%-8s %-20s %-40s %-8s %-20s\n", "ID", "EVENTS", "URL", "ACTIVE", "CONTENT_TYPE")
		fmt.Println("-------- -------------------- ---------------------------------------- -------- --------------------")

		for _, hook := range webhooks {
			url := ""
			contentType := ""

			if hook.Config != nil {
				if hook.Config.URL != nil {
					url = *hook.Config.URL
				}

				if hook.Config.ContentType != nil {
					contentType = *hook.Config.ContentType
				}
			}

			eventsStr := ""
			if len(hook.Events) > 0 {
				eventsStr = hook.Events[0]
				if len(hook.Events) > 1 {
					eventsStr += fmt.Sprintf(" (+%d more)", len(hook.Events)-1)
				}
			}

			fmt.Printf("%-8d %-20s %-40s %-8t %-20s\n",
				hook.GetID(),
				eventsStr,
				truncateString(url, 40),
				hook.GetActive(),
				contentType)
		}

		return nil
	}
}

// displayWebhook displays a single webhook.
func displayWebhook(hook *github.Hook, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(webhookForDisplay(hook))
	case "yaml":
		config := webhookConfigForDisplay(hook)

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))

		return nil
	default:
		// Table format
		fmt.Printf("ID: %d\n", hook.GetID())
		fmt.Printf("URL: %s\n", safeStringFromPointer(hook.Config.URL))
		fmt.Printf("Events: %v\n", hook.Events)
		fmt.Printf("Active: %t\n", hook.GetActive())
		fmt.Printf("Content Type: %s\n", safeStringFromPointer(hook.Config.ContentType))
		fmt.Printf("Created: %s\n", hook.GetCreatedAt().Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", hook.GetUpdatedAt().Format("2006-01-02 15:04:05"))

		return nil
	}
}

// webhookForDisplay는 기존 JSON 필드를 보존하되 secret 필드가 없는 표시 전용 값을 만든다.
func webhookForDisplay(hook *github.Hook) *webhookDisplay {
	if hook == nil {
		return nil
	}

	display := &webhookDisplay{
		CreatedAt:    hook.CreatedAt,
		UpdatedAt:    hook.UpdatedAt,
		URL:          hook.URL,
		ID:           hook.ID,
		Type:         hook.Type,
		Name:         hook.Name,
		TestURL:      hook.TestURL,
		PingURL:      hook.PingURL,
		LastResponse: hook.LastResponse,
		Events:       hook.Events,
		Active:       hook.Active,
	}
	if hook.Config != nil {
		display.Config = &webhookDisplayConfig{
			ContentType: hook.Config.ContentType,
			InsecureSSL: hook.Config.InsecureSSL,
			URL:         hook.Config.URL,
		}
	}

	return display
}

// webhookConfigForDisplay는 기존 YAML 선택 경로의 축소형 출력 계약을 보존한다.
func webhookConfigForDisplay(hook *github.Hook) webhookConfigDisplay {
	config := webhookConfigDisplay{
		Events: hook.Events,
		Active: hook.Active,
	}

	if hook.Config != nil {
		if hook.Config.URL != nil {
			config.URL = *hook.Config.URL
		}

		if hook.Config.ContentType != nil {
			config.ContentType = *hook.Config.ContentType
		}
	}

	return config
}

// safeStringFromPointer safely gets a string value from a pointer.
func safeStringFromPointer(ptr *string) string {
	if ptr == nil {
		return ""
	}

	return *ptr
}
