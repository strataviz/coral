// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		return !strings.HasPrefix(k, stvziov1.LabelPrefix)
	})

	for k, v := range state {
		labels[stvziov1.HashedImageLabelKey(k)] = v
	}

	return labels
}
