package jwt

import "context"

type Verifier interface {
	Verify(ctx context.Context, jwt string) error
}
