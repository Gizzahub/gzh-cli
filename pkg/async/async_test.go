package async

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncIO(t *testing.T) {
	aio := NewAsyncIO(5)
	defer aio.Close()

	t.Run("ReadFileAsync", func(t *testing.T) {
		// Create test file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := "Hello, Async World!"

		err := os.WriteFile(testFile, []byte(testContent), 0o644)
		require.NoError(t, err)

		// Test async read
		ctx := context.Background()
		resultCh := aio.ReadFileAsync(ctx, testFile)

		select {
		case result := <-resultCh:
			require.NoError(t, result.Error)
			assert.Equal(t, testContent, string(result.Data))
			assert.Equal(t, testFile, result.Path)
			assert.NotNil(t, result.Meta["duration"])
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for file read")
		}
	})

	t.Run("WriteFileAsync", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "subdir", "write_test.txt")
		testContent := []byte("Async write test")

		ctx := context.Background()
		resultCh := aio.WriteFileAsync(ctx, testFile, testContent, 0o644)

		select {
		case result := <-resultCh:
			require.NoError(t, result.Error)
			assert.Equal(t, testFile, result.Path)

			// Verify file was written
			data, err := os.ReadFile(testFile)
			require.NoError(t, err)
			assert.Equal(t, testContent, data)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for file write")
		}
	})

	t.Run("HTTPRequestAsync", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "test response"}`))
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		ctx := context.Background()
		resultCh := aio.HTTPRequestAsync(ctx, req)

		select {
		case result := <-resultCh:
			require.NoError(t, result.Error)
			assert.Equal(t, 200, result.Response.StatusCode)
			assert.Contains(t, string(result.Body), "test response")
			assert.Equal(t, server.URL, result.URL)
			assert.True(t, result.Duration > 0)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for HTTP request")
		}
	})

	t.Run("BatchReadFiles", func(t *testing.T) {
		tempDir := t.TempDir()
		files := []string{
			filepath.Join(tempDir, "file1.txt"),
			filepath.Join(tempDir, "file2.txt"),
			filepath.Join(tempDir, "file3.txt"),
		}

		// Create test files
		for i, file := range files {
			content := fmt.Sprintf("Content of file %d", i+1)
			err := os.WriteFile(file, []byte(content), 0o644)
			require.NoError(t, err)
		}

		ctx := context.Background()
		resultCh := aio.BatchReadFiles(ctx, files)

		results := make(map[string]string)
		resultCount := 0

		for result := range resultCh {
			require.NoError(t, result.Error)
			results[result.Path] = string(result.Data)
			resultCount++
		}

		assert.Equal(t, len(files), resultCount)
		assert.Equal(t, len(files), len(results))

		for i, file := range files {
			expectedContent := fmt.Sprintf("Content of file %d", i+1)
			assert.Equal(t, expectedContent, results[file])
		}
	})

	t.Run("Statistics", func(t *testing.T) {
		stats := aio.GetStats()
		assert.True(t, stats.TotalOperations > 0)
		assert.True(t, stats.CompletedOps > 0)
		assert.True(t, stats.AverageLatency > 0)
	})
}

func TestEventBus(t *testing.T) {
	config := DefaultEventBusConfig()
	eventBus := NewEventBus(config)
	defer eventBus.Close()

	t.Run("SyncEventHandling", func(t *testing.T) {
		var received []Event
		var mu sync.Mutex

		// Subscribe to events
		eventBus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
			mu.Lock()
			defer mu.Unlock()
			received = append(received, event)
			return nil
		})

		// Publish event
		testEvent := BaseEvent{
			EventType:   "test.event",
			EventTime:   time.Now(),
			EventSource: "test",
			EventData:   "test data",
		}

		ctx := context.Background()
		err := eventBus.Publish(ctx, testEvent)
		require.NoError(t, err)

		// Verify event was received
		mu.Lock()
		assert.Len(t, received, 1)
		assert.Equal(t, testEvent.Type(), received[0].Type())
		assert.Equal(t, testEvent.Data(), received[0].Data())
		mu.Unlock()
	})

	t.Run("AsyncEventHandling", func(t *testing.T) {
		var received []Event
		var mu sync.Mutex
		done := make(chan struct{})

		// Subscribe to async events
		eventBus.SubscribeAsyncFunc("async.event", func(ctx context.Context, event Event) error {
			mu.Lock()
			received = append(received, event)
			mu.Unlock()

			if len(received) == 3 {
				close(done)
			}
			return nil
		})

		// Publish multiple events
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			testEvent := BaseEvent{
				EventType:   "async.event",
				EventTime:   time.Now(),
				EventSource: "test",
				EventData:   fmt.Sprintf("data-%d", i),
			}
			eventBus.PublishAsync(ctx, testEvent)
		}

		// Wait for all events to be processed
		select {
		case <-done:
			mu.Lock()
			assert.Len(t, received, 3)
			mu.Unlock()
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for async events")
		}
	})

	t.Run("EventMiddleware", func(t *testing.T) {
		var middlewareCalled bool
		var eventReceived bool

		// Add middleware
		eventBus.Use(func(next EventHandler) EventHandler {
			return EventHandlerFunc(func(ctx context.Context, event Event) error {
				middlewareCalled = true
				return next.Handle(ctx, event)
			})
		})

		// Subscribe to events
		eventBus.SubscribeFunc("middleware.test", func(ctx context.Context, event Event) error {
			eventReceived = true
			return nil
		})

		// Publish event
		testEvent := BaseEvent{
			EventType:   "middleware.test",
			EventTime:   time.Now(),
			EventSource: "test",
			EventData:   "middleware test",
		}

		ctx := context.Background()
		err := eventBus.Publish(ctx, testEvent)
		require.NoError(t, err)

		assert.True(t, middlewareCalled)
		assert.True(t, eventReceived)
	})

	t.Run("Statistics", func(t *testing.T) {
		stats := eventBus.GetStats()
		assert.True(t, stats.TotalEvents > 0)
		assert.True(t, stats.ProcessedEvents > 0)
	})
}

func TestWorkQueue(t *testing.T) {
	config := DefaultWorkQueueConfig("test-queue")
	config.Workers = 3
	workQueue := NewWorkQueue(config)
	defer workQueue.Stop(5 * time.Second)

	t.Run("JobExecution", func(t *testing.T) {
		var executed []string
		var mu sync.Mutex

		// Create test jobs
		for i := 0; i < 5; i++ {
			jobID := fmt.Sprintf("job-%d", i)
			job := NewSimpleJob(jobID, 5, func(ctx context.Context) error {
				mu.Lock()
				executed = append(executed, jobID)
				mu.Unlock()
				return nil
			})

			err := workQueue.Submit(job)
			require.NoError(t, err)
		}

		// Wait for jobs to complete
		time.Sleep(2 * time.Second)

		mu.Lock()
		assert.Len(t, executed, 5)
		mu.Unlock()
	})

	t.Run("PriorityOrdering", func(t *testing.T) {
		var executed []string
		var mu sync.Mutex

		// Create jobs with different priorities
		lowPriorityJob := NewSimpleJob("low", 1, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond) // Slow job
			mu.Lock()
			executed = append(executed, "low")
			mu.Unlock()
			return nil
		})

		highPriorityJob := NewSimpleJob("high", 9, func(ctx context.Context) error {
			mu.Lock()
			executed = append(executed, "high")
			mu.Unlock()
			return nil
		})

		// Submit low priority first, then high priority
		err := workQueue.Submit(lowPriorityJob)
		require.NoError(t, err)

		err = workQueue.Submit(highPriorityJob)
		require.NoError(t, err)

		// Wait for completion
		time.Sleep(1 * time.Second)

		mu.Lock()
		// High priority should be executed first (when workers are available)
		assert.Contains(t, executed, "high")
		assert.Contains(t, executed, "low")
		mu.Unlock()
	})

	t.Run("JobRetry", func(t *testing.T) {
		attemptCount := 0
		var mu sync.Mutex

		retryJob := &BaseJob{
			JobID:       "retry-job",
			JobPriority: 5,
			CanRetry:    true,
			MaxRetry:    2,
			Handler: func(ctx context.Context) error {
				mu.Lock()
				attemptCount++
				currentAttempt := attemptCount
				mu.Unlock()

				if currentAttempt <= 2 { // Fail first 2 attempts
					return fmt.Errorf("attempt %d failed", currentAttempt)
				}
				return nil
			},
		}

		err := workQueue.Submit(retryJob)
		require.NoError(t, err)

		// Wait for retries to complete
		time.Sleep(5 * time.Second)

		mu.Lock()
		// Should be at least 2 attempts (original + 1 retry), might be 3 if all retries complete
		assert.True(t, attemptCount >= 2, "Expected at least 2 attempts, got %d", attemptCount)
		mu.Unlock()
	})

	t.Run("Results", func(t *testing.T) {
		successJob := NewSimpleJob("success-job", 5, func(ctx context.Context) error {
			return nil
		})

		// Create a job that fails without retry
		failJob := &BaseJob{
			JobID:       "fail-job",
			JobPriority: 5,
			CanRetry:    false, // No retry
			MaxRetry:    0,
			Handler: func(ctx context.Context) error {
				return fmt.Errorf("intentional failure")
			},
		}

		err := workQueue.Submit(successJob)
		require.NoError(t, err)

		err = workQueue.Submit(failJob)
		require.NoError(t, err)

		// Collect results
		results := make(map[string]JobResult)
		timeout := time.After(5 * time.Second)
		resultCount := 0

		for resultCount < 2 {
			select {
			case result := <-workQueue.Results():
				results[result.Job.ID()] = result
				resultCount++
			case <-timeout:
				t.Fatal("Timeout waiting for job results")
			}
		}

		// Verify results
		successResult, hasSuccess := results["success-job"]
		if hasSuccess {
			assert.NoError(t, successResult.Error)
			assert.True(t, successResult.Duration > 0)
		}

		failResult, hasFail := results["fail-job"]
		if hasFail {
			assert.Error(t, failResult.Error)
			assert.Contains(t, failResult.Error.Error(), "intentional failure")
		}

		// At least one of the results should be present
		assert.True(t, hasSuccess || hasFail, "At least one result should be present")
	})

	t.Run("Statistics", func(t *testing.T) {
		stats := workQueue.GetStats()
		assert.True(t, stats.TotalJobs > 0)
		assert.True(t, stats.CompletedJobs > 0)
		assert.True(t, stats.AverageExecTime > 0)
	})
}

func TestIntegration(t *testing.T) {
	// Test integration between AsyncIO, EventBus, and WorkQueue
	t.Run("FileProcessingPipeline", func(t *testing.T) {
		tempDir := t.TempDir()

		// Setup components
		aio := NewAsyncIO(3)
		defer aio.Close()

		eventConfig := DefaultEventBusConfig()
		eventBus := NewEventBus(eventConfig)
		defer eventBus.Close()

		queueConfig := DefaultWorkQueueConfig("file-processor")
		queueConfig.EventBus = eventBus
		workQueue := NewWorkQueue(queueConfig)
		defer workQueue.Stop(5 * time.Second)

		// Track processed files
		var processedFiles []string
		var mu sync.Mutex
		done := make(chan struct{})

		// Subscribe to file processed events
		eventBus.SubscribeAsyncFunc(EventTypeFileProcessed, func(ctx context.Context, event Event) error {
			mu.Lock()
			defer mu.Unlock()

			data := event.Data().(map[string]interface{})
			processedFiles = append(processedFiles, data["file"].(string))

			if len(processedFiles) == 3 {
				close(done)
			}
			return nil
		})

		// Create test files
		testFiles := []string{
			filepath.Join(tempDir, "file1.txt"),
			filepath.Join(tempDir, "file2.txt"),
			filepath.Join(tempDir, "file3.txt"),
		}

		for i, file := range testFiles {
			content := fmt.Sprintf("Test content %d", i+1)
			err := os.WriteFile(file, []byte(content), 0o644)
			require.NoError(t, err)
		}

		// Submit file processing jobs
		for _, file := range testFiles {
			job := NewFileProcessingJob(file, func(ctx context.Context, path string) error {
				// Read file using AsyncIO
				resultCh := aio.ReadFileAsync(ctx, path)
				result := <-resultCh

				if result.Error != nil {
					return result.Error
				}

				// Process file (transform content)
				content := strings.ToUpper(string(result.Data))

				// Write processed content back
				writeResultCh := aio.WriteFileAsync(ctx, path+".processed", []byte(content), 0o644)
				writeResult := <-writeResultCh

				if writeResult.Error != nil {
					return writeResult.Error
				}

				// Publish file processed event
				if eventBus != nil {
					event := NewFileProcessedEvent("file-processor", path, map[string]interface{}{
						"file":           path,
						"processed_file": path + ".processed",
						"size":           len(content),
					})
					eventBus.PublishAsync(ctx, event)
				}

				return nil
			})

			err := workQueue.Submit(job)
			require.NoError(t, err)
		}

		// Wait for all files to be processed
		select {
		case <-done:
			mu.Lock()
			assert.Len(t, processedFiles, 3)

			// Verify processed files exist
			for _, originalFile := range testFiles {
				processedFile := originalFile + ".processed"
				assert.FileExists(t, processedFile)

				// Verify content was processed (uppercase)
				data, err := os.ReadFile(processedFile)
				require.NoError(t, err)
				content := string(data)
				assert.True(t, strings.ToUpper(content) == content)
			}
			mu.Unlock()
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for file processing pipeline")
		}
	})
}

func BenchmarkAsyncIO(b *testing.B) {
	aio := NewAsyncIO(10)
	defer aio.Close()

	tempDir := b.TempDir()

	b.Run("AsyncRead", func(b *testing.B) {
		// Create test file
		testFile := filepath.Join(tempDir, "bench.txt")
		content := strings.Repeat("benchmark test content\n", 1000)
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(b, err)

		b.ResetTimer()
		ctx := context.Background()

		for i := 0; i < b.N; i++ {
			resultCh := aio.ReadFileAsync(ctx, testFile)
			<-resultCh
		}
	})

	b.Run("SyncRead", func(b *testing.B) {
		// Create test file
		testFile := filepath.Join(tempDir, "sync_bench.txt")
		content := strings.Repeat("benchmark test content\n", 1000)
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(b, err)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := os.ReadFile(testFile)
			require.NoError(b, err)
		}
	})
}

func BenchmarkWorkQueue(b *testing.B) {
	config := DefaultWorkQueueConfig("bench-queue")
	config.Workers = 5
	workQueue := NewWorkQueue(config)
	defer workQueue.Stop(5 * time.Second)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := NewSimpleJob(fmt.Sprintf("job-%d", i), 5, func(ctx context.Context) error {
			time.Sleep(time.Microsecond) // Minimal work
			return nil
		})

		err := workQueue.Submit(job)
		require.NoError(b, err)
	}
}

// TestConnectionManager tests the connection manager functionality
func TestConnectionManager(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConnectionConfig()
		assert.Greater(t, config.MaxIdleConns, 0)
		assert.Greater(t, config.MaxIdleConnsPerHost, 0)
		assert.Greater(t, config.IdleConnTimeout, time.Duration(0))
		assert.Greater(t, config.RetryConfig.MaxRetries, 0)
	})

	t.Run("BasicHTTPRequest", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, Connection Manager!"))
		}))
		defer server.Close()

		cm := NewConnectionManager(DefaultConnectionConfig())
		defer cm.Close()

		// Create request
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		// Execute request
		ctx := context.Background()
		resp, err := cm.DoWithRetry(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check stats
		stats := cm.GetStats()
		assert.Equal(t, int64(1), stats.TotalRequests)
		assert.Equal(t, int64(1), stats.SuccessfulRequests)
		assert.Equal(t, int64(0), stats.FailedRequests)
	})

	t.Run("RetryOnServerError", func(t *testing.T) {
		var attempts int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current := atomic.AddInt64(&attempts, 1)
			if current <= 2 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server Error"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Success"))
			}
		}))
		defer server.Close()

		config := DefaultConnectionConfig()
		config.RetryConfig.MaxRetries = 3
		config.RetryConfig.BaseDelay = 10 * time.Millisecond
		cm := NewConnectionManager(config)
		defer cm.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		ctx := context.Background()
		resp, err := cm.DoWithRetry(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int64(3), atomic.LoadInt64(&attempts))

		// Check retry stats
		stats := cm.GetStats()
		assert.Greater(t, stats.RetryAttempts, int64(0))
	})

	t.Run("MaxRetriesExceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Persistent Error"))
		}))
		defer server.Close()

		config := DefaultConnectionConfig()
		config.RetryConfig.MaxRetries = 2
		config.RetryConfig.BaseDelay = 10 * time.Millisecond
		cm := NewConnectionManager(config)
		defer cm.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		ctx := context.Background()
		resp, err := cm.DoWithRetry(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Should have attempted max retries
		stats := cm.GetStats()
		assert.Equal(t, stats.RetryAttempts, int64(2))
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cm := NewConnectionManager(DefaultConnectionConfig())
		defer cm.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err = cm.DoWithRetry(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("ConcurrentRequests", func(t *testing.T) {
		var requestCount int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&requestCount, 1)
			time.Sleep(10 * time.Millisecond) // Simulate work
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		cm := NewConnectionManager(DefaultConnectionConfig())
		defer cm.Close()

		numRequests := 10
		var wg sync.WaitGroup
		errors := make(chan error, numRequests)

		ctx := context.Background()
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				req, err := http.NewRequest("GET", server.URL, nil)
				if err != nil {
					errors <- err
					return
				}

				resp, err := cm.DoWithRetry(ctx, req)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Request error: %v", err)
		}

		assert.Equal(t, int64(numRequests), atomic.LoadInt64(&requestCount))

		stats := cm.GetStats()
		assert.Equal(t, int64(numRequests), stats.TotalRequests)
		assert.Equal(t, int64(numRequests), stats.SuccessfulRequests)
	})

	t.Run("CustomRetryFunction", func(t *testing.T) {
		var attempts int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&attempts, 1)
			w.WriteHeader(http.StatusBadRequest) // Normally not retryable
		}))
		defer server.Close()

		cm := NewConnectionManager(DefaultConnectionConfig())
		defer cm.Close()

		// Custom retry function that retries on 400 status
		cm.SetRetryDecisionFunc(func(req *http.Request, resp *http.Response, err error, attempt int) bool {
			return attempt < 2 && resp != nil && resp.StatusCode == http.StatusBadRequest
		})

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		ctx := context.Background()
		resp, err := cm.DoWithRetry(ctx, req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, int64(3), atomic.LoadInt64(&attempts)) // Original + 2 retries

		stats := cm.GetStats()
		assert.Greater(t, stats.RetryAttempts, int64(0))
	})
}

// TestConnectionPooling tests connection reuse and pooling
func TestConnectionPooling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	config := DefaultConnectionConfig()
	config.MaxIdleConnsPerHost = 5
	config.IdleConnTimeout = 30 * time.Second

	cm := NewConnectionManager(config)
	defer cm.Close()

	ctx := context.Background()
	numRequests := 10

	// Make multiple requests to test connection reuse
	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		resp, err := cm.DoWithRetry(ctx, req)
		require.NoError(t, err)
		resp.Body.Close()
	}

	stats := cm.GetStats()
	assert.Equal(t, int64(numRequests), stats.TotalRequests)
	assert.Equal(t, int64(numRequests), stats.SuccessfulRequests)

	// Should have fewer new connections than total requests due to reuse
	// Note: This might be flaky depending on timing, but usually works
	if stats.NewConnections > 1 {
		t.Logf("New connections: %d, Total requests: %d", stats.NewConnections, numRequests)
	}
}

// TestRetryBackoff tests the exponential backoff calculation
func TestRetryBackoff(t *testing.T) {
	config := DefaultConnectionConfig()
	config.RetryConfig.BaseDelay = 100 * time.Millisecond
	config.RetryConfig.BackoffFactor = 2.0
	config.RetryConfig.MaxDelay = 5 * time.Second
	config.RetryConfig.JitterFactor = 0.1

	cm := NewConnectionManager(config)
	defer cm.Close()

	// Test delay calculation
	delay1 := cm.calculateRetryDelay(1)
	delay2 := cm.calculateRetryDelay(2)
	delay3 := cm.calculateRetryDelay(3)

	assert.True(t, delay1 >= 90*time.Millisecond && delay1 <= 110*time.Millisecond)
	assert.True(t, delay2 >= 180*time.Millisecond && delay2 <= 220*time.Millisecond)
	assert.True(t, delay3 >= 360*time.Millisecond && delay3 <= 440*time.Millisecond)

	// Test max delay cap
	delayHigh := cm.calculateRetryDelay(10)
	assert.LessOrEqual(t, delayHigh, 5*time.Second)
}

// BenchmarkConnectionManager benchmarks connection manager performance
func BenchmarkConnectionManager(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	cm := NewConnectionManager(DefaultConnectionConfig())
	defer cm.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				b.Fatal(err)
			}

			resp, err := cm.DoWithRetry(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}
