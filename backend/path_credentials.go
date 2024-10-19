package backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCredentials(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathCredentialsRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathCredentialsRead,
			},
		},
		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}
}

const pathCredentialsHelpSyn = `
Generate a Confluent API Key from a specific Vault role.
`

const pathCredentialsHelpDesc = `
This path generates Confluent API Keys for a particular service
account. A role can only represent a single service account
`

func (b *Backend) createApiKey(ctx context.Context, s logical.Storage, roleEntry *confluentRoleEntry) (*confluentApiKey, error) {
	client, err := b.getClient(ctx, s)
	if err != nil {
		return nil, err
	}

	var apiKey *confluentApiKey

	apiKey, err = createToken(ctx, client, roleEntry.ServiceAccount)
	if err != nil {
		return nil, fmt.Errorf("error creating Confluent API Key: %w", err)
	}

	if apiKey == nil {
		return nil, errors.New("error creating Confluent secret")
	}

	return apiKey, nil
}

func (b *Backend) createRoleCreds(ctx context.Context, req *logical.Request, role *confluentRoleEntry) (*logical.Response, error) {
	apiKey, err := b.createApiKey(ctx, req.Storage, role)
	if err != nil {
		return nil, err
	}

	resp := b.Secret(ConfluentApiKeyType).Response(map[string]interface{}{
		"api_key_id": apiKey.ApiKeyId,
		"api_key":    apiKey.ApiKey,
		"api_secret": apiKey.ApiSecret,
	}, map[string]interface{}{
		"api_key":    apiKey.ApiKey,
		"api_secret": apiKey.ApiSecret,
		"role":       role.ServiceAccount,
	})

	if role.TTL > 0 {
		resp.Secret.TTL = role.TTL
	}

	if role.MaxTTL > 0 {
		resp.Secret.MaxTTL = role.MaxTTL
	}

	return resp, nil
}

func (b *Backend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	roleEntry, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	return b.createRoleCreds(ctx, req, roleEntry)
}
