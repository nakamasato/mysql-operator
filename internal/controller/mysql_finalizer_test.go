package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
)

var _ = Describe("MySQL Finalizer", func() {
	const testNamespace = "default"
	ctx := context.Background()

	BeforeEach(func() {
		// Clean up any existing resources
		cleanUpMySQLUser(ctx, k8sClient, testNamespace)
		cleanUpMySQLDB(ctx, k8sClient, testNamespace)
		cleanUpMySQL(ctx, k8sClient, testNamespace)
	})

	AfterEach(func() {
		// Clean up test resources
		cleanUpMySQLUser(ctx, k8sClient, testNamespace)
		cleanUpMySQLDB(ctx, k8sClient, testNamespace)
		cleanUpMySQL(ctx, k8sClient, testNamespace)
	})

	Context("Finalizer Behavior", func() {
		It("Should add finalizer before creating dependent resources", func() {
			// Create MySQL without finalizer
			mysql := &mysqlv1alpha1.MySQL{
				TypeMeta: metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-finalizer",
					Namespace: testNamespace,
				},
				Spec: mysqlv1alpha1.MySQLSpec{
					Host:          "localhost",
					AdminUser:     mysqlv1alpha1.Secret{Name: "root", Type: "raw"},
					AdminPassword: mysqlv1alpha1.Secret{Name: "password", Type: "raw"},
				},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

			// Verify finalizer is added
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-mysql-finalizer", Namespace: testNamespace}, mysql)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(mysql, mysqlFinalizer)
			}, time.Second*10).Should(BeTrue())

			// Create dependent MySQLUser
			mysqlUser := &mysqlv1alpha1.MySQLUser{
				TypeMeta: metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-user",
					Namespace: testNamespace,
				},
				Spec: mysqlv1alpha1.MySQLUserSpec{
					MysqlName: "test-mysql-finalizer",
					Host:      "%",
				},
			}
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			// Delete MySQL
			Expect(k8sClient.Delete(ctx, mysql)).Should(Succeed())

			// Verify MySQL is not immediately deleted due to finalizer
			Consistently(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: "test-mysql-finalizer", Namespace: testNamespace}, mysql)
			}, time.Second*2).Should(Succeed())

			// Verify dependent resources are cleaned up
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: "test-user", Namespace: testNamespace}, mysqlUser)
			}, time.Second*10).ShouldNot(Succeed())

			// Verify MySQL is eventually deleted
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: "test-mysql-finalizer", Namespace: testNamespace}, mysql)
			}, time.Second*5).ShouldNot(Succeed())
		})

		It("Should handle deletion before finalizer addition", func() {
			// Create MySQL without finalizer
			mysql := &mysqlv1alpha1.MySQL{
				TypeMeta: metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-no-finalizer",
					Namespace: testNamespace,
				},
				Spec: mysqlv1alpha1.MySQLSpec{
					Host:          "localhost",
					AdminUser:     mysqlv1alpha1.Secret{Name: "root", Type: "raw"},
					AdminPassword: mysqlv1alpha1.Secret{Name: "password", Type: "raw"},
				},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

			// Delete immediately
			Expect(k8sClient.Delete(ctx, mysql)).Should(Succeed())

			// Verify it's deleted without issues
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: "test-mysql-no-finalizer", Namespace: testNamespace}, mysql)
			}, time.Second*5).ShouldNot(Succeed())
		})
	})
})
