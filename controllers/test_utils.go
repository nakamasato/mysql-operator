package controllers

import (
	"context"
	"fmt"
	"log"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func cleanUpMySQL(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQL{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
	mysqlList := &mysqlv1alpha1.MySQLList{}
	Eventually(func() int {
		err := k8sClient.List(ctx, mysqlList, &client.ListOptions{})
		if err != nil {
			return -1
		}
		return len(mysqlList.Items)
	}).Should(Equal(0))
}

func cleanUpMySQLUser(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLUser{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
	mysqlUserList := &mysqlv1alpha1.MySQLUserList{}
	Eventually(func() int {
		err := k8sClient.List(ctx, mysqlUserList, &client.ListOptions{})
		if err != nil {
			return -1
		}
		return len(mysqlUserList.Items)
	}).Should(Equal(0))
}

func cleanUpSecret(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &v1.Secret{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
}

func newMySQLUser(apiVersion, namespace, name, mysqlName string) *mysqlv1alpha1.MySQLUser {
	return &mysqlv1alpha1.MySQLUser{
		TypeMeta: metav1.TypeMeta{APIVersion: apiVersion, Kind: "MySQLUser"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec:   mysqlv1alpha1.MySQLUserSpec{MysqlName: mysqlName},
		Status: mysqlv1alpha1.MySQLUserStatus{},
	}
}

func addOwnerReferenceToMySQL(mysqlUser *mysqlv1alpha1.MySQLUser, mysql *mysqlv1alpha1.MySQL) *mysqlv1alpha1.MySQLUser {
	mysqlUser.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         "mysql.nakamasato.com/v1alpha1",
			Kind:               "MySQL",
			Name:               mysql.Name,
			UID:                mysql.UID,
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	}
	return mysqlUser
}

func StartDebugTool(ctx context.Context, cfg *rest.Config, scheme *runtime.Scheme) {
	fmt.Println("startDebugTool")
	// Set a mapper
	mapper, err := func(c *rest.Config) (meta.RESTMapper, error) {
		return apiutil.NewDynamicRESTMapper(c)
	}(cfg)
	if err != nil {
		log.Fatal("failed to create mapper")
	}

	// Create a cache
	cache, err := cache.New(cfg, cache.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		log.Fatal("failed to create cache")
	}

	secret := &v1.Secret{}
	mysqluser := &mysqlv1alpha1.MySQLUser{}

	// Start Cache
	go func() {
		if err := cache.Start(ctx); err != nil { // func (m *InformersMap) Start(ctx context.Context) error {
			log.Fatal("failed to start cache")
		}
	}()

	// create source
	kindWithCacheMysqlUser := source.NewKindWithCache(mysqluser, cache)
	kindWithCachesecret := source.NewKindWithCache(secret, cache)

	// create workqueue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test")

	// create eventhandler
	mysqlUserEventHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			fmt.Printf("[MySQLUser][Created] %s\n", e.Object.GetName())
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			fmt.Printf("[MySQLUser][Updated] %s\n", e.ObjectNew.GetName())
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			fmt.Printf("[MySQLUser][Deleted] %s\n", e.Object.GetName())
		},
	}
	secretEventHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			fmt.Printf("[Secret][Created] %s\n", e.Object.GetName())
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			fmt.Printf("[Secret][Updated] %s\n", e.ObjectNew.GetName())
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			fmt.Printf("[Secret][Deleted] %s\n", e.Object.GetName())
		},
	}

	// start kind
	fmt.Println("cache starting")
	if err := kindWithCacheMysqlUser.Start(ctx, mysqlUserEventHandler, queue); err != nil {
		log.Fatal("failed to start kind")
	}
	if err := kindWithCachesecret.Start(ctx, secretEventHandler, queue); err != nil {
		log.Fatal("failed to start kind")
	}

	// wait cache to be synced
	fmt.Println("waiting for cache to be synced")
	if err := kindWithCacheMysqlUser.WaitForSync(ctx); err != nil {
		log.Fatal("failed to wait cache")
	}
	if err := kindWithCachesecret.WaitForSync(ctx); err != nil {
		log.Fatal("failed to wait cache")
	}
	fmt.Println("cache is synced")

	// wait until canceled
	go func() {
		<-ctx.Done()
	}()
}
