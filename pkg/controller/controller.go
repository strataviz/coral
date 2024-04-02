package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"stvz.io/coral/pkg/controller/image"
)

type ControllerOpts struct{}

type Controller struct{}

func SetupWithManager(mgr ctrl.Manager) (err error) {
	if err = image.SetupWithManager(mgr); err != nil {
		return
	}

	return
}
