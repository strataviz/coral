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
					Images: []ImageSpecImages{
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
					Images: []ImageSpecImages{
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
					Images: []ImageSpecImages{
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
					Images: []ImageSpecImages{
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
					Images: []ImageSpecImages{
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
