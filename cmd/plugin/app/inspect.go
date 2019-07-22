package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fatsheep9146/kubectl-pvc/pkg/plugin"
)

var (
	inspectExample = `
	
`
)

type InspectOption struct {
	pvcname string
	pctx    *plugin.PvcContext
}

func NewInspectOption() *InspectOption {
	return &InspectOption{}
}

func NewInspectCommand() *cobra.Command {
	opts := NewInspectOption()

	cmd := &cobra.Command{
		Use:     "inspect",
		Short:   "inspect one specific pvc status",
		Example: inspectExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(pctx); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}

			if err := opts.Run(args); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func (opts *InspectOption) Complete(pctx *plugin.PvcContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *InspectOption) Validate() error {
	return nil
}

func (opts *InspectOption) Run(args []string) (err error) {
	if len(args) == 0 {
		return fmt.Errorf("user should input one pvc to inspect")
	}

	pvcStatus, err := opts.pctx.GetPvcDetail(args[0])
	if err != nil {
		return err
	}

	plugin.FormatPvcDetail(os.Stdout, pvcStatus)

	return nil
}
