package controllers

import (
	"context"
	"time"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	. "github.com/nakamasato/mysql-operator/internal/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("MySQLDB controller", func() {
	Context("With available MySQL", func() {
		ctx := context.Background()
		var stopFunc func()
		BeforeEach(func() {
			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
			Expect(err).ToNot(HaveOccurred())
			err = (&MySQLDBReconciler{
				Client:             k8sManager.GetClient(),
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
			cleanUpMySQL(ctx, k8sClient, Namespace)
		})
		It("Should have finalizer", func() {
			By("By creating a new MySQL")
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
			db := &mysqlv1alpha1.MySQLDB{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLDB"},
				ObjectMeta: metav1.ObjectMeta{Name: "sample-db", Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLDBSpec{DBName: "sample_db", MysqlName: MySQLName},
			}
			Expect(k8sClient.Create(ctx, db)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: "sample-db"}, db)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(db, mysqlDBFinalizer)
			}, 5*time.Second).Should(BeTrue())
		})
		AfterEach(func() {
			cleanUpMySQL(ctx, k8sClient, Namespace)
			stopFunc()
			time.Sleep(100 * time.Millisecond)
		})
	})
})
