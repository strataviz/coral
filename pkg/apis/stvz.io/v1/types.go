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

// RepositoryOn is the watch configuration for a git repository.  There's quite
// a few combinations of things that can be watched, so I'm going to start with
// the simplest which will be pushes to a branch, the creation of releases, and
// the creation of tags matching a specific pattern.
// TODO: I'll need additional fields for manipulating the release names for the
// builds.  By default we'll use the tag, branch, or release name as the version
// for the build.  I think we can add a regex to extract version info from the
// name or the actually source.
type Watch struct {
	// +optional
	// DryRun is a flag that indicates whether or not to actually run the build.
	// If set to true, then a change event will be logged, but the build will not
	// be kicked off.
	DryRun *bool `json:"dryRun"`
	// +optional
	// Branches are the branches to watch.
	Branches []string `json:"branches"`
	// +optional
	// Tags contain the patterns to match for tags.
	Tags []string `json:"tags"`
	// +optional
	// Releases is the wildcard patterns to match for release events.
	Releases []string `json:"releases"`
}

// Repository is the configuration for a git repository.
type Repository struct {
	// +optional
	Watch *Watch `json:"watch"`
}

// BuilderSpec is the spec for a Builder resource.
type BuilderSpec struct {
	// +optional
	// Secret is the name of the secret resource that contains the credentials
	// for accessing a git repository.  In the future, I'll pull this into vendor
	// specific secrets.
	SecretName string `json:"secretName"`
	// +optional
	// API is the base URL for the git (in this case github) API.
	URL string `json:"url"`
	// +optional
	// Type is the type of git repository.  Currently only github is supported.
	Type string `json:"type"`
	// +optional
	// Watch is a list of repositories to watch.
	Repository *Repository `json:"repository"`
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
