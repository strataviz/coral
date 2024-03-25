package image

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/monitor"
)

type Controller struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Monitor  *monitor.Manager
}

func SetupWithManager(mgr ctrl.Manager, mtr *monitor.Manager) error {
	c := &Controller{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("image-controller"),
		Monitor:  mtr,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&stvziov1.Image{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(c)
}

// +kubebuilder:rbac:groups=stvz.io,resources=images,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stvz.io,resources=images/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stvz.io,resources=images/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

// Reconcile is the main controller loop for the image controller.
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
		return ctrl.Result{
			RequeueAfter: 10 * time.Second,
		}, err
	}

	if observed.image == nil {
		return ctrl.Result{}, nil
	}

	logger.Info("observing state",
		"image", observed.image.Spec,
	)

	finalizer := &finalizer{
		client: c.Client,
		image:  observed.image,
	}
	// TODO: revisit this, I want to make this more clear that doesn't actually finalize if
	// it's not needed.  Otherwise, it's just a no-op.  I need to potentially catch other
	// errors for requeue.
	requeue, err := finalizer.reconcile(ctx)
	if err != nil {
		return ctrl.Result{
			RequeueAfter: 10 * time.Second,
		}, err
	} else if requeue {
		return ctrl.Result{}, nil
	}

	// Place the image resource under monitoring
	c.Monitor.AddImage(ctx, observed.image.DeepCopy())

	return ctrl.Result{}, nil
}
