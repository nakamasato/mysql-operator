package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/redhat-cop/operator-utils/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	. "github.com/nakamasato/mysql-operator/internal/mysql"
)

var _ = Describe("MySQLUser controller", func() {

	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		err = (&MySQLUserReconciler{
			ReconcilerBase: util.NewReconcilerBase(
				k8sManager.GetClient(),
				k8sManager.GetScheme(),
				k8sManager.GetConfig(),
				k8sManager.GetEventRecorderFor("mysqluser_controller"),
				k8sManager.GetAPIReader(),
			),
			Log:                nil,
			Scheme:             k8sManager.GetScheme(),
			MySQLClientFactory: NewFakeMySQLClient,
		}).SetupWithManager(k8sManager)
		Expect(err).ToNot(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err = k8sManager.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
		}()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	const (
		MySQLName     = "test-mysql"
		MySQLUserName = "test-mysql-user"
		Namespace     = "default"
	)

	Context("In normal case", func() {
		It("Should create Secret", func() {
			By("By creating a new MySQL")
			mysql := &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "localhost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

			By("By creating a new MySQLUser")
			mysqlUser := &mysqlv1alpha1.MySQLUser{
				TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQLUser"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLUserName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
				Status:     mysqlv1alpha1.MySQLUserStatus{},
			}
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			secret := &v1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: "mysql-test-mysql-" + MySQLUserName}, secret)
			}).Should(Succeed())
		})
	})
})
