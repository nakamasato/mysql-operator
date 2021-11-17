package e2e_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// "github.com/nakamasato/mysql-operator/e2e"
)

var _ = Describe("E2e", func() {
	BeforeEach(func() {
		fmt.Printf("before each in describe\n")
	})

	AfterEach(func() {
		fmt.Printf("after each in describe\n")
	})

	Describe("Test", func() {
		Context("context", func() {
			It("should pass", func() {
				Expect("test").To(Equal("test"))
			})
		})

		Context("context2", func() {
			It("should pass", func() {
				Expect("test").To(Equal("test"))
			})
		})
	})
})
