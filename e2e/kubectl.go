package e2e

import (
	"context"
	"log"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubectl struct {
	clientset *kubernetes.Clientset
}

func newKubectl(kubeconfigPath string) *Kubectl {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return &Kubectl{clientset: clientset}
}

func (k *Kubectl) GetDeployment(namespace, name string) (*v1.Deployment, error) {
	v1AppsV1Interface := k.clientset.AppsV1()
	v1DeploymentInterface := v1AppsV1Interface.Deployments(namespace)
	deployment, err := v1DeploymentInterface.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
