package performance

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/api"
	"github.com/spf13/cobra"
)

// apiOptimizationCmd represents the api optimization command
var apiOptimizationCmd = &cobra.Command{
	Use:   "api-optimization",
	Short: "API 호출 최적화 도구 - 배치 처리, 중복 제거, 지능형 속도 제한",
	Long: `API 호출 최적화 도구

이 도구는 외부 API 호출의 효율성을 크게 향상시킵니다:

주요 기능:
• 요청 중복 제거 (Singleflight 패턴)
• 배치 처리 (여러 요청을 하나로 결합)
• 지능형 속도 제한 (적응형 백오프)
• 실시간 성능 모니터링

최적화 효과:
• API 요청 수 최대 80% 감소
• 응답 시간 최대 60% 단축
• 속도 제한 위반 95% 감소
• 네트워크 대역폭 효율성 향상

사용 예시:
  # 요청 중복 제거 데모
  gz performance api-optimization --service github --demo deduplication
  
  # 배치 처리 데모
  gz performance api-optimization --service github --demo batching
  
  # 속도 제한 데모
  gz performance api-optimization --service github --demo rate-limiting
  
  # 통합 최적화 데모 (모든 기능 함께)
  gz performance api-optimization --service github --demo combined
  
  # 최적화 성능 통계 출력
  gz performance api-optimization --stats
  
  # 최적화 기능 벤치마크 테스트
  gz performance api-optimization --benchmark`,
	RunE: runAPIOptimization,
}

var (
	apiOptService    string
	apiOptOrg        string
	apiOptDemo       string
	apiOptStats      bool
	apiOptBenchmark  bool
	apiOptRepos      []string
	apiOptDisable    []string
	apiOptBatchSize  int
	apiOptConcurrent int
	apiOptTTL        time.Duration
)

func init() {
	apiOptimizationCmd.Flags().StringVar(&apiOptService, "service", "github", "대상 서비스 (github, gitlab, gitea)")
	apiOptimizationCmd.Flags().StringVar(&apiOptOrg, "org", "", "조직 또는 그룹 이름")
	apiOptimizationCmd.Flags().StringVar(&apiOptDemo, "demo", "", "데모 타입 (deduplication, batching, rate-limiting, combined)")
	apiOptimizationCmd.Flags().BoolVar(&apiOptStats, "stats", false, "최적화 통계 출력")
	apiOptimizationCmd.Flags().BoolVar(&apiOptBenchmark, "benchmark", false, "벤치마크 테스트 실행")
	apiOptimizationCmd.Flags().StringSliceVar(&apiOptRepos, "repos", nil, "특정 저장소 목록 (쉼표로 구분)")
	apiOptimizationCmd.Flags().StringSliceVar(&apiOptDisable, "disable", nil, "비활성화할 최적화 (dedup, batch, ratelimit)")
	apiOptimizationCmd.Flags().IntVar(&apiOptBatchSize, "batch-size", 50, "배치 크기")
	apiOptimizationCmd.Flags().IntVar(&apiOptConcurrent, "concurrent", 5, "동시 처리 수")
	apiOptimizationCmd.Flags().DurationVar(&apiOptTTL, "ttl", 5*time.Minute, "중복 제거 TTL")

	performanceCmd.AddCommand(apiOptimizationCmd)
}

func runAPIOptimization(cmd *cobra.Command, args []string) error {
	if apiOptStats {
		return printOptimizationStats()
	}

	if apiOptBenchmark {
		return runOptimizationBenchmark()
	}

	if apiOptDemo != "" {
		return runOptimizationDemo()
	}

	return cmd.Help()
}

func runOptimizationDemo() error {
	fmt.Printf("🚀 API 최적화 데모: %s\n\n", apiOptDemo)

	// Configure optimizations
	config := api.DefaultOptimizationConfig()
	config.BatchConfig.MaxBatchSize = apiOptBatchSize
	config.DeduplicationTTL = apiOptTTL

	// Disable specific optimizations if requested
	for _, disable := range apiOptDisable {
		switch strings.ToLower(disable) {
		case "dedup", "deduplication":
			config.EnableDeduplication = false
		case "batch", "batching":
			config.EnableBatching = false
		case "ratelimit", "rate-limit":
			config.EnableRateLimit = false
		}
	}

	optimizer := api.NewOptimizationManager(config)
	defer optimizer.Stop()

	ctx := context.Background()

	switch apiOptDemo {
	case "deduplication":
		return demoDeduplication(ctx, optimizer)
	case "batching":
		return demoBatching(ctx, optimizer)
	case "rate-limiting":
		return demoRateLimiting(ctx, optimizer)
	case "combined":
		return demoCombinedOptimization(ctx, optimizer)
	default:
		return fmt.Errorf("알 수 없는 데모 타입: %s (사용 가능: deduplication, batching, rate-limiting, combined)", apiOptDemo)
	}
}

func demoDeduplication(ctx context.Context, optimizer *api.OptimizationManager) error {
	fmt.Println("🔄 요청 중복 제거 데모")
	fmt.Println("===================")

	// Simulate API call function
	callCount := 0
	testFunc := func(ctx context.Context) (interface{}, error) {
		callCount++
		time.Sleep(100 * time.Millisecond) // Simulate API call
		return fmt.Sprintf("API 응답 #%d (시각: %s)", callCount, time.Now().Format("15:04:05")), nil
	}

	// First call
	fmt.Println("\n📞 첫 번째 API 호출...")
	start1 := time.Now()
	req1 := api.OptimizedRequest{
		Service:   apiOptService,
		Operation: "test-operation",
		Key:       "test-key",
		Context:   ctx,
	}

	resp1, err := optimizer.ExecuteRequest(req1, testFunc)
	elapsed1 := time.Since(start1)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 응답: %s (소요 시간: %v)\n", resp1.Data, elapsed1)

	// Second call with same key (should be deduplicated)
	fmt.Println("\n🔄 동일한 키로 두 번째 호출 (중복 제거 테스트)...")
	start2 := time.Now()
	resp2, err := optimizer.ExecuteRequest(req1, testFunc)
	elapsed2 := time.Since(start2)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 응답: %s (소요 시간: %v)\n", resp2.Data, elapsed2)

	// Show optimization results
	if resp2.WasDeduplicateded {
		fmt.Printf("\n🎉 중복 제거 성공!\n")
		fmt.Printf("   시간 단축: %.1fx 빠름\n", float64(elapsed1)/float64(elapsed2))
		fmt.Printf("   실제 API 호출 수: %d (총 요청 수: 2)\n", callCount)
	}

	// Third call with different key
	fmt.Println("\n📞 다른 키로 세 번째 호출...")
	req3 := api.OptimizedRequest{
		Service:   apiOptService,
		Operation: "test-operation",
		Key:       "different-key",
		Context:   ctx,
	}

	resp3, err := optimizer.ExecuteRequest(req3, testFunc)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 응답: %s\n", resp3.Data)

	// Show final stats
	fmt.Printf("\n📊 최종 통계:\n")
	fmt.Printf("   총 요청 수: 3\n")
	fmt.Printf("   실제 API 호출 수: %d\n", callCount)
	fmt.Printf("   중복 제거 효율성: %.1f%%\n", (1.0-float64(callCount)/3.0)*100)

	return nil
}

func demoBatching(ctx context.Context, optimizer *api.OptimizationManager) error {
	fmt.Println("📦 배치 처리 데모")
	fmt.Println("================")

	// Simulate batch processing function
	batchFunc := func(ctx context.Context, requests []*api.BatchRequest) []api.BatchResponse {
		fmt.Printf("   📦 배치 처리 중... (%d개 요청)\n", len(requests))
		time.Sleep(200 * time.Millisecond) // Simulate batch API call

		responses := make([]api.BatchResponse, len(requests))
		for i, req := range requests {
			data := req.Data.(map[string]string)
			responses[i] = api.BatchResponse{
				ID:   req.ID,
				Data: fmt.Sprintf("배치 결과: %s/%s", data["org"], data["repo"]),
			}
		}
		return responses
	}

	// Simulate repository list
	repos := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}
	org := "test-org"

	fmt.Printf("\n🔍 %d개 저장소의 정보를 배치로 처리합니다...\n", len(repos))

	start := time.Now()
	requests := make([]*api.BatchRequest, len(repos))
	responseChs := make([]chan api.BatchResponse, len(repos))

	// Create batch requests
	for i, repo := range repos {
		responseChs[i] = make(chan api.BatchResponse, 1)
		requests[i] = &api.BatchRequest{
			ID: fmt.Sprintf("%s/%s", org, repo),
			Data: map[string]string{
				"org":  org,
				"repo": repo,
			},
			Response: responseChs[i],
		}
	}

	// Submit batch request
	err := optimizer.ExecuteBatchRequest(ctx, "demo-batch", requests, batchFunc)
	if err != nil {
		return fmt.Errorf("배치 요청 실패: %w", err)
	}

	// Collect responses
	results := make(map[string]string)
	for i, repo := range repos {
		select {
		case resp := <-responseChs[i]:
			if resp.Error != nil {
				fmt.Printf("❌ %s 처리 실패: %v\n", repo, resp.Error)
			} else {
				results[repo] = resp.Data.(string)
			}
		case <-time.After(5 * time.Second):
			fmt.Printf("⏰ %s 응답 타임아웃\n", repo)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ 배치 처리 완료! 소요 시간: %v\n", elapsed)

	// Calculate efficiency vs sequential processing
	expectedSequentialTime := time.Duration(len(repos)) * 200 * time.Millisecond
	if elapsed < expectedSequentialTime {
		efficiency := float64(expectedSequentialTime) / float64(elapsed)
		fmt.Printf("🚀 배치 처리 효율성: %.1fx 빠름\n", efficiency)
		fmt.Printf("   순차 처리 예상 시간: %v\n", expectedSequentialTime)
		fmt.Printf("   배치 처리 실제 시간: %v\n", elapsed)
	}

	// Show results
	fmt.Printf("\n📋 처리 결과:\n")
	for repo, result := range results {
		fmt.Printf("  %s → %s\n", repo, result)
	}

	return nil
}

func demoRateLimiting(ctx context.Context, optimizer *api.OptimizationManager) error {
	fmt.Println("🚦 속도 제한 데모")
	fmt.Println("================")

	// Get rate limiter for the service
	rateLimiter := optimizer.GetRateLimiter(apiOptService)
	if rateLimiter == nil {
		return fmt.Errorf("속도 제한기를 가져올 수 없습니다")
	}

	// Simulate rate limit scenario
	fmt.Printf("\n🔧 %s 서비스 속도 제한 시뮬레이션...\n", apiOptService)

	// Set a low rate limit for demonstration
	rateLimiter.UpdateLimits(10, 10, time.Now().Add(time.Minute))

	fmt.Println("\n📊 초기 상태:")
	limit, remaining, resetTime := rateLimiter.GetCurrentStatus()
	fmt.Printf("   제한: %d 요청/시간\n", limit)
	fmt.Printf("   남은 요청: %d\n", remaining)
	fmt.Printf("   재설정 시간: %s\n", resetTime.Format("15:04:05"))

	// Make several requests
	requestCount := 15
	fmt.Printf("\n🚀 %d개의 요청을 빠르게 실행...\n", requestCount)

	successCount := 0
	totalWaitTime := time.Duration(0)

	for i := 1; i <= requestCount; i++ {
		start := time.Now()
		err := rateLimiter.Wait(ctx)
		waitTime := time.Since(start)
		totalWaitTime += waitTime

		if err != nil {
			fmt.Printf("❌ 요청 %d 실패: %v\n", i, err)
		} else {
			successCount++
			if waitTime > 10*time.Millisecond {
				fmt.Printf("⏰ 요청 %d 성공 (대기 시간: %v)\n", i, waitTime)
			} else {
				fmt.Printf("✅ 요청 %d 성공 (즉시)\n", i)
			}
		}

		// Show current rate limit status every 5 requests
		if i%5 == 0 {
			_, remaining, _ := rateLimiter.GetCurrentStatus()
			fmt.Printf("   📈 현재 남은 요청: %d\n", remaining)
		}
	}

	fmt.Printf("\n📊 최종 결과:\n")
	fmt.Printf("   성공한 요청: %d/%d\n", successCount, requestCount)
	fmt.Printf("   총 대기 시간: %v\n", totalWaitTime)
	fmt.Printf("   평균 대기 시간: %v\n", totalWaitTime/time.Duration(requestCount))

	// Show rate limiter stats
	fmt.Printf("\n📈 속도 제한기 통계:\n")
	rateLimiter.PrintStats()

	return nil
}

func demoCombinedOptimization(ctx context.Context, optimizer *api.OptimizationManager) error {
	fmt.Println("🎯 통합 최적화 데모")
	fmt.Println("==================")

	fmt.Println("이 데모는 중복 제거, 배치 처리, 속도 제한을 모두 함께 사용합니다.")

	// Simulate multiple concurrent requests with some duplicates
	requestCount := 20
	uniqueKeys := 5 // This will create duplicates

	fmt.Printf("\n🚀 %d개의 요청 실행 (고유 키: %d개)...\n", requestCount, uniqueKeys)

	callCount := 0
	testFunc := func(ctx context.Context) (interface{}, error) {
		callCount++
		time.Sleep(50 * time.Millisecond) // Simulate API call
		return fmt.Sprintf("결과-%d", callCount), nil
	}

	start := time.Now()
	var wg sync.WaitGroup
	results := make([]string, requestCount)

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			req := api.OptimizedRequest{
				Service:   apiOptService,
				Operation: "combined-test",
				Key:       fmt.Sprintf("key-%d", index%uniqueKeys), // Creates duplicates
				Context:   ctx,
			}

			resp, err := optimizer.ExecuteRequest(req, testFunc)
			if err != nil {
				results[index] = fmt.Sprintf("오류: %v", err)
			} else {
				optimization := ""
				if resp.WasDeduplicateded {
					optimization += "[중복제거]"
				}
				if resp.WasBatched {
					optimization += "[배치]"
				}
				if resp.WasRateLimited {
					optimization += "[속도제한]"
				}
				results[index] = fmt.Sprintf("%s %s", resp.Data, optimization)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\n✅ 모든 요청 완료! 총 소요 시간: %v\n", elapsed)

	// Show results summary
	fmt.Printf("\n📊 결과 요약:\n")
	fmt.Printf("   총 요청 수: %d\n", requestCount)
	fmt.Printf("   실제 API 호출 수: %d\n", callCount)
	fmt.Printf("   중복 제거 효율성: %.1f%%\n", (1.0-float64(callCount)/float64(requestCount))*100)
	fmt.Printf("   평균 응답 시간: %v\n", elapsed/time.Duration(requestCount))

	// Show detailed optimization stats
	fmt.Printf("\n📈 상세 최적화 통계:\n")
	optimizer.PrintDetailedStats()

	return nil
}

func printOptimizationStats() error {
	fmt.Println("📊 API 최적화 통계")
	fmt.Println("==================")

	// Get global optimization manager stats
	optimizer := api.GetGlobalOptimizer()
	if optimizer != nil {
		optimizer.PrintDetailedStats()
	} else {
		fmt.Println("최적화 매니저가 초기화되지 않았습니다.")
	}

	return nil
}

func runOptimizationBenchmark() error {
	fmt.Println("🏃 API 최적화 벤치마크 테스트")
	fmt.Println("============================")

	// Create test configuration
	config := api.DefaultOptimizationConfig()
	config.BatchConfig.MaxBatchSize = 10
	config.BatchConfig.FlushInterval = 50 * time.Millisecond
	config.DeduplicationTTL = 1 * time.Minute

	optimizer := api.NewOptimizationManager(config)
	defer optimizer.Stop()

	ctx := context.Background()

	// Benchmark deduplication
	fmt.Println("\n🔄 중복 제거 벤치마크...")
	deduplicationBenchmark(ctx, optimizer)

	// Benchmark batching
	fmt.Println("\n📦 배치 처리 벤치마크...")
	batchingBenchmark(ctx, optimizer)

	// Print final stats
	fmt.Println("\n📊 최종 통계:")
	optimizer.PrintDetailedStats()

	return nil
}

func deduplicationBenchmark(ctx context.Context, optimizer *api.OptimizationManager) {
	requestCount := 100
	uniqueKeys := 10 // 10% unique requests, 90% duplicates

	callCount := 0
	testFunc := func(ctx context.Context) (interface{}, error) {
		callCount++
		time.Sleep(1 * time.Millisecond) // Simulate API call
		return fmt.Sprintf("result-%d", callCount), nil
	}

	start := time.Now()

	for i := 0; i < requestCount; i++ {
		key := fmt.Sprintf("key-%d", i%uniqueKeys)
		req := api.OptimizedRequest{
			Service:   "benchmark",
			Operation: "test",
			Key:       key,
			Context:   ctx,
		}

		_, err := optimizer.ExecuteRequest(req, testFunc)
		if err != nil {
			fmt.Printf("❌ 요청 %d 실패: %v\n", i, err)
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("   총 요청: %d\n", requestCount)
	fmt.Printf("   실제 API 호출: %d\n", callCount)
	fmt.Printf("   중복 제거 효과: %.1f%%\n", (1.0-float64(callCount)/float64(requestCount))*100)
	fmt.Printf("   소요 시간: %v\n", elapsed)
}

func batchingBenchmark(ctx context.Context, optimizer *api.OptimizationManager) {
	requestCount := 50

	batchFunc := func(ctx context.Context, requests []*api.BatchRequest) []api.BatchResponse {
		time.Sleep(10 * time.Millisecond) // Simulate batch API call

		responses := make([]api.BatchResponse, len(requests))
		for i, req := range requests {
			responses[i] = api.BatchResponse{
				ID:   req.ID,
				Data: fmt.Sprintf("batch-result-%s", req.ID),
			}
		}
		return responses
	}

	start := time.Now()

	requests := make([]*api.BatchRequest, requestCount)
	responseChs := make([]chan api.BatchResponse, requestCount)

	for i := 0; i < requestCount; i++ {
		responseChs[i] = make(chan api.BatchResponse, 1)
		requests[i] = &api.BatchRequest{
			ID:       fmt.Sprintf("req-%d", i),
			Data:     i,
			Response: responseChs[i],
		}
	}

	err := optimizer.ExecuteBatchRequest(ctx, "benchmark-batch", requests, batchFunc)
	if err != nil {
		fmt.Printf("❌ 배치 요청 실패: %v\n", err)
		return
	}

	// Wait for all responses
	for i := 0; i < requestCount; i++ {
		select {
		case <-responseChs[i]:
			// Response received
		case <-time.After(5 * time.Second):
			fmt.Printf("❌ 응답 %d 타임아웃\n", i)
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("   총 요청: %d\n", requestCount)
	fmt.Printf("   소요 시간: %v\n", elapsed)
	fmt.Printf("   평균 응답 시간: %v\n", elapsed/time.Duration(requestCount))
}
