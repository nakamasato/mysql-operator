package controllers

import (
	. "github.com/onsi/ginkgo"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
)

var _ = Describe("MySQLUser controller", func() {

	const (
		MySQLUserName      = "test-mysql-user"
		MySQLUserNamespace = "default"
	)

	Context("In normal case", func() {
		It("Should create Secret", func() {
			By("By creating a new MySQLUser")
			mysqlUser := &mysqlv1alpha1.MySQLUser{
				TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQLUser"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLUserName, Namespace: MySQLUserNamespace},
				Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: "test-mysql"},
				Status:     mysqlv1alpha1.MySQLUserStatus{},
			}
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			secret := &v1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: MySQLUserNamespace, Name: "mysql-test-mysql-" + MySQLUserName}, secret)
			}).Should(Succeed())
		})
	})
})
