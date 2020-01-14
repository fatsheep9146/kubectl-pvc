package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/alauda/kubectl-captain/pkg/plugin"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/chartutil"
	"io/ioutil"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

var (
	importExample = `
	# import helm v2 release foo to a helmrequest 
 	kubectl captain import foo -n default 
`
)

type ImportOptions struct {
	pctx *plugin.CaptainContext

	repoName      string
	repoNamespace string

	helmBinPath string

	// useful in business cluster,
	createCR bool

	wait    bool
	timeout int
}

func NewImportOptions() *ImportOptions {
	return &ImportOptions{}
}

func NewImportCommand() *cobra.Command {
	opts := NewImportOptions()

	cmd := &cobra.Command{
		Use:        "import",
		Aliases:    nil,
		SuggestFor: nil,
		Short:      "import a helm release to helmrequest",
		Long:       "",
		Example:    importExample,
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
			klog.Info("Create helmrequest")

			return nil

		},
		SuggestionsMinimumDistance: 0,
		TraverseChildren:           false,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	}
	cmd.Flags().StringVarP(&opts.repoName, "repo", "r", "", "the repo name this chart belongs")
	cmd.Flags().StringVarP(&opts.repoNamespace, "repo-namespace", "", "alauda-system", "the ChartRepo resources' namespace")
	cmd.Flags().StringVarP(&opts.helmBinPath, "helm-bin-path", "", "/usr/local/bin/helm", "the helm binary path")
	cmd.Flags().BoolVarP(&opts.createCR, "create-chartrepo", "", true, "create chartrepo")
	cmd.Flags().BoolVarP(&opts.wait, "wait", "w", false, "wait for the helmrequest to be synced")
	cmd.Flags().IntVarP(&opts.timeout, "timeout", "t", 0, "timeout for the wait")
	return cmd
}

func (opts *ImportOptions) Complete(pctx *plugin.CaptainContext) error {
	opts.pctx = pctx
	return nil
}

func (opts *ImportOptions) Validate() error {
	return nil
}

// Run import a helm release to a HelmRequest, create ChartRepo if not exist
func (opts *ImportOptions) Run(args []string) (err error) {
	if opts.pctx == nil {
		klog.Errorf("ImportOptions.ctx shoud not be nil")
		return fmt.Errorf("ImportOtions.ctx shoud not be nil")
	}

	if len(args) == 0 {
		return fmt.Errorf("user should input a release name")
	}

	name := args[0]

	pctx := opts.pctx
	klog.Infof("Target namespace: %s", pctx.GetNamespace())
	// get values
	out, err := exec.Command(opts.helmBinPath, "get", "values", name).Output()
	if err != nil {
		return err
	}

	var values chartutil.Values
	err = yaml.Unmarshal(out, &values)
	if err != nil {
		return err
	}

	// get chart and version
	chart, version, err := opts.getChartVersion(name)
	if err != nil {
		return err
	}

	if opts.createCR {
		// check chartrepo exist
		_, err = opts.pctx.GetChartRepo(opts.repoName, opts.repoNamespace)
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.Info("Create chart repo: ", opts.repoName)
				if err := opts.createChartRepo(opts.repoName, opts.repoNamespace); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			klog.Info("Using exiting chartrepo")
		}

	}

	hr := v1alpha1.HelmRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HelmRequest",
			APIVersion: "app.alauda.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: opts.pctx.GetNamespace(),
		},
		Spec: v1alpha1.HelmRequestSpec{
			ClusterName:          "",
			InstallToAllClusters: false,
			Dependencies:         nil,
			ReleaseName:          name,
			Chart:                fmt.Sprintf("%s/%s", opts.repoName, chart),
			Version:              version,
			Namespace:            opts.pctx.GetNamespace(),
			ValuesFrom:           nil,
			HelmValues:           v1alpha1.HelmValues{Values: values},
		},
	}

	_, err = pctx.CreateHelmRequest(&hr)
	if !opts.wait {
		if err == nil {
			klog.Info("Create helmrequest: ", hr.GetName())
		}
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

type repo struct {
	CAFile   string `yaml:"caFile"`
	Cache    string `yaml:"cache"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Username string `yaml:"username"`
	URL      string `yaml:"url"`
}

type repoFile struct {
	APIVersion   string `yaml:"apiVersion"`
	Generated    string `yaml:"generated"`
	Repositories []repo `yaml:"repositories"`
}

func (opts *ImportOptions) createChartRepo(name, namespace string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	path := "/.helm/repository/repositories.yaml"
	yamlFile, err := ioutil.ReadFile(usr.HomeDir + path)
	if err != nil {
		return err
	}
	var repos repoFile
	if err := yaml.Unmarshal(yamlFile, &repos); err != nil {
		return err
	}

	for _, repo := range repos.Repositories {
		if repo.Name == name {
			klog.Info("Found repo in helm: ", name)
			secretName := ""
			if repo.Password != "" {
				klog.Info("Create secret for repo")
				if err := opts.createRepoSecret(repo.Username, repo.Password, name); err != nil {
					return err
				}
				secretName = name
			}
			if err := opts.createChartRepoResource(repo.URL, secretName); err != nil {
				return err
			}

		}
	}
	return nil

}

// createChartRepo create a new ChartRepo resource
func (opts *ImportOptions) createChartRepoResource(url string, secretName string) error {
	cr := v1alpha1.ChartRepo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChartRepo",
			APIVersion: "app.alauda.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.repoName,
			Namespace: opts.repoNamespace,
		},
		Spec: v1alpha1.ChartRepoSpec{
			URL: url,
		},
		Status: v1alpha1.ChartRepoStatus{
			Phase: "Pending",
		},
	}
	if secretName != "" {
		cr.Spec.Secret = &v1.SecretReference{
			Name: secretName,
		}
	}
	_, err := opts.pctx.CreateChartRepo(&cr)
	return err

}

func (opts *ImportOptions) createRepoSecret(username, password, name string) error {
	cli, err := kubernetes.NewForConfig(opts.pctx.GetRestConfig())
	if err != nil {
		return err
	}
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: opts.repoNamespace,
		},
		Data: make(map[string][]byte),
	}
	//secret.Data["username"] = []byte(base64.StdEncoding.EncodeToString([]byte(username)))
	//secret.Data["password"] = []byte(base64.StdEncoding.EncodeToString([]byte(password)))
	secret.Data["username"] = []byte(username)
	secret.Data["password"] = []byte(password)

	_, err = cli.CoreV1().Secrets(opts.repoNamespace).Create(&secret)
	return err

}

type release struct {
	Name       string  `json:"Name"`
	Revision   float64 `json:"revision"`
	Updated    string  `json:"string"`
	Status     string  `json:"Status"`
	Chart      string  `json:"Chart"`
	AppVersion string  `json:"AppVersion"`
	Namespace  string  `json:"Namespace"`
}

type releases struct {
	Next     string    `1json:"Next"`
	Releases []release `json:"Releases"`
}

func (opts *ImportOptions) getChartVersion(name string) (string, string, error) {
	out, err := exec.Command(opts.helmBinPath, "list", "--output", "json").Output()
	if err != nil {
		return "", "", err
	}

	var rels releases
	if err := json.Unmarshal(out, &rels); err != nil {
		return "", "", err
	}
	// klog.Infof("Get releases list: %+v", rels)

	var chartVersion string

	for _, rel := range rels.Releases {
		if rel.Name == name {
			chartVersion = rel.Chart
		}
	}

	if chartVersion == "" {
		return "", "", errors.New("release not found")
	}

	chart, version := parseVersion(chartVersion)
	klog.Infof("Parsed chart version: %s %s", chart, version)
	return chart, version, nil

}

func parseVersion(chartVersion string) (string, string) {
	result := strings.Split(chartVersion, "-v")
	version := result[len(result)-1]
	l := len(chartVersion) - len(version) - 1
	chart := chartVersion[:l-1]
	version = "v" + version
	return chart, version
}
