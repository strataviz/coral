package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"stvz.io/coral/pkg/controller/image"
	"stvz.io/coral/pkg/monitor"
)

type ControllerOpts struct{}

type Controller struct{}

func SetupWithManager(mgr ctrl.Manager, mtr *monitor.Manager) (err error) {
	if err = image.SetupWithManager(mgr, mtr); err != nil {
		return
	}

	return
}
