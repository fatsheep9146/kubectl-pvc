package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/fatsheep9146/kubectl-pvc/pkg/plugin"
)

var (
	lsExample = `
	# check all pvcs of given namespace
	kubectl pvc ls -n <namespace>

	# check all pvcs of given pod
	kubectl pvc ls -n <namespace> -p <pod>
`
)

type LsOption struct {
	podname string
	pctx    *plugin.PvcContext
}

func NewLsOption() *LsOption {
	return &LsOption{}
}

func NewLsCommand() *cobra.Command {
	opts := NewLsOption()

	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "list all pvcs of the whole namespace or one pod",
		Example: lsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(pctx); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}

			if err := opts.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.podname, "pod", "p", "", "the specific pod name you want to check")
	return cmd
}

func (opts *LsOption) Complete(pctx *plugin.PvcContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *LsOption) Validate() error {
	return nil
}

func (opts *LsOption) Run() (err error) {
	if opts.pctx == nil {
		klog.Errorf("LsOption.ctx should not be nil")
		return fmt.Errorf("LsOption.ctx should not be nil")
	}

	pctx := opts.pctx
	pvcs := make([]v1.PersistentVolumeClaim, 0)

	if opts.podname == "" {
		pvcs, err = pctx.ListPvcs()
		if err != nil {
			klog.Errorf("list pvcs of namespace %v failed, err %v", pctx.GetNamespace(), err)
		}
	} else {
		pvcs, err = pctx.ListPvcsByPod(opts.podname)
		if err != nil {
			klog.Errorf("list pvcs of pod %v/%v failed, err %v", pctx.GetNamespace(), opts.podname, err)
		}
	}

	plugin.Format(os.Stdout, pvcs)

	return nil
}
