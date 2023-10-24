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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:docs-gen:collapse=Go imports

const (
	DefaultWatchMaxAge              time.Duration = time.Hour
	DefaultWatchPollIntervalSeconds int           = 30
	DefaultRepositoryEnabled        bool          = true
	DefaultRepositoryDryRun         bool          = false
	DefaultBuilderEnabled           bool          = true
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

	for _, repo := range obj.Spec.Repositories {
		defaultedRepository(repo)
	}
}

// defaultedRepository defaults a Repository object
func defaultedRepository(obj *Repository) {
	if obj.DryRun == nil {
		obj.DryRun = new(bool)
		*obj.DryRun = DefaultRepositoryDryRun
	}

	if obj.Enabled == nil {
		obj.Enabled = new(bool)
		*obj.Enabled = DefaultRepositoryEnabled
	}

	defaultedWatch(obj.Watch)
}

// defaultedWatch defaults a Watch object
func defaultedWatch(obj *Watch) {
	if obj == nil {
		obj = &Watch{}
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

// Defaulted sets the resource defaults.
func Defaulted(obj client.Object) {
	switch obj := obj.(type) {
	case *Builder:
		defaultedBuilder(obj)
	}
}
