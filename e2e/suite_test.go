package e2e

import (
	"context"
	"fmt"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kindName = "mysql-operator-e2e"
	kubeconfigPath = ".kubeconfig"
	mysqlOperatorNamespace = "mysql-operator-system"
	mysqlOperatorDeploymentName = "mysql-operator-controller-manager"
)

var skaffold *Skaffold
var kind *Kind

func TestE2e(t *testing.T) {
    RegisterFailHandler(Fail) // Use Gomega with Ginkgo
    RunSpecs(t, "e2e suite") // tells Ginkgo to start the test suite.
}

var _ = BeforeSuite(func() {
	fmt.Println("Setup kind cluster and mysql-operator")
	// 1. TODO: Check if docker is running
	// 2. TODO: Check if kind is avaialble -> install kind if not available
	// 3. Start up kind cluster
	// 4. TODO: Check if kubeckl is available -> intall kubectl if not available
	// 5. Configure kubectl
	// 6. Deploy CRDs and controllers

	// prepare kind configuration
	// kindName := "mysql-operator-e2e"
	// kubeConfigPath := "./.kubeconfig"
	// kubernetesVersion := "v1.20.2"

	ctx := context.Background()
	kind = newKind(ctx, kindName, kubeconfigPath)

	// check kind version
	err := kind.checkVersion()
	if err != nil {
		log.Fatal(err)
	}

	// delete cluster: kind delete cluster --name mysql-operator-e2e
	err = kind.deleteCluster()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("kind deleted cluster")

	// create cluster: kind create cluster --name mysql-operator-e2e --kubeconfig ./.kubeconfig --image kindest/node:v1.20.2
	err = kind.createCluster()
	if err != nil {
		log.Fatal(fmt.Printf("failed to create kind cluster. error: %s\n", err))
	}
	fmt.Printf("kind created '%s'\n", kindName)

	// scaffold
	skaffold = &Skaffold{KubeconfigPath: kubeconfigPath}
	skaffold.run()

	// Wait until mysql operator is ready
	checkMySQLOperator()
	fmt.Println("Setup completed")
}, 60)

var _ = AfterSuite(func()  {
	fmt.Println("Clean up mysql-operator and kind cluster")
	// 1. Remove the deployed resources
	// 2. Stop kind cluster
	skaffold.delete()

	err := kind.deleteCluster()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("kind deleted '%s'\n", kindName)
})

func checkMySQLOperator()  {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	deployment, err := clientset.AppsV1().Deployments(mysqlOperatorNamespace).Get(context.TODO(), mysqlOperatorDeploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(fmt.Printf("failed to get %s", mysqlOperatorDeploymentName))
	}
	if deployment.Status.AvailableReplicas != *deployment.Spec.Replicas {
		log.Fatal(fmt.Printf("%s doesn't have the required replicss", mysqlOperatorDeploymentName))
	}
}
