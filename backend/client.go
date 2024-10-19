package backend

import (
	"errors"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

type client struct {
	iam     *iamv2.APIClient
	apikeys *apikeysv2.APIClient
}

func newApiClient() *client {
	// Figure out auth
	return &client{
		iam:     iamv2.NewAPIClient(nil),
		apikeys: apikeysv2.NewAPIClient(nil),
	}
}

func newClient(config *clientConfig) (*client, error) {
	if config == nil {
		return nil, errors.New("client configuration was nil")
	}

	if config.Username == "" {
		return nil, errors.New("client username was not defined")
	}

	if config.Password == "" {
		return nil, errors.New("client password was not defined")
	}

	if config.URL == "" {
		return nil, errors.New("client URL was not defined")
	}

	//c := newApiClient(&config.URL, &config.Username, &config.Password)
	c := newApiClient()
	return c, nil
}
