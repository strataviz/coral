package worker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	zapcore "go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"stvz.io/coral/pkg/util"
)

var _ = Describe("NodeState", func() {
	logger := zap.New(
		zap.Level(zapcore.Level(8) * -1),
	)

	Context("When a new NodeState is created", func() {
		It("should have the condition values", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
						{
							Type:   corev1.NodeDiskPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodePIDPressure,
							Status: corev1.ConditionFalse,
						},
					},
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
						{
							Names: []string{"image2"},
						},
					},
				},
			})
			Expect(state).NotTo(BeNil())
			Expect(state.Ready).To(BeTrue())
			Expect(state.DiskPressure).To(BeFalse())
			Expect(state.PidPressure).To(BeFalse())
		})
	})
	Context("When UpdateLabels is called", func() {
		It("should correctly add pending images not found in the node labels", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kubernetes.io/arch":     "arm64",
						"kubernetes.io/hostname": "coral-worker",
						"kubernetes.io/os":       "linux",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			}).WithLogger(logger)

			labels := state.GetUpdatedLabels(map[string]string{
				util.ImageHasher("image1"): "image1",
			})
			Expect(labels).To(Equal(map[string]string{
				"kubernetes.io/arch":                     "arm64",
				"kubernetes.io/hostname":                 "coral-worker",
				"kubernetes.io/os":                       "linux",
				LabelPrefix + util.ImageHasher("image1"): "pending",
			}))
		})

		It("should correctly update the state of images found in the node labels", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kubernetes.io/arch":                     "arm64",
						"kubernetes.io/hostname":                 "coral-worker",
						"kubernetes.io/os":                       "linux",
						LabelPrefix + util.ImageHasher("image1"): "pending",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			}).WithLogger(logger)

			labels := state.GetUpdatedLabels(map[string]string{
				util.ImageHasher("image1"): "image1",
			})

			Expect(labels).To(Equal(map[string]string{
				"kubernetes.io/arch":                     "arm64",
				"kubernetes.io/hostname":                 "coral-worker",
				"kubernetes.io/os":                       "linux",
				LabelPrefix + util.ImageHasher("image1"): "pending",
			}))
		})

		It("should correctly identify available images not found in the labels", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kubernetes.io/arch":     "arm64",
						"kubernetes.io/hostname": "coral-worker",
						"kubernetes.io/os":       "linux",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			}).WithLogger(logger)

			labels := state.GetUpdatedLabels(map[string]string{
				util.ImageHasher("image1"): "image1",
			})

			Expect(labels).To(Equal(map[string]string{
				"kubernetes.io/arch":                     "arm64",
				"kubernetes.io/hostname":                 "coral-worker",
				"kubernetes.io/os":                       "linux",
				LabelPrefix + util.ImageHasher("image1"): "available",
			}))
		})

		It("should correctly update available images found in the labels", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kubernetes.io/arch":                     "arm64",
						"kubernetes.io/hostname":                 "coral-worker",
						"kubernetes.io/os":                       "linux",
						LabelPrefix + util.ImageHasher("image1"): "available",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			}).WithLogger(logger)

			labels := state.GetUpdatedLabels(map[string]string{
				util.ImageHasher("image1"): "image1",
			})

			Expect(labels).To(Equal(map[string]string{
				"kubernetes.io/arch":                     "arm64",
				"kubernetes.io/hostname":                 "coral-worker",
				"kubernetes.io/os":                       "linux",
				LabelPrefix + util.ImageHasher("image1"): "available",
			}))
		})

		It("should correctly identify images that need to be deleted", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kubernetes.io/arch":                     "arm64",
						"kubernetes.io/hostname":                 "coral-worker",
						"kubernetes.io/os":                       "linux",
						LabelPrefix + util.ImageHasher("image1"): "available",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			}).WithLogger(logger)

			labels := state.GetUpdatedLabels(map[string]string{})

			Expect(labels).To(Equal(map[string]string{
				"kubernetes.io/arch":                     "arm64",
				"kubernetes.io/hostname":                 "coral-worker",
				"kubernetes.io/os":                       "linux",
				LabelPrefix + util.ImageHasher("image1"): "deleting",
			}))
		})

		It("should correctly remove labels for deleted images", func() {
			state := NewNodeState(&corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kubernetes.io/arch":                     "arm64",
						"kubernetes.io/hostname":                 "coral-worker",
						"kubernetes.io/os":                       "linux",
						LabelPrefix + util.ImageHasher("image1"): "deleting",
						// Though the image status of available would most likely not be
						// present in the labels at the time of deletion, this is a state
						// that could be reached if the image was deleted externally.
						LabelPrefix + util.ImageHasher("image2"): "available",
						// Pending images have not been fetched yet and if the image resource
						// has been removed, we'd also want to remove the label.
						LabelPrefix + util.ImageHasher("image3"): "pending",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			}).WithLogger(logger)

			labels := state.GetUpdatedLabels(map[string]string{})

			Expect(labels).To(Equal(map[string]string{
				"kubernetes.io/arch":     "arm64",
				"kubernetes.io/hostname": "coral-worker",
				"kubernetes.io/os":       "linux",
			}))
		})
	})

	Context("When NeedsImage is called", func() {
		It("should return true if the image is not found in the node images", func() {
			state := NewNodeState(&corev1.Node{
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			})

			Expect(state.NeedsImage("image2")).To(BeTrue())
		})

		It("should return false if the image is found in the node images", func() {
			state := NewNodeState(&corev1.Node{
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		It("should return true if the image label is pending", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "pending",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeTrue())
		})

		It("should return false if the image label is pending but the image is available on the node", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "pending",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		It("should return false if the image label is available", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "available",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		It("should return true if the image label is marked available but not available on the node", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "available",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeTrue())
		})

		It("should return false if the image is marked as deleting and is available on the node", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "deleting",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		It("should return false if the image is marked as deleting and is not available on the node", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "deleting",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		It("should return false if the image is marked as unknown and is not available on the node", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "unknown",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		It("should return false if the image is marked as unknown and is available on the node", func() {
			state := NewNodeState(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelPrefix + util.ImageHasher("image1"): "unknown",
					},
				},
				Status: corev1.NodeStatus{
					Images: []corev1.ContainerImage{
						{
							Names: []string{"image1"},
						},
					},
				},
			})

			Expect(state.NeedsImage("image1")).To(BeFalse())
		})

		// docker library images are not hashed correctly because the names are not 1-1 with the shortened
		// image name in the resource.  To solve this we need to make sure that if the image doesn't contain
		// the group/org in the name that we add 'library' as the org?
		//
		// Initially do we require a group/org in the image name?  I think it's ok to assume or give an option
		// to disable.
		// docker.io uses library
		// quay.io uses quay
		// marketplace.gcr.io uses google (note the marketplace subdomain) - you can assume that gcr.io would be bare.
		// gcr.io would never be modified as I'd expect groups to always be used.
		// public.ecr.aws aparently requires specific groups 'ubuntu/ubuntu:<tag>'
	})
})
