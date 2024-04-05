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
	"crypto/tls"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/controller"
	"stvz.io/coral/pkg/injector"
	"stvz.io/coral/pkg/monitor"
)

const (
	ControllerUsage     = "controller [ARG...]"
	ControllerShortDesc = "Start the build controller"
	ControllerLongDesc  = `Starts the build controller providing management of the
kubernetes resources and services.`
)

type Controller struct {
	certs              string
	leaderElection     bool
	skipInsecureVerify bool
	namespace          string

	scheme *runtime.Scheme

	logLevel int8
	// logEncoder string
}

func NewController() *Controller {
	return &Controller{
		scheme: runtime.NewScheme(),
	}
}

func (c *Controller) RunE(cmd *cobra.Command, args []string) error {
	_ = stvziov1.AddToScheme(c.scheme)
	_ = corev1.AddToScheme(c.scheme)
	_ = appsv1.AddToScheme(c.scheme)

	// TODO: more configurations to mirror bind flags.
	log := zap.New(
		zap.Level(zapcore.Level(c.logLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	log.Info("initializing manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           c.scheme,
		LeaderElection:   c.leaderElection,
		LeaderElectionID: "coral-leader-lock",
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: c.certs,
			Port:    9443,
			TLSOpts: []func(*tls.Config){
				func(config *tls.Config) {
					config.InsecureSkipVerify = c.skipInsecureVerify
				},
			},
		}),
	})

	if err != nil {
		log.Error(err, "unable to initialize manager")
		os.Exit(1)
	}

	mtr := monitor.NewMonitor(mgr.GetClient(), c.namespace, log)
	go mtr.Start(ctx)

	if err = (&stvziov1.Image{}).SetupWebhookWithManager(mgr); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "Image")
		os.Exit(1)
	}

	if err = injector.SetupWebhookWithManager(mgr); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "Pod")
		os.Exit(1)
	}

	if err = controller.SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to setup controllers")
		os.Exit(1)
	}

	// Start the manager process
	log.Info("starting manager")
	err = mgr.Start(ctx)
	mtr.Stop()

	return err
}

func (c *Controller) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ControllerUsage,
		Short: ControllerShortDesc,
		Long:  ControllerLongDesc,
		RunE:  c.RunE,
	}

	cmd.PersistentFlags().StringVarP(&c.certs, "certs", "c", DefaultCertDir, "specify the webhooks certs directory")
	cmd.PersistentFlags().BoolVarP(&c.leaderElection, "enable-leader-election", "", DefaultEnableLeaderElection, "enable leader election")
	cmd.PersistentFlags().BoolVarP(&c.skipInsecureVerify, "skip-insecure-verify", "", DefaultSkipInsecureVerify, "skip certificate verification for the webhooks")
	cmd.PersistentFlags().Int8VarP(&c.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().StringVarP(&c.namespace, "namespace", "n", DefaultNamespace, "limit the coral scope to a specific namespace")
	return cmd
}
