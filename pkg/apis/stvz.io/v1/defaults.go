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
	DefaultEnabled            bool          = false
	DefaultRegistry           string        = "docker.io"
	DefaultPollInterval       time.Duration = 30 * time.Second
	DefaultManagePullPolicies bool          = true
)

func defaultedImage(obj *Image) {
	if obj.Spec.Enabled == nil {
		obj.Spec.Enabled = new(bool)
		*obj.Spec.Enabled = DefaultEnabled
	}

	if obj.Spec.PollInterval == nil {
		obj.Spec.PollInterval = new(metav1.Duration)
		*obj.Spec.PollInterval = metav1.Duration{Duration: DefaultPollInterval}
	}

	if obj.Spec.Enabled == nil {
		obj.Spec.Enabled = new(bool)
		*obj.Spec.Enabled = true
	}

	if obj.Spec.ManagePullPolicies == nil {
		obj.Spec.ManagePullPolicies = new(bool)
		*obj.Spec.ManagePullPolicies = DefaultManagePullPolicies
	}
}

// Defaulted sets the resource defaults.
func Defaulted(obj client.Object) {
	switch obj := obj.(type) {
	case *Image:
		defaultedImage(obj)
	}
}
