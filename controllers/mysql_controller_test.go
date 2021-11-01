package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
)

var _ = Describe("MySQL controller", func() {

	const (
		MySQLName      = "sample-mysql"
		MySQLNamespace = "default"
		timeout        = time.Second * 10
		duration       = time.Second * 10
		interval       = time.Millisecond * 250
	)

	Context("When updating MySQL Status", func() {
		It("Should increase MySQL", func() {
			By("By creating a new MySQL")
			ctx := context.Background()
			mysql := &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: MySQLNamespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "localhost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

			lookUpKey := types.NamespacedName{Name: MySQLName, Namespace: MySQLNamespace}
			createdMySQL := &mysqlv1alpha1.MySQL{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookUpKey, createdMySQL)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdMySQL.Spec.Host).Should(Equal("localhost"))
		})
	})
})
