package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var (
	rollbackExample = `
	# rollback a helmerequest
	kubectl captain rollback -n <namespace> -r <helmrequest> 
`
)

type RollbackOption struct {
	pctx *plugin.CaptainContext
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

			if err := opts.Run(); err != nil {
				return err
			}
			return nil
		},
	}

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
func (opts *RollbackOption) Run() (err error) {
	if opts.pctx == nil {
		klog.Errorf("UpdateOption.ctx should not be nil")
		return fmt.Errorf("UpdateOption.ctx should not be nil")
	}

	pctx := opts.pctx
	hr, err := pctx.GetHelmRequest()
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
	return err

}