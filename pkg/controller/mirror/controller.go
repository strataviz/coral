package mirror

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		Recorder: mgr.GetEventRecorderFor("mirror-controller"),
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&stvziov1.Mirror{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(c)
}

// +kubebuilder:rbac:groups=stvz.io,resources=mirrors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stvz.io,resources=mirrors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stvz.io,resources=mirrors/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(6).Info("reconciling image", "request", req)

	// err := observer.observe(ctx, observed)
	// if err != nil {
	// 	logger.Error(err, "unable to observe state", "request", req)
	// 	return ctrl.Result{
	// 		RequeueAfter: 10 * time.Second,
	// 	}, err
	// }

	// The image has been deleted.
	// if observed.mirror == nil {
	// 	return ctrl.Result{}, nil
	// }

	// logger.V(8).Info("observed mirror", "obj", observed.mirror)

	// desired, err := GetDesiredState(observed)
	// if err != nil {
	// 	logger.Error(err, "unable to get desired state", "request", req)
	// 	return ctrl.Result{
	// 		RequeueAfter: 10 * time.Second,
	// 	}, err
	// }

	// logger.V(8).Info("desired state", "obj", desired)
	// err = c.reconcileDeployment(ctx, observed.deployment, desired.deployment)
	// if err != nil {
	// 	logger.Error(err, "unable to reconcile deployment", "request", req)
	// 	return ctrl.Result{
	// 		RequeueAfter: 10 * time.Second,
	// 	}, err
	// }

	return ctrl.Result{}, nil
}

// func (c *Controller) reconcileDeployment(ctx context.Context, observed *appsv1.Deployment, desired *appsv1.Deployment) error {
// 	if observed == nil && desired != nil {
// 		return c.Create(ctx, desired)
// 	}

// 	if observed != nil && desired != nil {
// 		return c.Update(ctx, observed)
// 	}

// 	return nil
// }
