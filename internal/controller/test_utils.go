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
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// typedRateLimitingQueue wraps a workqueue.RateLimitingInterface to implement workqueue.TypedRateLimitingInterface
type typedRateLimitingQueue struct {
	workqueue.RateLimitingInterface
}

func (q *typedRateLimitingQueue) AddRateLimited(item reconcile.Request) {
	q.RateLimitingInterface.AddRateLimited(item)
}

func (q *typedRateLimitingQueue) Get() (reconcile.Request, bool) {
	item, shutdown := q.RateLimitingInterface.Get()
	if item == nil {
		return reconcile.Request{}, shutdown
	}
	return item.(reconcile.Request), shutdown
}

func (q *typedRateLimitingQueue) Done(item reconcile.Request) {
	q.RateLimitingInterface.Done(item)
}

func (q *typedRateLimitingQueue) Forget(item reconcile.Request) {
	q.RateLimitingInterface.Forget(item)
}

func (q *typedRateLimitingQueue) Add(item reconcile.Request) {
	q.RateLimitingInterface.Add(item)
}

func (q *typedRateLimitingQueue) AddAfter(item reconcile.Request, duration time.Duration) {
	q.RateLimitingInterface.AddAfter(item, duration)
}

func (q *typedRateLimitingQueue) NumRequeues(item reconcile.Request) int {
	return q.RateLimitingInterface.NumRequeues(item)
}

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

func addOwnerReferenceToMySQL(mysqlUser *mysqlv1alpha1.MySQLUser, mysql *mysqlv1alpha1.MySQL) *mysqlv1alpha1.MySQLUser {
	mysqlUser.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         "mysql.nakamasato.com/v1alpha1",
			Kind:               "MySQL",
			Name:               mysql.Name,
			UID:                mysql.UID,
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
		},
	}
	return mysqlUser
}

func StartDebugTool(ctx context.Context, cfg *rest.Config, scheme *runtime.Scheme) {
	log := log.FromContext(ctx).WithName("DebugTool")
	fmt.Println("startDebugTool")
	// Set a mapper
	mapper, err := func(c *rest.Config) (meta.RESTMapper, error) {
		return apiutil.NewDynamicRESTMapper(c, nil)
	}(cfg)
	if err != nil {
		log.Error(err, "failed to create mapper")
	}

	// Create a cache
	cache, err := cache.New(cfg, cache.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		log.Error(err, "failed to create cache")
	}

	// Start Cache
	go func() {
		if err := cache.Start(ctx); err != nil { // func (m *InformersMap) Start(ctx context.Context) error {
			log.Error(err, "failed to start cache")
		}
	}()

	// create workqueue
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[interface{}]()
	baseQueue := workqueue.NewNamedRateLimitingQueue(rateLimiter, "test")
	queue := &typedRateLimitingQueue{baseQueue}

	// create eventhandler
	mysqlUserEventHandler := handler.TypedFuncs[*mysqlv1alpha1.MySQLUser, reconcile.Request]{
		CreateFunc: func(ctx context.Context, e event.TypedCreateEvent[*mysqlv1alpha1.MySQLUser], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[MySQLUser][Created]", "Name", e.Object.GetName())
			q.Add(reconcile.Request{})
		},
		UpdateFunc: func(ctx context.Context, e event.TypedUpdateEvent[*mysqlv1alpha1.MySQLUser], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[MySQLUser][Updated]", "Name", e.ObjectNew.GetName())
			q.Add(reconcile.Request{})
		},
		DeleteFunc: func(ctx context.Context, e event.TypedDeleteEvent[*mysqlv1alpha1.MySQLUser], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[MySQLUser][Deleted]", "Name", e.Object.GetName())
			q.Add(reconcile.Request{})
		},
	}
	mysqlEventHandler := handler.TypedFuncs[*mysqlv1alpha1.MySQL, reconcile.Request]{
		CreateFunc: func(ctx context.Context, e event.TypedCreateEvent[*mysqlv1alpha1.MySQL], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[MySQL][Created]", "Name", e.Object.GetName())
			q.Add(reconcile.Request{})
		},
		UpdateFunc: func(ctx context.Context, e event.TypedUpdateEvent[*mysqlv1alpha1.MySQL], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[MySQL][Updated]", "Name", e.ObjectNew.GetName())
			q.Add(reconcile.Request{})
		},
		DeleteFunc: func(ctx context.Context, e event.TypedDeleteEvent[*mysqlv1alpha1.MySQL], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[MySQL][Deleted]", "Name", e.Object.GetName())
			q.Add(reconcile.Request{})
		},
	}
	secretEventHandler := handler.TypedFuncs[*v1.Secret, reconcile.Request]{
		CreateFunc: func(ctx context.Context, e event.TypedCreateEvent[*v1.Secret], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[Secret][Created]", "Name", e.Object.GetName())
			q.Add(reconcile.Request{})
		},
		UpdateFunc: func(ctx context.Context, e event.TypedUpdateEvent[*v1.Secret], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[Secret][Updated]", "Name", e.ObjectNew.GetName())
			q.Add(reconcile.Request{})
		},
		DeleteFunc: func(ctx context.Context, e event.TypedDeleteEvent[*v1.Secret], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			log.Info("[Secret][Deleted]", "Name", e.Object.GetName())
			q.Add(reconcile.Request{})
		},
	}

	// create source with event handlers
	kindWithCacheMysqlUser := source.Kind(cache, &mysqlv1alpha1.MySQLUser{}, mysqlUserEventHandler)
	kindWithCacheMysql := source.Kind(cache, &mysqlv1alpha1.MySQL{}, mysqlEventHandler)
	kindWithCachesecret := source.Kind(cache, &v1.Secret{}, secretEventHandler)

	// start kind
	fmt.Println("cache starting")
	if err := kindWithCacheMysqlUser.Start(ctx, queue); err != nil {
		log.Error(err, "failed to start kindWithCacheMysqlUser")
	}
	if err := kindWithCacheMysql.Start(ctx, queue); err != nil {
		log.Error(err, "failed to start kindWithCacheMysql")
	}
	if err := kindWithCachesecret.Start(ctx, queue); err != nil {
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
