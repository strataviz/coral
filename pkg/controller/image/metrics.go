package image

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	observerError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coral_image_controller_observer_error",
			Help: "The number of errors that occurred while observing the state of an image.",
		},
		[]string{"name", "namespace"},
	)
)

func init() {
	metrics.Registry.MustRegister(observerError)
}
