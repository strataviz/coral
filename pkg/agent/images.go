package agent

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	secrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

// TODO: Secrets need to be per image rather than a global list.
type Image struct {
	keyring credentialprovider.DockerKeyring
	stvziov1.Image
}

func ListImages(ctx context.Context, c client.Client, ns string, nodeLabels map[string]string) ([]Image, error) {
	images := []Image{}

	imageList := stvziov1.ImageList{}
	err := c.List(ctx, &imageList, &client.ListOptions{
		Namespace: ns,
	})
	if err != nil {
		return nil, err
	}

	for _, image := range imageList.Items {
		matched, err := matched(image.Spec.Selector, nodeLabels)
		if err != nil {
			return nil, err
		}

		if matched {
			img := image.DeepCopy()
			wrapped := Image{Image: *img}

			secrets, err := getPullSecrets(ctx, c, img)
			if err != nil {
				return []Image{}, err
			}

			keyring, err := makeKeyring(secrets)
			if err != nil {
				return []Image{}, err
			}
			wrapped.keyring = keyring

			images = append(images, wrapped)
		}
	}

	return images, nil

}

func (i *Image) AuthLookup(name string) []credentialprovider.AuthConfig {
	auth, found := i.keyring.Lookup(name)
	if !found {
		return []credentialprovider.AuthConfig{}
	}

	return auth
}

func (i *Image) RuntimeAuthLookup(name string) []*runtime.AuthConfig {
	// TODO: should probably cache this, but for now, it's not super expensive.
	auth := i.AuthLookup(name)
	runtimeAuth := make([]*runtime.AuthConfig, len(auth))
	for i, v := range auth {
		runtimeAuth[i] = &runtime.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Auth:          v.Auth,
			ServerAddress: v.ServerAddress,
			IdentityToken: v.IdentityToken,
			RegistryToken: v.RegistryToken,
		}
	}

	return runtimeAuth
}

func matched(selectors []stvziov1.NodeSelector, nodeLabels map[string]string) (bool, error) {
	s := labels.NewSelector()
	for _, selector := range selectors {
		req, err := labels.NewRequirement(selector.Key, selector.Operator, selector.Values)
		if err != nil {
			return false, err
		}
		s = s.Add(*req)
	}

	return s.Matches(labels.Set(nodeLabels)), nil
}

func getPullSecrets(ctx context.Context, c client.Client, img *stvziov1.Image) ([]corev1.Secret, error) {
	secrets := []corev1.Secret{}

	if img.Spec.ImagePullSecrets == nil {
		return []corev1.Secret{}, nil
	}

	for _, s := range img.Spec.ImagePullSecrets {
		secret := &corev1.Secret{}
		err := c.Get(ctx, client.ObjectKey{Name: s.Name, Namespace: img.Namespace}, secret)
		if err != nil {
			return []corev1.Secret{}, err
		}

		secrets = append(secrets, *secret.DeepCopy())
	}

	return secrets, nil
}

func makeKeyring(pullSecrets []corev1.Secret) (credentialprovider.DockerKeyring, error) {
	defaultKeyring := credentialprovider.NewDockerKeyring()

	keyring, err := secrets.MakeDockerKeyring(pullSecrets, defaultKeyring)
	if err != nil {
		return nil, err
	}

	return keyring, nil
}
