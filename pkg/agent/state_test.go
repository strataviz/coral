package agent

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"stvz.io/coral/pkg/util"
)

var _ = Describe("Agent", func() {
	Describe("State", func() {
		Context("UpdateStateLabels", func() {
			It("should update the state labels when all are available", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "available",
				}
				nodeImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}
				managedImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image1"), "available"))
			})

			It("should set a label to pending if the image is managed and available, but the label does not exist", func() {
				nodeLabels := map[string]string{}
				nodeImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}
				managedImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image1"), "pending"))
			})

			It("should set a label to pending if the image is managed, but is not available", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "available",
				}
				nodeImages := map[string]string{}
				managedImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image1"), "pending"))
			})

			It("should set a label to available if the image is managed and available", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "pending",
				}
				nodeImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}
				managedImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image1"), "available"))
			})

			It("should set an image to delete if available and not managed", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "available",
				}
				nodeImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}
				managedImages := map[string]string{}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image1"), "deleting"))
			})

			It("should set an image to delete if pending and not managed", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "pending",
				}
				nodeImages := map[string]string{
					util.HashedImageLabelKey("image1"): "image1",
				}
				managedImages := map[string]string{}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image1"), "deleting"))
			})

			It("should remove a label if the image is not managed and not available", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "anything",
				}
				nodeImages := map[string]string{}
				managedImages := map[string]string{}

				labels := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
				Expect(labels).To(BeEmpty())
			})
		})

		Context("ReplaceImageLabels", func() {
			It("should replace all of the image labels with the new labels", func() {
				nodeLabels := map[string]string{
					util.HashedImageLabelKey("image1"): "available",
					"kubernetes.io/arch":               "arm64",
					"kubernetes.io/os":                 "linux",
				}
				imageLabels := map[string]string{
					util.HashedImageLabelKey("image2"): "pending",
				}

				labels := ReplaceImageLabels(nodeLabels, imageLabels)
				Expect(labels).To(HaveLen(3))
				Expect(labels).To(HaveKeyWithValue(util.HashedImageLabelKey("image2"), "pending"))
				Expect(labels).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
				Expect(labels).To(HaveKeyWithValue("kubernetes.io/os", "linux"))
			})
		})
	})
})
