package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

var (
	rollbackExample = `
	# rollback  helmerequest foo to last configurations
	kubectl captain rollback foo -n default
`
)

type RollbackOption struct {
	pctx *plugin.CaptainContext

	wait    bool
	timeout int
}

func NewRollbackOption() *RollbackOption {
	return &RollbackOption{}
}

func NewRollbackCommand() *cobra.Command {
	opts := NewRollbackOption()

	cmd := &cobra.Command{
		Use:     "rollback",
		Short:   "rollback a helmrequest",
		Example: rollbackExample,
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

func (opts *RollbackOption) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *RollbackOption) Validate() error {
	return nil
}

// Run rollback a helmrequest
func (opts *RollbackOption) Run(args []string) (err error) {
	if opts.pctx == nil {
		klog.Errorf("UpgradeOption.ctx should not be nil")
		return fmt.Errorf("UpgradeOption.ctx should not be nil")
	}

	if len(args) == 0 {
		return fmt.Errorf("user should input a helmrequest name  to rollback")
	}

	pctx := opts.pctx
	hr, err := pctx.GetHelmRequest(args[0])
	if err != nil {
		return err
	}

	key := "last-spec"

	if hr.Annotations == nil || hr.Annotations[key] == "" {
		return errors.New("no last configuration found")
	}

	data := hr.Annotations[key]

	var new v1alpha1.HelmRequestSpec
	if err = json.Unmarshal([]byte(data), &new); err != nil {
		return err
	}

	hr.Spec = new

	_, err = pctx.UpdateHelmRequest(hr)
	if !opts.wait {
		return err
	}

	klog.Info("Start wait for helmrequest to be synced")

	// TEST: should we update status too
	f := func() (done bool, err error) {
		result, err := pctx.GetHelmRequest(hr.GetName())
		if err != nil {
			return false, err
		}
		return result.Status.Phase == "Synced", nil
	}

	if opts.timeout != 0 {
		err = wait.Poll(1*time.Second, time.Duration(opts.timeout)*time.Second, f)
	} else {
		err = wait.PollInfinite(1*time.Second, f)
	}

	if err == nil {
		klog.Infof("Rollback  to version: %s", new.Version)
		message := fmt.Sprintf("Rollback helmrequest %s to version %s ", hr.Name, hr.Spec.Version)
		pctx.CreateEvent("Normal", "Synced", message, hr)
	} else {
		message := fmt.Sprintf("Rollback helmrequest %s to version %s error: %s", hr.Name, hr.Spec.Version, err.Error())
		pctx.CreateEvent("Warning", "FailedRollback", message, hr)
	}

	return err

}
