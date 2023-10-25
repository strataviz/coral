package builder

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type ObservedState struct {
	builder     *stvziov1.Builder
	token       *corev1.Secret
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

	var observedToken = new(corev1.Secret)
	nn := types.NamespacedName{
		// Consider using a ref for the builders secret instead of a string
		// to allow for cross namespace references.  I don't mind forcing
		// the secret to live in the same namespace as the builder for now.
		Namespace: o.Request.Namespace,
		Name:      *observed.builder.Spec.SecretName,
	}
	err = o.observeToken(nn, observedToken)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	}

	observed.token = observedToken
	return nil
}

func (o *StateObserver) observeBuilder(name client.ObjectKey, builder *stvziov1.Builder) error {
	return o.Client.Get(context.Background(), name, builder)
}

func (o *StateObserver) observeToken(name client.ObjectKey, token *corev1.Secret) error {
	return o.Client.Get(context.Background(), name, token)
}
