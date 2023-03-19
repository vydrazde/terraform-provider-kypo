package KYPOClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SandboxDefinition struct {
	Id        int64     `json:"id" tfsdk:"id"`
	Url       string    `json:"url" tfsdk:"url"`
	Name      string    `json:"name" tfsdk:"name"`
	Rev       string    `json:"rev" tfsdk:"rev"`
	CreatedBy UserModel `json:"created_by" tfsdk:"created_by"`
}

type SandboxDefinitionRequest struct {
	Url string `json:"url"`
	Rev string `json:"rev"`
}

func (c *Client) GetSandboxDefinition(definitionID int64) (*SandboxDefinition, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	definition := SandboxDefinition{}

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

func (c *Client) CreateSandboxDefinition(url, rev string) (*SandboxDefinition, error) {
	requestBody, err := json.Marshal(SandboxDefinitionRequest{url, rev})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions", c.Endpoint), strings.NewReader(string(requestBody)))
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

	definition := SandboxDefinition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func (c *Client) DeleteSandboxDefinition(definitionID int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions/%d", c.Endpoint, definitionID), nil)
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
