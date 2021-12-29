package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	Context("With available MySQL", func() {
		BeforeEach(func() {
			// Delete MySQLUser
			err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLUser{}, client.InNamespace(Namespace))
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: MySQLUserName, Namespace: Namespace}, mysqlUser)
			}).ShouldNot(Succeed())

			// Delete MySQL
			err = k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQL{}, client.InNamespace(Namespace))
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mysql)
			}).ShouldNot(Succeed())

			// Create MySQL
			mysql = &mysqlv1alpha1.MySQL{
				TypeMeta:   metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQL"},
				ObjectMeta: metav1.ObjectMeta{Name: MySQLName, Namespace: Namespace},
				Spec:       mysqlv1alpha1.MySQLSpec{Host: "localhost", AdminUser: "root", AdminPassword: "password"},
			}
			Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
		})
		AfterEach(func() {
			// Delete MySQLUser
			err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLUser{}, client.InNamespace(Namespace))
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: MySQLUserName, Namespace: Namespace}, mysqlUser)
			}).ShouldNot(Succeed())

			// Delete MySQL
			err = k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQL{}, client.InNamespace(Namespace))
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mysql)
			}).ShouldNot(Succeed())
		})
		It("Should have status.UserCount=0", func() {
			mySQL := &mysqlv1alpha1.MySQL{}
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mySQL)
				if err != nil {
					return -1
				}
				return mySQL.Status.UserCount
			}, timeout, interval).Should(Equal(int32(0)))
		})

		It("Should increase status.UserCount by one", func() {
			By("By creating a new MySQLUser")
			mysqlUser = &mysqlv1alpha1.MySQLUser{
				TypeMeta: metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: Namespace,
					Name:      MySQLUserName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         APIVersion,
							Kind:               "MySQL",
							Name:               mysql.Name,
							UID:                mysql.UID,
							BlockOwnerDeletion: pointer.BoolPtr(true),
							Controller:         pointer.BoolPtr(true),
						},
					},
				},
				Spec:   mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
				Status: mysqlv1alpha1.MySQLUserStatus{},
			}
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mysql)
				if err != nil {
					return -1
				}
				return mysql.Status.UserCount
			}, timeout, interval).Should(Equal(int32(1)))
		})

		It("Should decrease status.UserCount to zero", func() {
			By("By creating a new MySQLUser")
			mysqlUser = &mysqlv1alpha1.MySQLUser{
				TypeMeta: metav1.TypeMeta{APIVersion: APIVersion, Kind: "MySQLUser"},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: Namespace,
					Name:      MySQLUserName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         APIVersion,
							Kind:               "MySQL",
							Name:               mysql.Name,
							UID:                mysql.UID,
							BlockOwnerDeletion: pointer.BoolPtr(true),
							Controller:         pointer.BoolPtr(true),
						},
					},
				},
				Spec:   mysqlv1alpha1.MySQLUserSpec{MysqlName: MySQLName},
				Status: mysqlv1alpha1.MySQLUserStatus{},
			}
			Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())

			By("By deleting the MySQLUser")
			err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQL{}, client.InNamespace(Namespace))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: MySQLName, Namespace: Namespace}, mysql)
				if err != nil {
					return -1
				}
				return mysql.Status.UserCount
			}, timeout, interval).Should(Equal(int32(0)))
		})
	})
})
