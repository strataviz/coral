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
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"stvz.io/coral/pkg/mock"
)

var _ = Describe("Node", func() {
	Context("Get", func() {
		It("should return the wrapped node", func() {
			// TODO: probably pull this out into a larger environment setup
			// It would be pretty nice to even handle namespacing.
			By("mocking a new client")
			file := path.Join(fixtures, "nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "node1", c)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.Name).To(Equal("node1"))
		})
	})

	Context("HasImage", func() {
		It("should return true if the image is available", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "node1", c)
			Expect(err).ToNot(HaveOccurred())

			By("checking for an image")
			Expect(node.HasImage("docker.io/library/debian:bookworm-slim")).To(BeTrue())
			Expect(node.HasImage("docker.io/library/debian:bullseye-slim")).To(BeTrue())
		})

		It("should return false if the image is not available", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "node1", c)
			Expect(err).ToNot(HaveOccurred())

			By("checking for an image")
			Expect(node.HasImage("docker.io/library/notpresent")).To(BeFalse())
		})
	})

	Context("IsReady", func() {
		It("should return true if the node is ready", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "node1", c)
			Expect(err).ToNot(HaveOccurred())

			By("checking if the node is ready")
			Expect(node.IsReady()).To(BeTrue())
		})

		It("should return false if the node is not ready", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "not-ready-nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "notready", c)
			Expect(err).ToNot(HaveOccurred())

			By("checking if the node is ready")
			Expect(node.IsReady()).To(BeFalse())
		})

		It("should return false if there is disk pressure", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "not-ready-nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "diskpressure", c)
			Expect(err).ToNot(HaveOccurred())

			By("checking if the node is ready")
			Expect(node.IsReady()).To(BeFalse())
		})

		It("should return false if there is pid pressure", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "not-ready-nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "pidpressure", c)
			Expect(err).ToNot(HaveOccurred())

			By("checking if the node is ready")
			Expect(node.IsReady()).To(BeFalse())
		})
	})

	Context("Refresh", func() {
		It("should refresh the node", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "nodes.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the node")
			node, err := GetNode(ctx, "node1", c)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.conditionReady).To(BeTrue())

			By("modifying the node")
			conditions := []corev1.NodeCondition{
				{
					// NodeReady is now false instead of true
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   corev1.NodeDiskPressure,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   corev1.NodePIDPressure,
					Status: corev1.ConditionFalse,
				},
			}
			node.Status.Conditions = conditions
			err = node.StatusUpdate(ctx, c)
			Expect(err).ToNot(HaveOccurred())

			By("refreshing the node")
			err = node.Refresh(ctx, c)
			logger.Info("node", "obj", node.Node, "ready", node.conditionReady)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.conditionReady).To(BeFalse())
		})
	})
})
