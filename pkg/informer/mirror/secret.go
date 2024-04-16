package mirror

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"stvz.io/coral/pkg/credentials"
)

type SecretHandler struct {
	ManagedSecrets map[types.NamespacedName]bool
	Keyring        *credentials.Keyring
}

func (h *SecretHandler) OnAdd(obj interface{}, init bool) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return
	}

	nn := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	if h.Keyring.Has(nn) {
		h.Keyring.Add(nn)
	}
}

func (h *SecretHandler) OnUpdate(oldObj, newObj interface{}) {
	secret, ok := newObj.(*corev1.Secret)
	if !ok {
		return
	}

	nn := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	if h.Keyring.Has(nn) {
		h.Keyring.Add(nn)
	}
}

func (h *SecretHandler) OnDelete(obj interface{}) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return
	}

	nn := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	if h.Keyring.Has(nn) {
		h.Keyring.Remove(nn)
	}
}
