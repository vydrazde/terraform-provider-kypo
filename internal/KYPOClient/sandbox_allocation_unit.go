package KYPOClient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SandboxAllocationUnit struct {
	Id                int64          `json:"id" tfsdk:"id"`
	PoolId            int64          `json:"pool_id" tfsdk:"pool_id"`
	AllocationRequest SandboxRequest `json:"allocation_request" tfsdk:"allocation_request"`
	CleanupRequest    SandboxRequest `json:"cleanup_request" tfsdk:"cleanup_request"`
	CreatedBy         UserModel      `json:"created_by" tfsdk:"created_by"`
	Locked            bool           `json:"locked" tfsdk:"locked"`
}

type SandboxRequest struct {
	Id               int64    `json:"id" tfsdk:"id"`
	AllocationUnitId int64    `json:"allocation_unit_id" tfsdk:"allocation_unit_id"`
	Created          string   `json:"created" tfsdk:"created"`
	Stages           []string `json:"stages" tfsdk:"stages"`
}

func (c *Client) GetSandboxAllocationUnit(unitId int64) (*SandboxAllocationUnit, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d", c.Endpoint, unitId), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	allocationUnit := SandboxAllocationUnit{}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("allocationUnit %d %w", unitId, ErrNotFound)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	err = json.Unmarshal(body, &allocationUnit)
	if err != nil {
		return nil, err
	}

	return &allocationUnit, nil
}

func (c *Client) CreateSandboxAllocationUnits(poolId, count int64) ([]SandboxAllocationUnit, error) {
	// check if cleanup request is already created
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/pools/%d/sandbox-allocation-units?count=%d", c.Endpoint, poolId, count), nil)
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

	var allocationUnit []SandboxAllocationUnit
	err = json.Unmarshal(body, &allocationUnit)
	if err != nil {
		return nil, err
	}

	return allocationUnit, nil
}

func (c *Client) DeleteSandboxAllocationUnit(unitId int64) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d/cleanup-request", c.Endpoint, unitId), nil)
	if err != nil {
		return err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if status != http.StatusCreated && status != http.StatusNotFound {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}
	// unmarshall response and await deletion
	return nil
}
