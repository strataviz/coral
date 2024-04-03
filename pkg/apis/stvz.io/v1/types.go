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
	"k8s.io/apimachinery/pkg/selection"
)

// +kubebuilder:docs-gen:collapse=Go imports

const (
	Finalizer = "stvz.io/finalizer"
)

type NodeSelector struct {
	Key      string             `json:"key"`
	Operator selection.Operator `json:"operator"`
	Values   []string           `json:"values"`
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
	// +nullable
	// Selector defines which nodes the image should be synced to.
	Selector []NodeSelector `json:"selector"`
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	Images []ImageSpecImages `json:"images"`
	// +optional
	// +nullable
	// ImagePullSecrets is a list of secrets to use when pulling the image.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:validation:Required
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=img,singular=images
// +kubebuilder:printcolumn:name="Images",type="integer",JSONPath=".status.totalImages",description="The number of total images managed by the object"
// +kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableImages",description="The number of images that are currently available on the nodes"
// +kubebuilder:printcolumn:name="Pending",type="integer",JSONPath=".status.pendingImages",description="The number of images that are currently pending on the nodes"
// +kubebuilder:printcolumn:name="Deleting",type="integer",JSONPath=".status.deletingImages",description="The number of images that are currently pending deletion on the nodes"
// +kubebuilder:printcolumn:name="Unknown",type="integer",JSONPath=".status.unknownImages",description="The number of images that are in an unknown state on the nodes",priority=1
// +kubebuilder:printcolumn:name="Nodes",type="integer",JSONPath=".status.totalNodes",description="The number of nodes matching the selector (if any)",priority=1
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
	// TotalImages is the number of total images managed by the object.
	TotalImages int `json:"totalImages"`
	// +optional
	// AvailableImages is the number of images that are currently available on the nodes.
	AvailableImages int `json:"availableImages"`
	// +optional
	// PendingImages is the number of images that are currently pending on the nodes.
	PendingImages int `json:"pendingImages"`
	// +optional
	// DeletingImages is the number of images that are currently pending deletion on the nodes.
	DeletingImages int `json:"deletingImages"`
	// +optional
	// UnknownImages is the number of images that are in an unknown state on the nodes.
	UnknownImages int `json:"unknownImages"`
}
