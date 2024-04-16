// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("In the admission webhook", func() {
	When("validateSpec is called", func() {
		Context("and the spec is not valid", func() {
			It("should error if the name field isn't present", func() {
				spec := ImageSpec{
					Repositories: []RepositorySpec{
						{Tags: []string{"tag1"}},
					},
				}
				warnings, err := validateSpec(spec)
				Expect(len(warnings)).To(Equal(0))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name must be specified"))
			})

			It("should error if no tags are provided", func() {
				spec := ImageSpec{
					Repositories: []RepositorySpec{
						{Name: &[]string{"test"}[0]},
					},
				}
				warnings, err := validateSpec(spec)
				Expect(len(warnings)).To(Equal(0))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one tag must be specified"))
			})

			It("should error if there are duplicate tags", func() {
				spec := ImageSpec{
					Repositories: []RepositorySpec{
						{
							Name: &[]string{"test"}[0],
							Tags: []string{"tag1", "tag1", "tag2"},
						},
					},
				}
				warnings, err := validateSpec(spec)
				Expect(len(warnings)).To(Equal(0))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate tags found"))
			})

			It("should warn if the 'latest' tag is provided", func() {
				spec := ImageSpec{
					Repositories: []RepositorySpec{
						{
							Name: &[]string{"test"}[0],
							Tags: []string{"latest"},
						},
					},
				}
				warnings, err := validateSpec(spec)
				Expect(len(warnings)).To(Equal(1))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("and there are no errors with the spec", func() {
			It("should not return an error", func() {
				spec := ImageSpec{
					Repositories: []RepositorySpec{
						{
							Name: &[]string{"test"}[0],
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}
				warnings, err := validateSpec(spec)
				Expect(len(warnings)).To(Equal(0))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
