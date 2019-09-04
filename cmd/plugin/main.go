package main

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"

	"github.com/alauda/kubectl-captain/cmd/plugin/app"
)

func main() {
	klog.InitFlags(nil)

	cmd := app.NewCaptainCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
