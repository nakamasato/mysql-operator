package controllers

import (
	"context"
	"fmt"
	"time"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	}, 5*time.Second).Should(Equal(0))
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

func cleanUpMySQLDB(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLDB{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
	mysqlDBList := &mysqlv1alpha1.MySQLDBList{}
	Eventually(func() int {
		err := k8sClient.List(ctx, mysqlDBList, &client.ListOptions{})
		if err != nil {
			return -1
		}
		return len(mysqlDBList.Items)
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
		Spec: mysqlv1alpha1.MySQLUserSpec{MysqlName: mysqlName},
	}
}

func newMySQLDB(apiVersion, namespace, objName, dbName, mysqlName string) *mysqlv1alpha1.MySQLDB {
	return &mysqlv1alpha1.MySQLDB{
		TypeMeta: metav1.TypeMeta{APIVersion: apiVersion, Kind: "MySQLDB"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      objName,
		},
		Spec: mysqlv1alpha1.MySQLDBSpec{MysqlName: mysqlName, DBName: dbName},
	}
}

func StartDebugTool(ctx context.Context, cfg *rest.Config, scheme *runtime.Scheme) {
	log := log.FromContext(ctx).WithName("DebugTool")
	fmt.Println("startDebugTool")
	// Set a mapper
	mapper, err := func(c *rest.Config) (meta.RESTMapper, error) {
		return apiutil.NewDynamicRESTMapper(c)
	}(cfg)
	if err != nil {
		log.Error(err, "failed to create mapper")
	}

	// Create a cache
	cache, err := cache.New(cfg, cache.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		log.Error(err, "failed to create cache")
	}

	secret := &v1.Secret{}
	mysqluser := &mysqlv1alpha1.MySQLUser{}
	mysql := &mysqlv1alpha1.MySQL{}

	// Start Cache
	go func() {
		if err := cache.Start(ctx); err != nil { // func (m *InformersMap) Start(ctx context.Context) error {
			log.Error(err, "failed to start cache")
		}
	}()

	// create source
	kindWithCacheMysqlUser := source.NewKindWithCache(mysqluser, cache)
	kindWithCacheMysql := source.NewKindWithCache(mysql, cache)
	kindWithCachesecret := source.NewKindWithCache(secret, cache)

	// create workqueue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test")

	// create eventhandler
	mysqlUserEventHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			log.Info("[MySQLUser][Created]", "Name", e.Object.GetName())
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			log.Info("[MySQLUser][Updated]", "Name", e.ObjectNew.GetName())
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			log.Info("[MySQLUser][Deleted]", "Name", e.Object.GetName())
		},
	}
	mysqlEventHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			log.Info("[MySQL][Created]", "Name", e.Object.GetName())
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			log.Info("[MySQL][Updated]", "Name", e.ObjectNew.GetName())
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			log.Info("[MySQL][Deleted]", "Name", e.Object.GetName())
		},
	}
	secretEventHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			log.Info("[Secret][Created]", "Name", e.Object.GetName())
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			log.Info("[Secret][Updated]", "Name", e.ObjectNew.GetName())
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			log.Info("[Secret][Deleted]", "Name", e.Object.GetName())
		},
	}

	// start kind
	fmt.Println("cache starting")
	if err := kindWithCacheMysqlUser.Start(ctx, mysqlUserEventHandler, queue); err != nil {
		log.Error(err, "failed to start kindWithCacheMysqlUser")
	}
	if err := kindWithCacheMysql.Start(ctx, mysqlEventHandler, queue); err != nil {
		log.Error(err, "failed to start kindWithCacheMysql")
	}
	if err := kindWithCachesecret.Start(ctx, secretEventHandler, queue); err != nil {
		log.Error(err, "failed to start kindWithCachesecret")
	}

	// wait cache to be synced
	fmt.Println("waiting for cache to be synced")
	if err := kindWithCacheMysqlUser.WaitForSync(ctx); err != nil {
		log.Error(err, "failed to wait cache for kindWithCacheMysqlUser")
	}
	if err := kindWithCacheMysql.WaitForSync(ctx); err != nil {
		log.Error(err, "failed to wait cache for kindWithCacheMysql")
	}
	if err := kindWithCachesecret.WaitForSync(ctx); err != nil {
		log.Error(err, "failed to wait cache for kindWithCachesecret")
	}
	fmt.Println("cache is synced")

	// wait until canceled
	go func() {
		<-ctx.Done()
	}()
}
