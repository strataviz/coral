package agent

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	ctx      context.Context
	cancel   context.CancelFunc
	logger   logr.Logger
	fixtures = filepath.Join("..", "..", "fixtures", "agent_test")
)

func TestImageController(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.ParallelTotal = 1
	RunSpecs(t, "Image Controller Suite", suiteConfig)
}

var _ = BeforeSuite(func() {
	logger := zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	logf.SetLogger(logger)
	ctx, cancel = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	cancel()
})
