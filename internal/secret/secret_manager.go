package secret

import "context"

type SecretManager interface {
	GetSecret(ctx context.Context, name string) (string, error)
}
