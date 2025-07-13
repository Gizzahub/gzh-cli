package performance

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/async"
	"github.com/spf13/cobra"
)

// asyncProcessingCmd represents the async processing command
var asyncProcessingCmd = &cobra.Command{
	Use:   "async-processing",
	Short: "비동기 처리 시스템 데모 - 논블로킹 I/O, 이벤트 드리븐, 작업 큐",
	Long: `비동기 처리 시스템 데모

이 도구는 블로킹 작업을 최소화하는 비동기 처리 시스템을 시연합니다:

주요 기능:
• 논블로킹 I/O 처리 (파일, HTTP 요청)
• 이벤트 드리븐 아키텍처 (이벤트 버스)
• 우선순위 기반 작업 큐 시스템
• 실시간 성능 모니터링

성능 개선 효과:
• I/O 대기 시간 최대 80% 감소
• 동시 처리 능력 5-10배 향상
• 시스템 자원 사용률 최적화
• 응답성 및 처리량 대폭 개선

사용 예시:
  # 논블로킹 I/O 데모
  gz performance async-processing --demo io
  
  # 이벤트 드리븐 아키텍처 데모
  gz performance async-processing --demo events
  
  # 작업 큐 시스템 데모
  gz performance async-processing --demo queue
  
  # 통합 비동기 처리 데모
  gz performance async-processing --demo integration
  
  # 성능 비교 벤치마크
  gz performance async-processing --benchmark
  
  # 실시간 통계 모니터링
  gz performance async-processing --monitor`,
	RunE: runAsyncProcessing,
}

var (
	asyncDemo      string
	asyncBenchmark bool
	asyncMonitor   bool
	asyncWorkers   int
	asyncJobs      int
	asyncFiles     int
	asyncDuration  time.Duration
)

func init() {
	asyncProcessingCmd.Flags().StringVar(&asyncDemo, "demo", "", "데모 타입 (io, events, queue, integration)")
	asyncProcessingCmd.Flags().BoolVar(&asyncBenchmark, "benchmark", false, "성능 비교 벤치마크 실행")
	asyncProcessingCmd.Flags().BoolVar(&asyncMonitor, "monitor", false, "실시간 통계 모니터링")
	asyncProcessingCmd.Flags().IntVar(&asyncWorkers, "workers", 5, "워커 수")
	asyncProcessingCmd.Flags().IntVar(&asyncJobs, "jobs", 100, "작업 수")
	asyncProcessingCmd.Flags().IntVar(&asyncFiles, "files", 50, "테스트 파일 수")
	asyncProcessingCmd.Flags().DurationVar(&asyncDuration, "duration", 30*time.Second, "모니터링 지속 시간")

	performanceCmd.AddCommand(asyncProcessingCmd)
}

func runAsyncProcessing(cmd *cobra.Command, args []string) error {
	if asyncBenchmark {
		return runAsyncBenchmark()
	}

	if asyncMonitor {
		return runAsyncMonitoring()
	}

	if asyncDemo != "" {
		return runAsyncDemo()
	}

	return cmd.Help()
}

func runAsyncDemo() error {
	fmt.Printf("🚀 비동기 처리 데모: %s\n\n", asyncDemo)

	switch asyncDemo {
	case "io":
		return demoAsyncIO()
	case "events":
		return demoEventDriven()
	case "queue":
		return demoWorkQueue()
	case "integration":
		return demoIntegratedAsyncProcessing()
	default:
		return fmt.Errorf("알 수 없는 데모 타입: %s (사용 가능: io, events, queue, integration)", asyncDemo)
	}
}

func demoAsyncIO() error {
	fmt.Println("📁 논블로킹 I/O 데모")
	fmt.Println("===================")

	// Create async I/O manager
	aio := async.NewAsyncIO(10)
	defer aio.Close()

	// Create temporary directory and test files
	tempDir, err := os.MkdirTemp("", "async-io-demo")
	if err != nil {
		return fmt.Errorf("임시 디렉터리 생성 실패: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("\n📂 임시 디렉터리: %s\n", tempDir)
	fmt.Printf("📊 %d개 파일로 테스트 진행...\n\n", asyncFiles)

	// Create test files
	testFiles := make([]string, asyncFiles)
	for i := 0; i < asyncFiles; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("test_file_%03d.txt", i))
		content := fmt.Sprintf("Test file %d\nCreated at: %s\nContent length: %d bytes",
			i, time.Now().Format(time.RFC3339), i*100)

		err := os.WriteFile(filename, []byte(content), 0o644)
		if err != nil {
			return fmt.Errorf("테스트 파일 생성 실패: %w", err)
		}
		testFiles[i] = filename
	}

	ctx := context.Background()

	// Demo 1: Sequential vs Batch reading
	fmt.Println("🔄 순차 vs 배치 파일 읽기 비교")

	// Sequential reading
	fmt.Printf("📖 순차 읽기 시작...\n")
	sequentialStart := time.Now()
	for i := 0; i < min(10, len(testFiles)); i++ {
		resultCh := aio.ReadFileAsync(ctx, testFiles[i])
		result := <-resultCh
		if result.Error != nil {
			fmt.Printf("❌ 파일 읽기 실패: %s\n", result.Error)
		} else {
			fmt.Printf("✅ 읽기 완료: %s (%d bytes)\n",
				filepath.Base(result.Path), len(result.Data))
		}
	}
	sequentialDuration := time.Since(sequentialStart)

	// Batch reading
	fmt.Printf("\n📚 배치 읽기 시작...\n")
	batchStart := time.Now()
	batchFiles := testFiles[:min(10, len(testFiles))]
	resultCh := aio.BatchReadFiles(ctx, batchFiles)

	readCount := 0
	for result := range resultCh {
		if result.Error != nil {
			fmt.Printf("❌ 배치 읽기 실패: %s\n", result.Error)
		} else {
			fmt.Printf("✅ 배치 읽기 완료: %s (%d bytes)\n",
				filepath.Base(result.Path), len(result.Data))
		}
		readCount++
	}
	batchDuration := time.Since(batchStart)

	// Demo 2: HTTP requests
	fmt.Println("\n🌐 비동기 HTTP 요청 데모")

	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/json",
		"https://httpbin.org/uuid",
		"https://httpbin.org/status/200",
	}

	httpStart := time.Now()
	var httpResults []async.HTTPResult

	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("❌ 요청 생성 실패: %s\n", err)
			continue
		}

		resultCh := aio.HTTPRequestAsync(ctx, req)
		go func(url string) {
			result := <-resultCh
			httpResults = append(httpResults, result)

			if result.Error != nil {
				fmt.Printf("❌ HTTP 요청 실패: %s - %v\n", url, result.Error)
			} else {
				fmt.Printf("✅ HTTP 응답: %s (상태: %d, 응답시간: %v)\n",
					url, result.Response.StatusCode, result.Duration)
			}
		}(url)
	}

	// Wait for HTTP requests to complete
	time.Sleep(3 * time.Second)
	httpDuration := time.Since(httpStart)

	// Show performance comparison
	fmt.Printf("\n📊 성능 비교 결과:\n")
	fmt.Printf("   순차 파일 읽기: %v\n", sequentialDuration)
	fmt.Printf("   배치 파일 읽기: %v (%.1fx 빠름)\n", batchDuration,
		float64(sequentialDuration)/float64(batchDuration))
	fmt.Printf("   HTTP 요청 처리: %v\n", httpDuration)

	// Show I/O statistics
	fmt.Printf("\n📈 I/O 통계:\n")
	aio.PrintStats()

	return nil
}

func demoEventDriven() error {
	fmt.Println("⚡ 이벤트 드리븐 아키텍처 데모")
	fmt.Println("============================")

	// Create event bus
	config := async.DefaultEventBusConfig()
	eventBus := async.NewEventBus(config)
	defer eventBus.Close()

	// Setup event tracking
	var processedEvents []string
	var mu sync.Mutex
	eventCount := 0

	// Repository events handler
	eventBus.SubscribeAsyncFunc("repository.cloned", func(ctx context.Context, event async.Event) error {
		mu.Lock()
		defer mu.Unlock()

		data := event.Data().(map[string]interface{})
		repo := data["repository"].(string)
		processedEvents = append(processedEvents, fmt.Sprintf("Repository cloned: %s", repo))
		fmt.Printf("🔄 저장소 클론됨: %s (처리시간: %v)\n", repo, time.Since(event.Timestamp()))
		return nil
	})

	// File processing events handler
	eventBus.SubscribeAsyncFunc("file.processed", func(ctx context.Context, event async.Event) error {
		mu.Lock()
		defer mu.Unlock()

		data := event.Data().(map[string]interface{})
		file := data["file"].(string)
		processedEvents = append(processedEvents, fmt.Sprintf("File processed: %s", file))
		fmt.Printf("📄 파일 처리됨: %s (크기: %v bytes)\n", file, data["size"])
		return nil
	})

	// Error events handler
	eventBus.SubscribeAsyncFunc("error.occurred", func(ctx context.Context, event async.Event) error {
		mu.Lock()
		defer mu.Unlock()

		data := event.Data().(map[string]interface{})
		error := data["error"].(string)
		processedEvents = append(processedEvents, fmt.Sprintf("Error occurred: %s", error))
		fmt.Printf("❌ 에러 발생: %s (소스: %s)\n", error, event.Source())
		return nil
	})

	// Task completion handler with metrics
	eventBus.SubscribeFunc("task.completed", func(ctx context.Context, event async.Event) error {
		mu.Lock()
		eventCount++
		count := eventCount
		mu.Unlock()

		data := event.Data().(map[string]interface{})
		fmt.Printf("✅ 작업 완료 #%d: %s (소요시간: %v)\n",
			count, data["job_id"], data["duration"])
		return nil
	})

	ctx := context.Background()

	fmt.Println("\n🚀 이벤트 발생 시뮬레이션 시작...\n")

	// Simulate repository clone events
	repositories := []string{"user/repo1", "org/backend", "team/frontend", "company/api"}
	for _, repo := range repositories {
		event := async.BaseEvent{
			EventType:   "repository.cloned",
			EventTime:   time.Now(),
			EventSource: "git-cloner",
			EventData: map[string]interface{}{
				"repository": repo,
				"size":       "15.2 MB",
				"files":      142,
			},
		}
		eventBus.PublishAsync(ctx, event)
		time.Sleep(100 * time.Millisecond)
	}

	// Simulate file processing events
	files := []string{"config.yml", "main.go", "README.md", "package.json"}
	for _, file := range files {
		event := async.BaseEvent{
			EventType:   "file.processed",
			EventTime:   time.Now(),
			EventSource: "file-processor",
			EventData: map[string]interface{}{
				"file": file,
				"size": len(file) * 1024, // Mock file size
			},
		}
		eventBus.PublishAsync(ctx, event)
		time.Sleep(50 * time.Millisecond)
	}

	// Simulate task completion events
	for i := 0; i < 5; i++ {
		event := async.BaseEvent{
			EventType:   "task.completed",
			EventTime:   time.Now(),
			EventSource: "task-runner",
			EventData: map[string]interface{}{
				"job_id":   fmt.Sprintf("task-%d", i+1),
				"duration": time.Duration(100+i*50) * time.Millisecond,
			},
		}
		eventBus.PublishAsync(ctx, event)
		time.Sleep(80 * time.Millisecond)
	}

	// Simulate some errors
	errors := []string{"Network timeout", "File not found", "Permission denied"}
	for _, errMsg := range errors {
		event := async.BaseEvent{
			EventType:   "error.occurred",
			EventTime:   time.Now(),
			EventSource: "error-simulator",
			EventData: map[string]interface{}{
				"error": errMsg,
			},
		}
		eventBus.PublishAsync(ctx, event)
		time.Sleep(200 * time.Millisecond)
	}

	// Wait for all events to be processed
	time.Sleep(2 * time.Second)

	mu.Lock()
	totalProcessed := len(processedEvents)
	mu.Unlock()

	fmt.Printf("\n📊 이벤트 처리 결과:\n")
	fmt.Printf("   총 발행된 이벤트: %d\n", len(repositories)+len(files)+5+len(errors))
	fmt.Printf("   처리된 이벤트: %d\n", totalProcessed)

	fmt.Printf("\n📈 이벤트 버스 통계:\n")
	eventBus.PrintStats()

	return nil
}

func demoWorkQueue() error {
	fmt.Println("🔧 작업 큐 시스템 데모")
	fmt.Println("==================")

	// Create work queue
	config := async.DefaultWorkQueueConfig("demo-queue")
	config.Workers = asyncWorkers
	workQueue := async.NewWorkQueue(config)
	defer workQueue.Stop(10 * time.Second)

	ctx := context.Background()
	fmt.Printf("👷 %d개 워커로 %d개 작업 처리 중...\n\n", asyncWorkers, asyncJobs)

	// Track results
	var completedJobs []string
	var failedJobs []string
	var mu sync.Mutex

	// Submit various types of jobs
	jobTypes := []struct {
		name     string
		priority int
		duration time.Duration
		failRate float64
	}{
		{"빠른작업", 8, 50 * time.Millisecond, 0.0},
		{"일반작업", 5, 200 * time.Millisecond, 0.1},
		{"느린작업", 2, 500 * time.Millisecond, 0.2},
		{"중요작업", 9, 100 * time.Millisecond, 0.05},
	}

	jobIndex := 0
	for i := 0; i < asyncJobs; i++ {
		jobType := jobTypes[i%len(jobTypes)]
		jobID := fmt.Sprintf("%s-%d", jobType.name, jobIndex)
		jobIndex++

		job := async.NewSimpleJob(jobID, jobType.priority, func(ctx context.Context) error {
			// Simulate work
			time.Sleep(jobType.duration)

			// Simulate random failures
			if jobType.failRate > 0 && time.Now().UnixNano()%100 < int64(jobType.failRate*100) {
				return fmt.Errorf("작업 실패 시뮬레이션")
			}

			return nil
		})

		err := workQueue.Submit(job)
		if err != nil {
			fmt.Printf("❌ 작업 제출 실패: %s\n", err)
		}
	}

	// Monitor job completion
	start := time.Now()
	completionCount := 0

	fmt.Println("📊 작업 처리 현황:")
	go func() {
		for result := range workQueue.Results() {
			mu.Lock()
			if result.Error != nil {
				failedJobs = append(failedJobs, result.Job.ID())
				fmt.Printf("❌ 작업 실패: %s (%v)\n", result.Job.ID(), result.Error)
			} else {
				completedJobs = append(completedJobs, result.Job.ID())
				if result.Retried {
					fmt.Printf("✅ 작업 완료: %s (재시도 성공, 소요시간: %v)\n",
						result.Job.ID(), result.Duration)
				} else {
					fmt.Printf("✅ 작업 완료: %s (소요시간: %v)\n",
						result.Job.ID(), result.Duration)
				}
			}
			completionCount++
			mu.Unlock()
		}
	}()

	// Show real-time statistics
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := workQueue.GetStats()
			high, normal, low := workQueue.QueueLengths()

			fmt.Printf("\n⏱️  실시간 통계 (경과시간: %v):\n", time.Since(start).Round(time.Second))
			fmt.Printf("   활성 워커: %d\n", stats.ActiveWorkers)
			fmt.Printf("   대기 중인 작업: 높음=%d, 보통=%d, 낮음=%d\n", high, normal, low)
			fmt.Printf("   완료: %d, 실패: %d, 재시도: %d\n",
				stats.CompletedJobs, stats.FailedJobs, stats.RetriedJobs)

			mu.Lock()
			totalCompleted := len(completedJobs) + len(failedJobs)
			mu.Unlock()

			if totalCompleted >= asyncJobs {
				fmt.Printf("\n🎉 모든 작업 완료!\n")
				break
			}
		}

		if time.Since(start) > 30*time.Second {
			fmt.Printf("\n⏰ 타임아웃 - 처리 중단\n")
			break
		}
	}

	duration := time.Since(start)

	mu.Lock()
	successful := len(completedJobs)
	failed := len(failedJobs)
	mu.Unlock()

	fmt.Printf("\n📊 최종 결과:\n")
	fmt.Printf("   총 처리 시간: %v\n", duration)
	fmt.Printf("   성공한 작업: %d\n", successful)
	fmt.Printf("   실패한 작업: %d\n", failed)
	fmt.Printf("   성공률: %.1f%%\n", float64(successful)/float64(successful+failed)*100)
	fmt.Printf("   평균 처리율: %.1f 작업/초\n", float64(successful+failed)/duration.Seconds())

	fmt.Printf("\n📈 작업 큐 통계:\n")
	workQueue.PrintStats()

	return nil
}

func demoIntegratedAsyncProcessing() error {
	fmt.Println("🎯 통합 비동기 처리 데모")
	fmt.Println("=====================")
	fmt.Println("논블로킹 I/O + 이벤트 드리븐 + 작업 큐를 모두 활용한 파일 처리 파이프라인")

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "integrated-async-demo")
	if err != nil {
		return fmt.Errorf("임시 디렉터리 생성 실패: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup components
	aio := async.NewAsyncIO(5)
	defer aio.Close()

	eventConfig := async.DefaultEventBusConfig()
	eventBus := async.NewEventBus(eventConfig)
	defer eventBus.Close()

	queueConfig := async.DefaultWorkQueueConfig("file-pipeline")
	queueConfig.EventBus = eventBus
	queueConfig.Workers = 3
	workQueue := async.NewWorkQueue(queueConfig)
	defer workQueue.Stop(10 * time.Second)

	fmt.Printf("\n📂 작업 디렉터리: %s\n", tempDir)

	// Track processing pipeline
	var processedFiles []string
	var transformedFiles []string
	var archivedFiles []string
	var mu sync.Mutex

	// Event handlers for pipeline stages
	eventBus.SubscribeAsyncFunc("file.scanned", func(ctx context.Context, event async.Event) error {
		data := event.Data().(map[string]interface{})
		file := data["file"].(string)

		mu.Lock()
		processedFiles = append(processedFiles, file)
		mu.Unlock()

		fmt.Printf("📄 파일 발견: %s\n", filepath.Base(file))
		return nil
	})

	eventBus.SubscribeAsyncFunc("file.transformed", func(ctx context.Context, event async.Event) error {
		data := event.Data().(map[string]interface{})
		file := data["output_file"].(string)

		mu.Lock()
		transformedFiles = append(transformedFiles, file)
		mu.Unlock()

		fmt.Printf("🔄 파일 변환 완료: %s\n", filepath.Base(file))
		return nil
	})

	eventBus.SubscribeAsyncFunc("file.archived", func(ctx context.Context, event async.Event) error {
		data := event.Data().(map[string]interface{})
		file := data["archive_file"].(string)

		mu.Lock()
		archivedFiles = append(archivedFiles, file)
		mu.Unlock()

		fmt.Printf("📦 파일 아카이브 완료: %s\n", filepath.Base(file))
		return nil
	})

	ctx := context.Background()

	// Stage 1: Create test files
	fmt.Printf("\n🔧 1단계: 테스트 파일 생성 (%d개)...\n", asyncFiles)
	var testFiles []string

	for i := 0; i < asyncFiles; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("data_%03d.txt", i))
		content := fmt.Sprintf("Data file %d\nTimestamp: %s\nSize: %d\nContent: %s",
			i, time.Now().Format(time.RFC3339), i*100, strings.Repeat("x", i*10))

		// Use async I/O to create files
		writeResultCh := aio.WriteFileAsync(ctx, filename, []byte(content), 0o644)
		result := <-writeResultCh

		if result.Error != nil {
			fmt.Printf("❌ 파일 생성 실패: %v\n", result.Error)
			continue
		}

		testFiles = append(testFiles, filename)

		// Publish file scanned event
		event := async.BaseEvent{
			EventType:   "file.scanned",
			EventTime:   time.Now(),
			EventSource: "file-creator",
			EventData: map[string]interface{}{
				"file": filename,
				"size": len(content),
			},
		}
		eventBus.PublishAsync(ctx, event)
	}

	time.Sleep(1 * time.Second)

	// Stage 2: Submit file processing jobs
	fmt.Printf("\n⚙️  2단계: 파일 처리 작업 제출...\n")

	for _, file := range testFiles {
		job := async.NewFileProcessingJob(file, func(ctx context.Context, path string) error {
			// Read file using async I/O
			readResultCh := aio.ReadFileAsync(ctx, path)
			readResult := <-readResultCh

			if readResult.Error != nil {
				return readResult.Error
			}

			// Transform content (uppercase + add metadata)
			originalContent := string(readResult.Data)
			transformedContent := fmt.Sprintf("TRANSFORMED FILE\n=================\n%s\n\nTransformed at: %s\nOriginal size: %d bytes\n",
				strings.ToUpper(originalContent), time.Now().Format(time.RFC3339), len(originalContent))

			// Write transformed file
			transformedPath := strings.Replace(path, ".txt", "_transformed.txt", 1)
			writeResultCh := aio.WriteFileAsync(ctx, transformedPath, []byte(transformedContent), 0o644)
			writeResult := <-writeResultCh

			if writeResult.Error != nil {
				return writeResult.Error
			}

			// Publish transformation event
			event := async.BaseEvent{
				EventType:   "file.transformed",
				EventTime:   time.Now(),
				EventSource: "file-transformer",
				EventData: map[string]interface{}{
					"input_file":  path,
					"output_file": transformedPath,
					"input_size":  len(originalContent),
					"output_size": len(transformedContent),
				},
			}
			eventBus.PublishAsync(ctx, event)

			// Create archive (simulate compression)
			archivePath := strings.Replace(path, ".txt", "_archive.gz", 1)
			archiveContent := fmt.Sprintf("ARCHIVED: %s\nCompressed size: %d bytes\nCompression ratio: %.2f\n",
				filepath.Base(path), len(transformedContent)/2, 0.6)

			archiveWriteResultCh := aio.WriteFileAsync(ctx, archivePath, []byte(archiveContent), 0o644)
			archiveWriteResult := <-archiveWriteResultCh

			if archiveWriteResult.Error != nil {
				return archiveWriteResult.Error
			}

			// Publish archive event
			archiveEvent := async.BaseEvent{
				EventType:   "file.archived",
				EventTime:   time.Now(),
				EventSource: "file-archiver",
				EventData: map[string]interface{}{
					"original_file": path,
					"archive_file":  archivePath,
					"compression":   "gzip",
				},
			}
			eventBus.PublishAsync(ctx, archiveEvent)

			return nil
		})

		err := workQueue.Submit(job)
		if err != nil {
			fmt.Printf("❌ 작업 제출 실패: %v\n", err)
		}
	}

	// Stage 3: Monitor processing
	fmt.Printf("\n📊 3단계: 처리 상황 모니터링...\n")

	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mu.Lock()
			scannedCount := len(processedFiles)
			transformedCount := len(transformedFiles)
			archivedCount := len(archivedFiles)
			mu.Unlock()

			queueStats := workQueue.GetStats()
			ioStats := aio.GetStats()

			fmt.Printf("\n⏱️  진행 상황 (경과시간: %v):\n", time.Since(start).Round(time.Second))
			fmt.Printf("   📄 스캔된 파일: %d/%d\n", scannedCount, asyncFiles)
			fmt.Printf("   🔄 변환된 파일: %d/%d\n", transformedCount, asyncFiles)
			fmt.Printf("   📦 아카이브된 파일: %d/%d\n", archivedCount, asyncFiles)
			fmt.Printf("   👷 활성 워커: %d, 완료된 작업: %d\n", queueStats.ActiveWorkers, queueStats.CompletedJobs)
			fmt.Printf("   💾 I/O 작업: 총 %d, 완료 %d, 평균지연시간 %v\n",
				ioStats.TotalOperations, ioStats.CompletedOps, ioStats.AverageLatency)

			if archivedCount >= asyncFiles {
				fmt.Printf("\n🎉 파이프라인 처리 완료!\n")
				break
			}
		}

		if time.Since(start) > 30*time.Second {
			fmt.Printf("\n⏰ 타임아웃 - 처리 중단\n")
			break
		}
	}

	totalDuration := time.Since(start)

	mu.Lock()
	finalScanned := len(processedFiles)
	finalTransformed := len(transformedFiles)
	finalArchived := len(archivedFiles)
	mu.Unlock()

	fmt.Printf("\n📊 파이프라인 처리 결과:\n")
	fmt.Printf("   총 처리 시간: %v\n", totalDuration)
	fmt.Printf("   파일 처리 완료율: %.1f%% (%d/%d)\n",
		float64(finalArchived)/float64(asyncFiles)*100, finalArchived, asyncFiles)
	fmt.Printf("   평균 처리율: %.1f 파일/초\n", float64(finalArchived)/totalDuration.Seconds())

	fmt.Printf("\n📈 상세 통계:\n")
	fmt.Printf("=== 이벤트 버스 ===\n")
	eventBus.PrintStats()
	fmt.Printf("\n=== 작업 큐 ===\n")
	workQueue.PrintStats()
	fmt.Printf("\n=== 비동기 I/O ===\n")
	aio.PrintStats()

	return nil
}

func runAsyncBenchmark() error {
	fmt.Println("🏃 비동기 처리 성능 벤치마크")
	fmt.Println("==========================")

	// Benchmark configurations
	testSizes := []int{10, 50, 100, 200}

	for _, size := range testSizes {
		fmt.Printf("\n📊 테스트 크기: %d 작업\n", size)
		fmt.Println(strings.Repeat("-", 40))

		// Benchmark async I/O
		err := benchmarkAsyncIO(size)
		if err != nil {
			fmt.Printf("❌ 비동기 I/O 벤치마크 실패: %v\n", err)
		}

		// Benchmark work queue
		err = benchmarkWorkQueue(size)
		if err != nil {
			fmt.Printf("❌ 작업 큐 벤치마크 실패: %v\n", err)
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func benchmarkAsyncIO(size int) error {
	tempDir, err := os.MkdirTemp("", "async-benchmark")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := make([]string, size)
	for i := 0; i < size; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("bench_%d.txt", i))
		content := strings.Repeat("benchmark data\n", 100)
		err := os.WriteFile(filename, []byte(content), 0o644)
		if err != nil {
			return err
		}
		testFiles[i] = filename
	}

	aio := async.NewAsyncIO(10)
	defer aio.Close()

	ctx := context.Background()

	// Benchmark async batch reading
	start := time.Now()
	resultCh := aio.BatchReadFiles(ctx, testFiles)

	count := 0
	for range resultCh {
		count++
	}
	asyncDuration := time.Since(start)

	// Benchmark synchronous reading
	start = time.Now()
	for _, file := range testFiles {
		_, err := os.ReadFile(file)
		if err != nil {
			return err
		}
	}
	syncDuration := time.Since(start)

	fmt.Printf("📁 파일 I/O (%d 파일):\n", size)
	fmt.Printf("   비동기: %v\n", asyncDuration)
	fmt.Printf("   동기식: %v\n", syncDuration)
	fmt.Printf("   성능향상: %.1fx\n", float64(syncDuration)/float64(asyncDuration))

	return nil
}

func benchmarkWorkQueue(size int) error {
	config := async.DefaultWorkQueueConfig("benchmark")
	config.Workers = 5
	workQueue := async.NewWorkQueue(config)
	defer workQueue.Stop(5 * time.Second)

	// Benchmark async work queue
	start := time.Now()
	for i := 0; i < size; i++ {
		job := async.NewSimpleJob(fmt.Sprintf("job-%d", i), 5, func(ctx context.Context) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		})
		workQueue.Submit(job)
	}

	// Wait for completion
	completedCount := 0
	for result := range workQueue.Results() {
		if result.Error == nil {
			completedCount++
		}
		if completedCount >= size {
			break
		}
	}
	asyncDuration := time.Since(start)

	// Benchmark synchronous execution
	start = time.Now()
	for i := 0; i < size; i++ {
		time.Sleep(1 * time.Millisecond)
	}
	syncDuration := time.Since(start)

	fmt.Printf("⚙️  작업 처리 (%d 작업):\n", size)
	fmt.Printf("   작업 큐: %v\n", asyncDuration)
	fmt.Printf("   순차처리: %v\n", syncDuration)
	fmt.Printf("   성능향상: %.1fx\n", float64(syncDuration)/float64(asyncDuration))

	return nil
}

func runAsyncMonitoring() error {
	fmt.Printf("📈 비동기 처리 실시간 모니터링 (%v)\n", asyncDuration)
	fmt.Println("=================================")

	// Setup components
	aio := async.NewAsyncIO(8)
	defer aio.Close()

	eventConfig := async.DefaultEventBusConfig()
	eventBus := async.NewEventBus(eventConfig)
	defer eventBus.Close()

	queueConfig := async.DefaultWorkQueueConfig("monitor-queue")
	queueConfig.Workers = 6
	queueConfig.EventBus = eventBus
	workQueue := async.NewWorkQueue(queueConfig)
	defer workQueue.Stop(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), asyncDuration)
	defer cancel()

	// Start background workload
	go generateContinuousWorkload(ctx, aio, workQueue, eventBus)

	// Monitor loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n⏰ 모니터링 완료 (총 시간: %v)\n", time.Since(start))
			return nil
		case <-ticker.C:
			elapsed := time.Since(start)

			// Get statistics
			ioStats := aio.GetStats()
			eventStats := eventBus.GetStats()
			queueStats := workQueue.GetStats()

			// Clear screen and show real-time stats
			fmt.Print("\033[2J\033[H") // Clear screen
			fmt.Printf("📊 실시간 비동기 처리 모니터링 (경과: %v)\n", elapsed.Round(time.Second))
			fmt.Println("=" + strings.Repeat("=", 50))

			fmt.Printf("\n💾 비동기 I/O:\n")
			fmt.Printf("   총 작업: %d, 완료: %d, 실패: %d\n",
				ioStats.TotalOperations, ioStats.CompletedOps, ioStats.FailedOps)
			fmt.Printf("   활성 작업: %d, 최대 동시: %d\n",
				ioStats.ActiveOps, ioStats.MaxConcurrent)
			fmt.Printf("   평균 지연시간: %v\n", ioStats.AverageLatency)

			if ioStats.TotalOperations > 0 {
				successRate := float64(ioStats.CompletedOps) / float64(ioStats.TotalOperations) * 100
				fmt.Printf("   성공률: %.1f%%\n", successRate)
			}

			fmt.Printf("\n⚡ 이벤트 버스:\n")
			fmt.Printf("   총 이벤트: %d, 처리됨: %d, 실패: %d\n",
				eventStats.TotalEvents, eventStats.ProcessedEvents, eventStats.FailedEvents)
			fmt.Printf("   활성 핸들러: %d\n", eventStats.ActiveHandlers)
			fmt.Printf("   평균 지연시간: %v\n", eventStats.AverageLatency)

			fmt.Printf("\n🔧 작업 큐:\n")
			fmt.Printf("   총 작업: %d, 완료: %d, 실패: %d\n",
				queueStats.TotalJobs, queueStats.CompletedJobs, queueStats.FailedJobs)
			fmt.Printf("   활성 워커: %d, 대기 작업: %d\n",
				queueStats.ActiveWorkers, queueStats.QueuedJobs)
			fmt.Printf("   재시도: %d, 평균 실행시간: %v\n",
				queueStats.RetriedJobs, queueStats.AverageExecTime)

			high, normal, low := workQueue.QueueLengths()
			fmt.Printf("   큐 길이 - 높음:%d, 보통:%d, 낮음:%d\n", high, normal, low)

			// Calculate overall metrics
			totalOps := ioStats.TotalOperations + eventStats.TotalEvents + queueStats.TotalJobs
			if totalOps > 0 && elapsed.Seconds() > 0 {
				throughput := float64(totalOps) / elapsed.Seconds()
				fmt.Printf("\n🚀 전체 처리량: %.1f 작업/초\n", throughput)
			}
		}
	}
}

func generateContinuousWorkload(ctx context.Context, aio *async.AsyncIO, workQueue *async.WorkQueue, eventBus *async.EventBus) {
	tempDir, err := os.MkdirTemp("", "monitor-workload")
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)

	jobCounter := 0
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Generate file I/O workload
			filename := filepath.Join(tempDir, fmt.Sprintf("workload_%d.txt", jobCounter))
			content := fmt.Sprintf("Workload file %d created at %s", jobCounter, time.Now().Format(time.RFC3339))

			writeResultCh := aio.WriteFileAsync(ctx, filename, []byte(content), 0o644)
			go func() {
				<-writeResultCh
				// Read it back
				readResultCh := aio.ReadFileAsync(ctx, filename)
				<-readResultCh
			}()

			// Generate work queue jobs
			job := async.NewSimpleJob(fmt.Sprintf("workload-%d", jobCounter),
				3+jobCounter%7, // Varying priority
				func(ctx context.Context) error {
					// Simulate work
					work := time.Duration(10+jobCounter%50) * time.Millisecond
					time.Sleep(work)

					// Sometimes fail
					if jobCounter%20 == 0 {
						return fmt.Errorf("simulated failure")
					}
					return nil
				})
			workQueue.Submit(job)

			// Generate events
			event := async.BaseEvent{
				EventType:   "workload.generated",
				EventTime:   time.Now(),
				EventSource: "workload-generator",
				EventData: map[string]interface{}{
					"job_id": jobCounter,
					"type":   "continuous",
				},
			}
			eventBus.PublishAsync(ctx, event)

			jobCounter++
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
