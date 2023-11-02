// Copyright 2023 StrataViz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v1

// +kubebuilder:docs-gen:collapse=Apache License

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:docs-gen:collapse=Go imports

const (
	DefaultWatchMaxAge              time.Duration = time.Hour
	DefaultWatchPollIntervalSeconds int           = 30
	DefaultWatchEnabled             bool          = true
	DefaultWatchDryRun              bool          = false
	DefaultBuilderEnabled           bool          = true

	DefaultBuildQueueResourcesCPU    string = "200m"
	DefaultBuildQueueResourcesMemory string = "512Mi"
)

var (
	DefaultWatchBranches = []string{"main", "master"}
)

// defaultedBuilder defaults a Builder object
func defaultedBuilder(obj *Builder) {
	if obj.Spec.Enabled == nil {
		obj.Spec.Enabled = new(bool)
		*obj.Spec.Enabled = DefaultBuilderEnabled
	}

	if obj.Spec.SecretName == nil {
		obj.Spec.SecretName = new(string)
		*obj.Spec.SecretName = "coral-github-token"
	}

	for _, repo := range obj.Spec.Watches {
		defaultedWatch(&repo)
	}
}

// defaultedWatch defaults a Watch object
func defaultedWatch(obj *Watch) {
	if obj.DryRun == nil {
		obj.DryRun = new(bool)
		*obj.DryRun = DefaultWatchDryRun
	}

	if obj.Enabled == nil {
		obj.Enabled = new(bool)
		*obj.Enabled = DefaultWatchEnabled
	}

	defaultedOn(obj.On)
}

// defaultedOn defaults an On object
func defaultedOn(obj *On) {
	if obj == nil {
		obj = &On{}
	}

	if obj.MaxAge == nil {
		obj.MaxAge = new(metav1.Duration)
		*obj.MaxAge = metav1.Duration{Duration: DefaultWatchMaxAge}
	}

	if obj.PollIntervalSeconds == nil {
		obj.PollIntervalSeconds = new(int)
		*obj.PollIntervalSeconds = DefaultWatchPollIntervalSeconds
	}

	// Default will be to watch all pushes to the main branch.
	if obj.Branches == nil {
		obj.Branches = DefaultWatchBranches
	}

	if obj.Tags == nil {
		obj.Tags = []string{}
	}

	if obj.Releases == nil {
		obj.Releases = []string{}
	}
}

func defaultedBuildQueue(obj *BuildQueue) {
	if obj.Spec.Version == nil {
		obj.Spec.Version = new(string)
		*obj.Spec.Version = "2.10.4-alpine"
	}

	if obj.Spec.Resources == nil {
		obj.Spec.Resources = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(DefaultBuildQueueResourcesCPU),
				corev1.ResourceMemory: resource.MustParse(DefaultBuildQueueResourcesMemory),
			},
		}
	}

	// if obj.Spec.Volume == nil {
	// 	obj.Spec.Volume = &corev1.PersistentVolumeClaim{
	// 		ObjectMeta: metav1.ObjectMeta{
	// 			Name:      fmt.Sprintf("%s-data", obj.Name),
	// 			Namespace: obj.Namespace,
	// 		},
	// 		Spec: corev1.PersistentVolumeClaimSpec{
	// 			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
	// 			Resources: corev1.ResourceRequirements{
	// 				corev1.ResourceList{
	// 					corev1.ResourceStorage: resource.MustParse("1Gi"),
	// 				},
	// 			},
	// 		},
	// 	}
	// }
}

// Defaulted sets the resource defaults.
func Defaulted(obj client.Object) {
	switch obj := obj.(type) {
	case *Builder:
		defaultedBuilder(obj)
	case *BuildQueue:
		defaultedBuildQueue(obj)
	}
}
