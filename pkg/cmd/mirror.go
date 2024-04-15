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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/informer/mirror"
	command "stvz.io/coral/pkg/mirror"
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
	).WithName("mirror")

	scheme := runtime.NewScheme()
	_ = stvziov1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	log.Info("gathering host information")
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

	log.Info("initializing manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Error(err, "unable to initialize manager")
		os.Exit(1)
	}

	log.Info("setting up informer")
	informer, err := mirror.SetupWithManager(ctx, mgr, m.namespace, l)
	if err != nil {
		log.Error(err, "unable to setup informer")
		os.Exit(1)
	}

	// TODO: better/existent error handling
	go informer.Start(ctx) // nolint:errcheck

	// Think about moving all the command stuff into directories here in cmd...
	mirror := command.New(&command.Options{
		Scope:     m.scope,
		Namespace: m.namespace,
		Name:      m.name,
		Labels:    l,
		Informer:  informer,
	})

	log.Info("starting mirror watcher")
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
