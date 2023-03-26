package KYPOClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/exp/slices"
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
		return nil, fmt.Errorf("sandbox allocation unit %d %w", unitId, ErrNotFound)
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

func (c *Client) CreateSandboxAllocationUnitAwait(poolId int64) (*SandboxAllocationUnit, error) {
	units, err := c.CreateSandboxAllocationUnits(poolId, 1)
	if err != nil {
		return nil, err
	}
	if len(units) != 1 {
		return nil, fmt.Errorf("expected one allocation unit to be created, got %d instead", len(units))
	}
	unit := units[0]
	request, err := c.PollRequestFinished(unit.Id, 5*time.Second, "allocation")
	unit.AllocationRequest = *request
	return &unit, err
}

type valueOrError[T any] struct {
	err   error
	value T
}

func (c *Client) CreateSandboxAllocationUnitAwaitTimeout(poolId int64, timeout time.Duration) (*SandboxAllocationUnit, error) {
	resultChannel := make(chan valueOrError[*SandboxAllocationUnit], 1)
	go func() {
		res, err := c.CreateSandboxAllocationUnitAwait(poolId)
		resultChannel <- valueOrError[*SandboxAllocationUnit]{err: err, value: res}
	}()

	select {
	case result := <-resultChannel:
		return result.value, result.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("creating sandbox allocation unit %d %w %s", poolId, ErrTimeout, timeout)
	}
}

func (c *Client) CreateSandboxCleanupRequest(unitId int64) (*SandboxRequest, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d/cleanup-request", c.Endpoint, unitId), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("sandbox allocation unit %d %w", unitId, ErrNotFound)
	}

	if status != http.StatusCreated {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	sandboxRequest := SandboxRequest{}
	err = json.Unmarshal(body, &sandboxRequest)
	if err != nil {
		return nil, err
	}

	return &sandboxRequest, nil
}

func (c *Client) PollRequestFinished(unitId int64, pollTime time.Duration, requestType string) (*SandboxRequest, error) {
	ticker := time.Tick(pollTime)
	for range ticker {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d/%s-request", c.Endpoint, unitId, requestType), nil)
		if err != nil {
			return nil, err
		}

		body, status, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}

		if status == http.StatusNotFound {
			return nil, fmt.Errorf("sandbox request %d %w", unitId, ErrNotFound)
		}

		if status != http.StatusOK {
			return nil, fmt.Errorf("status: %d, body: %s", status, body)
		}
		sandboxRequest := SandboxRequest{}
		err = json.Unmarshal(body, &sandboxRequest)
		if err != nil {
			return nil, err
		}

		if !slices.Contains(sandboxRequest.Stages, "RUNNING") && !slices.Contains(sandboxRequest.Stages, "IN_QUEUE") {
			return &sandboxRequest, nil
		}
	}
	return nil, nil // Unreachable
}

func (c *Client) DeleteSandboxAllocationUnit(unitId int64) error {
	_, err := c.CreateSandboxCleanupRequest(unitId)
	if err != nil {
		return err
	}

	_, err = c.PollRequestFinished(unitId, 3*time.Second, "cleanup")
	// After cleanup is finished it deletes itself and 404 is thrown
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

var ErrTimeout = errors.New("has not finished within timeout")

func (c *Client) DeleteSandboxAllocationUnitWithTimeout(unitId int64, timeout time.Duration) error {
	resultChannel := make(chan error, 1)
	go func() {
		resultChannel <- c.DeleteSandboxAllocationUnit(unitId)
	}()

	select {
	case result := <-resultChannel:
		return result
	case <-time.After(timeout):
		return fmt.Errorf("deleting sandbox allocation unit %d %w %s", unitId, ErrTimeout, timeout)
	}
}