package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "v1.0.0"

func NewVersionCommand() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubectl-captain",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("kubectl-captain: " + version)
		},
	}
	return versionCmd
}
