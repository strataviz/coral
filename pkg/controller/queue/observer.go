package queue

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
	buildQueue      *stvziov1.BuildQueue
	statefulSet     *appsv1.StatefulSet
	headlessService *corev1.Service
	service         *corev1.Service
	configMap       *corev1.ConfigMap
	observeTime     time.Time
}

func NewObservedState() *ObservedState {
	return &ObservedState{
		buildQueue:      nil,
		statefulSet:     nil,
		headlessService: nil,
		service:         nil,
		configMap:       nil,
		observeTime:     time.Now(),
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
	observed.buildQueue = observedQueue

	var observedStatefulSet = new(appsv1.StatefulSet)
	err = o.observeStatefulSet(getStatefulSetNamespacedName(observedQueue), observedStatefulSet)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		observed.statefulSet = nil
	} else {
		observed.statefulSet = observedStatefulSet
	}

	var observedHeadlessService = new(corev1.Service)
	err = o.observeService(getHeadlessServiceNamespacedName(observedQueue), observedHeadlessService)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		observed.headlessService = nil
	} else {
		observed.headlessService = observedHeadlessService
	}

	var observedService = new(corev1.Service)
	err = o.observeService(getServiceNamespacedName(observedQueue), observedService)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		observed.service = nil
	} else {
		observed.service = observedService
	}

	var observedConfigMap = new(corev1.ConfigMap)
	err = o.observeConfigMap(getConfigMapNamespacedName(observedQueue), observedConfigMap)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		observed.configMap = nil
	} else {
		observed.configMap = observedConfigMap
	}

	return nil
}

func (o *StateObserver) observeQueue(name client.ObjectKey, builder *stvziov1.BuildQueue) error {
	return o.Client.Get(context.Background(), name, builder)
}

func (o *StateObserver) observeStatefulSet(name client.ObjectKey, statefulSet *appsv1.StatefulSet) error {
	return o.Client.Get(context.Background(), name, statefulSet)
}

func (o *StateObserver) observeService(name client.ObjectKey, service *corev1.Service) error {
	return o.Client.Get(context.Background(), name, service)
}

func (o *StateObserver) observeConfigMap(name client.ObjectKey, configMap *corev1.ConfigMap) error {
	return o.Client.Get(context.Background(), name, configMap)
}
