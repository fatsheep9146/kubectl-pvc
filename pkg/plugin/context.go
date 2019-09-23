package plugin

import (
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// CaptainContext holds context for captain command
type CaptainContext struct {
	flags     *genericclioptions.ConfigFlags
	cli       clientset.Interface
	config    *rest.Config
	namespace string
	name string
}

func NewCaptainContext(streams genericclioptions.IOStreams) *CaptainContext {
	return &CaptainContext{
		flags: genericclioptions.NewConfigFlags(true),
	}
}

func (p *CaptainContext) Complete(namespace, name string) (err error) {
	p.namespace = namespace
	p.name = name


	configLoader := p.flags.ToRawKubeConfigLoader()

	p.config, err = configLoader.ClientConfig()
	if err != nil {
		klog.Errorf("initial rest.Config obj config failed, err: %v", err)
		return err
	}

	p.cli, err = clientset.NewForConfig(p.config)
	if err != nil {
		klog.Errorf("initial kubernetes.clientset obj cli failed, err: %v", err)
	}
	return nil
}

func (p *CaptainContext) GetHelmRequest(name string) (*v1alpha1.HelmRequest, error) {
	return p.cli.AppV1alpha1().HelmRequests(p.namespace).Get(name, metav1.GetOptions{})
}

func (p *CaptainContext) UpdateHelmRequest(new *v1alpha1.HelmRequest) (*v1alpha1.HelmRequest, error) {
	return p.cli.AppV1alpha1().HelmRequests(p.namespace).Update(new)
}

func (p *CaptainContext) GetNamespace() string {
	return p.namespace
}
