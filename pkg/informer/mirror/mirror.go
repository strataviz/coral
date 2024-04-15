package mirror

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/credentials"
)

type MirrorHandler struct {
	Keyring *credentials.Keyring
	Mirrors map[client.ObjectKey]*stvziov1.Mirror
}

func (h *MirrorHandler) OnAdd(obj interface{}, init bool) {
	mirror, ok := obj.(*stvziov1.Mirror)
	if !ok {
		return
	}

	h.add(mirror)
}

func (h *MirrorHandler) OnUpdate(oldObj, newObj interface{}) {
	mirror, ok := newObj.(*stvziov1.Mirror)
	if !ok {
		return
	}

	h.add(mirror)
}

func (h *MirrorHandler) OnDelete(obj interface{}) {
	mirror, ok := obj.(*stvziov1.Mirror)
	if !ok {
		return
	}
	h.remove(mirror)
}

func (h *MirrorHandler) add(obj *stvziov1.Mirror) {
	h.Mirrors[client.ObjectKeyFromObject(obj)] = obj

	secrets := make([]client.ObjectKey, len(obj.Spec.ImagePullSecrets))
	for i, s := range obj.Spec.ImagePullSecrets {
		secrets[i] = types.NamespacedName{
			Name:      s.Name,
			Namespace: obj.Namespace,
		}
	}
	h.Keyring.Add(secrets...)
}

func (h *MirrorHandler) remove(obj *stvziov1.Mirror) {
	delete(h.Mirrors, client.ObjectKeyFromObject(obj))

	// Oh... how do I know if these are not in use by other mirrors?  I think
	secrets := make([]client.ObjectKey, len(obj.Spec.ImagePullSecrets))
	for i, s := range obj.Spec.ImagePullSecrets {
		secrets[i] = types.NamespacedName{
			Name:      s.Name,
			Namespace: obj.Namespace,
		}
	}
	h.Keyring.Remove(secrets...)
}
