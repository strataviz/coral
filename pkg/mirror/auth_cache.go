package mirror

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	"k8s.io/kubernetes/pkg/credentialprovider/secrets"
)

type AuthCache struct {
	secrets map[types.NamespacedName]bool
	keyring credentialprovider.DockerKeyring
	sync.Mutex
}

func NewAuthCache() *AuthCache {
	return &AuthCache{
		secrets: make(map[types.NamespacedName]bool),
		keyring: credentialprovider.NewDockerKeyring(),
	}
}

func (a *AuthCache) Add(sec ...corev1.Secret) error {
	a.Lock()
	defer a.Unlock()

	for _, s := range sec {
		nn := types.NamespacedName{
			Name:      s.Name,
			Namespace: s.Namespace,
		}
		a.secrets[nn] = true
	}

	keyring, err := secrets.MakeDockerKeyring(sec, a.keyring)
	if err != nil {
		return err
	}

	a.keyring = keyring
	return nil
}

func (a *AuthCache) Remove(sec ...types.NamespacedName) {
	a.Lock()
	defer a.Unlock()

	for _, s := range sec {
		nn := types.NamespacedName{
			Name:      s.Name,
			Namespace: s.Namespace,
		}
		delete(a.secrets, nn)
	}

	// TODO: I don't think there is a way to remove secrets from the
	// keyring without completely rebuilding it.  I don't see this growing
	// too large in practice, so I'm not going to worry about it for now.
	// This exists to give us a place to put the logic in the future.
}

func (a *AuthCache) Has(nn types.NamespacedName) bool {
	a.Lock()
	defer a.Unlock()

	_, found := a.secrets[nn]
	return found
}

func (a *AuthCache) Lookup(name string) []*runtime.AuthConfig {
	a.Lock()
	defer a.Unlock()

	auth, found := a.keyring.Lookup(name)
	if !found {
		return []*runtime.AuthConfig{}
	}

	rt := make([]*runtime.AuthConfig, len(auth))
	for i, v := range auth {
		rt[i] = &runtime.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Auth:          v.Auth,
			ServerAddress: v.ServerAddress,
			IdentityToken: v.IdentityToken,
			RegistryToken: v.RegistryToken,
		}
	}

	return rt
}
