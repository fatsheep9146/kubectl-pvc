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
