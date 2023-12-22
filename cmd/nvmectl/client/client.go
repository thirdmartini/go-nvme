package client

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/thirdmartini/go-nvme/api"
)

type Client struct {
	binPath string
	address string
}

func (c *Client) CreateVolume(name, description string, size uint64) (*api.Volume, error) {
	out, err := exec.Command(c.binPath, "-api", c.address, "target", "create", "-size", fmt.Sprintf("%d", size)).Output()

	var status api.Status
	err = json.Unmarshal(out, &status)
	if err != nil {
		return nil, err
	}

	if status.Message != "" {
		return nil, fmt.Errorf(status.Message)
	}

	var volume api.Volume
	err = json.Unmarshal(out, &volume)
	return &volume, err
}

func (c *Client) ListVolumes() ([]api.Volume, error) {
	out, err := exec.Command(c.binPath, "-api", c.address, "target", "list").Output()
	if err != nil {
		return nil, err
	}

	var status api.Status
	_ = json.Unmarshal(out, &status)

	if status.Message != "" {
		return nil, fmt.Errorf(status.Message)
	}

	volumes := make([]api.Volume, 0)
	err = json.Unmarshal(out, &volumes)
	return volumes, err
}

func (c *Client) DeleteVolume(UUID string) error {
	out, err := exec.Command(c.binPath, "-api", c.address, "target", "list", "-uuid", UUID).Output()
	if err != nil {
		return err
	}

	var status api.Status
	err = json.Unmarshal(out, &status)
	if err != nil {
		return err
	}

	if status.Message != "" {
		return fmt.Errorf(status.Message)
	}

	return nil
}

func NewClient(address, binPath string) *Client {
	if binPath == "" {
		binPath = "./nvmectl"
	}

	return &Client{
		binPath: binPath,
		address: address,
	}
}
