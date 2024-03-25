package worker

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	zapcore "go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

// +kubebuilder:docs-gen:collapse=Imports
const (
	timeout  = time.Second * 5
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Context("In a clean environment", func() {
	ctx := context.TODO()
	ns := SetupTest(ctx)
	logger := zap.New(
		zap.Level(zapcore.Level(8) * -1),
	)

	Describe("Worker Image Manager", func() {
		When("Included is called", func() {
			It("should return the correct value when matching the selector", func() {
				By("calling Included")
				imgs := NewImages()
				Expect(imgs.Included(map[string]string{
					"kubernetes.io/arch": "amd64",
				}, []stvziov1.NodeSelector{
					{
						Key:      "kubernetes.io/arch",
						Operator: "in",
						Values:   []string{"amd64"},
					},
				})).To(BeTrue())

				Expect(imgs.Included(map[string]string{
					"kubernetes.io/arch": "arm64",
				}, []stvziov1.NodeSelector{
					{
						Key:      "kubernetes.io/arch",
						Operator: "in",
						Values:   []string{"amd64", "arm64"},
					},
				})).To(BeTrue())

				Expect(imgs.Included(map[string]string{
					"kubernetes.io/arch": "nope",
				}, []stvziov1.NodeSelector{
					{
						Key:      "kubernetes.io/arch",
						Operator: "in",
						Values:   []string{"amd64", "arm64"},
					},
				})).To(BeFalse())
			})
		})

		When("GetImages is called", func() {
			It("should return all images that have no selector", func() {
				ensureImages(ctx, &stvziov1.Image{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "stvz.io/v1",
						Kind:       "Image",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: ns.Name,
					},
					Spec: stvziov1.ImageSpec{
						Images: []stvziov1.ImageSpecImages{
							{
								Name: &[]string{"test"}[0],
								Tags: []string{"tag1", "tag2"},
							},
						},
					},
				})

				By("calling GetImages")
				images := NewImages().WithLogger(logger).WithNamespace(ns.Name)
				err := images.GetImages(ctx, cli, map[string]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(images.List()).To(HaveLen(2))
				Expect(images.List()).To(ConsistOf("test:tag1", "test:tag2"))
			})

			It("should return all images that have a matching selector", func() {
				By("creating Images")
				ensureImages(ctx, &stvziov1.Image{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "stvz.io/v1",
						Kind:       "Image",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amd-images",
						Namespace: ns.Name,
					},
					Spec: stvziov1.ImageSpec{
						Images: []stvziov1.ImageSpecImages{
							{
								Name: &[]string{"test"}[0],
								Tags: []string{"amd1", "amd2"},
							},
						},
						Selector: []stvziov1.NodeSelector{
							{
								Key:      "kubernetes.io/arch",
								Operator: "in",
								Values:   []string{"amd64"},
							},
						},
					},
				}, &stvziov1.Image{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "stvz.io/v1",
						Kind:       "Image",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "arm-images",
						Namespace: ns.Name,
					},
					Spec: stvziov1.ImageSpec{
						Images: []stvziov1.ImageSpecImages{
							{
								Name: &[]string{"test"}[0],
								Tags: []string{"arm1", "arm2"},
							},
						},
						Selector: []stvziov1.NodeSelector{
							{
								Key:      "kubernetes.io/arch",
								Operator: "in",
								Values:   []string{"arm64"},
							},
						},
					},
				})

				By("calling GetImages")
				images := NewImages().WithLogger(logger).WithNamespace(ns.Name)
				err := images.GetImages(ctx, cli, map[string]string{
					"kubernetes.io/arch": "arm64",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(images.List()).To(HaveLen(2))
				Expect(images.List()).To(ConsistOf("test:arm1", "test:arm2"))
			})
		})

		// TODO: Test deduplication
	})
})
