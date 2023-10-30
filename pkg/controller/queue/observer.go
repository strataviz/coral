package queue

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type ObservedState struct {
	queue       *stvziov1.BuildQueue
	statefulSet []appsv1.StatefulSet
	observeTime time.Time
}

func NewObservedState() ObservedState {
	return ObservedState{
		queue:       nil,
		statefulSet: nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
}

func (o *StateObserver) observe(ctx context.Context, observed *ObservedState) error {
	var err error
	var observedQueue = new(stvziov1.BuildQueue)
	err = o.observeQueue(o.Request.NamespacedName, observedQueue)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return nil
	}

	stvziov1.Defaulted(observedQueue)
	observed.queue = observedQueue

	return nil
}

func (o *StateObserver) observeQueue(name client.ObjectKey, builder *stvziov1.BuildQueue) error {
	return o.Client.Get(context.Background(), name, builder)
}
