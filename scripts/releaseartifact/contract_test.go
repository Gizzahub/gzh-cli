package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// 권한과 채널은 실행 결과와 별개로 선언 자체에서 닫혀 있어야 한다.
func TestReleaseWorkflowPrivilegeBoundary(t *testing.T) {
	root := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Permissions map[string]string `yaml:"permissions"`
		Env         map[string]string `yaml:"env"`
		Jobs        map[string]struct {
			If          string            `yaml:"if"`
			Permissions map[string]string `yaml:"permissions"`
			Steps       []struct {
				Uses string `yaml:"uses"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &workflow); err != nil {
		t.Fatal(err)
	}
	if len(workflow.Permissions) != 1 || workflow.Permissions["contents"] != "read" || len(workflow.Jobs) != 2 || strings.Contains(string(b), "secrets.") {
		t.Fatal("release workflow must have only two credential-free build jobs with read-only defaults")
	}
	pins, err := os.ReadFile(filepath.Join(root, ".make", "tools.mk"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"GORELEASER_VERSION", "SYFT_VERSION", "COSIGN_VERSION"} {
		match := regexp.MustCompile(`(?m)^` + name + `\s*:=\s*(v[^\s]+)\s*$`).FindSubmatch(pins)
		if len(match) != 2 || workflow.Env[name] != string(match[1]) {
			t.Fatalf("workflow pin must match tools.mk: %s", name)
		}
	}
	for name, job := range workflow.Jobs {
		switch name {
		case "verify-release":
			if job.If != "github.event_name == 'workflow_dispatch'" || len(job.Permissions) != 1 || job.Permissions["contents"] != "read" {
				t.Fatal("manual preview must have no OIDC/write permission")
			}
		case "build-tag":
			if job.If != "github.event_name == 'push' && startsWith(github.ref, 'refs/tags/v')" || len(job.Permissions) != 2 || job.Permissions["contents"] != "read" || job.Permissions["id-token"] != "write" {
				t.Fatal("only tag-push build may receive OIDC; no publishing permission")
			}
		default:
			t.Fatalf("unreviewed job %s", name)
		}
		signers, builds, prepares, verifiers := 0, 0, 0, 0
		for _, step := range job.Steps {
			if step.Uses != "" && !strings.HasPrefix(step.Uses, "./") && !regexp.MustCompile(`@[0-9a-f]{40}$`).MatchString(step.Uses) {
				t.Fatalf("external action is not immutable: %s", step.Uses)
			}
			if strings.Contains(step.Run, "goreleaser release") {
				builds++
				if !strings.Contains(step.Run, "--skip=publish") {
					t.Fatal("every workflow build must skip publishing")
				}
			}
			if strings.Contains(step.Run, "releaseartifact prepare") {
				prepares++
				if name == "verify-release" && !strings.Contains(step.Run, "--snapshot") {
					t.Fatal("manual stage must be an unpublishable snapshot")
				}
			}
			if strings.Contains(step.Run, "cosign sign-blob") {
				signers++
			}
			if strings.Contains(step.Run, "releaseartifact verify") {
				verifiers++
			}
			if strings.Contains(step.Run, "releaseartifact publish") {
				t.Fatal("workflow cannot publish")
			}
		}
		if builds != 1 || prepares != 1 || (name == "verify-release" && (signers != 0 || verifiers != 0)) || (name == "build-tag" && (signers != 1 || verifiers != 1)) {
			t.Fatal("build/stage/sign/verify separation changed")
		}
	}
}

func TestReleaseChannelConfiguration(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".goreleaser.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(b, &config); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"nfpms", "dockers", "dockers_v2", "docker_signs", "scoops", "aurs", "chocolateys", "signs", "publishers"} {
		if _, exists := config[name]; exists {
			t.Fatalf("first-channel configuration cannot declare %s", name)
		}
	}
	release, ok := config["release"].(map[string]any)
	if !ok || release["disable"] != true || release["mode"] != "keep-existing" {
		t.Fatal("GoReleaser publication must be disabled and never replace existing releases")
	}
	casks, ok := config["homebrew_casks"].([]any)
	if !ok || len(casks) != 1 {
		t.Fatal("exactly one cask is required")
	}
	cask, ok := casks[0].(map[string]any)
	if !ok || cask["name"] != "gz" || cask["skip_upload"] != true || cask["license"] != "MIT" {
		t.Fatal("cask must remain generated-only with source MIT metadata")
	}
	url, ok := cask["url"].(map[string]any)
	if !ok || url["template"] != "https://github.com/Gizzahub/gzh-cli/releases/download/{{ .Tag }}/{{ .ArtifactName }}" {
		t.Fatal("disabled release requires an explicit canonical cask download URL")
	}
	license, err := os.ReadFile(filepath.Join("..", "..", "LICENSE"))
	if err != nil || !strings.HasPrefix(string(license), "MIT License") {
		t.Fatal("MIT metadata must match the actual source LICENSE")
	}
}
