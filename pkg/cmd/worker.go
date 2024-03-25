package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	criRun "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	v1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/worker"
)

const (
	WorkerUsage     = "worker [ARG...]"
	WorkerShortDesc = "Start the coral worker"
	WorkerLongDesc  = `Starts the coral worker which ensures the image state of the node is in sync with the configured images.`

	WorkerConnectionTimeout  time.Duration = 30 * time.Second
	WorkerMaxCallRecvMsgSize int           = 1024 * 1024 * 32
)

type Worker struct {
	logLevel       int8
	containerdAddr string
	pullInterval   time.Duration
	cleanInterval  time.Duration
	namespace      string
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) RunE(cmd *cobra.Command, args []string) error {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	log := zap.New(
		zap.Level(zapcore.Level(w.logLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()

	log.Info("starting worker")

	nodeName := os.Getenv("CORAL_NODE_NAME")
	if nodeName == "" {
		log.Error(nil, "CORAL_NODE_NAME must be set in the environment.")
		os.Exit(1)
	}

	ims, rts, err := w.getClient(ctx)
	if err != nil {
		log.Error(err, "unable to get CRI client")
		os.Exit(1)
	}

	kubeClient, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	werk := worker.NewWorker().
		WithImageServiceClient(ims).
		WithRuntimeServiceClient(rts).
		WithKubeClient(kubeClient).
		WithLogger(log.WithValues("node", nodeName)).
		WithName(nodeName).
		WithPullInterval(w.pullInterval).
		WithCleanInterval(w.cleanInterval).
		WithNamespace(w.namespace)

	werk.Start(ctx)

	return nil
}

func (w *Worker) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   WorkerUsage,
		Short: WorkerShortDesc,
		Long:  WorkerLongDesc,
		RunE:  w.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&w.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().DurationVarP(&w.pullInterval, "pull-interval", "", DefaultPullInterval, "set the puller poll interval")
	cmd.PersistentFlags().DurationVarP(&w.cleanInterval, "clean-interval", "", DefaultCleanInterval, "set the clean workers poll interval")
	cmd.PersistentFlags().StringVarP(&w.containerdAddr, "containerd-addr", "", DefaultContainerdAddr, "set the containerd address")
	cmd.PersistentFlags().StringVarP(&w.namespace, "namespace", "", DefaultNamespace, "limit the coral worker to images in a specific namespace")
	return cmd
}

func (w *Worker) getClient(ctx context.Context) (criRun.ImageServiceClient, criRun.RuntimeServiceClient, error) {
	addr, dialer, err := util.GetAddressAndDialer(w.containerdAddr)
	if err != nil {
		klog.ErrorS(err, "Get containerd address failed")
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, WorkerConnectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(WorkerMaxCallRecvMsgSize)),
	)
	if err != nil {
		klog.ErrorS(err, "Connect remote image service failed", "address", addr)
		return nil, nil, err
	}

	ims := criRun.NewImageServiceClient(conn)
	rts := criRun.NewRuntimeServiceClient(conn)
	return ims, rts, nil
}
