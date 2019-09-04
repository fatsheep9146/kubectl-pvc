package app

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/alauda/kubectl-captain/pkg/plugin"
)

var (
	updateExample = `
	# update one helmerequest
	kubectl captain update -n <namespace> --name <name> -v <version>
`
)

type UpdateOption struct {
	version string
	pctx    *plugin.CaptainContext
}

func NewUpdateOption() *UpdateOption {
	return &UpdateOption{}
}

func NewUpdateCommand() *cobra.Command {
	opts := NewUpdateOption()

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "update one helmrequest",
		Example: updateExample,
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

	cmd.Flags().StringVarP(&opts.version, "version", "v", "", "the chart version you want to use ")
	return cmd
}

func (opts *UpdateOption) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *UpdateOption) Validate() error {
	return nil
}

// Run do the real update
// 1. save the old spec to annotation
// 2. update
func (opts *UpdateOption) Run() (err error) {
	if opts.pctx == nil {
		klog.Errorf("UpdateOption.ctx should not be nil")
		return fmt.Errorf("UpdateOption.ctx should not be nil")
	}

	pctx := opts.pctx
	hr, err := pctx.GetHelmRequest()
	if err != nil {
		return err
	}

	// TODO: remove
	old, err := json.Marshal(hr.Spec)
	if err != nil {
		return err
	}

	if hr.Annotations == nil {
		hr.Annotations = make(map[string]string)
	}
	 hr.Annotations["last-spec"] = string(old)

	hr.Spec.Version = opts.version
	_, err = pctx.UpdateHelmRequest(hr)
	return err
}
