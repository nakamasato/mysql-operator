package e2e

import (
	"context"
	"fmt"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	mysqlName          = "mysql-sample"
	mysqlUserName      = "john"
	mysqlNamespace     = "default"
	mysqlUserFinalizer = "mysqluser.nakamasato.com/finalizer"
	mysqlFinalizer     = "mysql.nakamasato.com/finalizer"
	timeout            = 20 * time.Second
	interval           = 1 * time.Second
	secretName         = "mysql-" + mysqlName + "-" + mysqlUserName
)

var _ = Describe("E2e", func() {

	ctx := context.Background()
	BeforeEach(func() {
		// deleteMySQLDeploymentIfExist(ctx)
		// deleteMySQLServiceIfExist(ctx)
		deleteMySQLUserIfExist(ctx)
		deleteMySQLIfExist(ctx)
		deleteUserInMysql(mysqlUserName)
	})

	AfterEach(func() {
		// deleteMySQLDeploymentIfExist(ctx)
		// deleteMySQLServiceIfExist(ctx)
		deleteMySQLUserIfExist(ctx)
		deleteMySQLIfExist(ctx)
		deleteUserInMysql(mysqlUserName)
	})

	Describe("Creating and deleting MySQL/MySQLUser object", func() {
		Context("With the MySQL cluster", func() {
			BeforeEach(func() {
				// create mysql deployment & service
				createMySQLDeploymentAndService(ctx)

				// create mysql
				mysql := newMySQL(mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
				// create mysqluser
				mysqlUser := newMySQLUser(mysqlUserName, mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
			})
			It("Successfully create MySQL user and Secret", func() {

				// expect to have Secret
				secret := &corev1.Secret{}
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Namespace: mysqlNamespace, Name: secretName}, secret)
				}, timeout, interval).Should(Succeed())

				// expect to have mysql user in mysql
				Eventually(func() bool {
					res, _ := checkMySQLHasUser(mysqlUserName)
					return res
				}, timeout, interval).Should(BeTrue())
			})

			It("Successfully delete MySQL user and Secret", func() {
				By("Delete MySQLUser")
				mysqlUser, err := getMySQLUser(mysqlUserName, mysqlNamespace)
				if err != nil {
					return
				}
				Expect(k8sClient.Delete(ctx, mysqlUser)).Should(Succeed())

				// expect to delete Secret
				secret := &corev1.Secret{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: mysqlNamespace, Name: secretName}, secret)
					return errors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())

				// expect to delete mysql user in mysql
				Eventually(func() bool {
					res, _ := checkMySQLHasUser(mysqlUserName)
					return res
				}, timeout, interval).Should(BeFalse())
			})
		})

		Context("Without the MySQL cluster", func() {
			It("Fail to create MySQL", func() {
				// create mysql
				// mysql := newMySQL(mysqlName, mysqlNamespace)
				// Expect(k8sClient.Create(ctx, mysql)).Should(Fail())
			})
			It("Create MySQL and MySQLUser after MySQL cluster gets available", func() {
				// delete mysql
				deleteMySQLDeploymentIfExist(ctx)
				deleteMySQLServiceIfExist(ctx)
				// create mysql
				mysql := newMySQL(mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
				// create mysqluser
				mysqlUser := newMySQLUser(mysqlUserName, mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
				// expect not to have Secret
				secret := &corev1.Secret{}
				Consistently(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: mysqlNamespace, Name: secretName}, secret)
					return errors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())

				By("Create MySQL Deployment and Service")
				// create mysql deployment & service
				createMySQLDeploymentAndService(ctx)

				// expect to have Secret
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Namespace: mysqlNamespace, Name: secretName}, secret)
				}, timeout, interval).Should(Succeed())

				// expect to have mysql user in mysql
				Eventually(func() bool {
					res, _ := checkMySQLHasUser(mysqlUserName)
					return res
				}, timeout, interval).Should(BeTrue())
			})
		})
	})
})

func checkMySQLHasUser(mysqluser string) (bool, error) {
	db, err := sql.Open("mysql", "root:password@tcp(localhost:30306)/") // TODO: Make MySQL root user credentials configurable
	if err != nil {
		return false, err
	}
	defer db.Close()
	row := db.QueryRow("SELECT COUNT(*) FROM mysql.user where User = '" + mysqluser + "'")
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	} else {
		fmt.Printf("mysql.user count: %s, %d\n", mysqluser, count)
		return count > 0, nil
	}
}

func deleteUserInMysql(mysqluser string) {
	db, err := sql.Open("mysql", "root:password@tcp(localhost:30306)/") // TODO: Make MySQL root user credentials configurable
	if err != nil {
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM mysql.user where User = '" + mysqluser + "'")

	if err != nil {
		fmt.Printf("failed delete mysql.user %v\n", err)
	} else {
		fmt.Printf("successfully deleted mysql.user: %s\n", mysqluser)
	}
}

func deleteMySQLServiceIfExist(ctx context.Context) {
	svcNames := []string{"mysql", "mysql-nodeport"}
	for _, svcName := range svcNames {
		object, err := getService(svcName, mysqlNamespace)
		if err != nil {
			return
		}
		Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
	}
}

func deleteMySQLDeploymentIfExist(ctx context.Context) {
	object, err := getDeployment("mysql", mysqlNamespace)
	if err != nil {
		return
	}
	Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
}

func deleteMySQLIfExist(ctx context.Context) {
	object, err := getMySQL(mysqlName, mysqlNamespace) // TODO: enable to pass mysqlName and mysqlNamespace
	if err != nil {
		return
	}

	// remove finalizers
	if controllerutil.ContainsFinalizer(object, mysqlFinalizer) {
		controllerutil.RemoveFinalizer(object, mysqlFinalizer)
		err := k8sClient.Update(ctx, object)
		if err != nil {
			return
		}
	}

	// delete object if exist
	object, err = getMySQL(mysqlName, mysqlNamespace) // TODO: enable to pass mysqlName and mysqlNamespace
	if err != nil {
		return
	}
	Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
}

func deleteMySQLUserIfExist(ctx context.Context) {
	object, err := getMySQLUser(mysqlUserName, mysqlNamespace) // TODO: enable to pass mysqlUserName and mysqlNamespace
	if err != nil {
		return
	}

	// remove finalizers
	if controllerutil.ContainsFinalizer(object, mysqlUserFinalizer) {
		controllerutil.RemoveFinalizer(object, mysqlUserFinalizer)
		err := k8sClient.Update(ctx, object)
		if err != nil {
			return
		}
	}

	// delete object if exist
	object, err = getMySQLUser(mysqlUserName, mysqlNamespace) // TODO: enable to pass mysqlUserName and mysqlNamespace
	if err != nil {
		return
	}
	Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
}

func getDeployment(name, namespace string) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, deploy)
	if err != nil {
		return nil, err
	}
	return deploy, nil
}

func getService(name, namespace string) (*corev1.Service, error) {
	service := &corev1.Service{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func getMySQL(name, namespace string) (*mysqlv1alpha1.MySQL, error) {
	object := &mysqlv1alpha1.MySQL{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func getMySQLUser(name, namespace string) (*mysqlv1alpha1.MySQLUser, error) {
	object := &mysqlv1alpha1.MySQLUser{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func newMySQL(name, namespace string) *mysqlv1alpha1.MySQL {
	return &mysqlv1alpha1.MySQL{
		TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQL"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       mysqlv1alpha1.MySQLSpec{Host: "mysql.default", AdminUser: "root", AdminPassword: "password"},
	}
}

func newMySQLUser(name, mysqlName, namespace string) *mysqlv1alpha1.MySQLUser {
	return &mysqlv1alpha1.MySQLUser{
		TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQL"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: mysqlv1alpha1.MySQLUserSpec{
			MysqlName: mysqlName,
		},
	}
}

func newMySQLService() *corev1.Service {
	labels := map[string]string{
		"app": "mysql",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: mysqlNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "tcp",
					Protocol: "TCP",
					Port:     3306,
				},
			},
			Selector: labels,
			Type:     "ClusterIP",
		},
	}
}

func newMySQLServiceNodePort() *corev1.Service {
	labels := map[string]string{
		"app": "mysql",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-nodeport",
			Namespace: mysqlNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "tcp",
					Protocol: "TCP",
					Port:     3306,
					NodePort: 30306,
				},
			},
			Selector: labels,
			Type:     "NodePort",
		},
	}
}

func newMySQLDeployment() *appsv1.Deployment {
	labels := map[string]string{
		"app": "mysql",
	}
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: mysqlNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "mysql:8",
						Name:  "mysql",
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_ROOT_PASSWORD",
								Value: "password",
							},
						},
					}},
				},
			},
		},
	}
}

func createMySQLDeploymentAndService(ctx context.Context) {
	_, err := getDeployment("mysql", mysqlNamespace)
	if err != nil { // try to create for any error
		deploy := newMySQLDeployment()
		Expect(k8sClient.Create(ctx, deploy)).Should(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: mysqlNamespace, Name: "mysql"}, deploy)
			if err != nil {
				return false
			}
			return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas
		}, timeout, interval).Should(BeTrue())
	}

	_, err = getService("mysql", mysqlNamespace)
	if err != nil {
		service := newMySQLService()
		Expect(k8sClient.Create(ctx, service)).Should(Succeed())

		svcNodePort := newMySQLServiceNodePort()
		Expect(k8sClient.Create(ctx, svcNodePort)).Should(Succeed())
	}

	// wait until deployment is ready
	deployment := &appsv1.Deployment{}
	Eventually(func() bool {
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: "mysql", Namespace: mysqlNamespace}, deployment)
		return deployment.Status.AvailableReplicas == *deployment.Spec.Replicas
	}, timeout, interval).Should(BeTrue())
}
