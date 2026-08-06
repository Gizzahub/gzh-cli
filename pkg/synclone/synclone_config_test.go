//nolint:testpackage // White-box testing needed for internal function access
package bulkclone

import (
	"testing"

	"gopkg.in/yaml.v2"
)

// func (receiver ) name()  {
//
//}

func TestReadConfig(t *testing.T) {
	// use bulk-clone.yaml
	// call setclond_config.ReadConfig
	config := &bulkCloneConfig{}
	// bulkCloneConfig.ReadConfig("../../../test")
	// config.ReadConfig("../../../test")
	if err := config.ReadConfig("./"); err != nil {
		t.Logf("Warning: failed to read config: %v", err)
	}
	// t.Log(yaml.Marshal(config))
	// print unmarshal yaml format
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(yamlData))
}

// TestValidateConfigChecksRepoRootElements는 repo_roots 항목 하나하나가
// 실제로 검사되는지 본다.
//
// validator는 구조체 필드는 알아서 따라 들어가지만 조각(slice)은 dive를
// 붙여야 원소를 본다. 그 태그가 없던 동안 BulkCloneGithub의 required와
// oneof 네 개가 전부 죽어 있었고, 아무도 몰랐다. 태그만 봐서는 죽었는지
// 살았는지 알 수 없으니 여기서 값으로 확인한다.
func TestValidateConfigChecksRepoRootElements(t *testing.T) {
	// Default.Protocol은 그 자체로 required다. 여기서 보려는 것은
	// RepoRoots 쪽이므로 Default는 늘 옳게 채워 두고 시작한다.
	validDefault := bulkCloneDefault{Protocol: "https"}

	completeRoot := BulkCloneGithub{
		RootPath: "/tmp/repos",
		Provider: "github",
		Protocol: "https",
		OrgName:  "acme",
	}

	testCases := []struct {
		name      string
		repoRoots []BulkCloneGithub
		wantErr   bool
	}{
		{
			name:      "제대로 채운 항목은 통과한다",
			repoRoots: []BulkCloneGithub{completeRoot},
			wantErr:   false,
		},
		{
			name:      "항목이 없어도 통과한다",
			repoRoots: nil,
			wantErr:   false,
		},
		{
			name:      "빈 항목은 걸러야 한다",
			repoRoots: []BulkCloneGithub{{}},
			wantErr:   true,
		},
		{
			name: "org_name이 빠지면 걸러야 한다",
			repoRoots: []BulkCloneGithub{{
				RootPath: "/tmp/repos",
				Provider: "github",
				Protocol: "https",
			}},
			wantErr: true,
		},
		{
			name: "protocol이 허용된 셋 중 하나가 아니면 걸러야 한다",
			repoRoots: []BulkCloneGithub{{
				RootPath: "/tmp/repos",
				Provider: "github",
				Protocol: "carrier-pigeon",
				OrgName:  "acme",
			}},
			wantErr: true,
		},
		{
			// 앞 항목이 옳아도 뒤 항목이 틀리면 걸러야 한다. dive가 원소
			// 하나만 보고 마는 것이 아님을 적어 둔다.
			name:      "뒤쪽 항목이 틀려도 걸러야 한다",
			repoRoots: []BulkCloneGithub{completeRoot, {}},
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &bulkCloneConfig{
				Version:   "1.0.0",
				Default:   validDefault,
				RepoRoots: tc.repoRoots,
			}

			err := cfg.validateConfig()
			if tc.wantErr && err == nil {
				t.Errorf("오류가 나야 하는데 nil이 나왔다: repoRoots=%+v", tc.repoRoots)
			}

			if !tc.wantErr && err != nil {
				t.Errorf("통과해야 하는데 오류가 났다: %v", err)
			}
		})
	}
}
