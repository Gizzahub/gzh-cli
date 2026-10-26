// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package webhook

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

// NewWebhookCmd creates the webhook management command.
func NewWebhookCmd() *cobra.Command {
	return newWebhookCmdWithFactory(createWebhookService)
}

func newWebhookCmdWithFactory(factory func() github.WebhookService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "🔗 GitHub 웹훅 관리 도구",
		Long: `GitHub 웹훅 CRUD API 관리 도구

리포지토리 및 조직 웹훅을 생성, 조회, 수정, 삭제할 수 있습니다.
개별 CRUD와 대량 생성 작업을 제공합니다. 모니터링 및 config 실행은 미지원 오류를 반환합니다.

지원하는 기능:
• 개별 웹훅 CRUD 작업
• 명시한 리포지토리 대상 웹훅 대량 생성
• monitor/config 명령·플래그는 호환성을 위해 유지되며 실행은 지원하지 않음`,
	}

	// Repository webhook commands
	cmd.AddCommand(newRepositoryWebhookCmdWithFactory(factory))

	// Organization webhook commands
	cmd.AddCommand(newOrganizationWebhookCmd())

	// Bulk operations
	cmd.AddCommand(newBulkWebhookCmdWithFactory(factory))

	// Organization-wide configuration
	cmd.AddCommand(newWebhookConfigCmd())

	// Monitoring and testing
	cmd.AddCommand(newWebhookMonitorCmd())

	return cmd
}

// Repository webhook commands.
func newRepositoryWebhookCmd() *cobra.Command {
	return newRepositoryWebhookCmdWithFactory(createWebhookService)
}

func newRepositoryWebhookCmdWithFactory(factory func() github.WebhookService) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateRepositoryWebhookWithService(cmd, args, factory())
		},
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
	return newBulkWebhookCmdWithFactory(createWebhookService)
}

func newBulkWebhookCmdWithFactory(factory func() github.WebhookService) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBulkCreateWebhooksWithService(cmd, args, factory())
		},
	}
	createCmd.Flags().String("name", "", "웹훅 이름 (필수)")
	createCmd.Flags().String("url", "", "웹훅 URL (필수)")
	createCmd.Flags().StringSlice("events", []string{"push"}, "이벤트 목록")
	createCmd.Flags().Bool("active", true, "웹훅 활성화 여부")
	createCmd.Flags().StringSlice("repos", nil, "대상 리포지토리 목록 (기본 CLI에서는 명시 필요)")
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
		Long:  "웹훅 테스트와 배송 조회는 구현되지 않아 실행 시 명시적인 미지원 오류를 반환합니다.",
	}

	// Test webhook
	testCmd := &cobra.Command{
		Use:   "test <owner> <repo> <webhook-id>",
		Short: "웹훅 테스트",
		Long:  "웹훅 테스트는 구현되지 않아 실행 시 미지원 오류를 반환합니다.",
		Args:  cobra.ExactArgs(3),
		RunE:  runTestWebhook,
	}

	// Get deliveries
	deliveriesCmd := &cobra.Command{
		Use:   "deliveries <owner> <repo> <webhook-id>",
		Short: "배송 기록 조회",
		Long:  "배송 기록 조회는 구현되지 않아 실행 시 미지원 오류를 반환합니다.",
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
		Long:  "조직 웹훅 설정 CLI는 구현되지 않아 실행 시 명시적인 미지원 오류를 반환합니다. 현재 CLI 실행은 지원하지 않습니다.",
	}

	// Policy management
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "웹훅 정책 관리",
		Long:  "조직의 웹훅 정책을 생성, 조회, 수정, 삭제합니다. 현재 CLI 실행은 지원하지 않습니다.",
	}

	// Create policy
	createPolicyCmd := &cobra.Command{
		Use:   "create <organization> <policy-file>",
		Short: "웹훅 정책 생성",
		Long:  "YAML 파일에서 웹훅 정책을 생성합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(2),
		RunE:  runCreateWebhookPolicy,
	}

	// List policies
	listPoliciesCmd := &cobra.Command{
		Use:   "list <organization>",
		Short: "웹훅 정책 목록",
		Long:  "조직의 모든 웹훅 정책을 조회합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runListWebhookPolicies,
	}

	// Apply policies
	applyPoliciesCmd := &cobra.Command{
		Use:   "apply <organization>",
		Short: "웹훅 정책 적용",
		Long:  "조직의 모든 리포지토리에 웹훅 정책을 적용합니다. 현재 CLI 실행은 지원하지 않습니다.",
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
		Long:  "웹훅 정책 적용 결과를 미리 확인합니다. 현재 CLI 실행은 지원하지 않습니다.",
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
		Long:  "조직의 기본 웹훅 설정을 관리합니다. 현재 CLI 실행은 지원하지 않습니다.",
	}

	// Get org config
	getOrgConfigCmd := &cobra.Command{
		Use:   "get <organization>",
		Short: "조직 설정 조회",
		Long:  "조직의 웹훅 설정을 조회합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetOrganizationWebhookConfig,
	}

	// Update org config
	updateOrgConfigCmd := &cobra.Command{
		Use:   "update <organization> <config-file>",
		Short: "조직 설정 업데이트",
		Long:  "YAML 파일에서 조직의 웹훅 설정을 업데이트합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(2),
		RunE:  runUpdateOrganizationWebhookConfig,
	}

	// Validate org config
	validateOrgConfigCmd := &cobra.Command{
		Use:   "validate <config-file>",
		Short: "설정 검증",
		Long:  "웹훅 설정 파일을 검증합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runValidateWebhookConfig,
	}

	orgConfigCmd.AddCommand(getOrgConfigCmd, updateOrgConfigCmd, validateOrgConfigCmd)

	// Reporting and audit
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "웹훅 리포트 생성",
		Long:  "웹훅 사용 현황과 규정 준수 리포트를 생성합니다. 현재 CLI 실행은 지원하지 않습니다.",
	}

	// Compliance report
	complianceCmd := &cobra.Command{
		Use:   "compliance <organization>",
		Short: "규정 준수 리포트",
		Long:  "조직의 웹훅 규정 준수 상태를 리포트합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runWebhookComplianceReport,
	}

	// Inventory report
	inventoryCmd := &cobra.Command{
		Use:   "inventory <organization>",
		Short: "웹훅 인벤토리",
		Long:  "조직의 모든 웹훅 현황을 조회합니다. 현재 CLI 실행은 지원하지 않습니다.",
		Args:  cobra.ExactArgs(1),
		RunE:  runWebhookInventoryReport,
	}

	// Sync webhooks
	syncCmd := &cobra.Command{
		Use:   "sync <organization>",
		Short: "웹훅 동기화",
		Long:  "조직의 웹훅을 정책과 동기화합니다. 현재 CLI 실행은 지원하지 않습니다.",
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
	webhookService := createWebhookService()

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

	webhookService := createWebhookService()

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

	webhookService := createWebhookService()

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
	return runUpdateRepositoryWebhookWithService(cmd, args, createWebhookService())
}

func runUpdateRepositoryWebhookWithService(cmd *cobra.Command, args []string, webhookService github.WebhookService) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")

	request := &github.WebhookUpdateRequest{
		ID:     webhookID,
		Name:   name,
		URL:    url,
		Events: events,
		Config: github.WebhookConfig{URL: url},
	}

	if cmd.Flags().Changed("active") {
		request.Active = &active
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

	webhookService := createWebhookService()

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

	webhookService := createWebhookService()

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

	webhookService := createWebhookService()

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
	return runBulkCreateWebhooksWithService(cmd, args, createWebhookService())
}

func runBulkCreateWebhooksWithService(cmd *cobra.Command, args []string, webhookService github.WebhookService) error {
	org := args[0]

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")
	repos, _ := cmd.Flags().GetStringSlice("repos")

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

	if result.FailureCount > 0 {
		return fmt.Errorf("webhook creation failed for %d of %d repositories", result.FailureCount, result.TotalRepositories)
	}
	return nil
}

func runTestWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	webhookID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %s", args[2])
	}

	webhookService := createWebhookService()

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

	webhookService := createWebhookService()

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

// createWebhookService는 기존 CLI 토큰 우선순위로 실제 GitHub 서비스를 구성한다.
func createWebhookService() github.WebhookService {
	return createWebhookServiceWithBaseURL("")
}

func createWebhookServiceWithBaseURL(baseURL string) github.WebhookService {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	return github.NewWebhookServiceWithTokenAndBaseURL(nil, token, &mockLogger{}, baseURL)
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
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runListWebhookPolicies(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runApplyWebhookPolicies(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runPreviewWebhookPolicies(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runGetOrganizationWebhookConfig(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runUpdateOrganizationWebhookConfig(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runValidateWebhookConfig(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runWebhookComplianceReport(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runWebhookInventoryReport(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}

func runSyncWebhooks(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("webhook configuration CLI execution: %w", github.ErrNotImplemented)
}
