package builder

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type ObservedState struct {
	builder     *stvziov1.Builder
	observeTime time.Time
}

func NewObservedState() ObservedState {
	return ObservedState{
		builder:     nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
}

func (o *StateObserver) observe(ctx context.Context, observed *ObservedState) error {
	var err error
	var observedBuilder = new(stvziov1.Builder)
	err = o.observeBuilder(o.Request.NamespacedName, observedBuilder)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return nil
	}

	stvziov1.Defaulted(observedBuilder)
	observed.builder = observedBuilder

	return nil
}

func (o *StateObserver) observeBuilder(name client.ObjectKey, builder *stvziov1.Builder) error {
	return o.Client.Get(context.Background(), name, builder)
}
