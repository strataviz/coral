package injector

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"stvz.io/coral/pkg/injector/image"
)

type Injector struct{}

func SetupWebhookWithManager(mgr ctrl.Manager) (err error) {
	if err = image.SetupWebhookWithManager(mgr); err != nil {
		return
	}

	return
}
