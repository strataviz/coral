// Copyright 2024 Coral Authors
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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
)

// +kubebuilder:docs-gen:collapse=Go imports

const (
	Finalizer = "image.stvz.io/finalizer"
)

type NodeSelector struct {
	Key      string             `json:"key"`
	Operator selection.Operator `json:"operator"`
	Values   []string           `json:"values"`
}

type ListSelector string

const (
	ListSelectorAll ListSelector = "all"
)

// RepositorySpec is the spec for a Repository resource.
type RepositorySpec struct {
	// +required
	// Name is the repository name that will be used.
	Name *string `json:"name"`
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	// Tags are the repository tags that will be acted on.
	Tags []string `json:"tags"`
	// +optional
	// ListSelection is the type of selection used when syncing a repository.  It
	// is currently unused, however, in the future it will be used by the mirror
	// as a shortcut to sync all tags in a repository.  When 'all' is specified,
	// any tags defined are ignored. Most likely it will be excluded from pulls on
	// the node agent due to potential space constraints.
	ListSelection ListSelector `json:"listSelection"`
}

type Repositories []RepositorySpec

func (r Repositories) NormalizedList() ([]string, error) {
	var list []string
	for _, repo := range r {
		for _, tag := range repo.Tags {
			name, err := NormalizeRepoTag(*repo.Name, tag)
			if err != nil {
				return nil, err
			}
			list = append(list, name)
		}
	}
	return list, nil
}

// ImageSpec is the spec for a Image resource.
type ImageSpec struct {
	// +optional
	// +nullable
	// Selector defines which nodes the image should be synced to.
	Selector []NodeSelector `json:"selector"`
	// +required
	Repositories Repositories `json:"repositories"`
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
// +kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.condition.available",description="The number of images that are currently available on the nodes"
// +kubebuilder:printcolumn:name="Pending",type="integer",JSONPath=".status.condition.pending",description="The number of images that are currently pending on the nodes"
// +kubebuilder:printcolumn:name="Unknown",type="integer",JSONPath=".status.condition.unknown",description="The number of images that are in an unknown state on the nodes",priority=1
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
	ImageStateUnknown   ImageState = "unknown"
)

func (i ImageState) String() string {
	return string(i)
}

type ImageData struct {
	// +required
	// Name is the name of the image in NAME:TAG format.
	Name string `json:"name"`
	// +required
	// Label is the label that is used to track the image on the node.
	Label string `json:"label"`
}

type ImageCondition struct {
	// +required
	// Available is the number of images that are currently available on the nodes.
	Available int `json:"available"`
	// +required
	// Pending is the number of images that are currently pending on the nodes.
	Pending int `json:"pending"`
	// +required
	// Unknown is the number of images that are in an unknown state on the nodes.
	Unknown int `json:"unknown"`
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
	// Condition is the current state of the images on the nodes.
	Condition ImageCondition `json:"condition"`
	// +optional
	// Data is a list of image data that will be used to track the images on the nodes.
	Data []ImageData `json:"data"`
}

type RegistrySpec struct {
	// +required
	// Host is the hostname of the registry.
	Host string `json:"host"`
	// +optional
	// Port is the port of the registry.  It's default is 5000.
	Port int `json:"port"`
	// +optional
	// TLSVerify is a flag to enable or disable tls verification.  Default is true.
	TLSVerify bool `json:"tlsVerify"`
	// Maybe we should add the creds here.
}

func (r *RegistrySpec) URL() string {
	return fmt.Sprintf("docker://%s:%d", r.Host, r.Port)
}

type MirrorSpec struct {
	// +optional
	// Registry is the url to the local registry.  It's default is "localhost:5000".
	Registry *RegistrySpec `json:"registry"`
	// +required
	// Repositories is a list of repositories and associated tags that will be mirrored.
	Repositories Repositories `json:"repositories"`
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
// +kubebuilder:resource:scope=Namespaced,shortName=mi,singular=mirror
// +kubebuilder:printcolumn:name="Images",type="integer",JSONPath=".status.totalImages",description="The number of total images managed by the object"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Mirror is a resource defining images that will be mirrored to the local registry.  Currently
// the mirror requires that all images be defined.  In future iterations we'll support mirroring
// all images from an external registry.
type Mirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MirrorSpec `json:"spec"`
	// +optional
	Status MirrorStatus `json:"status"`
}

type MirrorStatus struct {
	// +optional
	// TotalImages is the number of images that are being mirrored.
	TotalImages int `json:"totalImages"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MirrorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mirror `json:"items"`
}
