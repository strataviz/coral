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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:docs-gen:collapse=Go imports

func defaultedImage(obj *Image) {}

func defaultedMirror(obj *Mirror) {
	spec := obj.Spec
	if spec.Registry == nil {
		spec.Registry = &RegistrySpec{
			Host:      "localhost",
			Port:      5000,
			TLSVerify: false,
		}
	}
}

// Defaulted sets the resource defaults.
func Defaulted(obj client.Object) {
	switch obj := obj.(type) { //nolint:gocritic
	case *Image:
		defaultedImage(obj)
	case *Mirror:
		defaultedMirror(obj)
	}
}
