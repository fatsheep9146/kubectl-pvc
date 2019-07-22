package plugin

import (
	"fmt"
	"strings"

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
		flags: genericclioptions.NewConfigFlags(),
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

type PvcPhaseName string

const (
	PvcProvision PvcPhaseName = "Provision"
	PvcBind      PvcPhaseName = "Bind"
	PvcAttach    PvcPhaseName = "Attach"
	PvcMount     PvcPhaseName = "Mount"
)

type PvcPhaseStatus string

const (
	PvcPhaseSuccess    PvcPhaseStatus = "success"
	PvcPhaseFail       PvcPhaseStatus = "fail"
	PvcPhasePartlyFail PvcPhaseStatus = "partly fail"
	PvcPhaseOndoing    PvcPhaseStatus = "ondoing"
)

type PvcPhase struct {
	Name   PvcPhaseName
	Status PvcPhaseStatus
	Detail string
}

type PvcStatus struct {
	Name     string
	PVStatus *PVStatus
	Nodes    []*Node
	Pods     []*Pod
	Phases   map[PvcPhaseName]*PvcPhase
}

type PVStatus struct {
	Name               string
	AttachedVolumeName string
}

type Node struct {
	Name string
}

func NewNode(n *corev1.Node) *Node {
	return &Node{
		Name: n.Name,
	}
}

type Pod struct {
	Name      string
	Volume    string
	Node      string
	PodStatus corev1.PodPhase
}

func NewPod(p *corev1.Pod, vol string) *Pod {
	return &Pod{
		Name:      p.Name,
		Volume:    vol,
		Node:      p.Spec.NodeName,
		PodStatus: p.Status.Phase,
	}
}

func (p *PvcContext) GetPvcDetail(pvcname string) (*PvcStatus, error) {
	pvcStatus := &PvcStatus{
		Name: pvcname,
		Phases: map[PvcPhaseName]*PvcPhase{
			PvcProvision: &PvcPhase{Name: PvcProvision},
			PvcBind:      &PvcPhase{Name: PvcBind},
			PvcAttach:    &PvcPhase{Name: PvcAttach},
			PvcMount:     &PvcPhase{Name: PvcMount},
		},
	}
	if p.k8scli == nil {
		return pvcStatus, fmt.Errorf("PvcContext.k8scli should not be nil")
	}

	cli := p.k8scli

	// check if persisentVolumeClaim's volumeName is set
	// if set, then it means this persistentVolumeClaim is
	pvc, err := cli.CoreV1().PersistentVolumeClaims(p.namespace).Get(pvcname, metav1.GetOptions{})
	if err != nil {
		return pvcStatus, fmt.Errorf("get info about pvc [%s/%s] failed, err: %v", p.namespace, pvcname, err)
	}

	pvname := pvc.Spec.VolumeName
	if pvname == "" {
		return pvcStatus, nil
	}
	pvcStatus.Phases[PvcProvision].Status = PvcPhaseSuccess
	pvcStatus.Phases[PvcBind].Status = PvcPhaseSuccess

	pv, err := cli.CoreV1().PersistentVolumes().Get(pvname, metav1.GetOptions{})
	if err != nil {
		return pvcStatus, fmt.Errorf("get info about pv [%s/%s] failed, err: %v", p.namespace, pvname, err)
	}

	attachedVolumeName, err := getAttachedVolumeName(pv)
	if err != nil {
		return pvcStatus, err
	}

	pvcStatus.PVStatus = &PVStatus{
		Name:               pvname,
		AttachedVolumeName: attachedVolumeName,
	}

	pods := make([]*Pod, 0)
	desiredNodes := make(map[string]struct{})
	podList, err := cli.CoreV1().Pods(p.namespace).List(metav1.ListOptions{})
	for _, pod := range podList.Items {
		if flag, vol := isPvcUsedByPod(pvcname, &pod); flag {
			np := NewPod(&pod, vol)
			desiredNodes[pod.Spec.NodeName] = struct{}{}
			pods = append(pods, np)
		}
	}

	pvcStatus.Pods = pods

	nodeList, err := cli.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return pvcStatus, fmt.Errorf("get info about nodes failed, err: %v", err)
	}

	nodes := make([]*Node, 0)

	for _, node := range nodeList.Items {
		if isPvAttachToNode(pvcStatus.PVStatus.AttachedVolumeName, &node) {
			n := NewNode(&node)
			nodes = append(nodes, n)
		}
	}

	pvcStatus.Nodes = nodes

	attachPhase := deducePhaseAttach(nodes, desiredNodes)
	pvcStatus.Phases[PvcAttach] = attachPhase

	mountPhase := deducePhaseMount(pvcname, pods)
	pvcStatus.Phases[PvcMount] = mountPhase

	return pvcStatus, nil
}

// get the name of this pv which is displayed on the volumesAttached of Node
func getAttachedVolumeName(pv *corev1.PersistentVolume) (string, error) {
	// Todo support none-CSI pv
	if pv.Spec.CSI == nil {
		return "", fmt.Errorf("pv which is not based on csi is now not supported")
	}

	return pv.Spec.CSI.VolumeHandle, nil
}

func isPvcUsedByPod(pvc string, p *corev1.Pod) (bool, string) {
	// klog.Infof("check pvc %s is used by pod %s", pvc, p.Name)
	for _, vol := range p.Spec.Volumes {
		if vol.VolumeSource.PersistentVolumeClaim != nil && vol.VolumeSource.PersistentVolumeClaim.ClaimName == pvc {
			return true, vol.Name
		}
	}
	return false, ""
}

// Todo: more precise way to determine the volume is mounted successfully to Pod
func isPvcMountedToPod(pvc string, pod *Pod) bool {
	if pod.PodStatus == corev1.PodPending {
		return false
	}
	return true
}

// Todo: consider non-CSI persistent volume
func isPvAttachToNode(name string, n *corev1.Node) bool {
	// klog.Infof("check pvc %s is attached by node %s", name, n.Name)
	for _, vol := range n.Status.VolumesAttached {
		if strings.Contains(string(vol.Name), name) {
			return true
		}
	}
	return false
}

func deducePhaseAttach(attachedNodes []*Node, desiredNodes map[string]struct{}) *PvcPhase {
	partly := false
	for _, node := range attachedNodes {
		if _, ok := desiredNodes[node.Name]; ok {
			delete(desiredNodes, node.Name)
			partly = true
		}
		// Todo if volume is not deattached from old node
	}

	p := &PvcPhase{
		Name: PvcAttach,
	}

	if len(desiredNodes) > 0 {
		// it means it has some volume not attached to desired nodes
		if partly {
			p.Status = PvcPhasePartlyFail
		} else {
			p.Status = PvcPhaseFail
		}
		p.Detail = formatUnattachedNodesMsg(desiredNodes)
	} else {
		p.Status = PvcPhaseSuccess
	}

	return p
}

func formatUnattachedNodesMsg(nodes map[string]struct{}) string {
	n := make([]string, 0)
	for node, _ := range nodes {
		n = append(n, node)
	}
	return fmt.Sprintf("nodes: [%s] are still not attached as desired", strings.Join(n, ","))
}

func deducePhaseMount(pvc string, pods []*Pod) *PvcPhase {
	partly := false
	fp := make([]string, 0)
	for _, pod := range pods {
		if !isPvcMountedToPod(pvc, pod) {
			fp = append(fp, pod.Name)
		} else {
			partly = true
		}
	}

	p := &PvcPhase{
		Name: PvcMount,
	}

	if len(fp) > 0 {
		if partly {
			p.Status = PvcPhasePartlyFail
		} else {
			p.Status = PvcPhaseFail
		}
		p.Detail = formatUnmountedPodsMsg(fp)
	} else {
		p.Status = PvcPhaseSuccess
	}
	return p
}

func formatUnmountedPodsMsg(pods []string) string {
	return fmt.Sprintf("pods: [%s] are still not mounted as desired", strings.Join(pods, ","))
}

func (p *PvcContext) GetNamespace() string {
	return p.namespace
}
