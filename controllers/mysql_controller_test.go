package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
)

var _ = Describe("MySQL controller", func() {

	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		// set index for mysqluser with spec.mysqlName
		cache := k8sManager.GetCache()
		indexFunc := func(obj client.Object) []string {
			return []string{obj.(*mysqlv1alpha1.MySQLUser).Spec.MysqlName}
		}
		if err := cache.IndexField(ctx, &mysqlv1alpha1.MySQLUser{}, "spec.mysqlName", indexFunc); err != nil {
			panic(err)
		}

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

	Context("With available MySQL", func() {
		BeforeEach(func() {
			cleanUpMySQLUser(ctx, k8sClient, Namespace)
			cleanUpMySQL(ctx, k8sClient, Namespace)

			// Create MySQL
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "localhost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
		})
		AfterEach(func() {
			cleanUpMySQLUser(ctx, k8sClient, Namespace)
			cleanUpMySQL(ctx, k8sClient, Namespace)
		})
		It("Should have status.UserCount=0", func() {
			checkMySQLUserCount(ctx, int32(0))
		})

		It("Should increase status.UserCount by one", func() {
			By("By creating a new MySQLUser")
			mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
			addOwnerReferenceToMySQL(mysqlUser, mysql)
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			checkMySQLUserCount(ctx, int32(1))
		})

		It("Should decrease status.UserCount to zero", func() {
			By("By creating a new MySQLUser")
			mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
			addOwnerReferenceToMySQL(mysqlUser, mysql)
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			By("By deleting the MySQLUser")
			cleanUpMySQLUser(ctx, k8sClient, Namespace)

			checkMySQLUserCount(ctx, int32(0))
		})
	})
})

func checkMySQLUserCount(ctx context.Context, expectedUserCount int32) {
	Eventually(func() int32 {
		err := k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mysql)
		if err != nil {
			return -1
		}
		return mysql.Status.UserCount
	}, timeout, interval).Should(Equal(expectedUserCount))
}
