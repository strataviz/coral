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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Go imports

// +kubebuilder:webhook:verbs=create;update,path=/mutate-stvz-io-v1-image,mutating=true,failurePolicy=fail,groups=stvz.io,resources=images,versions=v1,name=mimage.stvz.io,admissionReviewVersions=v1,sideEffects=none
// +kubebuilder:webhook:verbs=create;update,path=/validate-stvz-io-v1-image,mutating=false,failurePolicy=fail,groups=stvz.io,resources=images,versions=v1,name=vimage.stvz.io,admissionReviewVersions=v1,sideEffects=none

// SetupWebhookWithManager adds webhook for BuildSet.
func (i *Image) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(i).
		Complete()
}

func (i *Image) Default() {
	Defaulted(i)
}

func validateSpecImages(images []ImageSpecImages) (admission.Warnings, error) {
	warnings := make(admission.Warnings, 0)

	for _, image := range images {
		if image.Name == nil {
			return warnings, fmt.Errorf("name must be specified")
		}

		if len(image.Tags) < 1 {
			return warnings, fmt.Errorf("at least one tag must be specified")
		}

		tags := make(map[string]bool)
		duplicates := make([]string, 0)
		for _, tag := range image.Tags {
			if tag == "latest" {
				warnings = append(warnings, "tag 'latest' is not supported and will be ignored")
				continue
			}

			if _, ok := tags[tag]; ok {
				duplicates = append(duplicates, tag)
			} else {
				tags[tag] = true
			}
		}

		if len(duplicates) > 0 {
			return warnings, fmt.Errorf("duplicate tags found: %v", duplicates)
		}
	}

	return warnings, nil
}

func validateSpec(spec ImageSpec) (admission.Warnings, error) {
	return validateSpecImages(spec.Images)
}

// ValidateCreate implements webhook Validator.
func (i *Image) ValidateCreate() (admission.Warnings, error) {
	warnings := make(admission.Warnings, 0)

	specWarnings, err := validateSpec(i.Spec)
	if err != nil {
		return warnings, err
	}

	warnings = append(warnings, specWarnings...)
	return warnings, nil
}

// ValidateUpdate implements webhook Validator.
func (i *Image) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	warnings := make(admission.Warnings, 0)

	specWarnings, err := validateSpec(i.Spec)
	if err != nil {
		return warnings, err
	}

	warnings = append(warnings, specWarnings...)
	return warnings, nil
}

// ValidateDelete implements webhook Validator.
func (i *Image) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

var _ webhook.Defaulter = &Image{}
var _ webhook.Validator = &Image{}
