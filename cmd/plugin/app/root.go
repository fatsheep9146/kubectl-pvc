package app

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/fatsheep9146/kubectl-pvc/pkg/plugin"
)

var pctx *plugin.PvcContext = nil

func NewPvcCommand(streams genericclioptions.IOStreams) *cobra.Command {
	pctx = plugin.NewPvcContext(streams)
	var ns string

	cmd := &cobra.Command{
		Use:   "pvc",
		Short: "kubectl pvc: check info about pvc in faster way",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := pctx.Complete(ns)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&ns, "namespace", "n", "default", "the namespace you want to check")
	cmd.AddCommand(NewLsCommand())
	cmd.AddCommand(NewInspectCommand())

	return cmd
}
