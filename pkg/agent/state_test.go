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

package agent

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

var _ = Describe("Agent", func() {
	Describe("State", func() {
		Context("UpdateState", func() {
			It("should update the state when all are available", func() {
				nodeImages := map[string]string{
					"image1": stvziov1.HashedImageLabelKey("image1"),
				}
				managedImages := map[string]string{
					"image1": stvziov1.HashedImageLabelKey("image1"),
				}

				labels := UpdateState(nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue("image1", "available"))
			})

			It("should set a label to pending if the image is not available", func() {
				nodeImages := map[string]string{}
				managedImages := map[string]string{
					"image1": stvziov1.HashedImageLabelKey("image1"),
				}

				labels := UpdateState(nodeImages, managedImages)
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue("image1", "pending"))
			})
		})

		Context("ReplaceImageLabels", func() {
			It("should replace all of the image labels with the new labels", func() {
				nodeLabels := map[string]string{
					stvziov1.HashedImageLabelKey("image1"): "available",
					"kubernetes.io/arch":                   "arm64",
					"kubernetes.io/os":                     "linux",
				}
				state := map[string]string{
					"image1": "available",
					"image2": "pending",
				}

				labels := ReplaceImageLabels(nodeLabels, state)
				Expect(labels).To(HaveLen(4))
				Expect(labels).To(HaveKeyWithValue(stvziov1.HashedImageLabelKey("image1"), "available"))
				Expect(labels).To(HaveKeyWithValue(stvziov1.HashedImageLabelKey("image2"), "pending"))
				Expect(labels).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
				Expect(labels).To(HaveKeyWithValue("kubernetes.io/os", "linux"))
			})

			It("should remove any labels that are not managed", func() {
				nodeLabels := map[string]string{
					stvziov1.HashedImageLabelKey("image1"): "available",
					"kubernetes.io/arch":                   "arm64",
					"kubernetes.io/os":                     "linux",
				}
				state := map[string]string{}

				labels := ReplaceImageLabels(nodeLabels, state)
				Expect(labels).To(HaveLen(2))
				Expect(labels).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
				Expect(labels).To(HaveKeyWithValue("kubernetes.io/os", "linux"))
			})
		})
	})
})
