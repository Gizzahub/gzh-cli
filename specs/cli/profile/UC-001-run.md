# Command: gz profile

## Scenario: Run performance profiling on Go applications

### Input

**Command**:

```bash
gz profile --type cpu --duration 30s --output profile.pb.gz
```

**Prerequisites**:

- [ ] Go application running with pprof endpoints enabled
- [ ] Network access to target application
- [ ] Write permissions for output directory

### Expected Output

**CPU Profile Success**:

```text
🔬 Starting Go application profiling

🎯 Profile Configuration:
   📊 Type: CPU profiling
   ⏱️  Duration: 30 seconds
   🌐 Target: localhost:6060/debug/pprof/profile?seconds=30
   📁 Output: profile.pb.gz

🔍 Detecting application...
   ✅ Go pprof endpoints detected: http://localhost:6060/debug/pprof/
   ✅ Application responsive: 45ms response time
   📊 Runtime info: Go 1.21.5, 4 goroutines, 12.3MB heap

⏳ Collecting CPU profile data (30s)...
   [████████████████████████████████████████] 100% (30.0s/30.0s)

📊 Profile Collection Summary:
   ⏱️  Collection time: 30.05 seconds
   📦 Profile size: 2.3MB
   📈 Sample rate: 100Hz (3,005 samples collected)
   🎯 Top functions captured: 247 unique functions

✅ Profile saved: profile.pb.gz

🔍 Quick Analysis:
   🔥 Hottest function: main.processRequest (34.2% CPU time)
   📚 Top package: github.com/myorg/myapp/internal/handler (45.1% CPU time)
   🧵 Goroutines: 4-12 active during profiling
   💾 Memory: 12.3MB - 18.7MB heap usage

💡 Analyze profile:
   go tool pprof profile.pb.gz
   go tool pprof -http=:8080 profile.pb.gz
   go tool pprof -diff_base=previous.pb.gz profile.pb.gz

stderr: (empty)
Exit Code: 0
```

**Memory Profile Success**:

```text
🔬 Starting Go application profiling

🎯 Profile Configuration:
   📊 Type: Memory (heap) profiling
   🎯 Target: localhost:6060/debug/pprof/heap
   📁 Output: memory-profile.pb.gz

🔍 Collecting memory profile data...

📊 Memory Profile Summary:
   💾 Total allocations: 45.2MB since start
   📈 Current heap: 18.7MB
   🗂️  Objects in use: 156,234
   📦 Profile size: 892KB

✅ Profile saved: memory-profile.pb.gz

🔍 Memory Analysis:
   🔥 Largest allocator: internal/buffer.NewBuffer (12.3MB, 65.8%)
   📊 Top object type: []byte (34.5% of objects)
   🧹 GC stats: 15 cycles, avg 2.3ms pause
   📈 Growth rate: 1.2MB/hour steady state

💡 Memory optimization suggestions:
   • Consider object pooling for []byte allocations
   • Review buffer sizing in internal/buffer package
   • Monitor for memory leaks in long-running processes

stderr: (empty)
Exit Code: 0
```

**Goroutine Profile**:

```text
🔬 Starting Go application profiling

🎯 Profile Configuration:
   📊 Type: Goroutine profiling
   🎯 Target: localhost:6060/debug/pprof/goroutine
   📁 Output: goroutines.pb.gz

🔍 Collecting goroutine profile data...

📊 Goroutine Analysis:
   🧵 Active goroutines: 47
   🔄 Goroutine states:
      • Running: 2
      • Waiting: 41 (87.2%)
      • System: 4
   ⏱️  Longest running: 2h 15m 34s (main goroutine)

🔥 Goroutine Hotspots:
   • net/http.(*conn).serve: 23 goroutines (48.9%)
   • runtime.gopark: 18 goroutines (38.3%)
   • sync.(*WaitGroup).Wait: 3 goroutines (6.4%)

✅ Profile saved: goroutines.pb.gz

⚠️  Potential Issues Detected:
   • High number of waiting HTTP connection goroutines (23)
   • Possible goroutine leak in connection handling
   • 3 goroutines blocked on WaitGroup

💡 Recommendations:
   • Review HTTP connection pool settings
   • Check for proper connection cleanup
   • Monitor for goroutine growth over time

stderr: (empty)
Exit Code: 0
```

**Application Not Found**:

```text
🔬 Starting Go application profiling

🎯 Profile Configuration:
   📊 Type: CPU profiling
   🌐 Target: localhost:6060/debug/pprof/profile

🔍 Detecting application...

❌ Go application not found or not responding:
   • URL: http://localhost:6060/debug/pprof/
   • Error: connection refused
   • Status: No process listening on port 6060

💡 Troubleshooting:
   1. Start your Go application with pprof enabled:
      import _ "net/http/pprof"
      go func() { http.ListenAndServe(":6060", nil) }()

   2. Check if application is running:
      lsof -i :6060
      ps aux | grep your-app

   3. Specify custom endpoint:
      gz profile --endpoint http://localhost:8080/debug/pprof/

   4. Use external process profiling:
      gz profile --pid $(pgrep your-app)

🚫 Profiling failed - target application unreachable.

stderr: connection refused
Exit Code: 1
```

**Profile Collection Timeout**:

```text
🔬 Starting Go application profiling

🎯 Profile Configuration:
   📊 Type: CPU profiling
   ⏱️  Duration: 60 seconds
   🌐 Target: localhost:6060/debug/pprof/profile?seconds=60

⏳ Collecting CPU profile data (60s)...
   [████████████████████                    ] 67% (40.2s/60.0s)

❌ Profile collection failed:
   • Error: request timeout after 45 seconds
   • Partial data collected: 40.2 seconds
   • Possible cause: application under heavy load or deadlocked

🔍 Application Status:
   • pprof endpoint: reachable but slow
   • Last response: 45.3 seconds ago
   • HTTP status: connection timeout

💡 Recommendations:
   1. Reduce profiling duration: gz profile --duration 10s
   2. Check application health: gz profile --type goroutine
   3. Profile in shorter intervals during low load
   4. Check for deadlocks or infinite loops

⚠️  Partial profile may be available - check for temporary files.

stderr: request timeout
Exit Code: 1
```

### Side Effects

**Files Created**:

- Profile output file (e.g., `profile.pb.gz`, `memory-profile.pb.gz`)
- `~/.gzh/profile/session-<timestamp>.log` - Profiling session log
- `/tmp/gz-profile-*.tmp` - Temporary profiling data

**Files Modified**: None (read-only profiling)
**State Changes**: Profiling session recorded in history

### Validation

**Automated Tests**:

```bash
# Test profiling with mock HTTP server
(echo -e "HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n$(head -c 100 /dev/zero)" | nc -l 6060) &
server_pid=$!

result=$(gz profile --type cpu --duration 1s --output test-profile.pb.gz 2>&1)
exit_code=$?
kill $server_pid

assert_contains "$result" "Starting Go application profiling"
# Exit code: 0 (success), 1 (app not found/timeout), 2 (invalid config)

# Check profile file creation
if [ $exit_code -eq 0 ]; then
    assert_file_exists "test-profile.pb.gz"
    # Verify it's a valid pprof file
    file test-profile.pb.gz | grep -q "gzip compressed"
fi
```

**Manual Verification**:

1. Profile running Go application with pprof enabled
1. Test different profile types (cpu, memory, goroutine)
1. Verify profile files are valid and analyzable
1. Check error handling for unreachable applications
1. Test timeout scenarios with long-running collections
1. Validate profile analysis suggestions

### Edge Cases

**Application States**:

- Application under heavy load during profiling
- Application with disabled pprof endpoints
- Applications with custom pprof paths
- Multiple applications on different ports

**Profile Sizes**:

- Very large profiles (>100MB) from long-running apps
- Empty or minimal profiles from idle applications
- Corrupted profiles due to network issues
- Profile collection interrupted by application restart

**Network and Security**:

- Remote application profiling over network
- Applications behind authentication/proxy
- SSL/TLS enabled pprof endpoints
- Rate-limited or protected profiling endpoints

**Platform Differences**:

- Different Go runtime versions
- Custom pprof implementations
- Non-standard endpoint configurations
- Container-based applications with port mapping

### Performance Expectations

**Response Time**:

- Profile detection: < 2 seconds
- CPU profile (30s): 30-35 seconds total
- Memory profile: < 5 seconds
- Goroutine profile: < 3 seconds

**Resource Usage**:

- Memory: < 50MB during collection
- CPU: Minimal impact on profiled application
- Network: Varies by profile size (1-50MB typical)

**Profile Quality**:

- CPU sampling rate: 100Hz (configurable)
- Memory accuracy: All allocations >512 bytes
- Goroutine completeness: All active goroutines
- Time resolution: Microsecond precision

## Notes

- Integration with Go's built-in pprof package
- Support for all standard profile types (CPU, memory, goroutine, block, mutex)
- Remote profiling capabilities for production applications
- Profile analysis and comparison tools integration
- Automated performance insights and recommendations
- Historical profile tracking and comparison
- Integration with performance monitoring systems
- Safe profiling with minimal application impact
