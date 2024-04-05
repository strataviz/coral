package agent

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semaphore", func() {
	Context("Acquire", func() {
		It("should acquire a semaphore", func() {
			s := NewSemaphore()
			Expect(s.Acquire("key")).To(BeTrue())
		})

		It("should not acquire an already acquired semaphore", func() {
			s := NewSemaphore()
			s.Acquire("key")
			Expect(s.Acquire("key")).To(BeFalse())
		})

		It("should acquire a different semaphores", func() {
			s := NewSemaphore()
			s.Acquire("key1")
			Expect(s.Acquire("key2")).To(BeTrue())
		})
	})

	Context("Release/Acquired", func() {
		It("should recognize an acquired semaphore", func() {
			s := NewSemaphore()
			s.Acquire("key")
			Expect(s.Acquired("key")).To(BeTrue())
		})

		It("should recognize a released semaphore", func() {
			s := NewSemaphore()
			Expect(s.Acquired("key")).To(BeFalse())
		})
	})
})
