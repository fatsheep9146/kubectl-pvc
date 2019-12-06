package app

import (
	"fmt"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/chartutil"
	"helm.sh/helm/pkg/strvals"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

var (
	createExample = `
	# create helmrequest in default ns to set it's chart version to 1.5.0 and set value 'a=b'
	kubectl captain create foo --chart=stable/nginx-ingress -v 1.5.0 --set=a=b
`
)

type CreateOption struct {
	chart   string
	version string
	values  []string

	wait    bool
	timeout int

	pctx *plugin.CaptainContext
}

func NewCreateOption() *CreateOption {
	return &CreateOption{}
}

func NewCreateCommand() *cobra.Command {
	opts := NewCreateOption()

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "create a helmrequest",
		Example: createExample,
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

	cmd.Flags().StringArrayVarP(&opts.values, "set", "s", []string{}, "custom values")
	cmd.Flags().StringVarP(&opts.version, "version", "v", "", "the chart version you want to use ")
	cmd.Flags().BoolVarP(&opts.wait, "wait", "w", false, "wait for the helmrequest to be synced")
	cmd.Flags().IntVarP(&opts.timeout, "timeout", "t", 0, "timeout for the wait")
	cmd.Flags().StringVarP(&opts.chart, "chart", "c", "", "chart name, format: <repo>/<chart>")
	return cmd
}

func (opts *CreateOption) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *CreateOption) Validate() error {
	return nil
}

// Run do the real update
// 1. save the old spec to annotation
// 2. update
func (opts *CreateOption) Run(args []string) (err error) {
	if opts.pctx == nil {
		klog.Errorf("UpgradeOption.ctx should not be nil")
		return fmt.Errorf("UpgradeOption.ctx should not be nil")
	}

	if len(args) == 0 {
		return fmt.Errorf("user should input helmrequest name to create")
	}

	name := args[0]
	pctx := opts.pctx
	var hr v1alpha1.HelmRequest

	hr.Spec.Version = opts.version
	hr.Spec.Chart = opts.chart
	hr.Name = name
	hr.Namespace = pctx.GetNamespace()

	// merge values....oh,we have to import helm now....
	base := hr.Spec.Values.AsMap()
	for _, value := range opts.values {
		if err := strvals.ParseInto(value, base); err != nil {
			return errors.Wrap(err, "failed parsing --set data")
		}
	}

	hr.Spec.Values = chartutil.Values(base)

	_, err = pctx.CreateHelmRequest(&hr)
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
