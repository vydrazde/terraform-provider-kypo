package KYPOClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SandboxPool struct {
	Id            int64             `json:"id" tfsdk:"id"`
	DefinitionId  int64             `json:"definition_id" tfsdk:"definition_id"`
	Size          int64             `json:"size" tfsdk:"size"`
	MaxSize       int64             `json:"max_size" tfsdk:"max_size"`
	LockId        int64             `json:"lock_id" tfsdk:"lock_id"`
	Rev           string            `json:"rev" tfsdk:"rev"`
	RevSha        string            `json:"rev_sha" tfsdk:"rev_sha"`
	CreatedBy     UserModel         `json:"created_by" tfsdk:"created_by"`
	HardwareUsage HardwareUsage     `json:"hardware_usage" tfsdk:"hardware_usage"`
	Definition    SandboxDefinition `json:"definition" tfsdk:"definition"`
}

type SandboxPoolRequest struct {
	DefinitionId int64 `json:"definition_id"`
	MaxSize      int64 `json:"max_size"`
}

type HardwareUsage struct {
	Vcpu      string `json:"vcpu" tfsdk:"vcpu"`
	Ram       string `json:"ram" tfsdk:"ram"`
	Instances string `json:"instances" tfsdk:"instances"`
	Network   string `json:"network" tfsdk:"network"`
	Subnet    string `json:"subnet" tfsdk:"subnet"`
	Port      string `json:"port" tfsdk:"port"`
}

func (c *Client) GetSandboxPool(poolId int64) (*SandboxPool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/pools/%d", c.Endpoint, poolId), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	pool := SandboxPool{}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("pool %d %w", poolId, ErrNotFound)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	err = json.Unmarshal(body, &pool)
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

func (c *Client) CreateSandboxPool(definitionId, maxSize int64) (*SandboxPool, error) {
	requestBody, err := json.Marshal(SandboxPoolRequest{definitionId, maxSize})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/pools", c.Endpoint), strings.NewReader(string(requestBody)))
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

	pool := SandboxPool{}
	err = json.Unmarshal(body, &pool)
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

func (c *Client) DeleteSandboxPool(poolId int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/pools/%d", c.Endpoint, poolId), nil)
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
