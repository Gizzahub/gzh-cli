//nolint:testpackage // White-box testing needed for internal function access
package bulkclone

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// legacyExampleConfig는 이 패키지가 읽을 수 있는 유일한 배포 예시다.
//
// 예전에는 ../../examples/bulk-clone/bulk-clone-simple.yaml과
// bulk-clone-example.yaml을 가리켰지만 두 파일은 f4c8d5d(bulk-clone → synclone
// 이름 변경)에서 지워졌다. 그 뒤로 이 파일의 시험은 전부 "no such file or
// directory"로 실패하고 있었다.
//
// examples/synclone/synclone-simple.yaml은 여기서 쓸 수 없다. 그쪽은 새 v1.0
// 형식(github.organizations)이고 pkg/config가 다룬다. 이 패키지의
// bulkCloneConfig는 옛 형식(default/repo_roots)만 읽는다.
const legacyExampleConfig = "../../examples/synclone/synclone-example.yaml"

func TestExampleConfigs(t *testing.T) {
	t.Run(filepath.Base(legacyExampleConfig), func(t *testing.T) {
		cfg := &bulkCloneConfig{}
		err := cfg.ReadConfig(legacyExampleConfig)

		require.NoError(t, err, "Example config file should be valid: %s", legacyExampleConfig)

		// Basic validation
		assert.Equal(t, "1.0", cfg.Version, "Version should be 1.0")

		// Default protocol should be set
		assert.NotEmpty(t, cfg.Default.Protocol, "Default protocol should be set")

		// Should have at least one repo_root configured
		assert.NotEmpty(t, cfg.RepoRoots, "Should have at least one repo_root configured")

		// Each repo_root should have required fields
		for i, repo := range cfg.RepoRoots {
			assert.NotEmptyf(t, repo.RootPath, "repo_roots[%d].root_path should not be empty", i)
			assert.NotEmptyf(t, repo.Provider, "repo_roots[%d].provider should not be empty", i)
			assert.NotEmptyf(t, repo.Protocol, "repo_roots[%d].protocol should not be empty", i)
			// GitHub requires org_name, GitLab uses group_name (checked in config structure)
			assert.NotEmptyf(t, repo.OrgName, "repo_roots[%d].org_name should not be empty for GitHub", i)
		}
	})
}

func TestComprehensiveExampleConfig(t *testing.T) {
	cfg := &bulkCloneConfig{}
	err := cfg.ReadConfig(legacyExampleConfig)
	require.NoError(t, err)

	// 값을 하나하나 확인한다. 개수나 NotEmpty만 보면 키 표기가 어긋나 파일
	// 내용이 통째로 버려져도 알아채지 못한다 -- 실제로 구조체 태그가
	// camelCase였을 때 이 패키지의 시험은 전부 빈 설정을 보고 있었다.
	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, "https", cfg.Default.Protocol)
		assert.Equal(t, "$HOME/repos/github", cfg.Default.Github.RootPath)
		assert.Equal(t, "github", cfg.Default.Github.Provider)
		assert.Equal(t, "$HOME/repos/gitlab", cfg.Default.Gitlab.RootPath)
		assert.Equal(t, "https://gitlab.com", cfg.Default.Gitlab.URL)
	})

	t.Run("repo_roots", func(t *testing.T) {
		require.Len(t, cfg.RepoRoots, 5)

		assert.Equal(t, "$HOME/work/mycompany", cfg.RepoRoots[0].RootPath)
		assert.Equal(t, "github", cfg.RepoRoots[0].Provider)
		assert.Equal(t, "ssh", cfg.RepoRoots[0].Protocol)
		assert.Equal(t, "mycompany", cfg.RepoRoots[0].OrgName)

		assert.Equal(t, "$HOME/opensource/kubernetes", cfg.RepoRoots[1].RootPath)
		assert.Equal(t, "https", cfg.RepoRoots[1].Protocol)
		assert.Equal(t, "kubernetes", cfg.RepoRoots[1].OrgName)
	})

	t.Run("diverse_configurations", func(t *testing.T) {
		hasSSH := false
		hasHTTPS := false

		for _, repo := range cfg.RepoRoots {
			// All should be GitHub since repo_roots only supports GitHub currently
			assert.Equal(t, "github", repo.Provider)

			switch repo.Protocol {
			case "ssh":
				hasSSH = true
			case "https":
				hasHTTPS = true
			}
		}

		assert.True(t, hasSSH, "Should have SSH protocol examples")
		assert.True(t, hasHTTPS, "Should have HTTPS protocol examples")
	})
}

func TestIgnorePatternsValidity(t *testing.T) {
	cfg := &bulkCloneConfig{}
	err := cfg.ReadConfig(legacyExampleConfig)
	require.NoError(t, err)

	require.NotEmpty(t, cfg.IgnoreNameRegexes, "예시에 ignore_names가 있어야 한다")

	for _, pattern := range cfg.IgnoreNameRegexes {
		t.Run(pattern, func(t *testing.T) {
			// 실제로 컴파일해 본다. 예전에는 NotPanics 안에서 NotEmpty만
			// 확인해서 이름과 달리 무늬가 올바른지는 보지 않았고, 깨진
			// 정규식이 예시에 들어가도 통과했다.
			_, err := regexp.Compile(pattern)
			assert.NoError(t, err, "ignore_names 무늬는 Go 정규식(RE2)으로 컴파일돼야 한다")
		})
	}
}
