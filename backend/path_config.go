package backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configStoragePath = "config"
)

type clientConfig struct {
	URL         string `json:"url"`
	AccessToken string `json:"access_token,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

func getConfig(ctx context.Context, s logical.Storage) (*clientConfig, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	config := new(clientConfig)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("error reading root configuration: %w", err)
	}

	return config, nil
}

func pathConfig(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"username": {
				Type:        framework.TypeString,
				Description: "The username to access Confluent API",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Username",
					Sensitive: false,
				},
			},
			"password": {
				Type:        framework.TypeString,
				Description: "The user's password to access Confluent API",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Password",
					Sensitive: true,
				},
			},
			"access_token": {
				Type:        framework.TypeString,
				Description: "The access token to access Confluent API",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "AccessToken",
					Sensitive: true,
				},
			},
			"url": {
				Type:        framework.TypeString,
				Description: "The URL for the HashiCups Product API",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "URL",
					Sensitive: false,
				},
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},
		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    pathConfigHelpSynopsis,
		HelpDescription: pathConfigHelpDescription,
	}
}

func (b *Backend) pathConfigExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

const pathConfigHelpSynopsis = `Configure the Confluent backend.`

const pathConfigHelpDescription = `
The Confluent secret backend requires credentials for managing
IAM and API Key resources in Confluent Cloud.
`

func (b *Backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"username": config.Username,
			"url":      config.URL,
		},
	}, nil
}

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	createOperation := req.Operation == logical.CreateOperation

	if config == nil {
		if !createOperation {
			return nil, errors.New("config not found during update operation")
		}
		config = new(clientConfig)
	}

	authProvided := false
	if username, ok := data.GetOk("username"); ok {
		config.Username = username.(string)
	}

	if url, ok := data.GetOk("url"); ok {
		config.URL = url.(string)
	}

	if password, ok := data.GetOk("password"); ok {
		config.Password = password.(string)
		authProvided = true
	}

	if accessToken, ok := data.GetOk("access_token"); ok {
		config.AccessToken = accessToken.(string)
		authProvided = true
	}

	if createOperation && !authProvided {
		return nil, fmt.Errorf("authentication missing in configuration")
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	b.reset()

	return nil, nil
}

func (b *Backend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, configStoragePath)

	if err == nil {
		b.reset()
	}

	return nil, err
}
