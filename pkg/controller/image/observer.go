package image

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type ObservedState struct {
	image       *stvziov1.Image
	nodes       *corev1.NodeList
	observeTime time.Time
}

func NewObservedState() *ObservedState {
	return &ObservedState{
		image:       nil,
		nodes:       nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
}

func (o *StateObserver) observe(ctx context.Context, observed *ObservedState) error {
	var err error
	var observedImage = new(stvziov1.Image)
	err = o.Client.Get(ctx, o.Request.NamespacedName, observedImage)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return nil
	}
	stvziov1.Defaulted(observedImage)
	observed.image = observedImage

	return nil
}
