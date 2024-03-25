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
	"k8s.io/apimachinery/pkg/selection"
)

// +kubebuilder:docs-gen:collapse=Go imports

type RestartPolicy string

const (
	// RestartPolicyNever indicates that the resources using an updated image should never be restarted.
	RestartPolicyNever RestartPolicy = "Never"
	// RestartPolicyAlways indicates that the resources using an updated image should always be restarted.
	RestartPolicyAlways RestartPolicy = "Always"
	// RestartPolicyAnnotation indicates that the resources using an updated image should be restarted if the annotation is present.
	RestartPolicyAnnotation RestartPolicy = "Annotation"
)

type NodeSelector struct {
	Key      string             `json:"key"`
	Operator selection.Operator `json:"operator"`
	Values   []string           `json:"values"`
}

type ImageSpecRegistry struct {
	// +required
	// URL is the URL of the registry to mirror the image to.  The registry must be accessible from
	// the daemonsets that run on the nodes, and must also support the docker registry API V2.
	URL *string `json:"url"`
}

type ImageSpecImages struct {
	// +required
	// ImageName is the name of the image to mirror.
	Name *string `json:"name"`
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	// Tags are the tags of the image to mirror.
	Tags []string `json:"tags"`
	// +optional
	// pullSecrets is a list of secrets to use when pulling the image.
	// pullSecrets []corev1.LocalObjectReference `json:"pullSecrets"`
}

// ImageSpec is the spec for a Image resource.
type ImageSpec struct {
	// +optional
	// Enabled indicates whether the image synchronization is enabled.  This defaults to true.
	Enabled *bool `json:"enabled"`
	// +optional
	// ManagePullPolicies will adjust all resources that use the image to never pull the image.  This defaults to true.
	ManagePullPolicies *bool `json:"managePullPolicies"`
	// +optional
	// +nullable
	// Selector defines which nodes the image should be synced to.
	Selector []NodeSelector `json:"selector"`
	// +optional
	// PollInterval is the interval to poll the registry for new images.  This defaults to 5 minutes with a 1 minute splay.
	PollInterval *metav1.Duration `json:"pollInterval"`
	// +optional
	// +nullable
	// Registry provides details of an internal registry that will recieve the container images. If
	// an internal registry is set then the images will only be mirrored to the internal registry.
	// if the internal registry is set, the sync workers on the nodes will not pull the images from
	// external registries, but will only pull from the internal registry.
	Registry *ImageSpecRegistry `json:"registry"`
	// +optional
	// RestartPolicy is the policy to use when the image is updated.  I'm not sure that I want this though.  This defaults to Never.
	// RestartPolicy *RestartPolicy `json:"restartPolicy"`
	// +required
	Images []ImageSpecImages `json:"images"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:validation:Required
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=img,singular=images
// +kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".status.totalNodes",description="The number of nodes that should have the image prefetched"
// +kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableNodes",description="The number of nodes that have successfully fetched all tags"
// +kubebuilder:printcolumn:name="Pending",type="integer",JSONPath=".status.pendingNodes",description="The number of nodes that are pending fetchs of 1 or more tags"
// +kubebuilder:printcolumn:name="Deleting",type="integer",JSONPath=".status.deletingNodes",description="The number of nodes where images are waiting to be removed"
// +kubebuilder:printcolumn:name="Unknown",type="integer",JSONPath=".status.unknownNodes",description="The number of nodes where the images are in an unknown state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Image is an external image that will be mirrored to each configured node.
type Image struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ImageSpec `json:"spec"`
	// +optional
	Status ImageStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image `json:"items"`
}

type ImageState string

const (
	ImageStatePending   ImageState = "pending"
	ImageStateAvailable ImageState = "available"
	ImageStateDeleting  ImageState = "deleting"
	ImageStateUnknown   ImageState = "unknown"
)

func (i ImageState) String() string {
	return string(i)
}

func ImageStateFromString(i string) ImageState {
	switch i {
	case "available":
		return ImageStateAvailable
	case "pending":
		return ImageStatePending
	case "deleting":
		return ImageStateDeleting
	default:
		return ImageStateUnknown
	}
}

// WatchStatus is the status for a WatchSet resource.
type ImageStatus struct {
	// +optional
	// TotalNodes is the number of nodes that should have the image prefetched.
	TotalNodes int `json:"totalNodes"`
	// +optional
	// AvailableNodes is the number of nodes that have successfully fetched all images.
	AvailableNodes int `json:"availableNodes"`
	// +optional
	// PendingNodes is the number of nodes that are pending at least one image fetch.
	PendingNodes int `json:"pendingNodes"`
	// +optional
	// DeletingNodes is the number of nodes that are waiting for at least one image to be removed.
	DeletingNodes int `json:"deletingNodes"`
	// +optional
	// UnknownNodes is the number of nodes that are in an unknown state.
	UnknownNodes int `json:"unknownNodes"`
}
