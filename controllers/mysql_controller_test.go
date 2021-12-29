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

	Context("With available MySQL", func() {
		It("Should have status.UserCount=0", func() {
			By("By creating a new MySQL")
			mysql := &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: MySQLNamespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "localhost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

			lookUpKey := types.NamespacedName{Name: MySQLName, Namespace: MySQLNamespace}
			mySQL := &mysqlv1alpha1.MySQL{}
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, lookUpKey, mySQL)
				if err != nil {
					return -1
				}
				return mySQL.Status.UserCount
			}, timeout, interval).Should(Equal(int32(0)))

		})
	})
})
