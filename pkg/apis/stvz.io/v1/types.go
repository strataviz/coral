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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:docs-gen:collapse=Go imports

// On is the watch configuration for a git repository.
type On struct {
	// +optional
	// Branches are the branches to watch.
	Branches []string `json:"branches"`
	// +optional
	// MaxAge is the maximum age of a commit to consider.  This defaults to 1h.
	MaxAge *metav1.Duration `json:"maxAge"`
	// +optional
	// PollIntervalSeconds is the number of seconds to wait between polling the
	// repository events for changes.  This defaults to 10 seconds.
	PollIntervalSeconds *int `json:"pollIntervalSeconds"`
	// +optional
	// Tags contain the patterns to match for tags.
	Tags []string `json:"tags"`
	// +optional
	// Releases is the wildcard patterns to match for release events.
	Releases []string `json:"releases"`
}

// Watch is the watch configuration for a git repository.  In the future, we'll
// support other types of repositories, but for now, we're just going to support
// github.
type Watch struct {
	// +optional
	// +nullable
	// DryRun is a flag that indicates whether or not to actually run the build.
	// If set to true, then a change event will be logged, but the build will not
	// be kicked off.
	DryRun *bool `json:"dryRun"`
	// +optional
	// +nullable
	// Enabled indicates the watch should poll.
	Enabled *bool `json:"enabled"`
	// +required
	// Owner is the owner (user or organization) of the repository.
	Owner *string `json:"owner"`
	// +required
	// Repo is the name of the repository.
	Repo *string `json:"repo"`
	// +optional
	// +nullable
	On *On `json:"on"`
}

// BuilderSpec is the spec for a Builder resource.
type BuilderSpec struct {
	// +optional
	// +nullable
	// Enabled globally enables or disables the builder respositories.  This
	// defaults to true.
	Enabled *bool `json:"enabled"`
	// +optional
	// +nullable
	// Secret is the name of the secret resource that contains the credentials
	// for accessing a git repository.  In the future, I'll pull this into vendor
	// specific secrets.
	SecretName *string `json:"secretName"`
	// +required
	// Watches is a list of repositories to watch.
	Watches []Watch `json:"watches"`
}

// BuilderStatus is the status for a Builder resource.
type BuilderStatus struct {
	// +optional
	// Builds is the number of builds that have been kicked off.
	Builds *int64 `json:"builds"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=br,singular=builder
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Builder represents a service that watches for changes to a git repository and
// launches a build when a change is detected.
type Builder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BuilderSpec `json:"spec"`
	// +optional
	Status BuilderStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BuilderList is a list of Builder resources.
type BuilderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Builder `json:"items"`
}
