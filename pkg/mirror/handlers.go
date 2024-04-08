package mirror

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/hashring"
)

type MirrorHandler struct {
	Log         logr.Logger
	MirrorCache *MirrorCache
	AuthCache   *AuthCache
	Client      client.Client
}

func (mh *MirrorHandler) OnAdd(obj interface{}, isInInitialList bool) {
	mh.Log.Info("mirror added", "obj", obj.(client.Object))
	mirror, ok := obj.(*stvziov1.Mirror)
	if !ok {
		mh.Log.Error(nil, "failed to cast object to mirror")
		return
	}
	mh.MirrorCache.Add(mirror)

	secrets := make([]corev1.Secret, len(mirror.Spec.ImagePullSecrets))
	for i, s := range mirror.Spec.ImagePullSecrets {
		nn := types.NamespacedName{
			Name:      s.Name,
			Namespace: mirror.Namespace,
		}
		secret := &corev1.Secret{}
		if err := mh.Client.Get(context.TODO(), nn, secret); err != nil {
			mh.Log.Error(err, "failed to get secret", "secret", nn)
			continue
		}
		secrets[i] = *secret
	}

	mh.AuthCache.Add(secrets...)
}

func (mh *MirrorHandler) OnUpdate(oldObj, newObj interface{}) {
	mh.Log.Info("mirror updated", "obj", newObj.(client.Object))
	mirror, ok := newObj.(*stvziov1.Mirror)
	if !ok {
		mh.Log.Error(nil, "failed to cast object to mirror")
		return
	}

	// TODO: We can use this as a mechanism to trigger GC on the registry by
	// comparing the old and new objects.
	mh.MirrorCache.Add(mirror)

	secrets := make([]corev1.Secret, len(mirror.Spec.ImagePullSecrets))
	for i, s := range mirror.Spec.ImagePullSecrets {
		nn := types.NamespacedName{
			Name:      s.Name,
			Namespace: mirror.Namespace,
		}
		secret := &corev1.Secret{}
		if err := mh.Client.Get(context.TODO(), nn, secret); err != nil {
			mh.Log.Error(err, "failed to get secret", "secret", nn)
			continue
		}
		secrets[i] = *secret
	}

	mh.AuthCache.Add(secrets...)
}

func (mh *MirrorHandler) OnDelete(obj interface{}) {
	mh.Log.Info("mirror deleted", "obj", obj.(client.Object))
	mirror, ok := obj.(*stvziov1.Mirror)
	if !ok {
		mh.Log.Error(nil, "failed to cast object to mirror")
		return
	}
	// TODO: We can use this as a mechanism to trigger GC on the registry
	mh.MirrorCache.Remove(mirror)

	secrets := make([]types.NamespacedName, len(mirror.Spec.ImagePullSecrets))
	for i, s := range mirror.Spec.ImagePullSecrets {
		nn := types.NamespacedName{
			Name:      s.Name,
			Namespace: mirror.Namespace,
		}
		secrets[i] = nn
	}

	mh.AuthCache.Remove(secrets...)
}

type SecretHandler struct {
	Log       logr.Logger
	AuthCache *AuthCache
}

func (sh *SecretHandler) OnAdd(obj interface{}, isInInitialList bool) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		sh.Log.Error(nil, "failed to cast object to secret")
		return
	}

	nn := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}
	if !sh.AuthCache.Has(nn) {
		return
	}

	sh.AuthCache.Add(*secret)
}

func (sh *SecretHandler) OnUpdate(oldObj, newObj interface{}) {
	secret, ok := newObj.(*corev1.Secret)
	if !ok {
		sh.Log.Error(nil, "failed to cast object to secret")
		return
	}

	nn := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}
	if !sh.AuthCache.Has(nn) {
		return
	}

	sh.AuthCache.Add(*secret)
}

func (sh *SecretHandler) OnDelete(obj interface{}) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		sh.Log.Error(nil, "failed to cast object to secret")
		return
	}

	nn := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}
	if !sh.AuthCache.Has(nn) {
		return
	}

	sh.AuthCache.Remove(types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	})
}

type PodHandler struct {
	Log       logr.Logger
	Ring      *hashring.Ring
	Client    client.Client
	Namespace string
	Labels    labels.Selector
}

func (ph *PodHandler) OnAdd(obj interface{}, isInInitialList bool) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		ph.Log.Error(nil, "failed to cast object to pod")
		return
	}

	if !ph.owned(pod) {
		return
	}

	ph.Log.Info("adding pod to the ring", "pod", pod.Name, "namespace", pod.Namespace)
	ph.Ring.Add(pod.Name)
}

func (ph *PodHandler) OnUpdate(oldObj, newObj interface{}) {
	// Updates should not impact us since the name of the pod should never
	// change.  Ignore the updates.
}

func (ph *PodHandler) OnDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		ph.Log.Error(nil, "failed to cast object to pod")
		return
	}

	if !ph.owned(pod) {
		return
	}

	ph.Log.Info("removing pod from the ring", "obj", pod)
	ph.Ring.Remove(pod.Name)
}

func (ph *PodHandler) owned(pod *corev1.Pod) bool {
	// Only care about pods in our namespace
	if pod.Namespace != ph.Namespace {
		return false
	}

	// Only care about pods that match our deployment labels
	selector := ph.Labels
	return selector.Matches(labels.Set(pod.Labels))
}
