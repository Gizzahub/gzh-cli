// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package simpleprof

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type profileWriterFunc func(io.Writer, int) error

func (f profileWriterFunc) WriteTo(writer io.Writer, debug int) error {
	return f(writer, debug)
}

func testProfiler(outputDir string) *SimpleProfiler {
	profiler := NewSimpleProfiler(outputDir)
	profiler.lookupProfile = func(string) profileWriter {
		return profileWriterFunc(func(writer io.Writer, _ int) error {
			_, err := writer.Write([]byte("profile"))
			return err
		})
	}
	return profiler
}

func TestLastGCTime(t *testing.T) {
	tests := []struct {
		name   string
		lastGC uint64
		want   time.Time
	}{
		{name: "zero preserves Unix epoch", lastGC: 0, want: time.Unix(0, 0)},
		{name: "normal timestamp", lastGC: 1_234_567_890, want: time.Unix(0, 1_234_567_890)},
		{name: "maximum signed timestamp", lastGC: math.MaxInt64, want: time.Unix(9_223_372_036, 854_775_807)},
		{name: "signed boundary plus one", lastGC: uint64(math.MaxInt64) + 1, want: time.Unix(9_223_372_036, 854_775_808)},
		{name: "maximum unsigned timestamp", lastGC: math.MaxUint64, want: time.Unix(18_446_744_073, 709_551_615)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastGCTime(tt.lastGC); !got.Equal(tt.want) {
				t.Errorf("LastGCTime(%d) = %v, want %v", tt.lastGC, got, tt.want)
			}
		})
	}
}

func TestPprofServeMuxOwnsOnlyPprofRoutes(t *testing.T) {
	assertPprofServeMuxContract(t, newPprofServeMux())
}

func TestPprofServeMuxSupportsGo121CompatibilityMode(t *testing.T) {
	t.Setenv("GODEBUG", "httpmuxgo121=1")
	assertPprofServeMuxContract(t, newPprofServeMux())
}

func assertPprofServeMuxContract(t *testing.T, mux *http.ServeMux) {
	t.Helper()

	for _, path := range []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Errorf("mux has no handler for %q", path)
		}

		postRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, nil)
		postResponse := httptest.NewRecorder()
		mux.ServeHTTP(postResponse, postRequest)
		if postResponse.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s status = %d, want %d", path, postResponse.Code, http.StatusMethodNotAllowed)
		}
		if allow := postResponse.Header().Get("Allow"); allow != "GET, HEAD" {
			t.Errorf("POST %s Allow = %q, want %q", path, allow, "GET, HEAD")
		}
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unrelated endpoint status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestStartHTTPServerUsesDedicatedIPv4LoopbackHandler(t *testing.T) {
	profiler := NewSimpleProfiler(t.TempDir())
	if err := profiler.StartHTTPServer(0); err != nil {
		t.Fatalf("StartHTTPServer() error: %v", err)
	}
	t.Cleanup(func() {
		if err := profiler.StopHTTPServer(context.Background()); err != nil {
			t.Errorf("StopHTTPServer() error: %v", err)
		}
	})

	if profiler.server.Addr != "127.0.0.1:0" {
		t.Fatalf("server address = %q, want explicit IPv4 loopback", profiler.server.Addr)
	}
	if profiler.server.Handler == nil || profiler.server.Handler == http.DefaultServeMux {
		t.Fatalf("server handler = %#v, want dedicated mux", profiler.server.Handler)
	}
}

func TestStartProfileDoesNotMutateMutexProfileRate(t *testing.T) {
	const mutexRateSentinel = 37

	previousMutexRate := runtime.SetMutexProfileFraction(mutexRateSentinel)
	t.Cleanup(func() {
		runtime.SetMutexProfileFraction(previousMutexRate)
	})
	assertMutexRateSentinel := func() {
		t.Helper()
		if got := runtime.SetMutexProfileFraction(-1); got != mutexRateSentinel {
			t.Fatalf("mutex profile rate = %d, want %d", got, mutexRateSentinel)
		}
	}

	profiler := NewSimpleProfiler(t.TempDir())
	for _, profileType := range []ProfileType{ProfileTypeBlock, ProfileTypeMutex} {
		t.Run(string(profileType), func(t *testing.T) {
			filename, err := profiler.StartProfile(profileType, 0)
			if err != nil {
				t.Fatalf("StartProfile(%q) error: %v", profileType, err)
			}
			if filename == "" {
				t.Fatal("StartProfile() returned an empty filename")
			}
			if _, err := os.Stat(filename); err != nil {
				t.Fatalf("profile file %q was not saved: %v", filename, err)
			}
			assertMutexRateSentinel()
		})
	}
}

func TestStartProfileReservesUniqueTimestampSequenceNames(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	profiler.now = func() time.Time { return time.Date(2026, time.August, 29, 9, 30, 0, 0, time.UTC) }
	var sequence atomic.Uint64
	profiler.nextSequence = func() uint64 { return sequence.Add(1) }

	first, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if err != nil {
		t.Fatalf("first StartProfile() error: %v", err)
	}
	second, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if err != nil {
		t.Fatalf("second StartProfile() error: %v", err)
	}
	if first == second {
		t.Fatalf("profile filenames must be unique: %q", first)
	}

	matches, err := filepath.Glob(filepath.Join(profiler.outputDir, "memory_20260829_093000_*.prof"))
	if err != nil {
		t.Fatalf("Glob() error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("matching profile files = %d, want 2 (%v)", len(matches), matches)
	}
}

func TestStartProfileUsesExclusive0600Reservation(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	var gotFlags int
	var gotMode os.FileMode
	profiler.openFile = func(name string, flags int, mode os.FileMode) (*os.File, error) {
		gotFlags = flags
		gotMode = mode
		return os.CreateTemp(t.TempDir(), "reserved")
	}

	if _, err := profiler.StartProfile(ProfileTypeMemory, 0); err != nil {
		t.Fatalf("StartProfile() error: %v", err)
	}
	if want := os.O_WRONLY | os.O_CREATE | os.O_EXCL; gotFlags != want {
		t.Errorf("OpenFile flags = %#x, want %#x", gotFlags, want)
	}
	if gotMode != 0o600 {
		t.Errorf("OpenFile mode = %04o, want 0600", gotMode)
	}
}

func TestStartProfileConcurrentReservationsAreUnique(t *testing.T) {
	const workers = 32
	profiler := testProfiler(t.TempDir())
	profiler.now = func() time.Time { return time.Date(2026, time.August, 29, 9, 30, 0, 0, time.UTC) }

	filenames := make(chan string, workers)
	errorsByWorker := make(chan error, workers)
	var waitGroup sync.WaitGroup
	for range workers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			filename, err := profiler.StartProfile(ProfileTypeMemory, 0)
			if err != nil {
				errorsByWorker <- err
				return
			}
			filenames <- filename
		}()
	}
	waitGroup.Wait()
	close(filenames)
	close(errorsByWorker)
	for err := range errorsByWorker {
		t.Errorf("concurrent StartProfile() error: %v", err)
	}

	seen := make(map[string]struct{}, workers)
	for filename := range filenames {
		if _, exists := seen[filename]; exists {
			t.Errorf("duplicate filename %q", filename)
		}
		seen[filename] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("unique filenames = %d, want %d", len(seen), workers)
	}
}

func TestStartProfilePreservesExistingCandidate(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	profiler.now = func() time.Time { return time.Date(2026, time.August, 29, 9, 30, 0, 0, time.UTC) }
	var sequence atomic.Uint64
	profiler.nextSequence = func() uint64 { return sequence.Add(1) }
	sentinel := filepath.Join(profiler.outputDir, "memory_20260829_093000_1.prof")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	filename, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if err != nil {
		t.Fatalf("StartProfile() error: %v", err)
	}
	if filepath.Base(filename) != "memory_20260829_093000_2.prof" {
		t.Fatalf("filename = %q, want retry candidate", filename)
	}
	contents, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("ReadFile(sentinel) error: %v", err)
	}
	if string(contents) != "keep" {
		t.Fatalf("sentinel contents = %q, want preserved value", contents)
	}
}

func TestStartProfileFailsAfterReservationLimitWithoutChangingCandidates(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	profiler.now = func() time.Time { return time.Date(2026, time.August, 29, 9, 30, 0, 0, time.UTC) }
	var sequence atomic.Uint64
	profiler.nextSequence = func() uint64 { return sequence.Add(1) }
	for candidate := 1; candidate <= maxProfileReservationAttempts; candidate++ {
		filename := filepath.Join(profiler.outputDir, fmt.Sprintf("memory_20260829_093000_%d.prof", candidate))
		if err := os.WriteFile(filename, []byte(fmt.Sprintf("sentinel-%d", candidate)), 0o600); err != nil {
			t.Fatalf("WriteFile(%q) error: %v", filename, err)
		}
	}

	_, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("StartProfile() error = %v, want ErrExist", err)
	}
	for candidate := 1; candidate <= maxProfileReservationAttempts; candidate++ {
		filename := filepath.Join(profiler.outputDir, fmt.Sprintf("memory_20260829_093000_%d.prof", candidate))
		contents, readErr := os.ReadFile(filename)
		if readErr != nil || string(contents) != fmt.Sprintf("sentinel-%d", candidate) {
			t.Errorf("candidate %d changed: contents=%q error=%v", candidate, contents, readErr)
		}
	}
}

func TestStartProfileReturnsNonExistReservationErrorImmediately(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	reservationErr := errors.New("disk unavailable")
	var calls atomic.Int32
	profiler.openFile = func(string, int, os.FileMode) (*os.File, error) {
		calls.Add(1)
		return nil, reservationErr
	}

	_, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if !errors.Is(err, reservationErr) {
		t.Fatalf("StartProfile() error = %v, want reservation error", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("open attempts = %d, want 1", got)
	}
}

func TestStartProfileCleansReservationAfterSynchronousWriteFailure(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	writeErr := errors.New("write failed")
	profiler.lookupProfile = func(string) profileWriter {
		return profileWriterFunc(func(io.Writer, int) error { return writeErr })
	}

	filename, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if filename != "" {
		t.Fatalf("filename = %q, want empty on failure", filename)
	}
	if !errors.Is(err, writeErr) {
		t.Fatalf("StartProfile() error = %v, want write error", err)
	}
	entries, readErr := os.ReadDir(profiler.outputDir)
	if readErr != nil {
		t.Fatalf("ReadDir() error: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("reserved files after write failure = %v, want none", entries)
	}
}

func TestStartProfileJoinsPrimaryCloseAndRemoveErrorsInOrder(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	primaryErr := errors.New("write failed")
	closeErr := errors.New("close failed")
	removeErr := errors.New("remove failed")
	var reservedFile *os.File
	var order atomic.Int32
	profiler.openFile = func(name string, flags int, mode os.FileMode) (*os.File, error) {
		file, err := os.CreateTemp(t.TempDir(), "reserved")
		reservedFile = file
		return file, err
	}
	profiler.lookupProfile = func(string) profileWriter {
		return profileWriterFunc(func(io.Writer, int) error { return primaryErr })
	}
	profiler.closeFile = func(*os.File) error {
		if !order.CompareAndSwap(0, 1) {
			t.Errorf("close order = %d, want first", order.Load())
		}
		return closeErr
	}
	profiler.removeFile = func(string) error {
		if !order.CompareAndSwap(1, 2) {
			t.Errorf("remove order = %d, want after close", order.Load())
		}
		return removeErr
	}
	t.Cleanup(func() {
		if reservedFile != nil {
			_ = reservedFile.Close()
		}
	})

	_, err := profiler.StartProfile(ProfileTypeMemory, 0)
	for _, want := range []error{primaryErr, closeErr, removeErr} {
		if !errors.Is(err, want) {
			t.Errorf("StartProfile() error = %v, want errors.Is(..., %v)", err, want)
		}
	}
	if got := order.Load(); got != 2 {
		t.Errorf("cleanup order final state = %d, want 2", got)
	}
}

func TestStartProfileRejectsInvalidTypeBeforeReservation(t *testing.T) {
	profiler := testProfiler(t.TempDir())
	_, err := profiler.StartProfile(ProfileType("../memory"), 0)
	if err == nil {
		t.Fatal("StartProfile() error = nil, want unsupported profile type")
	}
	entries, readErr := os.ReadDir(profiler.outputDir)
	if readErr != nil {
		t.Fatalf("ReadDir() error: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("files created for invalid profile type = %v", entries)
	}
}

func TestStartProfileCPUFailureCleansOnlyItsReservation(t *testing.T) {
	profiler := NewSimpleProfiler(t.TempDir())
	competingStartErr := errors.New("CPU profiling is already in use")
	var starts atomic.Int32
	profiler.startCPU = func(io.Writer) error {
		if starts.Add(1) == 1 {
			return nil
		}
		return competingStartErr
	}
	done := make(chan struct{})
	profiler.stopCPU = func() {}
	profiler.sleep = func(time.Duration) {}
	profiler.cpuDone = func() { close(done) }

	first, err := profiler.StartProfile(ProfileTypeCPU, 0)
	if err != nil {
		t.Fatalf("first CPU StartProfile() error: %v", err)
	}

	second, err := profiler.StartProfile(ProfileTypeCPU, 0)
	if second != "" {
		t.Fatalf("second CPU filename = %q, want empty on competing start", second)
	}
	if !errors.Is(err, competingStartErr) {
		t.Fatalf("second CPU StartProfile() error = %v, want competing-start error", err)
	}
	if _, statErr := os.Stat(first); statErr != nil {
		t.Fatalf("first CPU reservation was removed: %v", statErr)
	}
	entries, readErr := filepath.Glob(filepath.Join(profiler.outputDir, "cpu_*.prof"))
	if readErr != nil {
		t.Fatalf("Glob() error: %v", readErr)
	}
	if len(entries) != 1 || entries[0] != first {
		t.Fatalf("CPU reservations = %v, want only %q", entries, first)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("CPU profile completion did not close its goroutine")
	}
}

func TestProfileReservationHelperProcess(t *testing.T) {
	if os.Getenv("SIMPLEPROF_HELPER") != "1" {
		return
	}

	outputDir := os.Getenv("SIMPLEPROF_OUTPUT_DIR")
	workerID := os.Getenv("SIMPLEPROF_WORKER_ID")
	readyPath := os.Getenv("SIMPLEPROF_READY_PATH")
	releasePath := os.Getenv("SIMPLEPROF_RELEASE_PATH")
	if err := os.WriteFile(readyPath, []byte(workerID), 0o600); err != nil {
		t.Fatalf("WriteFile(ready) error: %v", err)
	}
	for {
		if _, err := os.Stat(releasePath); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Stat(release) error: %v", err)
		}
		time.Sleep(time.Millisecond)
	}

	profiler := testProfiler(outputDir)
	profiler.now = func() time.Time { return time.Date(2026, time.August, 29, 9, 30, 0, 0, time.UTC) }
	var sequence atomic.Uint64
	profiler.nextSequence = func() uint64 { return sequence.Add(1) }
	profiler.lookupProfile = func(string) profileWriter {
		return profileWriterFunc(func(writer io.Writer, _ int) error {
			_, err := io.WriteString(writer, workerID)
			return err
		})
	}
	filename, err := profiler.StartProfile(ProfileTypeMemory, 0)
	if err != nil {
		t.Fatalf("StartProfile() error: %v", err)
	}
	fmt.Println(filename)
}

func TestStartProfileReservesAcrossHelperProcesses(t *testing.T) {
	const workers = 4
	outputDir := t.TempDir()
	fixedTimestamp := "20260829_093000"
	sentinel := filepath.Join(outputDir, "memory_"+fixedTimestamp+"_1.prof")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatalf("WriteFile(sentinel) error: %v", err)
	}
	releasePath := filepath.Join(outputDir, "release")

	type helperResult struct {
		id       int
		filename string
		err      error
	}
	results := make(chan helperResult, workers)
	for workerID := range workers {
		workerID := workerID
		go func() {
			commandContext, cancel := context.WithTimeout(t.Context(), 45*time.Second)
			defer cancel()
			readyPath := filepath.Join(outputDir, fmt.Sprintf("ready-%d", workerID))
			command := exec.CommandContext(commandContext, "go", "test", ".", "-run=^TestProfileReservationHelperProcess$", "-count=1", "-v")
			command.Env = append(os.Environ(),
				"SIMPLEPROF_HELPER=1",
				"SIMPLEPROF_OUTPUT_DIR="+outputDir,
				fmt.Sprintf("SIMPLEPROF_WORKER_ID=worker-%d", workerID),
				"SIMPLEPROF_READY_PATH="+readyPath,
				"SIMPLEPROF_RELEASE_PATH="+releasePath,
			)
			output, err := command.Output()
			filename := ""
			for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				if strings.HasSuffix(line, ".prof") {
					filename = line
					break
				}
			}
			results <- helperResult{id: workerID, filename: filename, err: err}
		}()
	}

	deadline := time.Now().Add(30 * time.Second)
	for workerID := range workers {
		readyPath := filepath.Join(outputDir, fmt.Sprintf("ready-%d", workerID))
		for {
			if _, err := os.Stat(readyPath); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("helper %d did not reach reservation barrier", workerID)
			}
			time.Sleep(time.Millisecond)
		}
	}
	if err := os.WriteFile(releasePath, nil, 0o600); err != nil {
		t.Fatalf("WriteFile(release) error: %v", err)
	}

	seen := make(map[string]struct{}, workers)
	for range workers {
		result := <-results
		if result.err != nil {
			t.Errorf("helper %d error: %v", result.id, result.err)
			continue
		}
		if _, exists := seen[result.filename]; exists {
			t.Errorf("duplicate helper filename %q", result.filename)
		}
		seen[result.filename] = struct{}{}
		contents, err := os.ReadFile(result.filename)
		if err != nil {
			t.Errorf("ReadFile(%q) error: %v", result.filename, err)
			continue
		}
		if want := fmt.Sprintf("worker-%d", result.id); string(contents) != want {
			t.Errorf("helper %d contents = %q, want %q", result.id, contents, want)
		}
	}
	if len(seen) != workers {
		t.Errorf("unique helper files = %d, want %d", len(seen), workers)
	}
	root, err := os.OpenRoot(outputDir)
	if err != nil {
		t.Fatalf("OpenRoot() error: %v", err)
	}
	defer root.Close()
	file, err := root.Open(filepath.Base(sentinel))
	if err != nil {
		t.Fatalf("Open(sentinel) error: %v", err)
	}
	contents, err := io.ReadAll(file)
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("ReadFile(sentinel) errors: read=%v close=%v", err, closeErr)
	}
	if string(contents) != "keep" {
		t.Errorf("sentinel contents = %q, want preserved value", contents)
	}
}
