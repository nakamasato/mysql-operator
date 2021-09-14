package controllers

import (
	"time"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("MySQL controller", func() {

	const (
		MySQLName      = "test-mysql-user"
		MySQLNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating MySQL Status", func() {
		It("Should increase MySQL", func() {
			By("By creating a new MySQL")
		})
	})
})
