package app

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/chartutil"
	"helm.sh/helm/pkg/strvals"
	"k8s.io/klog"

	"github.com/alauda/kubectl-captain/pkg/plugin"
)

var (
	updateExample = `
	# update one helmerequest
	kubectl captain upgrade -n <namespace> --name <name> -v <version> --set=<values>
`
)

type UpgradeOption struct {
	version string
	values []string
	pctx    *plugin.CaptainContext
}

func NewUpdateOption() *UpgradeOption {
	return &UpgradeOption{}
}

func NewUpgradeCommand() *cobra.Command {
	opts := NewUpdateOption()

	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "upgrade a helmrequest",
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

	cmd.Flags().StringArrayVarP(&opts.values, "set", "s", []string{}, "custom values")
	cmd.Flags().StringVarP(&opts.version, "version", "v", "", "the chart version you want to use ")
	return cmd
}

func (opts *UpgradeOption) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *UpgradeOption) Validate() error {
	return nil
}

// Run do the real update
// 1. save the old spec to annotation
// 2. update
func (opts *UpgradeOption) Run() (err error) {
	if opts.pctx == nil {
		klog.Errorf("UpgradeOption.ctx should not be nil")
		return fmt.Errorf("UpgradeOption.ctx should not be nil")
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


	// merge values....oh,we have to import helm now....
	base := hr.Spec.Values.AsMap()
	for _, value := range opts.values {
		if err := strvals.ParseInto(value, base); err != nil {
			return errors.Wrap(err, "failed parsing --set data")
		}
	}

	hr.Spec.Values = chartutil.Values(base)

	_, err = pctx.UpdateHelmRequest(hr)
	return err
}
