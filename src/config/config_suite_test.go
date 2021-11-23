package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dummy Test", func() {
	It("Should pass", func() {
		Expect(1).To(Equal(0))
	})
})

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}
