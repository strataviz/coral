package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// TODO: break me out to observer
	var observed stvziov1.Builder
	err := c.Get(ctx, req.NamespacedName, &observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
