package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	internalmysql "github.com/nakamasato/mysql-operator/internal/mysql"
	"github.com/nakamasato/mysql-operator/internal/secret"
)

var _ = Describe("MySQL controller", func() {

	ctx := context.Background()
	var stopFunc func()
	var mySQLClients internalmysql.MySQLClients

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
		indexFunc = func(obj client.Object) []string {
			return []string{obj.(*mysqlv1alpha1.MySQLDB).Spec.MysqlName}
		}
		if err := cache.IndexField(ctx, &mysqlv1alpha1.MySQLDB{}, "spec.mysqlName", indexFunc); err != nil {
			panic(err)
		}

		mySQLClients = internalmysql.MySQLClients{}
		err = (&MySQLReconciler{
			Client:          k8sManager.GetClient(),
			Scheme:          k8sManager.GetScheme(),
			MySQLClients:    mySQLClients,
			MySQLDriverName: "testdbdriver",
			SecretManagers:  map[string]secret.SecretManager{"raw": secret.RawSecretManager{}},
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
			cleanUpMySQLDB(ctx, k8sClient, Namespace)
			cleanUpMySQL(ctx, k8sClient, Namespace)

			// Create MySQL
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: mysqlv1alpha1.Secret{Name: "root", Type: "raw"}, AdminPassword: mysqlv1alpha1.Secret{Name: "password", Type: "raw"}},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
		})
		AfterEach(func() {
			cleanUpMySQLUser(ctx, k8sClient, Namespace)
			cleanUpMySQLDB(ctx, k8sClient, Namespace)
			cleanUpMySQL(ctx, k8sClient, Namespace)
		})
		It("Should have status.UserCount=0", func() {
			checkMySQLUserCount(ctx, int32(0))
		})

		It("Should increase status.UserCount by one", func() {
			By("By creating a new MySQLUser")
			mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
			Expect(controllerutil.SetOwnerReference(mysql, mysqlUser, scheme)).Should(Succeed())
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			checkMySQLUserCount(ctx, int32(1))
		})

		It("Should increase status.UserCount to two", func() {
			By("By creating a new MySQLUser")
			mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
			Expect(controllerutil.SetOwnerReference(mysql, mysqlUser, scheme)).Should(Succeed())
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
			checkMySQLUserCount(ctx, int32(1))

			By("By creating another MySQLUser")
			mysqlUser2 := newMySQLUser(APIVersion, Namespace, "mysql-test-user-2", MySQLName)
			Expect(controllerutil.SetOwnerReference(mysql, mysqlUser2, scheme)).Should(Succeed())
			Expect(k8sClient.Create(ctx, mysqlUser2)).Should(Succeed())

			// TODO: check why Owns doesn't trigger reconciliation
			mysql.Spec.Host = "localhost"
			Expect(k8sClient.Update(ctx, mysql)).Should(Succeed())

			checkMySQLUserCount(ctx, int32(2))
		})

		It("Should decrease status.UserCount to zero", func() {
			By("By creating a new MySQLUser")
			mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
			Expect(controllerutil.SetOwnerReference(mysql, mysqlUser, scheme)).Should(Succeed())
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
			checkMySQLUserCount(ctx, int32(1))

			By("By deleting the MySQLUser")
			Expect(k8sClient.Delete(ctx, mysqlUser)).Should(Succeed())

			// TODO: check why Owns doesn't trigger reconciliation
			mysql.Spec.Host = "localhost"
			Expect(k8sClient.Update(ctx, mysql)).Should(Succeed())

			checkMySQLUserCount(ctx, int32(0))
		})

		It("Should increase status.DBCount by one", func() {
			By("By creating a new MySQLDB")
			mysqlDB = newMySQLDB(APIVersion, Namespace, MySQLDBName, DatabaseName, MySQLName)
			Expect(controllerutil.SetOwnerReference(mysql, mysqlDB, scheme)).Should(Succeed())
			Expect(k8sClient.Create(ctx, mysqlDB)).Should(Succeed())

			checkMySQLDBCount(ctx, int32(1))
		})

		It("Should have finalizer", func() {
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLName}, mysql)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(mysql, mysqlFinalizer)
			}).Should(BeTrue())
		})

		It("Should update MySQLClient", func() {
			Eventually(func() error {
				_, err := mySQLClients.GetClient(mysql.GetKey())
				return err
			}).Should(BeNil())
			Eventually(func() int {
				return len(mySQLClients)
			}).Should(Equal(1))
		})

		It("Should clean up MySQLClient", func() {
			// Wait until MySQLClients is updated
			Eventually(func() error {
				_, err := mySQLClients.GetClient(mysql.GetKey())
				return err
			}).Should(BeNil())
			Eventually(func() int {
				return len(mySQLClients)
			}).Should(Equal(1))

			By("By deleting MySQL")
			Expect(k8sClient.Delete(ctx, mysql)).Should(Succeed())

			Eventually(func() int {
				return len(mySQLClients)
			}, 5*time.Second).Should(Equal(0))
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
	}, 5*time.Second).Should(Equal(expectedUserCount))
}

func checkMySQLDBCount(ctx context.Context, expectedDBCount int32) {
	Eventually(func() int32 {
		err := k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mysql)
		if err != nil {
			return -1
		}
		return mysql.Status.DBCount
	}, 5*time.Second).Should(Equal(expectedDBCount))
}
