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
	ApiKey    string `json:"api_key,omitempty"`
	ApiKeyId  string `json:"api_key_id"`
	ApiSecret string `json:"api_secret"`
	Url       string `json:"url"`
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
	apiKeyIdValue, ok := req.Secret.InternalData["api_key_id"]
	if ok {
		apiKeyId, ok = apiKeyIdValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for api_key_id in secret internal data")
		}
	}

	_, err = c.apikeys.APIKeysIamV2Api.DeleteIamV2ApiKey(ctx, apiKeyId).Execute()
	return nil, err
}

func createToken(ctx context.Context, c *client, roleName string) (*confluentApiKey, error) {
	sa, _, err := c.iam.ServiceAccountsIamV2Api.GetIamV2ServiceAccount(ctx, roleName).Execute()
	if err != nil {
		return nil, fmt.Errorf("error reading Confluent service account : %w", err)
	}

	// "spec": {
	//"display_name": "CI kafka access key",
	//"description": "This API key provides kafka access to cluster x",
	//"owner": {},
	//"resource": {}
	//}
	apiKey, _, err := c.apikeys.APIKeysIamV2Api.
		CreateIamV2ApiKey(ctx).
		IamV2ApiKey(v2.IamV2ApiKey{
			Spec: &v2.IamV2ApiKeySpec{
				Owner: &v2.ObjectReference{
					Id: *sa.Id,
				},
			},
		}).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("error creating Confluent API Key : %w", err)
	}

	return &confluentApiKey{
		ApiKeyId:  *apiKey.Id,
		ApiSecret: *apiKey.Spec.Secret,
		Url:       c.iam.GetConfig().Host,
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
