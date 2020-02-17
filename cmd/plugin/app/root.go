package app

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/alauda/kubectl-captain/pkg/plugin"
)

var pctx *plugin.CaptainContext = nil

// NewCaptainCommand init captain command
func NewCaptainCommand(streams genericclioptions.IOStreams) *cobra.Command {
	pctx = plugin.NewCaptainContext(streams)
	var ns string

	cmd := &cobra.Command{
		Use:   "captain",
		Short: "kubectl captain: access helmrequest/chartrepo resource",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := pctx.Complete(ns)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&ns, "namespace", "n", "default", "the working namespace")
	cmd.AddCommand(NewCreateRepoCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewUpgradeCommand())
	cmd.AddCommand(NewRollbackCommand())
	cmd.AddCommand(NewImportCommand())
	cmd.AddCommand(NewTriggerUpdateCommand())
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewResyncRepoCommand())

	return cmd
}
