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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:docs-gen:collapse=Go imports

// BuildQueueSpec is the spec for a BuildQueue resource. The BuildQueue currently
// supports NATs in a single server configuration.  Under the hood, we'll use stateful
// sets to manage pods in the cluster.  This will allow for us to come back through
// and add in clustering support later.  The build queue will be low throughput for the
// most part and should not require a lot of resources.
type BuildQueueSpec struct {
	// +optional
	// +nullable
	// Replicas is the number of cluster nodes to add.  Currently, we only support a
	// single node as we will not have the ability to cleanly update/restart.  In every
	// case this will be overridden to 1.
	Replicas *int32 `json:"replicas"`
	// +optional
	// +nullable
	// Version is the version of the NATs server to run.  This defaults to 2.10.4.
	Version *string `json:"version"`
	// +optional
	// +nullable
	// Resources is the resource limits and requests for the build queue pod(s).
	Resources *corev1.ResourceRequirements `json:"resources"`
	// +optional
	// +nullable
	// Volume is the persistent volume claim for the build queue.  If not defined
	// then the build queue will only use in memory storage which will not persist
	// across restarts.
	Volume *corev1.PersistentVolumeClaim `json:"volume,omitempty"`
}

type BuildQueueStatus struct {
	// +optional
	// ReadyReplicas is the number of ready replicas.
	ReadyReplicas *int `json:"readyReplicas"`
	// +optional
	// Replicas is the total number of replicas.
	Replicas *int `json:"replicas"`
	// +optional
	// UpdatedReplicas is the number of replicas that have been updated.
	UpdatedReplicas *int `json:"updatedReplicas"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=bq,singular=buildqueue
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas",description="The number of ready replicas"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas",description="The number of replicas"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// BuildQueue is the message queue for repo changes and builds
type BuildQueue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BuildQueueSpec `json:"spec"`
	// +optional
	Status BuildQueueStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BuildQueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BuildQueue `json:"items"`
}

// WatchSetSpec is the spec for a WatchSet resource.
type WatchSetSpec struct {
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
	// +optional
	// +nullable
	// Replicas is the number of replicas to run for the watch.
	Replicas *int `json:"replicas"`
}

// WatchStatus is the status for a WatchSet resource.
type WatchSetStatus struct {
	// +optional
	// Enabled indicates whether the watch is polling.
	Enabled *bool `json:"enabled"`
	// +optional
	// ReadyReplicas is the number of ready replicas.
	ReadyReplicas *int `json:"readyReplicas"`
	// +optional
	// Replicas is the total number of replicas.
	Replicas *int `json:"replicas"`
	// +optional
	// UpdatedReplicas is the number of replicas that have been updated.
	UpdatedReplicas *int `json:"updatedReplicas"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=ws,singular=watchset
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// WatchSet is a set of watches associated with the repository.
type WatchSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WatchSetSpec `json:"spec"`
	// +optional
	Status WatchSetStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WatchSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WatchSet `json:"items"`
}

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
	// +optional
	// +nullable
	// BuildQueueRef is the name of the build queue to use for streaming build
	// events between the watchers and builders.  If this is not set, the controller
	// will attempt to discover the build queue.  The first search will be for a
	// build queue with the same name as the builder.  If that is not found, then
	// a search for build queue called default within the same namespace as the
	// builder will be performed.  If that is not found, then the controller will
	// select the first build queue it finds in the builder's namespace.  If no
	// build queue is found then the apply will fail.
	// TODO: set up the webhook to validate that the build queue exists.
	BuildQueueRef *corev1.LocalObjectReference `json:"buildQueueRef"`
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
