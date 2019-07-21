package main

import (
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog"

	"github.com/fatsheep9146/kubectl-pvc/cmd/plugin/app"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	cmd := app.NewPvcCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
