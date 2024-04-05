// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
)

func init() {
	metrics.Registry.MustRegister(agentError)
	metrics.Registry.MustRegister(agentImageError)
	metrics.Registry.MustRegister(agentRunDuration)
	metrics.Registry.MustRegister(agentImagePulls)
}
