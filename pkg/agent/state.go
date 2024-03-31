package agent

import (
	"strings"

	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

func UpdateStateLabels(nodeLabels map[string]string, nodeImages map[string]string, managedImages map[string]string) map[string]string {
	labels := make(map[string]string)

	imageLabels := util.FilterMapFunc(nodeLabels, func(k string, v string) bool {
		return strings.HasPrefix(k, util.LabelPrefix)
	})

	// Default everything to pending
	for k := range managedImages {
		labels[k] = string(stvziov1.ImageStatePending)
	}

	for k := range imageLabels {
		_, managed := managedImages[k]
		_, available := nodeImages[k]
		if managed && available {
			labels[k] = string(stvziov1.ImageStateAvailable)
		} else if managed && !available {
			labels[k] = string(stvziov1.ImageStatePending)
		} else if !managed && available {
			labels[k] = string(stvziov1.ImageStateDeleting)
		}
	}

	return labels
}

func ReplaceImageLabels(nodeLabels map[string]string, imageLabels map[string]string) map[string]string {
	// Copy in the non-image labels
	labels := util.FilterMapFunc(nodeLabels, func(k string, v string) bool {
		return !strings.HasPrefix(k, util.LabelPrefix)
	})

	for k, v := range imageLabels {
		labels[k] = v
	}

	return labels
}
