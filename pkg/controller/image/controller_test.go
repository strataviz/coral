package image

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/mock"
	"stvz.io/coral/pkg/util"
)

var _ = Describe("Controller", func() {
	Context("Reconcile", func() {
		It("should add the finalizer to a new object", func() {
			nn := types.NamespacedName{
				Namespace: "default",
				Name:      "base",
			}

			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(
				path.Join(fixtures, "image_step_1.yaml"),
			)

			image := &stvziov1.Image{}

			// Sanity check to make sure that we are starting with the correct state.
			By("ensuring the initial object does not have the finalizer")
			err := c.Get(ctx, nn, image)
			Expect(err).ToNot(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(image, stvziov1.Finalizer)).To(BeFalse())

			By("creating a new controller")
			controller := &Controller{
				Client: c,
			}

			By("reconciling the object")
			response, err := controller.Reconcile(ctx, reconcile.Request{
				NamespacedName: nn,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Requeue).To(BeFalse())

			By("checking if the object has the finalizer")
			err = c.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "base",
			}, image)
			Expect(err).ToNot(HaveOccurred())
			Expect(image.ObjectMeta.Finalizers).To(ContainElement(stvziov1.Finalizer))
		})

		It("should add a monitor for the new object after it has been updated with the finalizer", func() {
			nn := types.NamespacedName{
				Namespace: "default",
				Name:      "base",
			}

			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(
				path.Join(fixtures, "image_step_2.yaml"),
			)

			image := &stvziov1.Image{}

			// Sanity check to make sure that we are starting with the correct state.
			By("ensuring the initial object does not have the finalizer")
			err := c.Get(ctx, nn, image)
			Expect(err).ToNot(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(image, stvziov1.Finalizer)).To(BeTrue())

			By("creating a new controller")
			controller := &Controller{
				Client: c,
			}

			By("reconciling the object")
			response, err := controller.Reconcile(ctx, reconcile.Request{
				NamespacedName: nn,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Requeue).To(BeFalse())
		})

		It("should wait for the nodes to clean up it's images before removing the finalizer", func() {
			nn := types.NamespacedName{
				Namespace: "default",
				Name:      "base",
			}

			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(
				path.Join(fixtures, "image_step_3.yaml"),
			)

			image := &stvziov1.Image{}

			// Sanity check to make sure that we are starting with the correct state.
			By("ensuring the initial object has the finalizer")
			err := c.Get(ctx, nn, image)
			Expect(err).ToNot(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(image, stvziov1.Finalizer)).To(BeTrue())

			By("adding the image labels to the nodes")
			labels := map[string]string{
				util.HashedImageLabelKey("docker.io/library/debian:bookworm-slim"): "available",
				util.HashedImageLabelKey("docker.io/library/debian:bullseye-slim"): "available",
			}

			var node1 = &corev1.Node{}
			err = c.Get(ctx, types.NamespacedName{Name: "node1"}, node1)
			Expect(err).ToNot(HaveOccurred())

			var node2 = &corev1.Node{}
			err = c.Get(ctx, types.NamespacedName{Name: "node2"}, node2)
			Expect(err).ToNot(HaveOccurred())

			node1.SetLabels(labels)
			err = c.Update(ctx, node1)
			Expect(err).ToNot(HaveOccurred())

			node2.SetLabels(labels)
			err = c.Update(ctx, node2)
			Expect(err).ToNot(HaveOccurred())

			By("creating a new controller")
			controller := &Controller{
				Client: c,
			}

			By("reconciling the object initially")
			response, err := controller.Reconcile(ctx, reconcile.Request{
				NamespacedName: nn,
			})
			Expect(err).ToNot(HaveOccurred())
			// We expect the controller to requeue after 10 seconds from the finalizer add.
			Expect(response.RequeueAfter).ToNot(BeNil())

			By("deleting the image")
			err = c.Delete(ctx, image)
			Expect(err).ToNot(HaveOccurred())

			By("reconciling the object while the nodes still have the image labels")
			response, err = controller.Reconcile(ctx, reconcile.Request{
				NamespacedName: nn,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.RequeueAfter).ToNot(BeNil())

			By("checking if the object still has the finalizer while the nodes still are labeled")
			err = c.Get(ctx, nn, image)
			Expect(err).ToNot(HaveOccurred())
			Expect(image.DeletionTimestamp.IsZero()).To(BeFalse())
			Expect(controllerutil.ContainsFinalizer(image, stvziov1.Finalizer)).To(BeTrue())

			By("removing the label from a single node, leaving the other node available")
			node1.SetLabels(map[string]string{})
			err = c.Update(ctx, node1)
			Expect(err).ToNot(HaveOccurred())

			By("reconciling the object after one node has been cleaned up")
			response, err = controller.Reconcile(ctx, reconcile.Request{
				NamespacedName: nn,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.RequeueAfter).ToNot(BeNil())

			By("checking if the object still has the finalizer while the last node is still labeled")
			err = c.Get(ctx, nn, image)
			Expect(err).ToNot(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(image, stvziov1.Finalizer)).To(BeTrue())

			By("removing the label from the final node")
			node2.SetLabels(map[string]string{})
			err = c.Update(ctx, node2)
			Expect(err).ToNot(HaveOccurred())

			By("reconciling the object after the final node has been cleaned up")
			response, err = controller.Reconcile(ctx, reconcile.Request{
				NamespacedName: nn,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.RequeueAfter).ToNot(BeNil())

			By("checking if the object has been deleted")
			err = c.Get(ctx, nn, image)
			Expect(client.IgnoreNotFound(err)).To(BeNil())
		})
	})
})
