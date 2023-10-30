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

// PodSpec represents a modified corev1.PodSpec allowing modification of a limited subset of fields.
type PodSpec struct {
	// +optional
	// +nullable
	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
	// +optional
	// +nullable
	// List of containers belonging to the pod.  There must be at least one container in a Pod.
	// This is featured for the ability to add sidecars if needed.  The main process will always
	// be injected into the pod as the first container and will have the name of `nats` (if the
	// nats container does not exist).  Otherwise if a container called `nats` exists, then the
	// container fields that are not defined will be merged in to the user defined container.
	Containers []corev1.Container `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// +optional
	// +nullable
	// List of ephemeral containers run in this pod. Ephemeral containers may be run in an existing
	// pod to perform user-initiated actions such as debugging. This list cannot be specified when
	// creating a pod, and it cannot be modified by updating the pod spec. In order to add an
	// ephemeral container to an existing pod, use the pod's ephemeralcontainers subresource.
	EphemeralContainers []corev1.EphemeralContainer `json:"ephemeralContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,34,rep,name=ephemeralContainers"`
	// +optional
	// +nullable
	// Optional duration in seconds the pod needs to terminate gracefully.
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty" protobuf:"varint,4,opt,name=terminationGracePeriodSeconds"`
	// Optional duration in seconds the pod may be active on the node relative to
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	// +optional
	// +nullable
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"varint,5,opt,name=activeDeadlineSeconds"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// If specified, indicates the pod's priority. Follows the format of the standard PriorityClassName
	// defined in core/v1 PodSpec.
	// https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
	// The priority value.  The higher the value, the higher the priority.
	// +optional
	Priority *int32 `json:"priority,omitempty" protobuf:"bytes,25,opt,name=priority"`
	// PreemptionPolicy is the Policy for preempting pods with lower priority.
	// +optional
	PreemptionPolicy *corev1.PreemptionPolicy `json:"preemptionPolicy,omitempty" protobuf:"bytes,31,opt,name=preemptionPolicy"`
	// ResourceClaims defines which ResourceClaims must be allocated
	// and reserved before the Pod is allowed to start. The resources
	// will be made available to those containers which consume them
	// by name.
	ResourceClaims []corev1.PodResourceClaim `json:"resourceClaims,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,39,rep,name=resourceClaims"`
}

// PodTemplateSpec is the spec for a PodTemplateSpec resource.  We define our own
// spec as a way to remove the confusion between the native PodTemplateSpec and the
// stripped down version that we need for the BuildQueue.
type PodTemplateSpec struct {
	Spec PodSpec `json:"spec"`
}

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
	Replicas *int `json:"replicas"`
	// +optional
	// +nullable
	// Template is the pod template for the build queue.  There will be several fields
	// that we will override regardless of the settings.  I will document those here as
	// we get to them.  It would also be nice to log if the setting is being overriden
	// just as a way to help with debugging, however, I don't really care about that yet.
	// TODO: I may just make my own template spec and then copy the fields that I want
	// to remove any confusion.
	Template PodTemplateSpec `json:"template"`
	// +optional
	// +nullable
	// VolumeClaimTemplates is a list of claims that will be created for the build queue
	// pod.
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
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
