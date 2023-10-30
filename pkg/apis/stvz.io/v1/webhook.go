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

// SetupWebhookWithManager adds webhook for Builder.
func (b *Builder) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(b).
		Complete()
}

func (b *Builder) Default() {
	Defaulted(b)
}

// ValidateCreate implements webhook Validator.
func (b *Builder) ValidateCreate() (admission.Warnings, error) {
	// TODO: check to see if the secret exists, then grab the credentials to make
	// sure that everything is defined correctly.

	// TODO: actually go out to github and see if we have access to the repo and if
	// it exists.  We should be able to do that if we have the credentials.
	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook Validator.
func (b *Builder) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	// TODO: check to see if the secret exists, then grab the credentials to make
	// sure that everything is defined correctly.

	// TODO: actually go out to github and see if we have access to the repo and if
	// it exists.  We should be able to do that if we have the credentials.
	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook Validator.
func (b *Builder) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

var _ webhook.Defaulter = &Builder{}
var _ webhook.Validator = &Builder{}
