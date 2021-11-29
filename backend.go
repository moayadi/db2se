package db2se

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// Factory returns a new backend as logical.Backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// hashiCupsBackend defines an object that
// extends the Vault backend and stores the
// target API's client.
type db2Backend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *Client
}

// backend defines the target API backend
// for Vault. It must include each path
// and the secrets it will store.
func backend() *db2Backend {
	var b = db2Backend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{},
			SealWrapStorage: []string{
				"config",
				"role/*",
				"static-role",
			},
		},
		Paths:       framework.PathAppend(
			b.pathConfig(),
			),
		Secrets:     []*framework.Secret{},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}
	return &b
}

// reset clears any client configuration for a new
// backend to be configured
func (b *db2Backend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

// invalidate clears an existing client configuration in
// the backend
func (b *db2Backend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

// getClient locks the backend as it configures and creates a
// a new client for the target API
func (b *db2Backend) getClient(ctx context.Context, s logical.Storage) (*Client, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	if b.client != nil {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	return nil, fmt.Errorf("need to return client")
}

// backendHelp should contain help information for the backend
const backendHelp = `
The HashiCups secrets backend dynamically generates user tokens.
After mounting this backend, credentials to manage HashiCups user tokens
must be configured with the "config/" endpoints.
`
