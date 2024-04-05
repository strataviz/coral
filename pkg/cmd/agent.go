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
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crun "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"stvz.io/coral/pkg/agent"
	v1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

const (
	AgentUsage     = "agent [ARG...]"
	AgentShortDesc = "Start the coral agent"
	AgentLongDesc  = `Starts the coral agent which ensures the image state of the node is in sync with the configured images.`

	AgentConnectionTimeout  time.Duration = 30 * time.Second
	AgentMaxCallRecvMsgSize int           = 1024 * 1024 * 32

	ConnectionTimeout  time.Duration = 30 * time.Second
	MaxCallRecvMsgSize int           = 1024 * 1024 * 32
)

type Agent struct {
	logLevel       int8
	parallel       int
	containerdAddr string
	pollInterval   time.Duration
	namespace      string
}

func NewAgent() *Agent {
	return &Agent{}
}

func (a *Agent) RunE(cmd *cobra.Command, args []string) error {
	log := zap.New(
		zap.Level(zapcore.Level(a.logLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	log.Info("starting agent")

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		log.Error(nil, "NODE_NAME must be set in the environment.")
		os.Exit(1)
	}

	ims, rts, err := a.connectContainerRuntime(ctx, a.containerdAddr)
	if err != nil {
		log.Error(err, "failed to connect to container runtime")
		os.Exit(1)
	}

	c, err := a.connectKubeClient()
	if err != nil {
		log.Error(err, "failed to connect to kube client")
		os.Exit(1)
	}

	metrics, err := metricsserver.NewServer(
		metricsserver.Options{
			BindAddress: ":8080",
		},
		nil, nil,
	)
	if err != nil {
		log.Error(err, "failed to create metrics server")
		os.Exit(1)
	}

	// The built-in metrics server doesn't appear to have any mechanism for stopping
	// gracefully, so we'll just start it and forget about it.  I may come back to this
	// and strip out the built-in for strata-go, though I'd be missing out on some
	// good default metrics.
	go func() {
		err := metrics.Start(ctx)
		if err != nil {
			log.Error(err, "failed to start metrics server")
			os.Exit(1)
		}
	}()

	options := &agent.AgentOptions{
		Log:                  log,
		WorkerProcesses:      a.parallel,
		Namespace:            a.namespace,
		PollInterval:         a.pollInterval,
		ImageServiceClient:   ims,
		RuntimeServiceClient: rts,
		Client:               c,
		NodeName:             nodeName,
	}

	agent := agent.NewAgent(options)
	agent.Start(ctx)

	return nil
}

func (w *Agent) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   AgentUsage,
		Short: AgentShortDesc,
		Long:  AgentLongDesc,
		RunE:  w.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&w.logLevel, "log-level", "v", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().DurationVarP(&w.pollInterval, "poll-interval", "i", DefaultPollInterval, "set the puller poll interval")
	cmd.PersistentFlags().StringVarP(&w.containerdAddr, "containerd-addr", "A", DefaultContainerdAddr, "set the containerd address")
	cmd.PersistentFlags().StringVarP(&w.namespace, "namespace", "n", DefaultNamespace, "limit the coral agent to images in a specific namespace")
	cmd.PersistentFlags().IntVarP(&w.parallel, "parallel", "p", DefaultParallel, "set the number of parallel workers")
	return cmd
}

func (a *Agent) connectContainerRuntime(ctx context.Context, addr string) (crun.ImageServiceClient, crun.RuntimeServiceClient, error) {
	addr, dialer, err := util.GetAddressAndDialer(addr)
	if err != nil {
		klog.ErrorS(err, "Get container runtime address failed")
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, ConnectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize)),
	)
	if err != nil {
		klog.ErrorS(err, "Connect remote image service failed", "address", addr)
		return nil, nil, err
	}

	ims := crun.NewImageServiceClient(conn)
	rts := crun.NewRuntimeServiceClient(conn)
	return ims, rts, nil
}

func (a *Agent) connectKubeClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	c, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
