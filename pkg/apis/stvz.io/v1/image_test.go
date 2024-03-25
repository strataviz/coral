package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Image functions:", func() {
	When("GetImages is called", func() {

		Context("and there are no tags", func() {
			It("should return no images", func() {
				image := Image{
					Spec: ImageSpec{
						Images: []ImageSpecImages{
							{
								Name: &[]string{"test"}[0],
								Tags: []string{},
							},
						},
					},
				}
				images := image.GetImages()
				Expect(images).To(HaveLen(0))
			})
		})

		Context("and there is a tag", func() {
			It("should return images", func() {
				image := Image{
					Spec: ImageSpec{
						Images: []ImageSpecImages{
							{
								Name: &[]string{"test"}[0],
								Tags: []string{"tag1"},
							},
						},
					},
				}
				images := image.GetImages()
				Expect(images).To(HaveLen(1))
				Expect(images[0]).To(Equal("test:tag1"))
			})
		})

		Context("and there are multiple tags", func() {
			It("should return images", func() {
				image := Image{
					Spec: ImageSpec{
						Images: []ImageSpecImages{
							{
								Name: &[]string{"test"}[0],
								Tags: []string{"tag1", "tag2", "tag3"},
							},
						},
					},
				}
				images := image.GetImages()
				Expect(images).To(HaveLen(3))
				Expect(images[0]).To(Equal("test:tag1"))
				Expect(images[1]).To(Equal("test:tag2"))
				Expect(images[2]).To(Equal("test:tag3"))
			})
		})
	})
})
