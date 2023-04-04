package secret

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type gcpSecretManager struct {
	projectId string
	client    *secretmanager.Client
}

func NewGCPSecretManager(ctx context.Context) (*gcpSecretManager, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return nil, err
	}
	if credentials.ProjectID == "" {
		return nil, fmt.Errorf("ProjectID must be provided")
	}

	return &gcpSecretManager{
		projectId: credentials.ProjectID,
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
