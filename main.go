package main

import (
	"crypto/tls"
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	v1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/controller"
)

const (
	DefaultCertDir              string = "/etc/webhook/tls"
	DefaultEnableLeaderElection bool   = false
	DefaultSkipInsecureVerify   bool   = false
)

var (
	// Temporary logger for initial setup
	setupLog           = ctrl.Log.WithName("setup")
	scheme             = runtime.NewScheme()
	certDir            string
	leaderElection     bool
	skipInsecureVerify bool
)

func init() {
	_ = v1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	flag.StringVar(&certDir, "certs", DefaultCertDir, "specify the cert directory")
	flag.BoolVar(&leaderElection, "enable-leader-election", DefaultEnableLeaderElection, "enable leader election")
	flag.BoolVar(&skipInsecureVerify, "skip-insecure-verify", DefaultSkipInsecureVerify, "skip insecure verify")
}

func main() {
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	log := zap.New(zap.UseFlagOptions(&opts))

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	setupLog.Info("initializing manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   leaderElection,
		LeaderElectionID: "coral-lock",
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: certDir,
			Port:    9443,
			TLSOpts: []func(*tls.Config){
				func(config *tls.Config) {
					config.InsecureSkipVerify = skipInsecureVerify
				},
			},
		}),
	})

	if err != nil {
		log.Error(err, "unable to initialize manager")
		os.Exit(1)
	}

	if err = (&v1.Builder{}).SetupWebhookWithManager(mgr); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "Builder")
		os.Exit(1)
	}

	if err = controller.SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to setup controller")
		os.Exit(1)
	}

	// Start the manager process
	log.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
