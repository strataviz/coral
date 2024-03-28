package image

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestV1API(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Image Injector Suite")
}
