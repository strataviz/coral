package image

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Image functions:", func() {
	logger := zap.New(
		zap.Level(zapcore.Level(8) * -1),
	)

	Context("FromReq", func() {
		var (
			mutator *Mutator
			decoder *admission.Decoder
			req     admission.Request
		)

		BeforeEach(func() {
			By("creating a new decoder")
			decoder = admission.NewDecoder(scheme.Scheme)
			Expect(decoder).NotTo(BeNil())

			mutator = NewMutator(logger)

			req = admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind:    "Pod",
						Version: "v1",
						Group:   "",
					},
					OldObject: runtime.RawExtension{
						Object: nil,
					},
				},
			}
		})

		It("should not set policy or selectors if annotation is not present", func() {
			req.Object = runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "v1",
					"kind": "Pod",
					"metadata": {
						"name": "test",
						"namespace": "default"
					},
					"spec": {
						"containers": [{
							"name": "test",
							"image": "docker.io/library/debian:bookworm-slim"
						}]
					}
				}`),
			}

			err := mutator.FromReq(req, decoder)
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator.policy).To(BeFalse())
			Expect(mutator.selectors).To(BeFalse())
		})

		It("should correctly identify annotations to manage pull policies and node selectors", func() {
			req.Object = runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "v1",
					"kind": "Pod",
					"metadata": {
						"name": "test",
						"namespace": "default",
						"annotations": {
							"image.stvz.io/inject": "pull-policy,selectors"
						}
					},
					"spec": {
						"containers": [{
							"name": "test",
							"image": "docker.io/library/debian:bookworm-slim"
						}]
					}
				}`),
			}

			err := mutator.FromReq(req, decoder)
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator.policy).To(BeTrue())
			Expect(mutator.selectors).To(BeTrue())
		})

		It("should set included and excluded pull policies from annotations", func() {
			req.Object = runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "v1",
					"kind": "Pod",
					"metadata": {
						"name": "test",
						"namespace": "default",
						"annotations": {
							"image.stvz.io/inject": "pull-policy,selectors",
							"image.stvz.io/included": "name1",
							"image.stvz.io/excluded": "name2"
						}
					},
					"spec": {
						"containers": [{
							"name": "test",
							"image": "docker.io/library/debian:bookworm-slim"
						}]
					}
				}`),
			}

			decoder := admission.NewDecoder(scheme.Scheme)
			mutator := NewMutator(logger)
			Expect(mutator.FromReq(req, decoder)).To(Succeed())
			Expect(mutator.policy).To(BeTrue())
			Expect(mutator.include).To(ConsistOf("name1"))
			Expect(mutator.exclude).To(ConsistOf("name2"))
		})

		It("should set included and excluded selectors from annotations", func() {
			req.Object = runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "v1",
					"kind": "Pod",
					"metadata": {
						"name": "test",
						"namespace": "default",
						"annotations": {
							"image.stvz.io/inject": "pull-policy,selectors",
							"image.stvz.io/included": "name1",
							"image.stvz.io/excluded": "name2"
						}
					},
					"spec": {
						"containers": [{
							"name": "test",
							"image": "docker.io/library/debian:bookworm-slim"
						}]
					}
				}`),
			}
			err := mutator.FromReq(req, decoder)
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator.include).To(ConsistOf("name1"))
			Expect(mutator.exclude).To(ConsistOf("name2"))
		})
	})

	Context("mutate", func() {
		It("it should return a deployment if the object is a deployment", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(deployment)
			Expect(got).To(BeAssignableToTypeOf(&appsv1.Deployment{}))
		})

		It("it should return a job if the object is a job", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(job)
			Expect(got).To(BeAssignableToTypeOf(&batchv1.Job{}))
		})

		It("it should return a daemonset if the object is a daemonset", func() {
			daemonset := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(daemonset)
			Expect(got).To(BeAssignableToTypeOf(&appsv1.DaemonSet{}))
		})

		It("it should return a pod if the object is a pod", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "docker.io/library/debian:bookworm-slim",
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(pod)
			Expect(got).To(BeAssignableToTypeOf(&corev1.Pod{}))
		})

		It("it should return a replica set if the object is a replica set", func() {
			replicaset := &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(replicaset)
			Expect(got).To(BeAssignableToTypeOf(&appsv1.ReplicaSet{}))
		})

		It("it should return a replication controller if the object is a replication controller", func() {
			replicationcontroller := &corev1.ReplicationController{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(replicationcontroller)
			Expect(got).To(BeAssignableToTypeOf(&corev1.ReplicationController{}))
		})

		It("it should return a stateful set if the object is a stateful set", func() {
			statefulset := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := NewMutator(logger)
			got := m.mutate(statefulset)
			Expect(got).To(BeAssignableToTypeOf(&appsv1.StatefulSet{}))
		})

		It("it should return the original pod if has a ref", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "docker.io/library/debian:bookworm-slim",
						},
					},
				},
			}
			m := Mutator{
				policy:    true,
				selectors: true,
			}
			got := m.mutate(pod)
			Expect(got).To(BeAssignableToTypeOf(&corev1.Pod{}))
			Expect(got.(*corev1.Pod)).To(Equal(pod))
		})

		It("it should return the original replicaset if it has a ref", func() {
			replicaset := &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "docker.io/library/debian:bookworm-slim",
								},
							},
						},
					},
				},
			}
			m := Mutator{
				policy:    true,
				selectors: true,
			}
			got := m.mutate(replicaset)
			Expect(got).To(BeAssignableToTypeOf(&appsv1.ReplicaSet{}))
			Expect(got.(*appsv1.ReplicaSet)).To(Equal(replicaset))
		})
	})

	Context("manage", func() {
		It("it should return the original spec if no selectors or pull policies are set", func() {
			mutator := &Mutator{
				selectors: false,
				policy:    false,
				include:   []string{},
				exclude:   []string{},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image:           "docker.io/library/debian:bookworm-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
					{
						Image:           "docker.io/library/debian:bullseye-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
				NodeSelector: map[string]string{
					"kubernetes.io/arch": "arm64",
				},
			}

			got := mutator.manage(spec)
			Expect(got).To(Equal(spec))
		})

		It("it should return the mutated spec when both pull-policy and selectors are true", func() {
			mutator := &Mutator{
				selectors: true,
				policy:    true,
				include:   []string{"bookworm"},
				exclude:   []string{},
				log:       logger,
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "bookworm",
						Image:           "docker.io/library/debian:bookworm-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
					{
						Name:            "bullseye",
						Image:           "docker.io/library/debian:bullseye-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
				NodeSelector: map[string]string{
					"kubernetes.io/arch": "arm64",
				},
			}

			got := mutator.manage(spec)
			Expect(got.NodeSelector).NotTo(BeNil())
			Expect(got.NodeSelector).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).NotTo(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))
			Expect(got.Containers[0].ImagePullPolicy).To(Equal(corev1.PullNever))
			Expect(got.Containers[1].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		})
	})

	Context("managedSelectors", func() {
		It("it should return selectors on all containers if no include or exclude", func() {
			mutator := &Mutator{
				include: []string{},
				exclude: []string{},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "docker.io/library/debian:bookworm-slim",
					},
				},
			}
			got := mutator.manageSelectors(spec)
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
		})

		It("it should return selectors on included containers, but not on others", func() {
			mutator := &Mutator{
				include: []string{"bookworm"},
				exclude: []string{},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "bookworm",
						Image: "docker.io/library/debian:bookworm-slim",
					},
					{
						Name:  "bullseye",
						Image: "docker.io/library/debian:bullseye-slim",
					},
				},
			}
			got := mutator.manageSelectors(spec)
			Expect(got.NodeSelector).NotTo(BeNil())
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).NotTo(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))
		})

		It("it should return selectors on excluded containers, but not on others", func() {
			mutator := &Mutator{
				include: []string{},
				exclude: []string{"bookworm"},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "bookworm",
						Image: "docker.io/library/debian:bookworm-slim",
					},
					{
						Name:  "bullseye",
						Image: "docker.io/library/debian:bullseye-slim",
					},
				},
			}
			got := mutator.manageSelectors(spec)
			Expect(got.NodeSelector).NotTo(BeNil())
			Expect(got.NodeSelector).NotTo(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))
		})

		It("it should return selectors on included containers if both included and excluded is set", func() {
			mutator := &Mutator{
				include: []string{"bookworm"},
				exclude: []string{"bookworm"},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "bookworm",
						Image: "docker.io/library/debian:bookworm-slim",
					},
					{
						Name:  "bullseye",
						Image: "docker.io/library/debian:bullseye-slim",
					},
				},
			}
			got := mutator.manageSelectors(spec)
			Expect(got.NodeSelector).NotTo(BeNil())
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).NotTo(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))
		})

		It("it should return selectors to merge with existing node selectors", func() {
			mutator := &Mutator{}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "bookworm",
						Image: "docker.io/library/debian:bookworm-slim",
					},
					{
						Name:  "bullseye",
						Image: "docker.io/library/debian:bullseye-slim",
					},
				},
				NodeSelector: map[string]string{
					"kubernetes.io/arch": "arm64",
				},
			}
			got := mutator.manageSelectors(spec)
			Expect(got.NodeSelector).NotTo(BeNil())
			Expect(got.NodeSelector).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))
		})

		It("it should not contain previous selectors when the image has been updated", func() {
			mutator := &Mutator{}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "bullseye",
						Image: "docker.io/library/debian:bullseye-slim",
					},
				},
				NodeSelector: map[string]string{
					"kubernetes.io/arch": "arm64",
				},
			}
			got := mutator.manageSelectors(spec)
			Expect(got.NodeSelector).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
			//Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))

			spec.Containers[0].Image = "docker.io/library/debian:bookworm-slim"
			spec.NodeSelector = map[string]string{
				"kubernetes.io/arch":                             "arm64",
				"image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac": "available",
			}

			got = mutator.manageSelectors(spec)
			Expect(got.NodeSelector).To(HaveKeyWithValue("kubernetes.io/arch", "arm64"))
			Expect(got.NodeSelector).To(HaveKeyWithValue("image.stvz.io/e28d47094db7c64507211886dcba74c9", "available"))
			Expect(got.NodeSelector).NotTo(HaveKeyWithValue("image.stvz.io/52eacfd06bb1d06c9b440400a88c6fac", "available"))
		})
	})

	Context("managePullPolicy", func() {
		It("it should set the pull policy to Never on all containers if no include or exclude", func() {
			mutator := &Mutator{
				include: []string{},
				exclude: []string{},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "bookworm",
						Image:           "docker.io/library/debian:bookworm-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
					{
						Name:            "bullseye",
						Image:           "docker.io/library/debian:bullseye-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
			}
			got := mutator.manageImagePullPolicy(spec)
			Expect(got.Containers).To(HaveLen(2))
			Expect(got.Containers[0].ImagePullPolicy).To(Equal(corev1.PullNever))
			Expect(got.Containers[1].ImagePullPolicy).To(Equal(corev1.PullNever))
		})

		It("it should set the pull policy to Never on included containers, but not on others", func() {
			mutator := &Mutator{
				include: []string{"bookworm"},
				exclude: []string{},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "bookworm",
						Image:           "docker.io/library/debian:bookworm-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
					{
						Name:            "bullseye",
						Image:           "docker.io/library/debian:bullseye-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
			}
			got := mutator.manageImagePullPolicy(spec)
			Expect(got.Containers).To(HaveLen(2))
			Expect(got.Containers[0].ImagePullPolicy).To(Equal(corev1.PullNever))
			Expect(got.Containers[1].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		})

		It("it should set the pull policy to Never on excluded containers, but not on others", func() {
			mutator := &Mutator{
				include: []string{},
				exclude: []string{"bookworm"},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "bookworm",
						Image:           "docker.io/library/debian:bookworm-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
					{
						Name:            "bullseye",
						Image:           "docker.io/library/debian:bullseye-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
			}
			got := mutator.manageImagePullPolicy(spec)
			Expect(got.Containers).To(HaveLen(2))
			Expect(got.Containers[0].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
			Expect(got.Containers[1].ImagePullPolicy).To(Equal(corev1.PullNever))
		})

		It("it should set the pull policy to Never on included containers if both included and excluded is set", func() {
			mutator := &Mutator{
				include: []string{"bookworm"},
				exclude: []string{"bookworm"},
			}

			spec := corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "bookworm",
						Image:           "docker.io/library/debian:bookworm-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
					{
						Name:            "bullseye",
						Image:           "docker.io/library/debian:bullseye-slim",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
			}
			got := mutator.manageImagePullPolicy(spec)
			Expect(got.Containers).To(HaveLen(2))
			Expect(got.Containers[0].ImagePullPolicy).To(Equal(corev1.PullNever))
			Expect(got.Containers[1].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		})
	})
})
