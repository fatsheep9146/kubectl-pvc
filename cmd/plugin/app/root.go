package app

import (
	"github.com/spf13/cobra"
)

type PvcOptions struct {
	flags  *genericclioptions.ConfigFlags
	k8scli *kubernetes.Clientset
	config *restclient.Config
}

func NewPvcCommand() *cobra.Command {

	return nil
}
