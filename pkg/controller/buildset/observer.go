package buildset

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type ObservedState struct {
	buildSet    *stvziov1.BuildSet
	deployment  *appsv1.Deployment
	observeTime time.Time
}

func NewObservedState() *ObservedState {
	return &ObservedState{
		buildSet:    nil,
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
	var observedBuildSet = new(stvziov1.BuildSet)
	err = o.observeBuildSet(o.Request.NamespacedName, observedBuildSet)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return nil
	}
	stvziov1.Defaulted(observedBuildSet)
	observed.buildSet = observedBuildSet

	var observedDeployment = new(appsv1.Deployment)
	err = o.observeDeployment(getDeploymentNamespacedName(observedBuildSet), observedDeployment)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		observed.deployment = nil
	} else {
		observed.deployment = observedDeployment
	}

	return nil
}

func (o *StateObserver) observeBuildSet(name client.ObjectKey, builder *stvziov1.BuildSet) error {
	return o.Client.Get(context.Background(), name, builder)
}

func (o *StateObserver) observeDeployment(name client.ObjectKey, deployment *appsv1.Deployment) error {
	return o.Client.Get(context.Background(), name, deployment)
}
