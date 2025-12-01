// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package extensions

// Config는 사용자 정의 확장 설정을 나타냅니다
type Config struct {
	Aliases  map[string]AliasConfig  `yaml:"aliases"`  // 명령어 별칭
	External []ExternalCommandConfig `yaml:"external"` // 외부 명령어 통합
}

// AliasConfig는 명령어 별칭 설정을 나타냅니다
type AliasConfig struct {
	Command     string   `yaml:"command,omitempty"`     // 단순 별칭 명령어
	Description string   `yaml:"description"`           // 설명
	Steps       []string `yaml:"steps,omitempty"`       // 다단계 워크플로우 (미래 확장)
	Params      []Param  `yaml:"params,omitempty"`      // 파라미터 (미래 확장)
}

// Param은 별칭 명령어의 파라미터를 나타냅니다
type Param struct {
	Name        string `yaml:"name"`        // 파라미터 이름
	Description string `yaml:"description"` // 파라미터 설명
	Required    bool   `yaml:"required"`    // 필수 여부
}

// ExternalCommandConfig는 외부 명령어 통합 설정을 나타냅니다
type ExternalCommandConfig struct {
	Name        string   `yaml:"name"`                  // 명령어 이름
	Command     string   `yaml:"command"`               // 실행할 명령어 경로
	Description string   `yaml:"description"`           // 설명
	Passthrough bool     `yaml:"passthrough,omitempty"` // 인자 통과 여부
	Args        []string `yaml:"args,omitempty"`        // 기본 인자
}
