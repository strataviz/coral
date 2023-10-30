package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	builder "stvz.io/coral/pkg/controller/builder"
	"stvz.io/coral/pkg/controller/queue"
)

type ControllerOpts struct{}

type Controller struct{}

func SetupWithManager(mgr ctrl.Manager) (err error) {
	if err = builder.SetupWithManager(mgr); err != nil {
		return
	}

	if err = queue.SetupWithManager(mgr); err != nil {
		return
	}

	return
}
