package panel

import (
	"fmt"
	"time"

	"encoding/json"
)

type MachineNode struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type MachineNodesResponse struct {
	Nodes      []MachineNode `json:"nodes"`
	BaseConfig *BaseConfig   `json:"base_config"`
}

type MachineResource struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

type MachineStatus struct {
	CPU  float64         `json:"cpu"`
	Mem  MachineResource `json:"mem"`
	Swap MachineResource `json:"swap"`
	Disk MachineResource `json:"disk"`
}

func (r *MachineNodesResponse) PushInterval() time.Duration {
	if r == nil || r.BaseConfig == nil {
		return 0
	}
	return intervalToTime(r.BaseConfig.PushInterval)
}

func (r *MachineNodesResponse) PullInterval() time.Duration {
	if r == nil || r.BaseConfig == nil {
		return 0
	}
	return intervalToTime(r.BaseConfig.PullInterval)
}

func (c *Client) GetMachineNodes() (*MachineNodesResponse, error) {
	const path = "/api/v2/server/machine/nodes"
	r, err := c.client.R().
		ForceContentType("application/json").
		Post(path)
	if err = c.checkResponse(r, path, err); err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("received nil response")
	}
	resp := &MachineNodesResponse{}
	if err = json.Unmarshal(r.Body(), resp); err != nil {
		return nil, fmt.Errorf("decode machine nodes error: %s", err)
	}
	return resp, nil
}

func (c *Client) ReportMachineStatus(status *MachineStatus) error {
	const path = "/api/v2/server/machine/status"
	r, err := c.client.R().
		SetBody(status).
		ForceContentType("application/json").
		Post(path)
	return c.checkResponse(r, path, err)
}
