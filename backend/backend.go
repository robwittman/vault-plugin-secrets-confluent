package backend

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"strings"
	"sync"
)

type Backend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *client
}

const backendHelp = `
The Confluent secrets backend dynamically generates Cloud API Keys.
After mounting this backend, credentials to manage Confluent API Keys
must be configured with the "config/" endpoints.
`

func New() *Backend {
	var b = Backend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{},
			SealWrapStorage: []string{
				"config",
				"role/*",
			},
		},
		Paths: framework.PathAppend(
			pathRole(&b),
			[]*framework.Path{
				pathConfig(&b),
				pathCredentials(&b),
			},
		),
		Secrets: []*framework.Secret{
			b.confluentApiKey(),
		},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}
	return &b
}

func Factory(ctx context.Context, config *logical.BackendConfig) (logical.Backend, error) {
	b := New()
	if err := b.Setup(ctx, config); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Backend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

func (b *Backend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

func (b *Backend) getClient(ctx context.Context, s logical.Storage) (*client, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	if b.client != nil {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = new(clientConfig)
	}

	b.client, err = newClient(config)
	if err != nil {
		return nil, err
	}

	return b.client, nil
}
