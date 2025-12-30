package gitsync

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/app"

	"github.com/gizzahub/gzh-cli-gitforge/pkg/reposync"
	"github.com/gizzahub/gzh-cli-gitforge/pkg/reposynccli"
)

// NewGitSyncCmd는 `gz git-sync` 커맨드를 생성합니다.
func NewGitSyncCmd(_ *app.AppContext) *cobra.Command {
	planner := reposync.FSPlanner{}
	executor := reposync.GitExecutor{}
	state := reposync.NewInMemoryStateStore()
	orchestrator := reposync.NewOrchestrator(planner, executor, state)

	factory := reposynccli.CommandFactory{
		Use:          "git-sync",
		Short:        "Git repository synchronization",
		Orchestrator: orchestrator,
		SpecLoader:   reposynccli.FileSpecLoader{},
	}

	return factory.NewRootCmd()
}
