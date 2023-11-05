package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"stvz.io/coral/pkg/controller/buildqueue"
	"stvz.io/coral/pkg/controller/buildset"
	"stvz.io/coral/pkg/controller/watchset"
)

type ControllerOpts struct{}

type Controller struct{}

func SetupWithManager(mgr ctrl.Manager) (err error) {
	if err = watchset.SetupWithManager(mgr); err != nil {
		return
	}

	if err = buildset.SetupWithManager(mgr); err != nil {
		return
	}

	if err = buildqueue.SetupWithManager(mgr); err != nil {
		return
	}

	return
}
