package KYPOClient

import (
	"encoding/json"
	"errors"
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

func NewClient(endpoint, token string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		HTTPClient: http.DefaultClient,
		Token:      token,
	}

	return &client, nil
}

var ErrNotFound = errors.New("not found")

func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, res.StatusCode, nil
}

type Definition struct {
	Id        int64     `json:"id" tfsdk:"id"`
	Url       string    `json:"url" tfsdk:"url"`
	Name      string    `json:"name" tfsdk:"name"`
	Rev       string    `json:"rev" tfsdk:"rev"`
	CreatedBy UserModel `json:"created_by" tfsdk:"created_by"`
}

type UserModel struct {
	Id         int64  `json:"id" tfsdk:"id"`
	Sub        string `json:"sub" tfsdk:"sub"`
	FullName   string `json:"full_name" tfsdk:"full_name"`
	GivenName  string `json:"given_name" tfsdk:"given_name"`
	FamilyName string `json:"family_name" tfsdk:"family_name"`
	Mail       string `json:"mail" tfsdk:"mail"`
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

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	definition := Definition{}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("definition %d %w", definitionID, ErrNotFound)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

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

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status != http.StatusCreated {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	definition := Definition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func (c *Client) DeleteDefinition(definitionID int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if status != http.StatusNoContent && status != http.StatusNotFound {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}

	return nil
}
