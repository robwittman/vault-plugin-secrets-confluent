package backend

import (
	"context"
	"errors"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

type client struct {
	iam     *iamv2.APIClient
	apikeys *apikeysv2.APIClient

	authContext func() context.Context
}

func newClient(config *clientConfig) (*client, error) {
	if config == nil {
		return nil, errors.New("client configuration was nil")
	}

	var credentialHelper func() context.Context

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("both username and password must be provided")
	}

	credentialHelper = func() context.Context {
		return context.WithValue(context.Background(), apikeysv2.ContextBasicAuth, apikeysv2.BasicAuth{
			UserName: config.Username,
			Password: config.Password,
		})
	}

	if config.URL != "" {
		// TODO: Override the URL
	}

	c := &client{
		iam:     iamv2.NewAPIClient(iamv2.NewConfiguration()),
		apikeys: apikeysv2.NewAPIClient(apikeysv2.NewConfiguration()),

		authContext: credentialHelper,
	}
	return c, nil
}
