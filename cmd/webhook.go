package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

// NewWebhookCmd creates the webhook management command
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

	// Monitoring and testing
	cmd.AddCommand(newWebhookMonitorCmd())

	return cmd
}

// Repository webhook commands
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
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("url")

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

// Organization webhook commands
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
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("url")

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

// Bulk operations
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
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("url")

	cmd.AddCommand(createCmd)
	return cmd
}

// Monitoring and testing
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

// Command implementations

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

	webhook, err := webhookService.CreateRepositoryWebhook(context.Background(), owner, repo, request)
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
	webhooks, err := webhookService.ListRepositoryWebhooks(context.Background(), owner, repo, nil)
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
	webhook, err := webhookService.GetRepositoryWebhook(context.Background(), owner, repo, webhookID)
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

	webhook, err := webhookService.UpdateRepositoryWebhook(context.Background(), owner, repo, request)
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
	err = webhookService.DeleteRepositoryWebhook(context.Background(), owner, repo, webhookID)
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

	webhook, err := webhookService.CreateOrganizationWebhook(context.Background(), org, request)
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
	webhooks, err := webhookService.ListOrganizationWebhooks(context.Background(), org, nil)
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
	result, err := webhookService.BulkCreateWebhooks(context.Background(), request)
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
	result, err := webhookService.TestWebhook(context.Background(), owner, repo, webhookID)
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
	deliveries, err := webhookService.GetWebhookDeliveries(context.Background(), owner, repo, webhookID)
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

// Helper function to create a mock webhook service for demo purposes
func createMockWebhookService() github.WebhookService {
	// In real implementation, this would be properly injected with real dependencies
	return github.NewWebhookService(nil, &mockLogger{})
}

// Mock logger for demo purposes
type mockLogger struct{}

func (l *mockLogger) Debug(msg string, fields ...interface{}) {
	// No-op for demo
}

func (l *mockLogger) Info(msg string, fields ...interface{}) {
	// No-op for demo
}

func (l *mockLogger) Warn(msg string, fields ...interface{}) {
	// No-op for demo
}

func (l *mockLogger) Error(msg string, fields ...interface{}) {
	// No-op for demo
}
