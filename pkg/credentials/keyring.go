package credentials

import (
	"context"
	"sync"
	"sync/atomic"

	corev1 "k8s.io/api/core/v1"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	"k8s.io/kubernetes/pkg/credentialprovider/secrets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretsRef struct {
	Object     *corev1.Secret
	References int
}

func (s *SecretsRef) Add() {
	s.Object = nil
	s.References++
}

type Keyring struct {
	secrets map[client.ObjectKey]*SecretsRef
	ring    credentialprovider.DockerKeyring

	modified atomic.Bool
	client.Client
	sync.Mutex
}

func NewKeyring(c client.Client) *Keyring {
	return &Keyring{
		secrets: make(map[client.ObjectKey]*SecretsRef),
		ring:    credentialprovider.NewDockerKeyring(),
		Client:  c,
	}
}

func (k *Keyring) Lookup(ctx context.Context, name string) ([]*runtimev1.AuthConfig, bool, error) {
	k.Lock()
	defer k.Unlock()

	if !k.modified.Load() {
		auths, found := k.ring.Lookup(name)
		return toRuntimeAuthConfig(auths), found, nil
	}
	// Get our secrets from the cache to see if anything has changed.  I may revisit this
	// as it's pretty heavy for something that is being called quite often.
	secs, err := k.getSecrets(ctx)
	if err != nil {
		return []*runtimev1.AuthConfig{}, false, err
	}

	keyring, err := secrets.MakeDockerKeyring(secs, credentialprovider.NewDockerKeyring())
	if err != nil {
		return []*runtimev1.AuthConfig{}, false, err
	}
	k.ring = keyring
	k.modified.Store(false)

	auths, found := k.ring.Lookup(name)
	return toRuntimeAuthConfig(auths), found, nil
}

func (k *Keyring) Add(sec ...client.ObjectKey) {
	k.Lock()
	defer k.Unlock()

	// For each of the secrets that we are adding, we reset the secret to nil
	// so we can re-fetch it from the API server on demand.  It is expected that
	// both the Add and Remove methods are being called from the informers which
	// will have limited capabilities on what they can do with the objects.  In
	// our case, we needed to be able to look up the secrets that both the Mirror
	// and Image objects reference while having access to request scoped values
	// that are not available to the handlers in the informers.
	for _, s := range sec {
		if _, found := k.secrets[s]; found {
			k.secrets[s].Add()
		} else {
			k.secrets[s] = &SecretsRef{
				Object:     nil,
				References: 1,
			}
		}
	}

	k.modified.Store(true)
}

func (k *Keyring) Remove(sec ...client.ObjectKey) {
	k.Lock()
	defer k.Unlock()

	for _, s := range sec {
		if _, found := k.secrets[s]; found {
			k.secrets[s].References--
			if k.secrets[s].References == 0 {
				delete(k.secrets, s)
			}
		}
	}

	k.modified.Store(true)
}

func (k *Keyring) Has(nn client.ObjectKey) bool {
	k.Lock()
	defer k.Unlock()

	_, found := k.secrets[nn]
	return found
}

func (k *Keyring) getSecrets(ctx context.Context) ([]corev1.Secret, error) {
	secrets := make([]corev1.Secret, 0)
	for key, v := range k.secrets {
		var secret *corev1.Secret
		// Add is called every time the informer notices an update and it will set the secret
		// to nil to reflect that it needs to be re-fetched.  If the secret is not nil, we can
		// reuse the value.
		if v == nil {
			if err := k.Get(ctx, key, secret); err != nil {
				if client.IgnoreNotFound(err) != nil {
					continue
				}
				return nil, err
			}
			k.secrets[key].Object = secret
		} else {
			secret = v.Object
		}

		k.secrets[key].Object = secret
	}
	return secrets, nil
}

func toRuntimeAuthConfig(cfgs []credentialprovider.AuthConfig) []*runtimev1.AuthConfig {
	rt := make([]*runtimev1.AuthConfig, len(cfgs))
	for i, v := range cfgs {
		rt[i] = &runtimev1.AuthConfig{
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
