package doctor

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type doctorCmdProvider struct{}

func (doctorCmdProvider) Command() *cobra.Command {
	DoctorCmd.Hidden = true
	return DoctorCmd
}

func init() {
	registry.Register(doctorCmdProvider{})
}
