package watchset

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type ObservedState struct {
	watchSet    *stvziov1.WatchSet
	deployment  *appsv1.Deployment
	token       *corev1.Secret
	observeTime time.Time
}

func NewObservedState() *ObservedState {
	return &ObservedState{
		watchSet:    nil,
		deployment:  nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
}

func (o *StateObserver) observe(ctx context.Context, observed *ObservedState) error {
	var err error
	var observedWatchSet = new(stvziov1.WatchSet)
	err = o.observeWatchSet(o.Request.NamespacedName, observedWatchSet)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return nil
	}
	stvziov1.Defaulted(observedWatchSet)
	observed.watchSet = observedWatchSet

	var observedDeployment = new(appsv1.Deployment)
	err = o.observeDeployment(getDeploymentNamespacedName(observedWatchSet), observedDeployment)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		observed.deployment = nil
	} else {
		observed.deployment = observedDeployment
	}

	var observedToken = new(corev1.Secret)
	err = o.observeSecret(getSecretNamespacedName(observedWatchSet), observedToken)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	}

	observed.token = observedToken

	return nil
}

func (o *StateObserver) observeWatchSet(name client.ObjectKey, builder *stvziov1.WatchSet) error {
	return o.Client.Get(context.Background(), name, builder)
}

func (o *StateObserver) observeDeployment(name client.ObjectKey, deployment *appsv1.Deployment) error {
	return o.Client.Get(context.Background(), name, deployment)
}

func (o *StateObserver) observeSecret(name client.ObjectKey, token *corev1.Secret) error {
	return o.Client.Get(context.Background(), name, token)
}
