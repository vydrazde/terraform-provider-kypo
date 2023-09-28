package KYPOClient

import (
	"net/http"
)

type Client struct {
	Endpoint   string
	ClientID   string
	HTTPClient *http.Client
	Token      string
	Username   string
	Password   string
}

func NewClientWithToken(endpoint, clientId, token string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Token:      token,
	}

	return &client, nil
}

func NewClient(endpoint, clientId, username, password string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Username:   username,
		Password:   password,
	}
	token, err := client.signIn()
	if err != nil {
		return nil, err
	}
	client.Token = token
	return &client, nil
}

func NewClientKeycloak(endpoint, clientId, username, password string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Username:   username,
		Password:   password,
	}
	token, err := client.authenticateKeycloak()
	if err != nil {
		return nil, err
	}
	client.Token = token
	return &client, nil
}
