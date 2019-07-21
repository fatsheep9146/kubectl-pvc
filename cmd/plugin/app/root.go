package app

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PvcOptions struct {
	flags  *genericclioptions.ConfigFlags
	k8scli *kubernetes.Clientset
	config *restclient.Config
}

func NewPvcCommand() *cobra.Command {

	return nil
}
