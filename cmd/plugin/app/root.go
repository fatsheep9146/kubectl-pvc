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
	var name string

	cmd := &cobra.Command{
		Use:   "cpatain",
		Short: "kubectl captain: access helmrequest resource",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := pctx.Complete(ns, name)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&ns, "namespace", "n", "default", "the namespace you want to check")
	cmd.PersistentFlags().StringVarP(&name, "name", "", "", "the specific helmrequest name you want to operate on")
	cmd.AddCommand(NewUpgradeCommand())
	cmd.AddCommand(NewRollbackCommand())

	return cmd
}
