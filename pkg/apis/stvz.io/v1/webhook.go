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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Go imports

// +kubebuilder:webhook:verbs=create;update,path=/mutate-stvz-io-v1-builder,mutating=true,failurePolicy=fail,groups=stvz.io,resources=builders,versions=v1,name=mbuilder.stvz.io,admissionReviewVersions=v1,sideEffects=none
// +kubebuilder:webhook:verbs=create;update,path=/validate-stvz-io-v1-builder,mutating=false,failurePolicy=fail,groups=stvz.io,resources=builders,versions=v1,name=vbuilder.stvz.io,admissionReviewVersions=v1,sideEffects=none

// SetupWebhookWithManager adds webhook for BuildSet.
func (b *BuildSet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(b).
		Complete()
}

func (b *BuildSet) Default() {
	Defaulted(b)
}

// ValidateCreate implements webhook Validator.
func (b *BuildSet) ValidateCreate() (admission.Warnings, error) {
	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook Validator.
func (b *BuildSet) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook Validator.
func (b *BuildSet) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

// SetupWebhookWithManager adds webhook for BuildSet.
func (w *WatchSet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(w).
		Complete()
}

func (w *WatchSet) Default() {
	Defaulted(w)
}

// ValidateCreate implements webhook Validator.
func (w *WatchSet) ValidateCreate() (admission.Warnings, error) {
	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook Validator.
func (w *WatchSet) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook Validator.
func (w *WatchSet) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

var _ webhook.Defaulter = &BuildSet{}
var _ webhook.Defaulter = &WatchSet{}
var _ webhook.Validator = &BuildSet{}
var _ webhook.Validator = &WatchSet{}
