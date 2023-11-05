package buildqueue

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type Controller struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func SetupWithManager(mgr ctrl.Manager) error {
	c := &Controller{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("buildqueue-controller"),
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&stvziov1.BuildQueue{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Complete(c)
}

// +kubebuilder:rbac:groups=stvz.io,resources=buildqueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stvz.io,resources=buildqueues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stvz.io,resources=buildqueues/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get

// Reconcile is the main controller loop for the queue controller.  Though it's more
// flexible to have a seperate controller for the queue, this does raise the issue of
// potential issues if the queue has been deleted but the builder is still running.
// For now, I'll just make the assumption that once we validate that the queue exists
// when the builder starts, that we won't have to check again and if the queue is
// deleted then we'll let the builder continue to run and fail once it tries to
// produce or consume from the queue.  We'll be able to capture this state in metrics
// and alert on it.
func (c Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	observed := NewObservedState()
	observer := StateObserver{
		Client:  c.Client,
		Request: req,
	}

	err := observer.observe(ctx, observed)
	if err != nil {
		logger.Error(err, "unable to observe state", "request", req)
		return ctrl.Result{}, err
	}

	logger.Info("observed",
		"statefulSet", observed.statefulSet,
		"service", observed.headlessService,
		"buildQueue", observed.buildQueue,
	)

	if observed.buildQueue == nil {
		logger.Info("queue has been deleted, cleaning up", "request", req)
		return ctrl.Result{}, nil
	}

	desired, err := GetDesiredState(observed)
	if err != nil {
		logger.Error(err, "unable to get desired state", "request", req)
		return ctrl.Result{}, err
	}

	logger.Info("desired", "statefulset", desired.StatefulSet, "service", desired.HeadlessService)

	err = c.reconcileConfigMap(ctx, observed.configMap, desired.ConfigMap)
	if err != nil {
		logger.Error(err, "unable to reconcile configmap", "request", req)
		return ctrl.Result{}, err
	}

	err = c.reconcileService(ctx, observed.headlessService, desired.HeadlessService)
	if err != nil {
		logger.Error(err, "unable to reconcile headless service", "request", req)
		return ctrl.Result{}, err
	}

	err = c.reconcileService(ctx, observed.service, desired.Service)
	if err != nil {
		logger.Error(err, "unable to reconcile headless service", "request", req)
		return ctrl.Result{}, err
	}

	err = c.reconcileStatefulSet(ctx, observed.statefulSet, desired.StatefulSet)
	if err != nil {
		logger.Error(err, "unable to reconcile statefulset", "request", req)
		return ctrl.Result{}, err
	}

	// TODO: create or update the builder
	return ctrl.Result{}, nil
}

func (c *Controller) reconcileStatefulSet(
	ctx context.Context,
	observed *appsv1.StatefulSet,
	desired *appsv1.StatefulSet) error {

	if observed == nil && desired != nil {
		return c.Create(ctx, desired)
	}

	if observed != nil && desired != nil {
		return c.Update(ctx, desired)
	}

	return nil
}

func (c *Controller) reconcileService(
	ctx context.Context,
	observed *corev1.Service,
	desired *corev1.Service) error {

	if observed == nil && desired != nil {
		return c.Create(ctx, desired)
	}

	if observed != nil && desired != nil {
		return c.Update(ctx, desired)
	}

	return nil
}

func (c *Controller) reconcileConfigMap(
	ctx context.Context,
	observed *corev1.ConfigMap,
	desired *corev1.ConfigMap) error {

	if observed == nil && desired != nil {
		return c.Create(ctx, desired)
	}

	if observed != nil && desired != nil {
		return c.Update(ctx, desired)
	}

	return nil
}
