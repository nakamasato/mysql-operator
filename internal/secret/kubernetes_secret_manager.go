package secret

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sSecretManager struct {
	namespace string
	client    client.Client
}

// Initialize SecretManager with projectId
func Newk8sSecretManager(ctx context.Context, ns string, c client.Client) (*k8sSecretManager, error) {

	return &k8sSecretManager{
		namespace: ns,
		client:    c,
	}, nil
}

// Get latest version from SecretManager
func (s k8sSecretManager) GetSecret(ctx context.Context, name string) (string, error) {
	secret := &corev1.Secret{}
	err := s.client.Get(ctx, client.ObjectKey{
		Namespace: s.namespace,
		Name:      name,
	}, secret)
	if err != nil {
		return "", err
	}
	stringKey := string(secret.Data["key"])
	return stringKey, nil
}
