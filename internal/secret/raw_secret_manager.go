package secret

import "context"

type RawSecretManager struct{}

// Return the name as secret value
func (r RawSecretManager) GetSecret(ctx context.Context, name string) (string, error) {
	return name, nil
}
