package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	. "github.com/nakamasato/mysql-operator/internal/mysql"
)

var _ = Describe("MySQLUser controller", func() {

	Context("With available MySQL", func() {

		ctx := context.Background()
		var stopFunc func()
		var close func() error
		BeforeEach(func() {
			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme,
			})
			Expect(err).ToNot(HaveOccurred())
			db, err := sql.Open("testdbdriver", "test")
			close = db.Close
			Expect(err).ToNot(HaveOccurred())
			err = (&MySQLUserReconciler{
				Client:       k8sManager.GetClient(),
				Scheme:       k8sManager.GetScheme(),
				MySQLClients: MySQLClients{fmt.Sprintf("%s-%s", Namespace, MySQLName): db},
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
			err := close()
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})

		When("Creating a MySQLUser", func() {
			AfterEach(func() {
				// Delete MySQLUser
				cleanUpMySQLUser(ctx, k8sClient, Namespace)
				// Delete MySQL
				cleanUpMySQL(ctx, k8sClient, Namespace)
			})
			It("Should create Secret and update mysqluser's status", func() {
				By("By creating a new MySQL")
				mysql = &mysqlv1alpha1.MySQL{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
					ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
					Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
				}
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

				By("By creating a new MySQLUser")
				mysqlUser = &mysqlv1alpha1.MySQLUser{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
					ObjectMeta: metav1.ObjectMeta{Namespace: Namespace, Name: MySQLUserName},
					Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
					Status:     mysqlv1alpha1.MySQLUserStatus{},
				}
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

				// secret should be created
				secret := &v1.Secret{}
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: getSecretName(MySQLName, MySQLUserName)}, secret)
				}).Should(Succeed())

				// status.phase should be ready
				Eventually(func() string {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return ""
					}
					return mysqlUser.Status.Phase
				}).Should(Equal(mysqlUserPhaseReady))

				// status.reason should be 'both secret and mysql user are successfully created.'
				Eventually(func() string {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return ""
					}
					return mysqlUser.Status.Reason
				}).Should(Equal(mysqlUserReasonCompleted))
			})

			It("Should have finalizer", func() {
				By("By creating a new MySQL")
				mysql = &mysqlv1alpha1.MySQL{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
					ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
					Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
				}
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
				By("By creating a new MySQLUser")
				mysqlUser = &mysqlv1alpha1.MySQLUser{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
					ObjectMeta: metav1.ObjectMeta{Namespace: Namespace, Name: MySQLUserName},
					Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
					Status:     mysqlv1alpha1.MySQLUserStatus{},
				}
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(mysqlUser, mysqlUserFinalizer)
				}).Should(BeTrue())
			})
		})

		When("Deleting a MySQLUser", func() {
			BeforeEach(func() {
				// Clean up MySQLUser
				cleanUpMySQLUser(ctx, k8sClient, Namespace)
				// Clean up MySQL
				cleanUpMySQL(ctx, k8sClient, Namespace)
				// Clean up Secret
				cleanUpSecret(ctx, k8sClient, Namespace)

				// Create resources
				mysql = &mysqlv1alpha1.MySQL{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
					ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
					Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
				}
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())

				mysqlUser = &mysqlv1alpha1.MySQLUser{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
					ObjectMeta: metav1.ObjectMeta{Name: MySQLUserName, Namespace: Namespace},
					Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
					Status:     mysqlv1alpha1.MySQLUserStatus{},
				}
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
			})
			AfterEach(func() {
				// Delete MySQL
				Expect(k8sClient.Delete(ctx, mysql)).Should(Succeed())
				// Remove finalizers from MySQL if exists
				if k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLName}, mysql) == nil {
					mysql.Finalizers = []string{}
					Eventually(k8sClient.Update(ctx, mysql)).Should(Succeed())
				}
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLName}, mysql)
				}).ShouldNot(Succeed())
			})
			It("Should delete Secret", func() {

				By("By deleting a MySQLUser")
				mysqlUser = &mysqlv1alpha1.MySQLUser{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
					ObjectMeta: metav1.ObjectMeta{Namespace: Namespace, Name: MySQLUserName},
					Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
					Status:     mysqlv1alpha1.MySQLUserStatus{},
				}
				Expect(k8sClient.Delete(ctx, mysqlUser)).To(Succeed())

				mysqlUser = &mysqlv1alpha1.MySQLUser{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					return errors.IsNotFound(err) // MySQLUser should not exist
				}).Should(BeTrue())

				secret := &v1.Secret{}
				secretName := getSecretName(MySQLName, MySQLUserName)
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: secretName}, secret)
					return errors.IsNotFound(err) // Secret should not exist
				}).Should(BeTrue())

				// MySQL should remain
				mysql = &mysqlv1alpha1.MySQL{}
				Consistently(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLName}, mysql)
				}).Should(Succeed())
			})

		})
	})

	Context("With MySQL with unnconnectable configuration", func() {
		ctx := context.Background()
		var stopFunc func()
		var close func() error
		BeforeEach(func() {
			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme,
			})
			Expect(err).ToNot(HaveOccurred())
			db, err := sql.Open("mysql", "test_user:password@tcp(nonexistinghost:3306)/")
			Expect(err).NotTo(HaveOccurred())
			close = db.Close
			err = (&MySQLUserReconciler{
				Client:       k8sManager.GetClient(),
				Scheme:       k8sManager.GetScheme(),
				MySQLClients: MySQLClients{fmt.Sprintf("%s-%s", Namespace, MySQLName): db},
			}).SetupWithManager(k8sManager)
			Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithCancel(ctx)
			stopFunc = cancel
			go func() {
				err = k8sManager.Start(ctx)
				Expect(err).ToNot(HaveOccurred())
			}()
			time.Sleep(100 * time.Millisecond)

			By("By creating a new MySQL")
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "nonexistinghost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
		})

		AfterEach(func() {
			// Clean up MySQL
			err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQL{}, client.InNamespace(Namespace))
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLName}, mysql)
			}).ShouldNot(Succeed())

			stopFunc()
			err = close()
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})

		When("Creating MySQLUser", func() {

			BeforeEach(func() {
				// Clean up MySQLUser
				err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLUser{}, client.InNamespace(Namespace))
				Expect(err).NotTo(HaveOccurred())

				// Clean up Secret
				err = k8sClient.DeleteAllOf(ctx, &v1.Secret{}, client.InNamespace(Namespace))
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				// Clean up MySQLUser
				err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLUser{}, client.InNamespace(Namespace))
				Expect(err).NotTo(HaveOccurred())

				// Clean up Secret
				err = k8sClient.DeleteAllOf(ctx, &v1.Secret{}, client.InNamespace(Namespace))
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should not create Secret", func() {
				By("By creating a new MySQLUser")
				mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

				// Secret should not be created
				secret := &v1.Secret{}
				Consistently(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: getSecretName(MySQLName, MySQLUserName)}, secret)
					return errors.IsNotFound(err)
				}).Should(BeTrue())
			})

			It("Should have NotReady status with reason 'failed to connect to mysql'", func() {
				By("By creating a new MySQLUser")
				mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

				// Secret should not be created
				secret := &v1.Secret{}
				Consistently(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: getSecretName(MySQLName, MySQLUserName)}, secret)
					return errors.IsNotFound(err)
				}).Should(BeTrue())

				// Status.Phase should be NotReady
				Eventually(func() string {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return ""
					}
					return mysqlUser.Status.Phase
				}).Should(Equal(mysqlUserPhaseNotReady))

				// Status.Reason should be 'failed to connect to mysql'
				Eventually(func() string {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return ""
					}
					return mysqlUser.Status.Reason
				}).Should(Equal(mysqlUserReasonMySQLFailedToCreateUser))
			})
		})

		When("Creating and deleting MySQLUser", func() {

			BeforeEach(func() {
				cleanUpMySQLUser(ctx, k8sClient, Namespace)
				cleanUpSecret(ctx, k8sClient, Namespace)
			})

			AfterEach(func() {
				cleanUpMySQLUser(ctx, k8sClient, Namespace)
				cleanUpSecret(ctx, k8sClient, Namespace)
			})

			It("Should delete MySQLUser", func() {
				By("By creating a new MySQLUser")
				mysqlUser = newMySQLUser(APIVersion, Namespace, MySQLUserName, MySQLName)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

				// Secret will not be created
				secret := &v1.Secret{}
				Consistently(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: getSecretName(MySQLName, MySQLUserName)}, secret)
					return errors.IsNotFound(err)
				}).Should(BeTrue())

				// Delete MySQLUser
				Expect(k8sClient.Delete(ctx, mysqlUser)).To(Succeed())
				Eventually(func() error {
					return k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
				}).Should(HaveOccurred())
			})
		})

		Context("With no MySQL found", func() {
			BeforeEach(func() {
				// Clean up MySQLUser
				cleanUpMySQLUser(ctx, k8sClient, Namespace)
				// Clean up MySQL
				cleanUpMySQL(ctx, k8sClient, Namespace)
			})
			It("Should have NotReady status with reason 'failed to fetch MySQL'", func() {
				By("By creating a new MySQLUser")
				mysqlUser = &mysqlv1alpha1.MySQLUser{
					TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
					ObjectMeta: metav1.ObjectMeta{Namespace: Namespace, Name: MySQLUserName},
					Spec:       mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
				}
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

				// Secret will not be created
				secret := &v1.Secret{}
				Consistently(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: getSecretName(MySQLName, MySQLUserName)}, secret)
					return errors.IsNotFound(err)
				}).Should(BeTrue())

				// Status.Phase should be NotReady
				Eventually(func() string {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return ""
					}
					return mysqlUser.Status.Phase
				}).Should(Equal(mysqlUserPhaseNotReady))

				// Status.Reason should be 'failed to fetch MySQL'
				Eventually(func() string {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: MySQLUserName}, mysqlUser)
					if err != nil {
						return ""
					}
					return mysqlUser.Status.Reason
				}).Should(Equal(mysqlUserReasonMySQLFetchFailed))
			})
		})
	})
})
