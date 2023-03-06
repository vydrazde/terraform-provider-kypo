package KYPOClient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client struct {
	Endpoint   string
	HTTPClient *http.Client
	Token      string
}

func NewClient(endpoint, token string) *Client {
	client := Client{
		Endpoint:   endpoint,
		HTTPClient: http.DefaultClient,
		Token:      token,
	}

	return &client
}

func (c *Client) doRequest(req *http.Request, expected_status int) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != expected_status {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

type Definition struct {
	Id        int64     `json:"id"`
	Url       string    `json:"url"`
	Name      string    `json:"name"`
	Rev       string    `json:"rev"`
	CreatedBy UserModel `json:"created_by"`
}

type UserModel struct {
	Id         int64  `json:"id"`
	Sub        string `json:"sub"`
	FullName   string `json:"full_name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Mail       string `json:"mail"`
}

type DefinitionRequest struct {
	Url string `json:"url"`
	Rev string `json:"rev"`
}

func (c *Client) GetDefinition(definitionID int64) (*Definition, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	definition := Definition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func (c *Client) CreateDefinition(url, rev string) (*Definition, error) {
	requestBody, err := json.Marshal(DefinitionRequest{url, rev})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/definitions", c.Endpoint), strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	definition := Definition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}
