package image

import (
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
)

// const (
// 	timeout  = time.Second * 5
// 	duration = time.Second * 10
// 	interval = time.Millisecond * 250
// )

// var tagLabels = map[string]string{
// 	"test:tag1": "image.stvz.io/69d3a74873a698ff1b861ce171fad465",
// 	"test:tag2": "image.stvz.io/642a0b75bac87e3b38fccdcf6c7b0722",
// }

var _ = Context("In a clean environment", func() {
	// ctx := context.TODO()
	// ns := SetupTest(ctx)

	Describe("Image Controller", func() {
		BeforeEach(func() {
			// setupNodes(ctx)
		})

		AfterEach(func() {
			// tearDownNodes(ctx)
		})

		// Right now the controller is not doing any modifications to the image
		// so no need to test the reconcile function.  Once we add the monitors
		// we may be able to test the monitor startup.
	})
})
