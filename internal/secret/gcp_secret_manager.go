package secret

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type gcpSecretManager struct {
	projectId string
	client    *secretmanager.Client
}

// Initialize SecretManager with projectId
func NewGCPSecretManager(ctx context.Context, projectId string) (*gcpSecretManager, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	if projectId == "" {
		return nil, fmt.Errorf("ProjectID must not be empty")
	}

	return &gcpSecretManager{
		projectId: projectId,
		client:    c,
	}, nil
}

// Get latest version from SecretManager
func (s gcpSecretManager) GetSecret(ctx context.Context, name string) (string, error) {
	res, err := s.client.AccessSecretVersion(
		ctx,
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", s.projectId, name),
		},
	)
	if err != nil {
		return "", err
	}
	return string(res.Payload.Data), nil
}

// Close secretmanager's client
func (s gcpSecretManager) Close() {
	s.client.Close()
}
