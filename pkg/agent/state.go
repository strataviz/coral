package agent

import (
	"strings"

	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

func UpdateState(nodeImages map[string]string, managedImages map[string]string) map[string]string {
	state := make(map[string]string)

	for k := range managedImages {
		_, available := nodeImages[k]

		switch {
		case available:
			state[k] = string(stvziov1.ImageStateAvailable)
		default:
			state[k] = string(stvziov1.ImageStatePending)
		}
	}

	return state
}

func ReplaceImageLabels(nodeLabels map[string]string, state map[string]string) map[string]string {
	// Copy in the non-image labels
	labels := util.FilterMapFunc(nodeLabels, func(k string, v string) bool {
		return !strings.HasPrefix(k, util.LabelPrefix)
	})

	for k, v := range state {
		labels[util.HashedImageLabelKey(k)] = v
	}

	return labels
}
