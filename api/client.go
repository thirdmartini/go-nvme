package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type Client interface {
	CreateVolume(name, description string, size uint64) (*Volume, error)
	ListVolumes() ([]Volume, error)
	DeleteVolume(UUID string) error
}

type HTTPClient struct {
	address string
	version string
}

func (c *HTTPClient) request(req interface{}, resp interface{}) error {
	arr := strings.SplitN(reflect.TypeOf(req).String(), ".", 2)
	if len(arr) != 2 {
		return ErrUnsupportedRequest
	}

	url := fmt.Sprintf("%s/api/%s/%s", c.address, c.version, arr[1])
	//fmt.Printf("HTTPClient Request To: %s\n", url)

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.New(r.Status)
	}
	defer r.Body.Close()

	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	status := Status{}
	err = json.Unmarshal(data, &status)
	if err != nil {
		return err
	}

	if status.Message != "" {
		return errors.New(status.Message)
	}

	return json.Unmarshal(data, resp)
}

func (c *HTTPClient) CreateVolume(name, description string, size uint64) (*Volume, error) {
	req := &CreateVolumeRequest{
		Name:        name,
		Description: description,
		Size:        size,
	}
	resp := &CreateVolumeResponse{}

	err := c.request(req, resp)
	if err != nil {
		return nil, err
	}

	return resp.Volume, nil
}

func (c *HTTPClient) ListVolumes() ([]Volume, error) {
	req := &ListVolumeRequest{}
	resp := &ListVolumeResponse{}

	err := c.request(req, resp)
	if err != nil {
		return nil, err
	}

	return resp.Volumes, nil
}

func (c *HTTPClient) DeleteVolume(UUID string) error {
	req := &DeleteVolumeRequest{
		UUID: UUID,
	}
	resp := &DeleteVolumeRequest{}

	err := c.request(req, resp)
	if err != nil {
		return err
	}

	return nil
}

func NewHTTPClient(address string) *HTTPClient {
	return &HTTPClient{
		address: address,
		version: Version,
	}
}
