package e2e

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	kindName                    = "mysql-operator-e2e"
	kubeconfigPath              = "kubeconfig"
	mysqlOperatorNamespace      = "mysql-operator-system"
	mysqlOperatorDeploymentName = "mysql-operator-controller-manager"
)

var skaffold *Skaffold
var kind *Kind
var k8sClient client.Client
var cancel context.CancelFunc

var (
	log    = logf.Log.WithName("mysql-operator-e2e")
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(mysqlv1alpha1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func TestE2e(t *testing.T) {
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	RegisterFailHandler(Fail) // Use Gomega with Ginkgo
	RunSpecs(t, "e2e suite")  // tells Ginkgo to start the test suite.
}

var _ = BeforeSuite(func() {
	log.Info("Setup kind cluster and mysql-operator")
	// 1. TODO: Check if docker is running.
	// 2. TODO: Check if kind is avaialble -> install kind if not available.

	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)
	kind = newKind(
		ctx,
		kindName,
		kubeconfigPath,
		true,
	)
	// 3. Start up kind cluster.
	prepareKind(kind)

	// 4. set k8sclient
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	os.Setenv("KUBECONFIG", path.Join(mydir, kubeconfigPath))
	cfg, err := config.GetConfigWithContext("kind-" + kindName)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	err = mysqlv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	deleteMySQLUserIfExist(ctx)
	deleteMySQLIfExist(ctx)

	// 5. TODO: Check if skaffold is available -> intall skaffold if not available.

	// 6. Set up skaffold
	skaffold = &Skaffold{KubeconfigPath: kubeconfigPath}

	// 7. Deploy CRDs and controllers with skaffold.
	err = skaffold.run(ctx) // To check log during running tests
	Expect(err).To(BeNil())

	// 8. Check if mysql-operator is running.
	checkMySQLOperator() // check if mysql-operator is running

	// 9. Start debug tool
	startDebugTool(ctx, cfg, scheme)
	fmt.Println("Setup completed")
}, 120)

var _ = AfterSuite(func() {
	fmt.Println("Clean up mysql-operator and kind cluster")
	cancel()
	// 1. Remove the deployed resources
	skaffold.cleanup()

	// 2. Stop kind cluster
	cleanUpKind(kind)
})

func setUpK8sClient() {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	os.Setenv("KUBECONFIG", path.Join(mydir, kubeconfigPath))
	cfg, err := config.GetConfigWithContext("kind-" + kindName)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	err = mysqlv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
}

func prepareKind(kind *Kind) {
	// check kind version
	err := kind.checkVersion()
	if err != nil {
		log.Error(err, "failed to check version")
	}

	isDeleted, err := kind.deleteCluster()
	if err != nil {
		log.Error(err, "failed to delete cluster")
	} else if isDeleted {
		fmt.Println("kind deleted cluster")
	}

	// create cluster
	isCreated, err := kind.createCluster()
	if err != nil {
		log.Error(err, "failed to create kind cluster.")
	} else if isCreated {
		fmt.Printf("kind created '%s'\n", kindName)
	}
}

func cleanUpKind(kind *Kind) {
	isDeleted, err := kind.deleteCluster()
	if err != nil {
		log.Error(err, "failed to clean up cluster")
	} else if isDeleted {
		fmt.Printf("kind deleted '%s'\n", kindName)
	}
}

func checkMySQLOperator() {
	deployment := &appsv1.Deployment{}
	Eventually(func() error {
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: mysqlOperatorNamespace, Name: mysqlOperatorDeploymentName}, deployment)
		log.Info("waiting until mysqlOperator Deployment is deployed")
		return err
	}, timeout, interval).Should(BeNil())

	Eventually(func() bool {
		k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: mysqlOperatorNamespace, Name: mysqlOperatorDeploymentName}, deployment)
		log.Info("waiting until mysqlOperator Pods get ready", "Replicas", *deployment.Spec.Replicas, "AvailableReplicas", deployment.Status.AvailableReplicas)
		return deployment.Status.AvailableReplicas == *deployment.Spec.Replicas
	}, timeout, interval).Should(BeTrue())
}

func startDebugTool(ctx context.Context, cfg *rest.Config, scheme *runtime.Scheme) {
	fmt.Println("startDebugTool")
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		log.Error(err, "failed to create mapper")
	}

	cache, err := cache.New(cfg, cache.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		log.Error(err, "failed to create cache")
	}
	mysqluser := &mysqlv1alpha1.MySQLUser{}
	cache.Get(ctx, client.ObjectKeyFromObject(mysqluser), mysqluser)
	// Start Cache
	go func() {
		if err := cache.Start(ctx); err != nil { // func (m *InformersMap) Start(ctx context.Context) error {
			log.Error(err, "failed to start cache")
		}
	}()
	log.Info("cache is started")

	kindWithCacheMysqlUser := source.NewKindWithCache(mysqluser, cache)
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test")
	eventHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			log.Info("CreateFunc is called", "Name", e.Object.GetName(), "finalizers", e.Object.GetFinalizers())
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			log.Info("UpdateFunc is called", "Name", e.ObjectNew.GetName(), "finalizers", e.ObjectNew.GetFinalizers())
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			log.Info("DeleteFunc is called", "Name", e.Object.GetName(), "finalizers", e.Object.GetFinalizers())
		},
	}
	log.Info("kindWithCacheMysqlUser starting")
	if err := kindWithCacheMysqlUser.Start(ctx, eventHandler, queue); err != nil {
		log.Error(err, "failed to start kind")
	}
	log.Info("waiting for kindWithCacheMysqlUser to be synced")
	if err := kindWithCacheMysqlUser.WaitForSync(ctx); err != nil {
		log.Error(err, "failed to wait cache")
	}
	log.Info("kindWithCacheMysqlUser is synced")
	go func() {
		// run until ctx is done.
		<-ctx.Done()
	}()
}
