//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSshOptions(t *testing.T) {
	// BaseOptions의 기본값 테스트
	opts := &BaseOptions{}

	// 기본값들이 empty/false인지 확인
	assert.Empty(t, opts.ConfigPath) // 기본값은 비어있음
	assert.Empty(t, opts.StorePath)  // 기본값은 비어있음
	assert.False(t, opts.Force)      // 기본값은 false
	assert.False(t, opts.ListAll)    // 기본값은 false
}

func TestNewSshCmd(t *testing.T) {
	cmd := newSshCmd() // 함수명은 newSshCmd

	assert.Equal(t, "ssh", cmd.Use)
	assert.Contains(t, cmd.Short, "ssh")
	assert.NotEmpty(t, cmd.Long)

	// BaseCommand를 사용하므로 save, load, list 서브커맨드가 자동 생성됨
	subcommands := cmd.Commands()
	assert.True(t, len(subcommands) >= 3) // save, load, list

	// 서브커맨드 존재 확인
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Use] = true
	}

	assert.True(t, subcommandNames["save"], "save subcommand should exist")
	assert.True(t, subcommandNames["load"], "load subcommand should exist")
	assert.True(t, subcommandNames["list"], "list subcommand should exist")
}
