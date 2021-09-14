package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
)

var _ = Describe("MySQLUser controller", func() {

	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		err = (&MySQLReconciler{
			Client: k8sManager.GetClient(),
			Scheme: k8sManager.GetScheme(),
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
		MySQLUserName      = "test-mysql-user"
		MySQLUserNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating MySQLUser Status", func() {
		It("Should increase MySQLUser Status.Active count when new Jobs are created", func() {
			By("By creating a new MySQLUser")
			mysqlUser := &mysqlv1alpha1.MySQLUser{
				TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQLUser"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLUserName, Namespace: MySQLUserNamespace},
				Spec:       mysqlv1alpha1.MySQLUserSpec{Name: MySQLUserName, MysqlName: "test-mysql"},
				Status:     mysqlv1alpha1.MySQLUserStatus{},
			}
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
		})
	})
})
