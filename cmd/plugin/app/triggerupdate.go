package app

import (
	"fmt"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/chartutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

var (
	forceUpdateExample = `
	# trigger helmrequest in default ns
	kubectl captain trigger-update foo -n default 
`
)

// ForceUpdateOption force trigger a update on target HelmRequest.
type TriggerUpdateOption struct {
	wait    bool
	timeout int

	pctx *plugin.CaptainContext
}

func (opts *TriggerUpdateOption) Complete(context *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *TriggerUpdateOption) Validate(args []string) error {
	if opts.pctx == nil {
		klog.Errorf("UpgradeOption.ctx should not be nil")
		return fmt.Errorf("UpgradeOption.ctx should not be nil")
	}
	if len(args) == 0 {
		return fmt.Errorf("user should input helmrequest name to upgrade")
	}
	return nil

}

func (opts *TriggerUpdateOption) Run(args []string) error {
	pctx := opts.pctx
	hr, err := pctx.GetHelmRequest(args[0])
	if err != nil {
		return err
	}

	// merge values....oh,we have to import helm now....
	if hr.Spec.Values == nil {
		hr.Spec.Values = chartutil.Values{}
	}
	hr.Spec.Values["_retry"] = 1

	_, err = pctx.UpdateHelmRequest(hr)
	if err != nil {
		return err
	}

	// get again
	hr, err = pctx.GetHelmRequest(args[0])
	if err != nil {
		return err
	}
	if hr.Spec.Values != nil {
		delete(hr.Spec.Values, "_retry")
	}
	hr.Status.Phase = v1alpha1.HelmRequestPending

	_, err = pctx.UpdateHelmRequestStatus(hr)
	if !opts.wait {
		return err
	}

	_, err = pctx.UpdateHelmRequest(hr)
	if !opts.wait {
		return err
	}

	klog.Info("Start wait for helmrequest to be synced")

	f := func() (done bool, err error) {
		result, err := pctx.GetHelmRequest(hr.GetName())
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

func NewTriggerUpdateOption() *TriggerUpdateOption {
	return &TriggerUpdateOption{}
}

// NewTriggerUpdateCommand ...
// We need this command because captain will stop sync helmrequest resource after retries. If we want to force trigger
// a update, we need to update the spec part of HelmRequest for now. So this command will perform this task. The
// imperfect part is, we have to trigger the update twice in this command, one the set and custom value(no effect on the
// real deploy) and one to reset this update.
func NewTriggerUpdateCommand() *cobra.Command {
	opts := NewTriggerUpdateOption()

	cmd := &cobra.Command{
		Use:     "trigger-update",
		Short:   "trigger update on helmrequest",
		Example: forceUpdateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(pctx); err != nil {
				return err
			}

			if err := opts.Validate(args); err != nil {
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
