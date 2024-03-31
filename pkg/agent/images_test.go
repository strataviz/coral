package agent

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/credentialprovider"

	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/mock"
)

var _ = Describe("Images", func() {

	Context("Get", func() {
		It("should return the images across all namespaces matching all nodes", func() {
			By("mocking a new client")
			file := path.Join(fixtures, "images.yaml")
			c := mock.NewClient().WithLogger(logger).WithFixtureOrDie(file)

			By("getting the images")
			images, err := ListImages(ctx, c, "", map[string]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(images).To(HaveLen(2))
			Expect(images).To(MatchElements(func(element interface{}) string {
				return element.(Image).Name
			}, IgnoreMissing, Elements{
				"base":      HaveField("ObjectMeta.Name", "base"),
				"strataviz": HaveField("ObjectMeta.Name", "strataviz"),
			}))
		})

		It("should return the images across all namespaces matching node labels", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(path.Join(fixtures, "images_selector.yaml"))

			By("getting the images")
			images, err := ListImages(ctx, c, "", map[string]string{
				"service": "analytics",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(images).To(HaveLen(1))
			Expect(images).To(MatchElements(func(element interface{}) string {
				return element.(Image).Name
			}, IgnoreMissing, Elements{
				"strataviz": HaveField("ObjectMeta.Name", "strataviz"),
			}))
		})

		It("should return the images restricted to a namespace", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(path.Join(fixtures, "images.yaml"))

			By("getting the images")
			images, err := ListImages(ctx, c, "default", map[string]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(images).To(HaveLen(1))
			Expect(images).To(MatchElements(func(element interface{}) string {
				return element.(Image).Name
			}, IgnoreMissing, Elements{
				"base": HaveField("ObjectMeta.Name", "base"),
			}))
		})

		It("should not return the image if the pull secrets are not available", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(path.Join(fixtures, "images_secret_fake.yaml"))

			By("getting the images")
			images, err := ListImages(ctx, c, "", map[string]string{})
			Expect(err).To(HaveOccurred())
			Expect(images).To(BeEmpty())
		})
	})

	Context("getPullSecrets", func() {
		It("should get the pull secrets for an image", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(
					path.Join(fixtures, "images_secret_fake.yaml"),
					path.Join(fixtures, "secrets_fake.yaml"),
				)

			By("getting the image")
			image := stvziov1.Image{}
			err := c.Get(ctx, types.NamespacedName{Name: "strataviz", Namespace: "analytics"}, &image)
			Expect(err).ToNot(HaveOccurred())

			By("getting the secrets")
			secrets, err := getPullSecrets(ctx, c, &image)
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets).To(HaveLen(1))
			Expect(secrets[0].Name).To(Equal("regcred"))
		})

		It("should return an error if the pull secret is outside of the namespace", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(
					path.Join(fixtures, "images_secret_fake_v2.yaml"),
					path.Join(fixtures, "secrets_fake.yaml"),
				)

			By("getting the image")
			image := stvziov1.Image{}
			err := c.Get(ctx, types.NamespacedName{Name: "strataviz", Namespace: "default"}, &image)
			Expect(err).ToNot(HaveOccurred())

			By("getting the secrets")
			secrets, err := getPullSecrets(ctx, c, &image)
			Expect(err).To(MatchError("secrets \"regcred\" not found"))
			Expect(secrets).To(HaveLen(0))
		})

		It("should return an error if the pull secret does not exist", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(
					path.Join(fixtures, "images_secret_fake_v2.yaml"),
				)

			By("getting the image")
			image := stvziov1.Image{}
			err := c.Get(ctx, types.NamespacedName{Name: "strataviz", Namespace: "default"}, &image)
			Expect(err).ToNot(HaveOccurred())

			By("getting the secrets")
			secrets, err := getPullSecrets(ctx, c, &image)
			Expect(err).To(MatchError("secrets \"regcred\" not found"))
			Expect(secrets).To(HaveLen(0))
		})
	})

	Context("matched", func() {
		It("should match the node labels with the selectors", func() {
			By("matching the labels")
			matched, err := matched([]stvziov1.NodeSelector{
				{
					Key:      "service",
					Operator: "in",
					Values:   []string{"analytics"},
				},
			}, map[string]string{
				"service": "analytics",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(matched).To(BeTrue())
		})

		It("should no match the node labels with the selectors", func() {
			By("matching the labels")
			matched, err := matched([]stvziov1.NodeSelector{
				{
					Key:      "service",
					Operator: "in",
					Values:   []string{"analytics"},
				},
			}, map[string]string{
				"service": "monitoring",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(matched).To(BeFalse())
		})
	})

	Context("makeKeyring", func() {
		It("should make a keyring from the secrets", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(
					path.Join(fixtures, "secrets_fake.yaml"),
				)

			By("getting the secrets")
			secrets := corev1.SecretList{}
			err := c.List(ctx, &secrets)
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets.Items).To(HaveLen(1))

			By("making the keyring")
			keyring, err := makeKeyring(secrets.Items)
			Expect(err).ToNot(HaveOccurred())
			Expect(keyring).To(HaveLen(2)) // Basic and caching keyring

			By("looking up an indexed value")
			auth, found := keyring.Lookup("docker.io/strataviz/pyflink:1.17")
			Expect(found).To(BeTrue())
			// Lookup returns both the auth with the credentials and what appears to be an
			// anonymous auth.  Order does not appear to change (anon is last).
			Expect(auth).To(Equal([]credentialprovider.AuthConfig{
				{
					Username:      "testing",
					Password:      "thisisnotmypassword",
					Auth:          "",
					Email:         "testing@example.com",
					ServerAddress: "",
					IdentityToken: "",
					RegistryToken: "",
				},
				{
					Username:      "",
					Password:      "",
					Auth:          "",
					Email:         "",
					ServerAddress: "",
					IdentityToken: "",
					RegistryToken: "",
				},
			}))

			By("looking up a non-indexed value")
			auth, found = keyring.Lookup("gcr.io/spark-operator/spark-operator:v2.2.0")
			Expect(found).To(BeFalse())
			Expect(auth).To(BeEmpty())
		})
	})

	Context("AuthLookup", func() {
		It("should lookup the authentication config for an image", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(
					path.Join(fixtures, "images_secret_fake.yaml"),
					path.Join(fixtures, "secrets_fake.yaml"),
				)

			By("getting the images")
			image, err := ListImages(ctx, c, "", map[string]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(HaveLen(1))

			auth := image[0].AuthLookup("docker.io/strataviz/pyflink:1.17")
			// Lookup returns both the auth with the credentials and what appears to be an
			// anonymous auth.  Order does not appear to change (anon is last).
			Expect(auth).To(Equal([]credentialprovider.AuthConfig{
				{
					Username:      "testing",
					Password:      "thisisnotmypassword",
					Auth:          "",
					Email:         "testing@example.com",
					ServerAddress: "",
					IdentityToken: "",
					RegistryToken: "",
				},
				{
					Username:      "",
					Password:      "",
					Auth:          "",
					Email:         "",
					ServerAddress: "",
					IdentityToken: "",
					RegistryToken: "",
				},
			}))
		})

		It("should return an empty authentication config list if the lookup fails", func() {
			By("mocking a new client")
			c := mock.NewClient().WithLogger(logger).
				WithFixtureOrDie(
					path.Join(fixtures, "images_secret_fake.yaml"),
					path.Join(fixtures, "secrets_fake.yaml"),
				)

			By("getting the images")
			image, err := ListImages(ctx, c, "", map[string]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(HaveLen(1))

			auth := image[0].AuthLookup("gcr.io/spark-operator/spark-operator:v2.2.0")
			Expect(auth).To(BeEmpty())
		})
	})
})
