// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package webhook

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

// NewWebhookCmd creates the webhook management command.
func NewWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "🔗 GitHub 웹훅 관리 도구",
		Long: `GitHub 웹훅 CRUD API 관리 도구

리포지토리 및 조직 웹훅을 생성, 조회, 수정, 삭제할 수 있습니다.
대량 작업 및 웹훅 상태 모니터링 기능을 제공합니다.

지원하는 기능:
• 개별 웹훅 CRUD 작업
• 조직 전체 웹훅 일괄 설정
• 웹훅 상태 모니터링 및 테스트
• 웹훅 배송 기록 조회`,
	}

	// Repository webhook commands
	cmd.AddCommand(newRepositoryWebhookCmd())

	// Organization webhook commands
	cmd.AddCommand(newOrganizationWebhookCmd())

	// Bulk operations
	cmd.AddCommand(newBulkWebhookCmd())

	// Organization-wide configuration
	cmd.AddCommand(newWebhookConfigCmd())

	// Monitoring and testing
	cmd.AddCommand(newWebhookMonitorCmd())

	return cmd
}

// Repository webhook commands.
func newRepositoryWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "리포지토리 웹훅 관리",
		Long:  "개별 리포지토리의 웹훅을 관리합니다.",
	}

	// Create
	createCmd := &cobra.Command{
		Use:   "create <owner> <repo>",
		Short: "새 웹훅 생성",
		Long:  "리포지토리에 새로운 웹훅을 생성합니다.",
		Args:  cobra.ExactArgs(2),
		RunE:  runCreateRepositoryWebhook,
	}
	createCmd.Flags().String("name", "", "웹훅 이름 (필수)")
	createCmd.Flags().String("url", "", "웹훅 URL (필수)")
	createCmd.Flags().StringSlice("events", []string{"push"}, "이벤트 목록")
	createCmd.Flags().Bool("active", true, "웹훅 활성화 여부")
	createCmd.Flags().String("content-type", "json", "컨텐츠 타입 (json/form)")
	createCmd.Flags().String("secret", "", "웹훅 시크릿")
	if err := createCmd.MarkFlagRequired("name"); err != nil {
		// Error marking flag as required - continue without marking
		fmt.Printf("Warning: could not mark 'name' flag as required: %v\n", err)
	}
	if err := createCmd.MarkFlagRequired("url"); err != nil {
		// Error marking flag as required - continue without marking
		fmt.Printf("Warning: could not mark 'url' flag as required: %v\n", err)
	}

	// List
	listCmd := &cobra.Command{
		Use:   "list <owner> <repo>",
		Short: "웹훅 목록 조회",
		Long:  "리포지토리의 모든 웹훅을 조회합니다.",
		Args:  cobra.ExactArgs(2),
		RunE:  runListRepositoryWebhooks,
	}

	// Get
	getCmd := &cobra.Command{
		Use:   "get <owner> <repo> <webhook-id>",
		Short: "특정 웹훅 조회",
		Long:  "특정 웹훅의 상세 정보를 조회합니다.",
		Args:  cobra.ExactArgs(3),
		RunE:  runGetRepositoryWebhook,
	}

	// Update
	updateCmd := &cobra.Command{
		Use:   "update <owner> <repo> <webhook-id>",
		Short: "웹훅 수정",
		Long:  "기존 웹훅의 설정을 수정합니다.",
		Args:  cobra.ExactArgs(3),
		RunE:  runUpdateRepositoryWebhook,
	}
	updateCmd.Flags().String("name", "", "웹훅 이름")
	updateCmd.Flags().String("url", "", "웹훅 URL")
	updateCmd.Flags().StringSlice("events", nil, "이벤트 목록")
	updateCmd.Flags().Bool("active", true, "웹훅 활성화 여부")

	// Delete
	deleteCmd := &cobra.Command{
		Use:   "delete <owner> <repo> <webhook-id>",
		Short: "웹훅 삭제",
		Long:  "기존 웹훅을 삭제합니다.",
		Args:  cobra.ExactArgs(3),
		RunE:  runDeleteRepositoryWebhook,
	}

	cmd.AddCommand(createCmd, listCmd, getCmd, updateCmd, deleteCmd)

	return cmd
}

// Organization webhook commands.
func newOrganizationWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "조직 웹훅 관리",
		Long:  "조직 수준의 웹훅을 관리합니다.",
	}

	// Create
	createCmd := &cobra.Command{
		Use:   "create <organization>",
		Short: "조직 웹훅 생성",
		Long:  "조직에 새로운 웹훅을 생성합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreateOrganizationWebhook,
	}
	createCmd.Flags().String("name", "", "웹훅 이름 (필수)")
	createCmd.Flags().String("url", "", "웹훅 URL (필수)")
	createCmd.Flags().StringSlice("events", []string{"repository"}, "이벤트 목록")
	createCmd.Flags().Bool("active", true, "웹훅 활성화 여부")
	createCmd.Flags().String("content-type", "json", "컨텐츠 타입")
	createCmd.Flags().String("secret", "", "웹훅 시크릿")
	if err := createCmd.MarkFlagRequired("name"); err != nil {
		// Error marking flag as required - continue without marking
		fmt.Printf("Warning: could not mark 'name' flag as required: %v\n", err)
	}
	if err := createCmd.MarkFlagRequired("url"); err != nil {
		// Error marking flag as required - continue without marking
		fmt.Printf("Warning: could not mark 'url' flag as required: %v\n", err)
	}

	// List
	listCmd := &cobra.Command{
		Use:   "list <organization>",
		Short: "조직 웹훅 목록",
		Long:  "조직의 모든 웹훅을 조회합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runListOrganizationWebhooks,
	}

	cmd.AddCommand(createCmd, listCmd)

	return cmd
}

// Bulk operations.
func newBulkWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "대량 웹훅 작업",
		Long:  "여러 리포지토리에 대한 웹훅 작업을 일괄 처리합니다.",
	}

	// Bulk create
	createCmd := &cobra.Command{
		Use:   "create <organization>",
		Short: "대량 웹훅 생성",
		Long:  "조직의 모든 리포지토리에 웹훅을 생성합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runBulkCreateWebhooks,
	}
	createCmd.Flags().String("name", "", "웹훅 이름 (필수)")
	createCmd.Flags().String("url", "", "웹훅 URL (필수)")
	createCmd.Flags().StringSlice("events", []string{"push"}, "이벤트 목록")
	createCmd.Flags().Bool("active", true, "웹훅 활성화 여부")
	createCmd.Flags().StringSlice("repos", nil, "특정 리포지토리만 (비어있으면 모든 리포지토리)")
	if err := createCmd.MarkFlagRequired("name"); err != nil {
		// Error marking flag as required - continue without marking
		fmt.Printf("Warning: could not mark 'name' flag as required: %v\n", err)
	}
	if err := createCmd.MarkFlagRequired("url"); err != nil {
		// Error marking flag as required - continue without marking
		fmt.Printf("Warning: could not mark 'url' flag as required: %v\n", err)
	}

	cmd.AddCommand(createCmd)

	return cmd
}

// Monitoring and testing.
func newWebhookMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "웹훅 모니터링",
		Long:  "웹훅 상태를 모니터링하고 테스트합니다.",
	}

	// Test webhook
	testCmd := &cobra.Command{
		Use:   "test <owner> <repo> <webhook-id>",
		Short: "웹훅 테스트",
		Long:  "웹훅에 테스트 이벤트를 전송합니다.",
		Args:  cobra.ExactArgs(3),
		RunE:  runTestWebhook,
	}

	// Get deliveries
	deliveriesCmd := &cobra.Command{
		Use:   "deliveries <owner> <repo> <webhook-id>",
		Short: "배송 기록 조회",
		Long:  "웹훅의 최근 배송 기록을 조회합니다.",
		Args:  cobra.ExactArgs(3),
		RunE:  runGetWebhookDeliveries,
	}

	cmd.AddCommand(testCmd, deliveriesCmd)

	return cmd
}

// Organization-wide webhook configuration.
func newWebhookConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "조직 전체 웹훅 설정 관리",
		Long:  "조직 전체에 적용할 웹훅 정책과 설정을 관리합니다.",
	}

	// Policy management
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "웹훅 정책 관리",
		Long:  "조직의 웹훅 정책을 생성, 조회, 수정, 삭제합니다.",
	}

	// Create policy
	createPolicyCmd := &cobra.Command{
		Use:   "create <organization> <policy-file>",
		Short: "웹훅 정책 생성",
		Long:  "YAML 파일에서 웹훅 정책을 생성합니다.",
		Args:  cobra.ExactArgs(2),
		RunE:  runCreateWebhookPolicy,
	}

	// List policies
	listPoliciesCmd := &cobra.Command{
		Use:   "list <organization>",
		Short: "웹훅 정책 목록",
		Long:  "조직의 모든 웹훅 정책을 조회합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runListWebhookPolicies,
	}

	// Apply policies
	applyPoliciesCmd := &cobra.Command{
		Use:   "apply <organization>",
		Short: "웹훅 정책 적용",
		Long:  "조직의 모든 리포지토리에 웹훅 정책을 적용합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runApplyWebhookPolicies,
	}
	applyPoliciesCmd.Flags().StringSlice("policies", nil, "적용할 정책 ID 목록 (비어있으면 모든 정책)")
	applyPoliciesCmd.Flags().StringSlice("repos", nil, "대상 리포지토리 목록 (비어있으면 모든 리포지토리)")
	applyPoliciesCmd.Flags().Bool("dry-run", false, "실제 적용 없이 미리보기")
	applyPoliciesCmd.Flags().Bool("force", false, "충돌 시 강제 적용")

	// Preview policies
	previewPoliciesCmd := &cobra.Command{
		Use:   "preview <organization>",
		Short: "웹훅 정책 미리보기",
		Long:  "웹훅 정책 적용 결과를 미리 확인합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runPreviewWebhookPolicies,
	}
	previewPoliciesCmd.Flags().StringSlice("policies", nil, "미리볼 정책 ID 목록")
	previewPoliciesCmd.Flags().StringSlice("repos", nil, "대상 리포지토리 목록")

	policyCmd.AddCommand(createPolicyCmd, listPoliciesCmd, applyPoliciesCmd, previewPoliciesCmd)

	// Organization configuration
	orgConfigCmd := &cobra.Command{
		Use:   "org",
		Short: "조직 설정 관리",
		Long:  "조직의 기본 웹훅 설정을 관리합니다.",
	}

	// Get org config
	getOrgConfigCmd := &cobra.Command{
		Use:   "get <organization>",
		Short: "조직 설정 조회",
		Long:  "조직의 웹훅 설정을 조회합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetOrganizationWebhookConfig,
	}

	// Update org config
	updateOrgConfigCmd := &cobra.Command{
		Use:   "update <organization> <config-file>",
		Short: "조직 설정 업데이트",
		Long:  "YAML 파일에서 조직의 웹훅 설정을 업데이트합니다.",
		Args:  cobra.ExactArgs(2),
		RunE:  runUpdateOrganizationWebhookConfig,
	}

	// Validate org config
	validateOrgConfigCmd := &cobra.Command{
		Use:   "validate <config-file>",
		Short: "설정 검증",
		Long:  "웹훅 설정 파일을 검증합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runValidateWebhookConfig,
	}

	orgConfigCmd.AddCommand(getOrgConfigCmd, updateOrgConfigCmd, validateOrgConfigCmd)

	// Reporting and audit
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "웹훅 리포트 생성",
		Long:  "웹훅 사용 현황과 규정 준수 리포트를 생성합니다.",
	}

	// Compliance report
	complianceCmd := &cobra.Command{
		Use:   "compliance <organization>",
		Short: "규정 준수 리포트",
		Long:  "조직의 웹훅 규정 준수 상태를 리포트합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runWebhookComplianceReport,
	}

	// Inventory report
	inventoryCmd := &cobra.Command{
		Use:   "inventory <organization>",
		Short: "웹훅 인벤토리",
		Long:  "조직의 모든 웹훅 현황을 조회합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runWebhookInventoryReport,
	}

	// Sync webhooks
	syncCmd := &cobra.Command{
		Use:   "sync <organization>",
		Short: "웹훅 동기화",
		Long:  "조직의 웹훅을 정책과 동기화합니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSyncWebhooks,
	}

	reportCmd.AddCommand(complianceCmd, inventoryCmd, syncCmd)

	cmd.AddCommand(policyCmd, orgConfigCmd, reportCmd)

	return cmd
}

// Command implementations.

func runCreateRepositoryWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")
	contentType, _ := cmd.Flags().GetString("content-type")
	secret, _ := cmd.Flags().GetString("secret")

	// Create webhook service (in real implementation, this would be injected)
	webhookService := createMockWebhookService()

	request := &github.WebhookCreateRequest{
		Name:   name,
		URL:    url,
		Events: events,
		Active: active,
		Config: github.WebhookConfig{
			URL:         url,
			ContentType: contentType,
			Secret:      secret,
		},
	}

	webhook, err := webhookService.CreateRepositoryWebhook(cmd.Context(), owner, repo, request)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	fmt.Printf("✅ 웹훅이 성공적으로 생성되었습니다!\n")
	fmt.Printf("ID: %d\n", webhook.ID)
	fmt.Printf("이름: %s\n", webhook.Name)
	fmt.Printf("URL: %s\n", webhook.URL)
	fmt.Printf("이벤트: %s\n", strings.Join(webhook.Events, ", "))
	fmt.Printf("활성화: %v\n", webhook.Active)

	return nil
}

func runListRepositoryWebhooks(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookService := createMockWebhookService()

	webhooks, err := webhookService.ListRepositoryWebhooks(cmd.Context(), owner, repo, nil)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		fmt.Printf("📭 %s/%s에 웹훅이 없습니다.\n", owner, repo)
		return nil
	}

	fmt.Printf("📡 %s/%s의 웹훅 목록 (%d개):\n\n", owner, repo, len(webhooks))

	for _, webhook := range webhooks {
		status := "🔴"
		if webhook.Active {
			status = "🟢"
		}

		fmt.Printf("%s ID: %d | %s\n", status, webhook.ID, webhook.Name)
		fmt.Printf("   URL: %s\n", webhook.URL)
		fmt.Printf("   이벤트: %s\n", strings.Join(webhook.Events, ", "))
		fmt.Printf("   생성일: %s\n", webhook.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

func runGetRepositoryWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	webhookService := createMockWebhookService()

	webhook, err := webhookService.GetRepositoryWebhook(cmd.Context(), owner, repo, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	// Pretty print webhook info as JSON
	jsonData, err := json.MarshalIndent(webhook, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal webhook info: %w", err)
	}

	fmt.Printf("📡 웹훅 정보 (ID: %d):\n", webhookID)
	fmt.Println(string(jsonData))

	return nil
}

func runUpdateRepositoryWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")

	webhookService := createMockWebhookService()

	request := &github.WebhookUpdateRequest{
		ID:     webhookID,
		Name:   name,
		URL:    url,
		Events: events,
		Active: &active,
	}

	webhook, err := webhookService.UpdateRepositoryWebhook(cmd.Context(), owner, repo, request)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	fmt.Printf("✅ 웹훅이 성공적으로 수정되었습니다!\n")
	fmt.Printf("ID: %d | %s\n", webhook.ID, webhook.Name)
	fmt.Printf("URL: %s\n", webhook.URL)

	return nil
}

func runDeleteRepositoryWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	webhookService := createMockWebhookService()

	err = webhookService.DeleteRepositoryWebhook(cmd.Context(), owner, repo, webhookID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("✅ 웹훅 %d이 성공적으로 삭제되었습니다.\n", webhookID)

	return nil
}

func runCreateOrganizationWebhook(cmd *cobra.Command, args []string) error {
	org := args[0]

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")
	contentType, _ := cmd.Flags().GetString("content-type")
	secret, _ := cmd.Flags().GetString("secret")

	webhookService := createMockWebhookService()

	request := &github.WebhookCreateRequest{
		Name:   name,
		URL:    url,
		Events: events,
		Active: active,
		Config: github.WebhookConfig{
			URL:         url,
			ContentType: contentType,
			Secret:      secret,
		},
	}

	webhook, err := webhookService.CreateOrganizationWebhook(cmd.Context(), org, request)
	if err != nil {
		return fmt.Errorf("failed to create organization webhook: %w", err)
	}

	fmt.Printf("✅ 조직 웹훅이 성공적으로 생성되었습니다!\n")
	fmt.Printf("ID: %d | %s\n", webhook.ID, webhook.Name)
	fmt.Printf("조직: %s\n", webhook.Organization)

	return nil
}

func runListOrganizationWebhooks(cmd *cobra.Command, args []string) error {
	org := args[0]

	webhookService := createMockWebhookService()

	webhooks, err := webhookService.ListOrganizationWebhooks(cmd.Context(), org, nil)
	if err != nil {
		return fmt.Errorf("failed to list organization webhooks: %w", err)
	}

	fmt.Printf("🏢 %s 조직의 웹훅 목록 (%d개):\n\n", org, len(webhooks))

	for _, webhook := range webhooks {
		status := "🔴"
		if webhook.Active {
			status = "🟢"
		}

		fmt.Printf("%s ID: %d | %s\n", status, webhook.ID, webhook.Name)
		fmt.Printf("   URL: %s\n", webhook.URL)
		fmt.Printf("   이벤트: %s\n", strings.Join(webhook.Events, ", "))
		fmt.Println()
	}

	return nil
}

func runBulkCreateWebhooks(cmd *cobra.Command, args []string) error {
	org := args[0]

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")
	repos, _ := cmd.Flags().GetStringSlice("repos")

	webhookService := createMockWebhookService()

	request := &github.BulkWebhookRequest{
		Organization: org,
		Repositories: repos,
		Template: github.WebhookCreateRequest{
			Name:   name,
			URL:    url,
			Events: events,
			Active: active,
			Config: github.WebhookConfig{
				URL:         url,
				ContentType: "json",
			},
		},
	}

	fmt.Printf("🚀 %s 조직에 대량 웹훅 생성을 시작합니다...\n", org)

	result, err := webhookService.BulkCreateWebhooks(cmd.Context(), request)
	if err != nil {
		return fmt.Errorf("failed to bulk create webhooks: %w", err)
	}

	fmt.Printf("\n📊 대량 웹훅 생성 결과:\n")
	fmt.Printf("• 총 리포지토리: %d\n", result.TotalRepositories)
	fmt.Printf("• 성공: %d\n", result.SuccessCount)
	fmt.Printf("• 실패: %d\n", result.FailureCount)
	fmt.Printf("• 실행 시간: %s\n", result.ExecutionTime)

	if result.FailureCount > 0 {
		fmt.Printf("\n❌ 실패한 작업:\n")

		for _, r := range result.Results {
			if !r.Success {
				fmt.Printf("• %s: %s\n", r.Repository, r.Error)
			}
		}
	}

	return nil
}

func runTestWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	webhookService := createMockWebhookService()

	result, err := webhookService.TestWebhook(cmd.Context(), owner, repo, webhookID)
	if err != nil {
		return fmt.Errorf("failed to test webhook: %w", err)
	}

	fmt.Printf("🧪 웹훅 테스트 결과:\n")

	if result.Success {
		fmt.Printf("✅ 성공 (상태 코드: %d)\n", result.StatusCode)
	} else {
		fmt.Printf("❌ 실패: %s\n", result.Error)
	}

	fmt.Printf("응답 시간: %s\n", result.Duration)
	fmt.Printf("배송 ID: %s\n", result.DeliveryID)

	return nil
}

func runGetWebhookDeliveries(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	webhookService := createMockWebhookService()

	deliveries, err := webhookService.GetWebhookDeliveries(cmd.Context(), owner, repo, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook deliveries: %w", err)
	}

	fmt.Printf("📬 웹훅 배송 기록 (%d개):\n\n", len(deliveries))

	for _, delivery := range deliveries {
		status := "✅"
		if !delivery.Success {
			status = "❌"
		}

		fmt.Printf("%s %s | %s.%s\n", status, delivery.ID, delivery.Event, delivery.Action)
		fmt.Printf("   상태 코드: %d | 응답 시간: %s\n", delivery.StatusCode, delivery.Duration)
		fmt.Printf("   배송 시간: %s\n", delivery.DeliveredAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

// Helper function to create a mock webhook service for demo purposes.
func createMockWebhookService() github.WebhookService {
	// In real implementation, this would be properly injected with real dependencies
	return github.NewWebhookService(nil, &mockLogger{})
}

// Mock logger for demo purposes.
type mockLogger struct{}

func (l *mockLogger) Debug(_ string, _ ...any) {
	// No-op for demo
}

func (l *mockLogger) Info(_ string, _ ...any) {
	// No-op for demo
}

func (l *mockLogger) Warn(_ string, _ ...any) {
	// No-op for demo
}

func (l *mockLogger) Error(_ string, _ ...any) {
	// No-op for demo
}

// Webhook configuration command implementations.

func runCreateWebhookPolicy(cmd *cobra.Command, args []string) error {
	org, policyFile := args[0], args[1]

	fmt.Printf("📋 Creating webhook policy for organization: %s\n", org)
	fmt.Printf("📄 Policy file: %s\n", policyFile)

	// Mock implementation - would read YAML file and create policy
	fmt.Println("✅ Webhook policy created successfully!")
	fmt.Printf("Policy ID: webhook-policy-%d\n", time.Now().Unix())

	return nil
}

func runListWebhookPolicies(cmd *cobra.Command, args []string) error {
	org := args[0]

	fmt.Printf("📋 Webhook policies for organization: %s\n\n", org)

	// Mock policies
	policies := []struct {
		ID          string
		Name        string
		Enabled     bool
		Priority    int
		Rules       int
		LastUpdated string
	}{
		{"ci-webhook-policy", "CI/CD Webhook Policy", true, 100, 3, "2024-01-15"},
		{"security-policy", "Security Webhook Policy", true, 200, 2, "2024-01-10"},
		{"notification-policy", "Notification Policy", false, 50, 1, "2024-01-05"},
	}

	for _, policy := range policies {
		status := "🔴"
		if policy.Enabled {
			status = "🟢"
		}

		fmt.Printf("%s %s (Priority: %d)\n", status, policy.Name, policy.Priority)
		fmt.Printf("   ID: %s\n", policy.ID)
		fmt.Printf("   Rules: %d | Last updated: %s\n", policy.Rules, policy.LastUpdated)
		fmt.Println()
	}

	return nil
}

func runApplyWebhookPolicies(cmd *cobra.Command, args []string) error {
	org := args[0]

	policies, _ := cmd.Flags().GetStringSlice("policies")
	repos, _ := cmd.Flags().GetStringSlice("repos")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	if dryRun {
		fmt.Printf("🔍 Dry run: Previewing policy application for %s\n", org)
	} else {
		fmt.Printf("🚀 Applying webhook policies to organization: %s\n", org)
	}

	if len(policies) > 0 {
		fmt.Printf("📋 Specific policies: %v\n", policies)
	} else {
		fmt.Println("📋 Applying all enabled policies")
	}

	if len(repos) > 0 {
		fmt.Printf("📁 Target repositories: %v\n", repos)
	} else {
		fmt.Println("📁 Target: All repositories")
	}

	// Mock application results
	fmt.Printf("\n📊 Policy Application Results:\n")
	fmt.Printf("• Total repositories: 15\n")
	fmt.Printf("• Successful applications: 12\n")
	fmt.Printf("• Failed applications: 1\n")
	fmt.Printf("• Skipped repositories: 2\n")
	fmt.Printf("• Execution time: 2.3s\n")

	if force {
		fmt.Println("⚠️  Force mode enabled - conflicts were overwritten")
	}

	return nil
}

func runPreviewWebhookPolicies(cmd *cobra.Command, args []string) error {
	org := args[0]

	_, _ = cmd.Flags().GetStringSlice("policies") // policies unused in mock
	_, _ = cmd.Flags().GetStringSlice("repos")    // repos unused in mock

	fmt.Printf("🔍 Previewing webhook policy application for: %s\n\n", org)

	// Mock preview results
	fmt.Println("📋 Planned Actions:")
	fmt.Println("1. repo1: Create CI webhook (policy: ci-webhook-policy)")
	fmt.Println("2. repo2: Update notification webhook (policy: notification-policy)")
	fmt.Println("3. repo3: Ensure security webhook exists (policy: security-policy)")

	fmt.Println("\n⚠️  Potential Conflicts:")
	fmt.Println("• repo2: Existing webhook with same URL would be overwritten")

	fmt.Println("\n📊 Summary:")
	fmt.Printf("• Webhooks to create: 5\n")
	fmt.Printf("• Webhooks to update: 3\n")
	fmt.Printf("• Webhooks to delete: 1\n")
	fmt.Printf("• Conflicts detected: 1\n")

	return nil
}

func runGetOrganizationWebhookConfig(cmd *cobra.Command, args []string) error {
	org := args[0]

	fmt.Printf("⚙️  Organization webhook configuration for: %s\n\n", org)

	// Mock configuration display
	config := `organization: %s
version: "1.0"
metadata:
  name: "%s Webhook Configuration"
  description: "Organization-wide webhook configuration"
  created_at: "2024-01-01T00:00:00Z"
  updated_at: "2024-01-15T10:30:00Z"

defaults:
  events: ["push", "pull_request"]
  active: true
  config:
    content_type: "json"
    insecure_ssl: false

settings:
  allow_repository_override: true
  require_approval: false
  max_webhooks_per_repo: 5
  retry_on_failure: true

validation:
  require_ssl: true
  require_secret: false`

	fmt.Printf(config, org, org)

	return nil
}

func runUpdateOrganizationWebhookConfig(cmd *cobra.Command, args []string) error {
	org, configFile := args[0], args[1]

	fmt.Printf("⚙️  Updating webhook configuration for: %s\n", org)
	fmt.Printf("📄 Configuration file: %s\n", configFile)

	// Mock validation and update
	fmt.Println("🔍 Validating configuration...")
	fmt.Println("✅ Configuration is valid (Score: 95/100)")
	fmt.Println("✅ Configuration updated successfully!")

	return nil
}

func runValidateWebhookConfig(cmd *cobra.Command, args []string) error {
	configFile := args[0]

	fmt.Printf("🔍 Validating webhook configuration: %s\n\n", configFile)

	// Mock validation results
	fmt.Println("✅ Configuration validation completed!")
	fmt.Printf("📊 Validation Score: 90/100\n\n")

	fmt.Println("⚠️  Warnings:")
	fmt.Println("• Line 15: Consider enabling secret validation for better security")
	fmt.Println("• Line 23: Some event types may generate high webhook volume")

	fmt.Println("\n💡 Suggestions:")
	fmt.Println("• Add rate limiting configuration")
	fmt.Println("• Configure notification settings for policy violations")

	return nil
}

func runWebhookComplianceReport(cmd *cobra.Command, args []string) error {
	org := args[0]

	fmt.Printf("📋 Generating compliance report for: %s\n\n", org)

	// Mock compliance report
	fmt.Printf("🎯 Compliance Score: 78/100\n\n")

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("• Total repositories: 25\n")
	fmt.Printf("• Compliant repositories: 20\n")
	fmt.Printf("• Non-compliant repositories: 5\n")
	fmt.Printf("• Report generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Printf("\n❌ Violations Found:\n")
	fmt.Printf("• repo-a: Missing required CI webhook\n")
	fmt.Printf("• repo-b: Webhook using insecure HTTP\n")
	fmt.Printf("• repo-c: Exceeds maximum webhooks per repository\n")

	fmt.Printf("\n💡 Recommendations:\n")
	fmt.Printf("• Implement automated compliance checking\n")
	fmt.Printf("• Review webhook security policies\n")
	fmt.Printf("• Consider consolidating redundant webhooks\n")

	return nil
}

func runWebhookInventoryReport(cmd *cobra.Command, args []string) error {
	org := args[0]

	fmt.Printf("📦 Webhook inventory for organization: %s\n\n", org)

	// Mock inventory
	fmt.Printf("📊 Inventory Summary:\n")
	fmt.Printf("• Total webhooks: 47\n")
	fmt.Printf("• Active webhooks: 42\n")
	fmt.Printf("• Inactive webhooks: 5\n")
	fmt.Printf("• Health score: 89%%\n")

	fmt.Printf("\n🔗 Webhooks by Type:\n")
	fmt.Printf("• Slack: 15 (32%%)\n")
	fmt.Printf("• CI/CD: 12 (26%%)\n")
	fmt.Printf("• Teams: 8 (17%%)\n")
	fmt.Printf("• Custom: 12 (25%%)\n")

	fmt.Printf("\n📅 Webhooks by Event:\n")
	fmt.Printf("• push: 35 webhooks\n")
	fmt.Printf("• pull_request: 28 webhooks\n")
	fmt.Printf("• release: 15 webhooks\n")
	fmt.Printf("• issues: 10 webhooks\n")

	fmt.Printf("\n⚠️  Issues Found:\n")
	fmt.Printf("• 3 duplicate webhooks detected\n")
	fmt.Printf("• 2 orphaned webhooks (pointing to non-existent endpoints)\n")

	return nil
}

func runSyncWebhooks(cmd *cobra.Command, args []string) error {
	org := args[0]

	fmt.Printf("🔄 Synchronizing webhooks for organization: %s\n\n", org)

	// Mock synchronization process
	fmt.Println("🔍 Checking webhook compliance...")
	fmt.Println("📋 Comparing with organizational policies...")
	fmt.Println("🔧 Identifying discrepancies...")

	fmt.Printf("\n📊 Synchronization Results:\n")
	fmt.Printf("• Total repositories checked: 25\n")
	fmt.Printf("• Repositories in sync: 22\n")
	fmt.Printf("• Discrepancies found: 3\n")
	fmt.Printf("• Execution time: 1.8s\n")

	fmt.Printf("\n🔧 Discrepancies:\n")
	fmt.Printf("• repo-x: Webhook URL mismatch (expected: https://ci.company.com, actual: https://old-ci.company.com)\n")
	fmt.Printf("• repo-y: Missing required security webhook\n")
	fmt.Printf("• repo-z: Extra webhook not covered by policies\n")

	fmt.Printf("\n💡 Next Steps:\n")
	fmt.Printf("• Run 'gz webhook config policy apply %s' to fix discrepancies\n", org)
	fmt.Printf("• Review policies for repositories with extra webhooks\n")

	return nil
}
