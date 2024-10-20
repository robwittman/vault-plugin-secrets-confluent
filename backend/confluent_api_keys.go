package backend

import (
	"context"
	"errors"
	"fmt"
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	ConfluentApiKeyType = "confluent_api_key"
)

type confluentApiKey struct {
	ApiKey    string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
}

func (b *Backend) confluentApiKey() *framework.Secret {
	return &framework.Secret{
		Type: ConfluentApiKeyType,
		Fields: map[string]*framework.FieldSchema{
			"api_key": {
				Type:        framework.TypeString,
				Description: "Confluent API Key",
			},
			"api_secret": {
				Type:        framework.TypeString,
				Description: "Confluent API Secret",
			},
		},
		Revoke: b.apiKeyRevoke,
		Renew:  b.tokenRenew,
	}
}

func (b *Backend) apiKeyRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	c, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	apiKeyId := ""
	apiKeyIdValue, ok := req.Secret.InternalData["api_key"]
	if ok {
		apiKeyId, ok = apiKeyIdValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for api_key in secret internal data")
		}
	}

	auth := c.authContext()

	_, err = c.apikeys.APIKeysIamV2Api.DeleteIamV2ApiKey(auth, apiKeyId).Execute()
	return nil, err
}

func createToken(ctx context.Context, c *client, roleName string) (*confluentApiKey, error) {
	ownerKind := "service-account"
	spec := v2.NewIamV2ApiKeySpec()
	spec.SetDisplayName("Vault generated token")
	spec.SetDescription("Vault generated token")
	spec.SetOwner(v2.ObjectReference{Id: roleName, Kind: &ownerKind})
	createApiKeyRequest := v2.IamV2ApiKey{Spec: spec}
	auth := c.authContext()

	apiKey, _, err := c.apikeys.APIKeysIamV2Api.
		CreateIamV2ApiKey(auth).
		IamV2ApiKey(createApiKeyRequest).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("error creating Confluent API Key : %w", err)
	}

	return &confluentApiKey{
		ApiKey:    *apiKey.Id,
		ApiSecret: *apiKey.Spec.Secret,
	}, nil
}

func (b *Backend) tokenRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleRaw, ok := req.Secret.InternalData["role"]
	if !ok {
		return nil, fmt.Errorf("secret is missing role internal data")
	}

	role := roleRaw.(string)
	roleEntry, err := b.getRole(ctx, req.Storage, role)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	resp := &logical.Response{Secret: req.Secret}

	if roleEntry.TTL > 0 {
		resp.Secret.TTL = roleEntry.TTL
	}
	if roleEntry.MaxTTL > 0 {
		resp.Secret.MaxTTL = roleEntry.MaxTTL
	}

	return resp, nil
}
