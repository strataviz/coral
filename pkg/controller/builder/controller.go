package builder

import (
	"context"

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

func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&stvziov1.Builder{}).Complete(c)
}

// +kubebuilder:rbac:groups=stvz.io,resources=builders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stvz.io,resources=builders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stvz.io,resources=builders/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=jobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (c Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	observed := NewObservedState()
	observer := StateObserver{
		Client:  c.Client,
		Request: req,
	}

	err := observer.observe(ctx, &observed)
	if err != nil {
		logger.Error(err, "unable to observe state", "request", req)
		return ctrl.Result{}, err
	}

	if observed.builder == nil {
		logger.Info("builder has been deleted, cleaning up", "request", req)
		// TODO: cleanup
		return ctrl.Result{}, nil
	}

	logger.Info("reconciling builder", "request", req, "spec", observed.builder.Spec)

	var token string
	if observed.token == nil {
		logger.Info("unable to find the secret for the builder, private repos will not be accessible", "request", req)
	} else {
		if tok, ok := observed.token.Data["token"]; !ok {
			logger.Info("unable to find the token for the builder, private repos will not be accessible", "request", req)
		} else {
			token = string(tok)
			logger.Info("found the token for the builder", "request", req, "token", token)
		}
	}

	// TODO: create or update the builder
	return ctrl.Result{}, nil
}
