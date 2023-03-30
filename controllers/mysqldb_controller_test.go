package controllers

import (
	"context"
	"database/sql"
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
	It("Should create database", func() {
		db, err := sql.Open("testdbdriver", "test")
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()
		ctx := context.Background()
		_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS test_database")
		Expect(err).ToNot(HaveOccurred())
	})
	Context("With available MySQL", func() {
		ctx := context.Background()
		var stopFunc func()
		var close func() error
		BeforeEach(func() {
			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
			Expect(err).ToNot(HaveOccurred())
			db, err := sql.Open("testdbdriver", "test")
			close = db.Close
			Expect(err).ToNot(HaveOccurred())
			err = (&MySQLDBReconciler{
				Client:       k8sManager.GetClient(),
				Scheme:       k8sManager.GetScheme(),
				MySQLClients: MySQLClients{MySQLName: db},
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
		It("Should have mysqlDBFinalizer", func() {
			By("By creating a new MySQL")
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
			mysqlDB := &mysqlv1alpha1.MySQLDB{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLDB"},
				ObjectMeta: metav1.ObjectMeta{Name: "sample-db", Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLDBSpec{DBName: "sample_db", MysqlName: MySQLName},
			}
			Expect(k8sClient.Create(ctx, mysqlDB)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: "sample-db"}, mysqlDB)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(mysqlDB, mysqlDBFinalizer)
			}).Should(BeTrue())
		})

		It("Should be ready", func() {
			By("By creating a new MySQL")
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
			mysqlDB := &mysqlv1alpha1.MySQLDB{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLDB"},
				ObjectMeta: metav1.ObjectMeta{Name: "sample-db", Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLDBSpec{DBName: "sample_db", MysqlName: MySQLName},
			}
			Expect(k8sClient.Create(ctx, mysqlDB)).Should(Succeed())
			Eventually(func() string {
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: "sample-db"}, mysqlDB)
				if err != nil {
					return ""
				}
				return mysqlDB.Status.Phase
			}).Should(Equal(mysqlDBPhaseReady))
		})

		AfterEach(func() {
			cleanUpMySQLDB(ctx, k8sClient, Namespace)
			cleanUpMySQL(ctx, k8sClient, Namespace)
			stopFunc()
			err := close()
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})
	})
})
