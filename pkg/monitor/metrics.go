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

package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	monitorError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coral_monitor_error",
			Help: "The number of errors that occurred while monitoring an image.",
		},
		[]string{"name", "namespace", "error"},
	)

	monitorDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "coral_monitor_duration_seconds",
			Help:    "The duration of monitoring run for an image.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"name", "namespace"},
	)

	monitorImagesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "coral_monitor_images_total",
			Help: "The total number of nodes that have the image",
		},
		[]string{"name", "namespace"},
	)

	monitorNodesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "coral_monitor_nodes_total",
			Help: "The total number of nodes under management for the image",
		},
		[]string{"name", "namespace"},
	)

	monitorImagesPending = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "coral_monitor_images_pending",
			Help: "The number of nodes that have the image pending",
		},
		[]string{"name", "namespace"},
	)

	monitorImagesAvailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "coral_monitor_images_available",
			Help: "The number of nodes that have the image available",
		},
		[]string{"name", "namespace"},
	)

	monitorImagesDeleting = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "coral_monitor_images_deleting",
			Help: "The number of nodes that have the image deleting",
		},
		[]string{"name", "namespace"},
	)

	monitorImagesUnknown = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "coral_monitor_images_unknown",
			Help: "The number of nodes that have the image in an unknown state",
		},
		[]string{"name", "namespace"},
	)
)

func init() {
	metrics.Registry.MustRegister(monitorError)
	metrics.Registry.MustRegister(monitorDuration)
	metrics.Registry.MustRegister(monitorImagesPending)
	metrics.Registry.MustRegister(monitorImagesAvailable)
	metrics.Registry.MustRegister(monitorImagesDeleting)
	metrics.Registry.MustRegister(monitorImagesUnknown)
	metrics.Registry.MustRegister(monitorImagesTotal)
	metrics.Registry.MustRegister(monitorNodesTotal)
}
