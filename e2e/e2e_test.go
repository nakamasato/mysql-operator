package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
		mysqlName      = "mysql-sample"
		mysqlUserName  = "john"
		mysqlNamespace = "default"
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
			It("successfully create MySQL object", func() {
				// create mysql deployment & service
				deploy := newMySQLDeployment()
				Expect(k8sClient.Create(ctx, deploy)).Should(Succeed())
				service := newMySQLService()
				Expect(k8sClient.Create(ctx, service)).Should(Succeed())

				// create mysql
				mysql := newMySQL(mysqlName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysql)).Should(Succeed())
				// create mysqluser
				mysqlUser := newMySQLUser(mysqlUserName, mysqlNamespace)
				Expect(k8sClient.Create(ctx, mysqlUser)).Should(Succeed())
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

func deleteMySQLServiceIfExist(ctx context.Context) {
	object, err := getService("mysql", mysqlNamespace)
	if err != nil {
		return
	}
	Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
}

func deleteMySQLDeploymentIfExist(ctx context.Context) {
	object, err := getDeployment("mysql", mysqlNamespace)
	if err != nil {
		return
	}
	Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
}

func deleteMySQLIfExist(ctx context.Context) {
	object, err := getMySQL("mysql", mysqlNamespace)
	if err != nil {
		return
	}
	Expect(k8sClient.Delete(ctx, object)).Should(Succeed())
}

func deleteMySQLUserIfExist(ctx context.Context) {
	object, err := getMySQLUser("mysql", mysqlNamespace)
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
		Spec:       mysqlv1alpha1.MySQLSpec{Host: "mysql", AdminUser: "root", AdminPassword: "password"},
	}
}

func newMySQLUser(name, namespace string) *mysqlv1alpha1.MySQLUser {
	return &mysqlv1alpha1.MySQLUser{
		TypeMeta:   metav1.TypeMeta{APIVersion: "mysql.nakamasato.com/v1alphav1", Kind: "MySQL"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       mysqlv1alpha1.MySQLUserSpec{
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
			HealthCheckNodePort:           0,
			PublishNotReadyAddresses:      false,
			SessionAffinityConfig:         &corev1.SessionAffinityConfig{},
			AllocateLoadBalancerNodePorts: new(bool),
			LoadBalancerClass:             new(string),
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
					}},
				},
			},
		},
	}
}
