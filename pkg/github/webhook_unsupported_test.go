// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type unsupportedWebhookTransport struct{ calls int }

func (tr *unsupportedWebhookTransport) RoundTrip(*http.Request) (*http.Response, error) {
	tr.calls++
	return nil, ErrNotImplemented
}

// 직접 SDK 호출도 HTTP·logger·storage 접근 없이 미지원임을 알린다.
func TestWebhookUnsupportedSDKBeforeDependencies(t *testing.T) {
	transport := &unsupportedWebhookTransport{}
	service := NewWebhookService(nil, nil).(*webhookServiceImpl)
	service.httpClient = &http.Client{Transport: transport}
	ping, err := service.TestWebhook(context.Background(), "owner", "repo", 7)
	require.ErrorIs(t, err, ErrNotImplemented)
	require.Nil(t, ping)
	deliveries, err := service.GetWebhookDeliveries(context.Background(), "owner", "repo", 7)
	require.ErrorIs(t, err, ErrNotImplemented)
	require.Nil(t, deliveries)
	require.Zero(t, transport.calls)
	config := NewWebhookConfigurationService(nil, nil, nil, nil)
	migration, err := config.MigrateExistingWebhooks(context.Background(), nil)
	require.ErrorIs(t, err, ErrNotImplemented)
	require.Nil(t, migration)
	syncResult, err := config.SyncOrganizationWebhooks(context.Background(), "org")
	require.ErrorIs(t, err, ErrNotImplemented)
	require.Nil(t, syncResult)
	compliance, err := config.GenerateComplianceReport(context.Background(), "org")
	require.ErrorIs(t, err, ErrNotImplemented)
	require.Nil(t, compliance)
	inventory, err := config.GetWebhookInventory(context.Background(), "org")
	require.ErrorIs(t, err, ErrNotImplemented)
	require.Nil(t, inventory)
}
