package e2e

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

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

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail) // Use Gomega with Ginkgo
	RunSpecs(t, "e2e suite")  // tells Ginkgo to start the test suite.
}

var _ = BeforeSuite(func() {
	fmt.Println("Setup kind cluster and mysql-operator")
	// 1. TODO: Check if docker is running.
	// 2. TODO: Check if kind is avaialble -> install kind if not available.

	ctx := context.Background()
	kind = newKind(
		ctx,
		kindName,
		kubeconfigPath,
		false,
	)
	// 3. Start up kind cluster.
	prepareKind(kind)

	// 4. set k8sclient
	setUpK8sClient()
	deleteMySQLUserIfExist(ctx)
	deleteMySQLIfExist(ctx)

	// 5. TODO: Check if skaffold is available -> intall skaffold if not available.

	// 6. Set up skaffold
	skaffold = &Skaffold{KubeconfigPath: kubeconfigPath}

	// 7. Deploy CRDs and controllers with skaffold.
	skaffold.run()

	// 8. Check if mysql-operator is running.
	checkMySQLOperator() // check if mysql-operator is running

	fmt.Println("Setup completed")
}, 60)

var _ = AfterSuite(func() {
	fmt.Println("Clean up mysql-operator and kind cluster")
	// 1. Remove the deployed resources
	skaffold.delete()

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
	err = mysqlv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
}

func prepareKind(kind *Kind) {
	// check kind version
	err := kind.checkVersion()
	if err != nil {
		log.Fatal(err)
	}

	isDeleted, err := kind.deleteCluster()
	if err != nil {
		log.Fatal(err)
	} else if isDeleted {
		fmt.Println("kind deleted cluster")
	}

	// create cluster
	isCreated, err := kind.createCluster()
	if err != nil {
		log.Fatal(fmt.Printf("failed to create kind cluster. error: %s\n", err))
	} else if isCreated {
		fmt.Printf("kind created '%s'\n", kindName)
	}
}

func cleanUpKind(kind *Kind) {
	isDeleted, err := kind.deleteCluster()
	if err != nil {
		log.Fatal(err)
	} else if isDeleted {
		fmt.Printf("kind deleted '%s'\n", kindName)
	}
}

func checkMySQLOperator() {
	deployment := &appsv1.Deployment{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: mysqlOperatorNamespace, Name: mysqlOperatorDeploymentName}, deployment)
	if err != nil {
		log.Fatal(fmt.Printf("failed to get %s", mysqlOperatorDeploymentName))
	}
	if deployment.Status.AvailableReplicas != *deployment.Spec.Replicas {
		log.Fatal(fmt.Printf("%s doesn't have the required replicss", mysqlOperatorDeploymentName))
	}
}
