package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
)

var _ = Describe("MySQL controller", func() {

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
		MySQLName      = "sample-mysql"
		MySQLNamespace = "default"
		timeout        = time.Second * 10
		interval       = time.Millisecond * 250
	)

	Context("When updating MySQL Status", func() {
		It("Should increase MySQL", func() {
			By("By creating a new MySQL")
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
