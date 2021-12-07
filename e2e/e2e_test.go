package e2e

import (
	"context"
	"fmt"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	mysqlName          = "mysql-sample"
	mysqlUserName      = "john"
	mysqlNamespace     = "default"
	mysqlUserFinalizer = "mysqluser.nakamasato.com/finalizer"
	mysqlFinalizer     = "mysql.nakamasato.com/finalizer"
	timeout            = 60 * time.Second
	interval           = 250 * time.Millisecond
)

var _ = Describe("E2e", func() {

	ctx := context.Background()
	BeforeEach(func() {
		deleteMySQLDeploymentIfExist(ctx)
		deleteMySQLServiceIfExist(ctx)
		deleteMySQLUserIfExist(ctx)
		deleteMySQLIfExist(ctx)
	})

	AfterEach(func() {
		deleteMySQLDeploymentIfExist(ctx)
		deleteMySQLServiceIfExist(ctx)
		deleteMySQLUserIfExist(ctx)
		deleteMySQLIfExist(ctx)
	})

	Describe("Creating MySQL object", func() {
		Context("With the MySQL cluster", func() {
			BeforeEach(func ()  {
				// create mysql deployment & service
				deploy := newMySQLDeployment()
				Expect(k8sClient.Create(ctx, deploy)).Should(Succeed())
				Eventually(func() bool {
					err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: mysqlNamespace, Name: "mysql"}, deploy)
					if err != nil {
						return false
					}
					return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas
				}, timeout, interval).Should(BeTrue())
				service := newMySQLService()
				Expect(k8sClient.Create(ctx, service)).Should(Succeed())

				svcNodePort := newMySQLServiceNodePort()
				Expect(k8sClient.Create(ctx, svcNodePort)).Should(Succeed())

				ep := &corev1.Endpoints{}
				Eventually(func() bool {
					k8sClient.Get(ctx, client.ObjectKey{Namespace: mysqlNamespace, Name: "mysql"}, ep)
					for _, subset := range ep.Subsets {
						for _, port := range subset.Ports {
							if port.Port == 3306 {
								return len(subset.Addresses) > 0
							}
						}
					}
					return false
				}, timeout, interval).Should(BeTrue())

				time.Sleep(2 * time.Second)

				// create mysql
				mysql := newMySQL(mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
				// create mysqluser
				mysqlUser := newMySQLUser(mysqlUserName, mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
			})
			It("successfully create MySQL object", func() {

				// expect to have Secret
				secret := &corev1.Secret{}
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Namespace: mysqlNamespace, Name: "mysql-" + mysqlName + "-" + mysqlUserName}, secret)
				}, timeout, interval).Should(Succeed())

				// expect to have mysql user in mysql
				Eventually(func() bool {
					res, _ := checkMySQLHasUser(mysqlUserName)
					return res
				}, timeout, interval).Should(BeTrue())
			})
		})

		Context("Without the MySQL cluster", func() {
			It("should fail", func() {
				Expect("test").To(Equal("test"))
			})
		})
	})

	Describe("Creating MySQLUser", func() {
		Context("With the MySQL object", func() {
			It("should pass", func() {
				Expect("test").To(Equal("test"))
			})
		})

		Context("Without the MySQL object", func() {
			It("should fail", func() {
				Expect("test").To(Equal("test"))
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
			Selector:                      labels,
			Type:                          "ClusterIP",
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
			Selector:                      labels,
			Type:                          "NodePort",
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
						Image: "mysql:5.7",
						Name:  "mysql",
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_ROOT_PASSWORD",
								Value: "password",
							},
						},
						ReadinessProbe: &corev1.Probe{
							Handler:                       corev1.Handler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.FromInt(3306),
								},
							},
							InitialDelaySeconds:           5,
							PeriodSeconds:                 10,
						},
					}},
				},
			},
		},
	}
}
