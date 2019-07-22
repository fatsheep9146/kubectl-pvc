package main

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"

	"github.com/fatsheep9146/kubectl-pvc/cmd/plugin/app"
)

func main() {
	klog.InitFlags(nil)

	cmd := app.NewPvcCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
