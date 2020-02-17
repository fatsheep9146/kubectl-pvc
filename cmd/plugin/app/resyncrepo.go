package app

import (
	"fmt"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

var (
	resyncRepoExample = `
	# resync chart repo foo 
	kubectl captain resync-repo foo -n system
`
)

type ResyncRepoOption struct {
	wait    bool
	timeout int

	pctx *plugin.CaptainContext
}

func NewResyncRepoOption() *ResyncRepoOption {
	return &ResyncRepoOption{}
}

func NewResyncRepoCommand() *cobra.Command {
	opts := NewResyncRepoOption()

	cmd := &cobra.Command{
		Use:     "resync-repo",
		Short:   "resync a chart repo",
		Example: resyncRepoExample,
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

	cmd.Flags().BoolVarP(&opts.wait, "wait", "w", false, "wait for the helmrequest to be synced")
	cmd.Flags().IntVarP(&opts.timeout, "timeout", "t", 0, "timeout for the wait")
	return cmd
}

func (opts *ResyncRepoOption) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *ResyncRepoOption) Validate() error {
	return nil
}

// Run do the real resync
// update status to pending and wati for it to be ready
func (opts *ResyncRepoOption) Run(args []string) (err error) {
	if opts.pctx == nil {
		klog.Errorf("ResyncOption.ctx should not be nil")
		return fmt.Errorf("ResyncOption.ctx should not be nil")
	}

	if len(args) == 0 {
		return fmt.Errorf("user should input chartrepo name to resync")
	}

	name := args[0]
	pctx := opts.pctx
	namespace := pctx.GetNamespace()

	repo, err := pctx.GetChartRepo(name, namespace)
	if err != nil {
		klog.Error("Get chartrepo error: ", err)
		return err
	}

	repo.Status.Phase = "Pending"

	_, err = pctx.UpdateChartRepo(repo)
	if err != nil {
		klog.Error("Update chartrepo error: ", err)
		return err
	}

	if !opts.wait {
		return err
	}

	klog.Info("Start wait for chartrepo to be synced")

	f := func() (done bool, err error) {
		result, err := pctx.GetChartRepo(name, pctx.GetNamespace())
		if err != nil {
			return false, err
		}
		return result.Status.Phase == "Synced", nil
	}

	if opts.timeout != 0 {
		return wait.Poll(1*time.Second, time.Duration(opts.timeout)*time.Second, f)
	} else {
		return wait.PollInfinite(1*time.Second, f)
	}

}
