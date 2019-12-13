package app

import (
	"fmt"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

var (
	createRepoExample = `
	# create a chartrepo with username and password
	kubectl captain create-repo foo --url=www.example.org --username=tom --password=lisa
`
)

type CreateRepoOption struct {
	url string

	username string
	password string

	wait    bool
	timeout int

	pctx *plugin.CaptainContext
}

func NewCreateRepoOption() *CreateRepoOption {
	return &CreateRepoOption{}
}

func NewCreateRepoCommand() *cobra.Command {
	opts := NewCreateRepoOption()

	cmd := &cobra.Command{
		Use:     "create-repo",
		Short:   "create a chartrepo",
		Example: createRepoExample,
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
	cmd.Flags().StringVarP(&opts.url, "url", "", "", "repo url")
	cmd.Flags().StringVarP(&opts.username, "username", "u", "", "repo username")
	cmd.Flags().StringVarP(&opts.password, "password", "p", "", "repo password")
	return cmd
}

func (opts *CreateRepoOption) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *CreateRepoOption) Validate() error {
	return nil
}

// Run do the real update
// 1. save the old spec to annotation
// 2. update
func (opts *CreateRepoOption) Run(args []string) (err error) {
	if opts.pctx == nil {
		klog.Errorf("UpgradeOption.ctx should not be nil")
		return fmt.Errorf("UpgradeOption.ctx should not be nil")
	}

	if len(args) == 0 {
		return fmt.Errorf("user should input chartrepo name to create")
	}

	name := args[0]
	pctx := opts.pctx
	var cr v1alpha1.ChartRepo
	cr.Spec.URL = opts.url
	cr.Namespace = pctx.GetNamespace()
	cr.Name = name

	if opts.username != "" && opts.password != "" {
		cr.Spec.Secret = &v1.SecretReference{
			Name:      name,
			Namespace: pctx.GetNamespace(),
		}

		importOptions := ImportOptions{
			repoNamespace: pctx.GetNamespace(),
			pctx:          pctx,
		}

		if err := importOptions.createRepoSecret(opts.username, opts.password, name); err != nil {
			return err
		}
	}

	_, err = pctx.CreateChartRepo(&cr)
	if !opts.wait {
		return err
	}

	if err != nil {
		klog.Error("Create chartrepo error: ", err)
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
