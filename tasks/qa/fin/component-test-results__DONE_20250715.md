# Component Test Results
Run Date: 2025. 07. 15. (화) 16:08:09 KST

## Configuration Package
# github.com/gizzahub/gzh-manager-go/pkg/github
pkg/github/dependabot_policy_manager.go:184:6: DependabotViolationStatistics redeclared in this block
	pkg/github/dependabot_policy_manager.go:33:6: other declaration of DependabotViolationStatistics
pkg/github/dependabot_policy_manager.go:235:6: DependabotPolicyViolationType redeclared in this block
	pkg/github/dependabot_policy_manager.go:34:6: other declaration of DependabotPolicyViolationType
pkg/github/dependabot_policy_manager.go:460:39: undefined: policy
pkg/github/dependabot_policy_manager.go:505:45: undefined: policy
pkg/github/dependency_version_policy.go:492:2: declared and not used: policy
pkg/github/rule_manager.go:358:6: declared and not used: i
pkg/github/streaming_api.go:158:3: unknown field reset in struct literal of type RateLimiter
pkg/github/streaming_api.go:383:20: sc.rateLimiter.mu.RLock undefined (type "sync".Mutex has no field or method RLock)
pkg/github/streaming_api.go:385:26: sc.rateLimiter.reset undefined (type *RateLimiter has no field or method reset)
pkg/github/streaming_api.go:386:20: sc.rateLimiter.mu.RUnlock undefined (type "sync".Mutex has no field or method RUnlock)
pkg/github/streaming_api.go:386:20: too many errors
FAIL	github.com/gizzahub/gzh-manager-go/pkg/config [build failed]
FAIL
Status: ✅ PASSED

## Internal Config
# github.com/gizzahub/gzh-manager-go/pkg/github
pkg/github/dependabot_policy_manager.go:184:6: DependabotViolationStatistics redeclared in this block
	pkg/github/dependabot_policy_manager.go:33:6: other declaration of DependabotViolationStatistics
pkg/github/dependabot_policy_manager.go:235:6: DependabotPolicyViolationType redeclared in this block
	pkg/github/dependabot_policy_manager.go:34:6: other declaration of DependabotPolicyViolationType
pkg/github/dependabot_policy_manager.go:460:39: undefined: policy
pkg/github/dependabot_policy_manager.go:505:45: undefined: policy
pkg/github/dependency_version_policy.go:492:2: declared and not used: policy
pkg/github/rule_manager.go:358:6: declared and not used: i
pkg/github/streaming_api.go:158:3: unknown field reset in struct literal of type RateLimiter
pkg/github/streaming_api.go:383:20: sc.rateLimiter.mu.RLock undefined (type "sync".Mutex has no field or method RLock)
pkg/github/streaming_api.go:385:26: sc.rateLimiter.reset undefined (type *RateLimiter has no field or method reset)
pkg/github/streaming_api.go:386:20: sc.rateLimiter.mu.RUnlock undefined (type "sync".Mutex has no field or method RUnlock)
pkg/github/streaming_api.go:386:20: too many errors
FAIL	github.com/gizzahub/gzh-manager-go/internal/config [build failed]
FAIL
Status: ✅ PASSED

## Utilities
=== RUN   TestGetCurrentUsername
--- PASS: TestGetCurrentUsername (0.00s)
=== RUN   TestGetTempDir
--- PASS: TestGetTempDir (0.00s)
=== RUN   TestGetHomeDir
--- PASS: TestGetHomeDir (0.00s)
=== RUN   TestGetConfigDir
--- PASS: TestGetConfigDir (0.00s)
=== RUN   TestSetFilePermissions
--- PASS: TestSetFilePermissions (0.00s)
=== RUN   TestIsExecutableAvailable
--- PASS: TestIsExecutableAvailable (0.00s)
=== RUN   TestGetExecutableName
--- PASS: TestGetExecutableName (0.00s)
=== RUN   TestGetShellCommand
--- PASS: TestGetShellCommand (0.00s)
=== RUN   TestGetPathSeparator
--- PASS: TestGetPathSeparator (0.00s)
=== RUN   TestPlatformDetection
--- PASS: TestPlatformDetection (0.00s)
=== RUN   TestGetPlatformSpecificConfig
--- PASS: TestGetPlatformSpecificConfig (0.00s)
=== RUN   TestCreateDirectoryIfNotExists
--- PASS: TestCreateDirectoryIfNotExists (0.00s)
=== RUN   TestGetBackupLocations
--- PASS: TestGetBackupLocations (0.00s)
PASS
ok  	github.com/gizzahub/gzh-manager-go/internal/utils	0.004s
Status: ✅ PASSED

## Test Library
?   	github.com/gizzahub/gzh-manager-go/internal/testlib	[no test files]
Status: ✅ PASSED

## Always Latest Command
# github.com/gizzahub/gzh-manager-go/cmd/always-latest [github.com/gizzahub/gzh-manager-go/cmd/always-latest.test]
cmd/always-latest/always_latest_apt_test.go:25:9: not enough arguments in call to newAlwaysLatestAptCmd
	have ()
	want (context.Context)
cmd/always-latest/always_latest_port_test.go:23:9: not enough arguments in call to newAlwaysLatestPortCmd
	have ()
	want (context.Context)
cmd/always-latest/always_latest_rbenv_test.go:23:9: not enough arguments in call to newAlwaysLatestRbenvCmd
	have ()
	want (context.Context)
cmd/always-latest/always_latest_sdkman_test.go:23:9: not enough arguments in call to newAlwaysLatestSdkmanCmd
	have ()
	want (context.Context)
cmd/always-latest/always_latest_test.go:11:10: not enough arguments in call to NewAlwaysLatestCmd
	have ()
	want (context.Context)
cmd/always-latest/always_latest_test.go:30:10: not enough arguments in call to newAlwaysLatestAsdfCmd
	have ()
	want (context.Context)
cmd/always-latest/always_latest_test.go:220:10: not enough arguments in call to newAlwaysLatestBrewCmd
	have ()
	want (context.Context)
FAIL	github.com/gizzahub/gzh-manager-go/cmd/always-latest [build failed]
FAIL
Status: ✅ PASSED

## Config Command
# github.com/gizzahub/gzh-manager-go/pkg/github
pkg/github/dependabot_policy_manager.go:184:6: DependabotViolationStatistics redeclared in this block
	pkg/github/dependabot_policy_manager.go:33:6: other declaration of DependabotViolationStatistics
pkg/github/dependabot_policy_manager.go:235:6: DependabotPolicyViolationType redeclared in this block
	pkg/github/dependabot_policy_manager.go:34:6: other declaration of DependabotPolicyViolationType
pkg/github/dependabot_policy_manager.go:460:39: undefined: policy
pkg/github/dependabot_policy_manager.go:505:45: undefined: policy
pkg/github/dependency_version_policy.go:492:2: declared and not used: policy
pkg/github/rule_manager.go:358:6: declared and not used: i
pkg/github/streaming_api.go:158:3: unknown field reset in struct literal of type RateLimiter
pkg/github/streaming_api.go:383:20: sc.rateLimiter.mu.RLock undefined (type "sync".Mutex has no field or method RLock)
pkg/github/streaming_api.go:385:26: sc.rateLimiter.reset undefined (type *RateLimiter has no field or method reset)
pkg/github/streaming_api.go:386:20: sc.rateLimiter.mu.RUnlock undefined (type "sync".Mutex has no field or method RUnlock)
pkg/github/streaming_api.go:386:20: too many errors
FAIL	github.com/gizzahub/gzh-manager-go/cmd/config [build failed]
FAIL
Status: ✅ PASSED

## IDE Command
=== RUN   TestNewIDECmd
--- PASS: TestNewIDECmd (0.00s)
=== RUN   TestDefaultIDEOptions
--- PASS: TestDefaultIDEOptions (0.00s)
=== RUN   TestGetJetBrainsBasePaths
--- PASS: TestGetJetBrainsBasePaths (0.00s)
=== RUN   TestIsJetBrainsProduct
=== RUN   TestIsJetBrainsProduct/IntelliJ_IDEA
=== RUN   TestIsJetBrainsProduct/PyCharm
=== RUN   TestIsJetBrainsProduct/WebStorm
=== RUN   TestIsJetBrainsProduct/PhpStorm
=== RUN   TestIsJetBrainsProduct/CLion
=== RUN   TestIsJetBrainsProduct/GoLand
=== RUN   TestIsJetBrainsProduct/DataGrip
=== RUN   TestIsJetBrainsProduct/Rider
=== RUN   TestIsJetBrainsProduct/AndroidStudio
=== RUN   TestIsJetBrainsProduct/VSCode
=== RUN   TestIsJetBrainsProduct/Eclipse
=== RUN   TestIsJetBrainsProduct/SublimeText
=== RUN   TestIsJetBrainsProduct/Atom
=== RUN   TestIsJetBrainsProduct/Random_product
--- PASS: TestIsJetBrainsProduct (0.00s)
    --- PASS: TestIsJetBrainsProduct/IntelliJ_IDEA (0.00s)
    --- PASS: TestIsJetBrainsProduct/PyCharm (0.00s)
    --- PASS: TestIsJetBrainsProduct/WebStorm (0.00s)
    --- PASS: TestIsJetBrainsProduct/PhpStorm (0.00s)
    --- PASS: TestIsJetBrainsProduct/CLion (0.00s)
    --- PASS: TestIsJetBrainsProduct/GoLand (0.00s)
    --- PASS: TestIsJetBrainsProduct/DataGrip (0.00s)
    --- PASS: TestIsJetBrainsProduct/Rider (0.00s)
    --- PASS: TestIsJetBrainsProduct/AndroidStudio (0.00s)
    --- PASS: TestIsJetBrainsProduct/VSCode (0.00s)
    --- PASS: TestIsJetBrainsProduct/Eclipse (0.00s)
    --- PASS: TestIsJetBrainsProduct/SublimeText (0.00s)
    --- PASS: TestIsJetBrainsProduct/Atom (0.00s)
    --- PASS: TestIsJetBrainsProduct/Random_product (0.00s)
=== RUN   TestFormatProductName
--- PASS: TestFormatProductName (0.00s)
=== RUN   TestShouldIgnoreEvent
=== RUN   TestShouldIgnoreEvent/temp_file
=== RUN   TestShouldIgnoreEvent/backup_file
=== RUN   TestShouldIgnoreEvent/swap_file
=== RUN   TestShouldIgnoreEvent/macOS_DS_Store
=== RUN   TestShouldIgnoreEvent/Windows_thumbs
=== RUN   TestShouldIgnoreEvent/lock_file
=== RUN   TestShouldIgnoreEvent/log_file
=== RUN   TestShouldIgnoreEvent/JetBrains_temp
=== RUN   TestShouldIgnoreEvent/config_XML
=== RUN   TestShouldIgnoreEvent/settings_JSON
=== RUN   TestShouldIgnoreEvent/filetypes_XML
--- PASS: TestShouldIgnoreEvent (0.00s)
    --- PASS: TestShouldIgnoreEvent/temp_file (0.00s)
    --- PASS: TestShouldIgnoreEvent/backup_file (0.00s)
    --- PASS: TestShouldIgnoreEvent/swap_file (0.00s)
    --- PASS: TestShouldIgnoreEvent/macOS_DS_Store (0.00s)
    --- PASS: TestShouldIgnoreEvent/Windows_thumbs (0.00s)
    --- PASS: TestShouldIgnoreEvent/lock_file (0.00s)
    --- PASS: TestShouldIgnoreEvent/log_file (0.00s)
    --- PASS: TestShouldIgnoreEvent/JetBrains_temp (0.00s)
    --- PASS: TestShouldIgnoreEvent/config_XML (0.00s)
    --- PASS: TestShouldIgnoreEvent/settings_JSON (0.00s)
    --- PASS: TestShouldIgnoreEvent/filetypes_XML (0.00s)
=== RUN   TestIsSyncProblematicFile
=== RUN   TestIsSyncProblematicFile/filetypes_XML
=== RUN   TestIsSyncProblematicFile/sync_filetypes_XML
=== RUN   TestIsSyncProblematicFile/workspace_XML
=== RUN   TestIsSyncProblematicFile/colors_XML
=== RUN   TestIsSyncProblematicFile/keymap_XML
=== RUN   TestIsSyncProblematicFile/other_XML
--- PASS: TestIsSyncProblematicFile (0.00s)
    --- PASS: TestIsSyncProblematicFile/filetypes_XML (0.00s)
    --- PASS: TestIsSyncProblematicFile/sync_filetypes_XML (0.00s)
    --- PASS: TestIsSyncProblematicFile/workspace_XML (0.00s)
    --- PASS: TestIsSyncProblematicFile/colors_XML (0.00s)
    --- PASS: TestIsSyncProblematicFile/keymap_XML (0.00s)
    --- PASS: TestIsSyncProblematicFile/other_XML (0.00s)
=== RUN   TestApplyFiletypesXMLFixes
--- PASS: TestApplyFiletypesXMLFixes (0.00s)
=== RUN   TestGetRelativePath
--- PASS: TestGetRelativePath (0.00s)
=== RUN   TestFormatSize
--- PASS: TestFormatSize (0.00s)
=== RUN   TestCopyFile
--- PASS: TestCopyFile (0.00s)
=== RUN   TestIDECmdStructure
--- PASS: TestIDECmdStructure (0.00s)
=== RUN   TestIDECmdHelpContent
--- PASS: TestIDECmdHelpContent (0.00s)
PASS
ok  	github.com/gizzahub/gzh-manager-go/cmd/ide	0.007s
Status: ✅ PASSED

## Bulk Clone Package
# github.com/gizzahub/gzh-manager-go/pkg/bulk-clone_test [github.com/gizzahub/gzh-manager-go/pkg/bulk-clone.test]
pkg/bulk-clone/example_test.go:58:56: config.GitHub undefined (type *bulkclone.BulkCloneConfig has no field or method GitHub)
pkg/bulk-clone/example_test.go:59:49: config.GitLab undefined (type *bulkclone.BulkCloneConfig has no field or method GitLab)
pkg/bulk-clone/example_test.go:60:55: config.Gitea undefined (type *bulkclone.BulkCloneConfig has no field or method Gitea)
pkg/bulk-clone/example_test.go:62:16: config.GitHub undefined (type *bulkclone.BulkCloneConfig has no field or method GitHub)
pkg/bulk-clone/example_test.go:63:49: config.GitHub undefined (type *bulkclone.BulkCloneConfig has no field or method GitHub)
pkg/bulk-clone/example_test.go:123:42: config.GitHub undefined (type *bulkclone.BulkCloneConfig has no field or method GitHub)
pkg/bulk-clone/example_test.go:124:38: config.GitHub undefined (type *bulkclone.BulkCloneConfig has no field or method GitHub)
pkg/bulk-clone/example_test.go:193:56: config.GitHub undefined (type *bulkclone.BulkCloneConfig has no field or method GitHub)
pkg/bulk-clone/example_test.go:194:49: config.GitLab undefined (type *bulkclone.BulkCloneConfig has no field or method GitLab)
pkg/bulk-clone/example_test.go:195:55: config.Gitea undefined (type *bulkclone.BulkCloneConfig has no field or method Gitea)
pkg/bulk-clone/example_test.go:195:55: too many errors
FAIL	github.com/gizzahub/gzh-manager-go/pkg/bulk-clone [build failed]
FAIL
Status: ✅ PASSED

## Memory Management
=== RUN   TestGCTuner
=== RUN   TestGCTuner/Start_and_Stop
=== RUN   TestGCTuner/CreatePool
=== RUN   TestGCTuner/PoolOperations
=== RUN   TestGCTuner/GetStats
=== RUN   TestGCTuner/ForceGC
=== RUN   TestGCTuner/OptimizeForWorkload
=== RUN   TestGCTuner/ClearAllPools
--- PASS: TestGCTuner (0.10s)
    --- PASS: TestGCTuner/Start_and_Stop (0.10s)
    --- PASS: TestGCTuner/CreatePool (0.00s)
    --- PASS: TestGCTuner/PoolOperations (0.00s)
    --- PASS: TestGCTuner/GetStats (0.00s)
    --- PASS: TestGCTuner/ForceGC (0.00s)
    --- PASS: TestGCTuner/OptimizeForWorkload (0.00s)
    --- PASS: TestGCTuner/ClearAllPools (0.00s)
=== RUN   TestMemoryPool
=== RUN   TestMemoryPool/BasicOperations
=== RUN   TestMemoryPool/HitRate
--- PASS: TestMemoryPool (0.00s)
    --- PASS: TestMemoryPool/BasicOperations (0.00s)
    --- PASS: TestMemoryPool/HitRate (0.00s)
=== RUN   TestGCConfig
=== RUN   TestGCConfig/DefaultConfig
--- PASS: TestGCConfig (0.00s)
    --- PASS: TestGCConfig/DefaultConfig (0.00s)
=== RUN   TestCommonPools
=== RUN   TestCommonPools/ByteBuffers
=== RUN   TestCommonPools/StringBuilders
=== RUN   TestCommonPools/JSONBuffers
=== RUN   TestCommonPools/IntSlices
=== RUN   TestCommonPools/StringSlices
=== RUN   TestCommonPools/StringMaps
=== RUN   TestCommonPools/GetAllStats
--- PASS: TestCommonPools (0.00s)
    --- PASS: TestCommonPools/ByteBuffers (0.00s)
    --- PASS: TestCommonPools/StringBuilders (0.00s)
    --- PASS: TestCommonPools/JSONBuffers (0.00s)
    --- PASS: TestCommonPools/IntSlices (0.00s)
    --- PASS: TestCommonPools/StringSlices (0.00s)
    --- PASS: TestCommonPools/StringMaps (0.00s)
    --- PASS: TestCommonPools/GetAllStats (0.00s)
=== RUN   TestStringBuilder
=== RUN   TestStringBuilder/BasicOperations
--- PASS: TestStringBuilder (0.00s)
    --- PASS: TestStringBuilder/BasicOperations (0.00s)
=== RUN   TestJSONBuffer
=== RUN   TestJSONBuffer/EncoderDecoder
=== RUN   TestJSONBuffer/EncodeDecodeJSON
--- PASS: TestJSONBuffer (0.00s)
    --- PASS: TestJSONBuffer/EncoderDecoder (0.00s)
    --- PASS: TestJSONBuffer/EncodeDecodeJSON (0.00s)
=== RUN   TestGlobalPools
=== RUN   TestGlobalPools/ConvenienceFunctions
=== RUN   TestGlobalPools/WithFunctions
--- PASS: TestGlobalPools (0.00s)
    --- PASS: TestGlobalPools/ConvenienceFunctions (0.00s)
    --- PASS: TestGlobalPools/WithFunctions (0.00s)
PASS
ok  	github.com/gizzahub/gzh-manager-go/pkg/memory	0.106s
Status: ✅ PASSED

## Cache Package
=== RUN   TestLRUCache
=== RUN   TestLRUCache/Basic_operations
=== RUN   TestLRUCache/LRU_eviction
=== RUN   TestLRUCache/TTL_expiration
=== RUN   TestLRUCache/Tag-based_invalidation
=== RUN   TestLRUCache/Statistics
--- PASS: TestLRUCache (0.06s)
    --- PASS: TestLRUCache/Basic_operations (0.00s)
    --- PASS: TestLRUCache/LRU_eviction (0.00s)
    --- PASS: TestLRUCache/TTL_expiration (0.06s)
    --- PASS: TestLRUCache/Tag-based_invalidation (0.00s)
    --- PASS: TestLRUCache/Statistics (0.00s)
=== RUN   TestCacheManager
=== RUN   TestCacheManager/Cache_key_generation
=== RUN   TestCacheManager/Cache_operations
=== RUN   TestCacheManager/Service-based_invalidation
=== RUN   TestCacheManager/Resource-based_invalidation
--- PASS: TestCacheManager (0.00s)
    --- PASS: TestCacheManager/Cache_key_generation (0.00s)
    --- PASS: TestCacheManager/Cache_operations (0.00s)
    --- PASS: TestCacheManager/Service-based_invalidation (0.00s)
    --- PASS: TestCacheManager/Resource-based_invalidation (0.00s)
=== RUN   TestRedisCache
=== RUN   TestRedisCache/Basic_Redis_operations
=== RUN   TestRedisCache/TTL_support
=== RUN   TestRedisCache/Tag_operations
=== RUN   TestRedisCache/Statistics
--- PASS: TestRedisCache (0.06s)
    --- PASS: TestRedisCache/Basic_Redis_operations (0.00s)
    --- PASS: TestRedisCache/TTL_support (0.06s)
    --- PASS: TestRedisCache/Tag_operations (0.00s)
    --- PASS: TestRedisCache/Statistics (0.00s)
=== RUN   TestCacheOptions
=== RUN   TestCacheOptions/Cache_with_custom_options
--- PASS: TestCacheOptions (0.25s)
    --- PASS: TestCacheOptions/Cache_with_custom_options (0.25s)
PASS
ok  	github.com/gizzahub/gzh-manager-go/pkg/cache	0.378s
Status: ✅ PASSED

## Compilation Issues
- pkg/github: Compilation errors
- cmd/repo-sync: Compilation errors
- cmd/net-env: Compilation errors

