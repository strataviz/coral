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

package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"stvz.io/coral/pkg/mirror"
	"stvz.io/hashring"
)

const (
	MirrorUsage     = "mirror [ARG...]"
	MirrorShortDesc = "Start the coral mirror utility"
	MirrorLongDesc  = `Starts the coral mirror utility which watches for changes in the mirror resources and copies external images to a local registry.`

	MirrorConnectionTimeout  time.Duration = 30 * time.Second
	MirrorMaxCallRecvMsgSize int           = 1024 * 1024 * 32
)

type Mirror struct {
	logLevel  int8
	namespace string
	scope     string
	labels    string
	name      string
}

func NewMirror() *Mirror {
	return &Mirror{}
}

func (m *Mirror) RunE(cmd *cobra.Command, args []string) error {
	log := zap.New(
		zap.Level(zapcore.Level(m.logLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	log.Info("starting mirror")

	metrics, err := metricsserver.NewServer(
		metricsserver.Options{
			BindAddress: ":9090",
		},
		nil, nil,
	)
	if err != nil {
		log.Error(err, "failed to create metrics server")
		os.Exit(1)
	}

	go func() {
		err := metrics.Start(ctx)
		if err != nil {
			log.Error(err, "failed to start metrics server")
			os.Exit(1)
		}
	}()

	if m.name == "" {
		// try to get the name from the environment
		m.name = os.Getenv("HOSTNAME")
		if m.name == "" {
			log.Error(nil, "pod name not provided through argument or HOSTNAME environment variable")
			os.Exit(1)
		}
	}

	l, err := labels.Parse(m.labels)
	if err != nil {
		log.Error(err, "failed to parse labels")
		os.Exit(1)
	}

	mirrorCache := mirror.NewMirrorCache()
	ring := hashring.NewRing(1, nil)

	mirror := mirror.New(&mirror.Options{
		Log:         log,
		Scope:       m.scope,
		Namespace:   m.namespace,
		Name:        m.name,
		Labels:      l,
		MirrorCache: mirrorCache,
		Ring:        ring,
	})

	return mirror.Start(ctx)
}

func (m *Mirror) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   MirrorUsage,
		Short: MirrorShortDesc,
		Long:  MirrorLongDesc,
		RunE:  m.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&m.logLevel, "log-level", "v", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().StringVarP(&m.scope, "scope", "", DefaultScope, "limit the coral mirror to images in a specific namespace")
	cmd.PersistentFlags().StringVarP(&m.namespace, "namespace", "", DefaultNamespace, "the namespace of the deployment to watch for pod changes")
	cmd.PersistentFlags().StringVarP(&m.labels, "labels", "", DefaultLabels, "the match labels used to identify pods used by the mirror")
	cmd.PersistentFlags().StringVarP(&m.name, "name", "", "", "the pod name")
	return cmd
}
