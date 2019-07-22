package plugin

import (
	"fmt"
	"io"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
)

func Format(out io.Writer, pvcs []corev1.PersistentVolumeClaim) {
	w := tabwriter.NewWriter(out, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tVOLUME")
	for _, pvc := range pvcs {
		s := formatPvc(pvc)
		fmt.Fprintln(w, s)
	}
	w.Flush()
}

func formatPvc(pvc corev1.PersistentVolumeClaim) string {
	return fmt.Sprintf("%s\t%s", pvc.Name, pvc.Spec.VolumeName)
}

func FormatPvcDetail(out io.Writer, status *PvcStatus) {
	w := tabwriter.NewWriter(out, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "DESIRED POD\tDESIRED NODE")
	for _, pod := range status.Pods {
		s := fmt.Sprintf("%s\t%s", pod.Name, pod.Node)
		fmt.Fprintln(w, s)
	}
	w.Flush()
	w = tabwriter.NewWriter(out, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "PHASE\tSTATUS\tDETAIL")
	phaseProvision := fmt.Sprintf("%s\t%s\t%s", PvcProvision, string(status.Phases[PvcProvision].Status), status.Phases[PvcProvision].Detail)
	fmt.Fprintln(w, phaseProvision)
	phaseBind := fmt.Sprintf("%s\t%s\t%s", PvcBind, string(status.Phases[PvcBind].Status), status.Phases[PvcBind].Detail)
	fmt.Fprintln(w, phaseBind)
	phaseAttach := fmt.Sprintf("%s\t%s\t%s", PvcAttach, string(status.Phases[PvcAttach].Status), status.Phases[PvcAttach].Detail)
	fmt.Fprintln(w, phaseAttach)
	phaseMount := fmt.Sprintf("%s\t%s\t%s", PvcMount, string(status.Phases[PvcMount].Status), status.Phases[PvcMount].Detail)
	fmt.Fprintln(w, phaseMount)
	w.Flush()
}
