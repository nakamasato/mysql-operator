package secret

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type GCPSecretManager struct {
	ProjectId string
	Client    *secretmanager.Client
}

// Get latest version from SecretManager
func (s GCPSecretManager) GetSecret(ctx context.Context, name string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", s.ProjectId, name),
	}
	res, err := s.Client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}
	return string(res.Payload.Data), nil
}
