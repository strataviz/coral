package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	agentError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coral_agent_error",
			Help: "Errors that occurred while the agent is running.",
		},
		[]string{"error"},
	)

	agentImageError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coral_agent_image_error",
			Help: "Errors that occurred while the agent is running image processing.",
		},
		[]string{"image", "error"},
	)

	agentRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "coral_agent_run_duration_ms",
			Help:    "The duration of the agent run.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	agentImagePulls = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "coral_agent_image_pulls",
			Help: "The number of image pulls.",
		},
	)

	agentImageRemovals = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "coral_agent_image_removals",
			Help: "The number of image removals.",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(agentError)
	metrics.Registry.MustRegister(agentImageError)
	metrics.Registry.MustRegister(agentRunDuration)
	metrics.Registry.MustRegister(agentImagePulls)
	metrics.Registry.MustRegister(agentImageRemovals)
}
