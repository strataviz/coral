package mirror

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"stvz.io/hashring"
)

type PodHandler struct {
	Log        logr.Logger
	ServerRing *hashring.Ring
	Namespace  string
	Labels     labels.Selector
}

func (ph *PodHandler) OnAdd(obj interface{}, isInInitialList bool) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	if !ph.owned(pod) {
		return
	}

	ph.ServerRing.Add(pod.Name)
	ph.Log.Info("Added pod to hashring", "pod", pod.Name)
}

func (ph *PodHandler) OnUpdate(oldObj, newObj interface{}) {
	// Updates should not impact us since the name of the pod should never
	// change.  Ignore the updates.
}

func (ph *PodHandler) OnDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	if !ph.owned(pod) {
		return
	}

	ph.ServerRing.Remove(pod.Name)
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
