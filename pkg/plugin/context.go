package plugin

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type PvcContext struct {
	flags     *genericclioptions.ConfigFlags
	k8scli    *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

func NewPvcContext(streams genericclioptions.IOStreams) *PvcContext {
	return &PvcContext{
		flags: genericclioptions.NewConfigFlags(false),
	}
}

func (p *PvcContext) Complete(namespace string) (err error) {
	p.namespace = namespace

	configLoader := p.flags.ToRawKubeConfigLoader()

	p.config, err = configLoader.ClientConfig()
	if err != nil {
		klog.Errorf("initial rest.Config obj config failed, err: %v", err)
		return err
	}

	p.k8scli, err = kubernetes.NewForConfig(p.config)
	if err != nil {
		klog.Errorf("initial kubernetes.clientset obj k8scli failed, err: %v", err)
	}
	return nil
}

func (p *PvcContext) ListPvcs() (pvcs []corev1.PersistentVolumeClaim, err error) {
	pvcs = make([]corev1.PersistentVolumeClaim, 0)
	if p.k8scli == nil {
		return pvcs, fmt.Errorf("PvcContext.k8scli should not be nil")
	}

	cli := p.k8scli

	pvclist, err := cli.CoreV1().PersistentVolumeClaims(p.namespace).List(metav1.ListOptions{})
	if err != nil {
		return pvcs, fmt.Errorf("list pvcs from kubernetes apiserver failed, err %v", err)
	}

	pvcs = append(pvcs, pvclist.Items...)

	return pvcs, nil
}

func (p *PvcContext) ListPvcsByPod(podname string) (pvcs []corev1.PersistentVolumeClaim, err error) {
	pvcs = make([]corev1.PersistentVolumeClaim, 0)
	if p.k8scli == nil {
		return pvcs, fmt.Errorf("PvcContext.k8scli should not be nil")
	}

	cli := p.k8scli

	pod, err := cli.CoreV1().Pods(p.namespace).Get(podname, metav1.GetOptions{})
	if err != nil {
		return pvcs, fmt.Errorf("get pod [%v/%v] info from kubernetes apiserver failed, err: %v", p.namespace, podname, err)
	}

	volumes := pod.Spec.Volumes
	for _, vol := range volumes {
		if vol.PersistentVolumeClaim != nil {
			pvc, err := cli.CoreV1().PersistentVolumeClaims(p.namespace).Get(vol.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
			if err != nil {
				return pvcs, fmt.Errorf("get pvc [%v/%v] info from kubernetes apiserver failed, err: %v", p.namespace, vol.PersistentVolumeClaim.ClaimName, err)
			}
			pvcs = append(pvcs, *pvc)
		}
	}

	return pvcs, nil
}

func (p *PvcContext) GetNamespace() string {
	return p.namespace
}
