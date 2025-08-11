package compat

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Rule represents a compatibility rule for a package manager and optional plugin.
// If Plugin is empty, the rule applies to all plugins/targets under the Manager.
// Env contains environment variables that should be applied to child processes.
// Note provides a brief explanation about why the rule exists.
type Rule struct {
	Manager string            `yaml:"manager"`
	Plugin  string            `yaml:"plugin"`
	Env     map[string]string `yaml:"env"`
	Note    string            `yaml:"note"`
}

// registry holds built-in compatibility rules. This can be extended in future to be
// loaded from user configuration files if needed.
var registry = []Rule{
	{
		Manager: "asdf",
		Plugin:  "rust",
		Env: map[string]string{
			// rustup-init blocks when rust exists in PATH (e.g., from asdf shims)
			// Skipping PATH check prevents install from failing during automated runs
			"RUSTUP_INIT_SKIP_PATH_CHECK": "yes",
			// Non-interactive install for CI/automation scenarios
			"RUSTUP_INIT_YES": "1",
		},
		Note: "자동 호환성 규칙 적용: asdf rust 설치 시 rustup PATH 체크를 우회했습니다.",
	},
}

// GetEnvFor returns the environment variables and an optional note for a manager/plugin pair.
// If multiple rules match (specific plugin and generic manager), variables are merged with
// plugin-specific values taking precedence.
func GetEnvFor(manager, plugin string) (env map[string]string, note string) {
	merged := make(map[string]string)
	var collectedNote string

	for _, r := range registry {
		if r.Manager != manager {
			continue
		}
		if r.Plugin != "" && r.Plugin != plugin {
			continue
		}
		for k, v := range r.Env {
			merged[k] = v
		}
		if collectedNote == "" && r.Note != "" {
			collectedNote = r.Note
		}
	}

	if len(merged) == 0 {
		return nil, ""
	}
	return merged, collectedNote
}

// MergeWithProcessEnv merges the provided environment variables with the current process
// environment and returns a slice in KEY=VALUE form suitable for exec.Cmd.Env.
func MergeWithProcessEnv(custom map[string]string) []string {
	if len(custom) == 0 {
		return os.Environ()
	}
	existing := os.Environ()
	// Create a map for quick override of existing keys
	resultMap := make(map[string]string, len(existing)+len(custom))

	// Populate with existing env
	for _, kv := range existing {
		// Split only at first '=' to preserve values containing '='
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				key := kv[:i]
				val := kv[i+1:]
				resultMap[key] = val
				break
			}
		}
	}

	// Apply custom overrides
	for k, v := range custom {
		resultMap[k] = v
	}

	// Rebuild slice
	merged := make([]string, 0, len(resultMap))
	for k, v := range resultMap {
		merged = append(merged, k+"="+v)
	}
	return merged
}

// === 사용자 정의 필터 로딩 ===

// userFilterYAML 은 사용자 정의 필터 설정의 YAML 스키마다.
type userFilterYAML struct {
	Filters []struct {
		Manager string            `yaml:"manager"`
		Plugin  string            `yaml:"plugin"`
		Kind    string            `yaml:"kind"` // advisory|conflict
		Level   string            `yaml:"level"`
		Env     map[string]string `yaml:"env"`
		Warning string            `yaml:"warning"`
		When    struct {
			OS           []string `yaml:"os"`
			Arch         []string `yaml:"arch"`
			HasCommand   []string `yaml:"has_command"`
			VersionRange struct {
				Manager string `yaml:"manager"`
			} `yaml:"version_range"`
		} `yaml:"when"`
		MatchEnv struct {
			PathContains []string `yaml:"path_contains"`
		} `yaml:"match_env"`
		Post []struct {
			Command     []string          `yaml:"command"`
			Env         map[string]string `yaml:"env"`
			Description string            `yaml:"description"`
			IgnoreError bool              `yaml:"ignore_error"`
		} `yaml:"post"`
	} `yaml:"filters"`
}

// loadUserFilters 는 ~/.gzh/pm/compat.yml 을 읽어 사용자 정의 필터를 생성한다.
func loadUserFilters() []CompatibilityFilter {
	configPath := filepath.Join(os.Getenv("HOME"), ".gzh", "pm", "compat.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}
	var cfg userFilterYAML
	if yaml.Unmarshal(data, &cfg) != nil {
		return nil
	}

	out := make([]CompatibilityFilter, 0, len(cfg.Filters))
	for _, f := range cfg.Filters {
		kind := FilterKindAdvisory
		if f.Kind == string(FilterKindConflict) {
			kind = FilterKindConflict
		}
		out = append(out, &userDefinedFilter{
			manager:  f.Manager,
			plugin:   f.Plugin,
			kind:     kind,
			env:      f.Env,
			warning:  f.Warning,
			postlist: convertUserPostToActions(f.Post),
			whenOS:   f.When.OS,
			whenArch: f.When.Arch,
			hasCmd:   f.When.HasCommand,
			pathHas:  f.MatchEnv.PathContains,
		})
	}
	return out
}

func convertUserPostToActions(in []struct {
	Command     []string          `yaml:"command"`
	Env         map[string]string `yaml:"env"`
	Description string            `yaml:"description"`
	IgnoreError bool              `yaml:"ignore_error"`
},
) []PostAction {
	out := make([]PostAction, 0, len(in))
	for _, p := range in {
		out = append(out, PostAction{
			Command:     p.Command,
			Env:         p.Env,
			Description: p.Description,
			IgnoreError: p.IgnoreError,
		})
	}
	return out
}

// userDefinedFilter 는 구성 파일에서 로드된 필터 구현체다.
type userDefinedFilter struct {
	manager  string
	plugin   string
	kind     FilterKind
	env      map[string]string
	warning  string
	postlist []PostAction
	whenOS   []string
	whenArch []string
	hasCmd   []string
	pathHas  []string
}

func (f *userDefinedFilter) Applies(manager, plugin string) bool {
	if f.manager != manager {
		return false
	}
	if f.plugin != "" && f.plugin != plugin {
		return false
	}
	// when OS
	if len(f.whenOS) > 0 {
		ok := false
		for _, osn := range f.whenOS {
			if strings.EqualFold(osn, runtime.GOOS) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	// when ARCH
	if len(f.whenArch) > 0 {
		ok := false
		for _, an := range f.whenArch {
			if strings.EqualFold(an, runtime.GOARCH) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	// has_command
	for _, c := range f.hasCmd {
		if _, err := execLookPath(c); err != nil {
			return false
		}
	}
	// path contains
	path := os.Getenv("PATH")
	for _, frag := range f.pathHas {
		if !strings.Contains(path, frag) {
			return false
		}
	}
	return true
}

// helper separated for testability
var execLookPath = func(file string) (string, error) { return exec.LookPath(file) }

func (f *userDefinedFilter) Env() map[string]string    { return f.env }
func (f *userDefinedFilter) Warning() string           { return f.warning }
func (f *userDefinedFilter) PostActions() []PostAction { return f.postlist }
func (f *userDefinedFilter) Kind() FilterKind          { return f.kind }
