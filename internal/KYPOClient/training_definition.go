package KYPOClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type TrainingDefinition struct {
	Id      int64  `json:"id" tfsdk:"id"`
	Content string `json:"content" tfsdk:"content"`
}

func (c *Client) GetTrainingDefinition(definitionID int64) (*TrainingDefinition, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/kypo-rest-training/api/v1/exports/training-definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/octet-stream")

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, &ErrNotFound{ResourceName: "training definition", Identifier: strconv.FormatInt(definitionID, 10)}
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	definition := TrainingDefinition{
		Id:      definitionID,
		Content: string(body),
	}

	return &definition, nil
}

func (c *Client) CreateTrainingDefinition(content string) (*TrainingDefinition, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/kypo-rest-training/api/v1/imports/training-definitions", c.Endpoint), strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	id := struct {
		Id int64 `json:"id"`
	}{}

	err = json.Unmarshal(body, &id)
	if err != nil {
		return nil, err
	}

	definition := TrainingDefinition{
		Id:      id.Id,
		Content: content,
	}

	return &definition, nil
}

func (c *Client) DeleteTrainingDefinition(definitionID int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/kypo-rest-training/api/v1/training-definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusNotFound {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}

	return nil
}
