package api

import (
	"errors"
)

const (
	Version = "v1"
)

var (
	ErrUnsupportedRequest = errors.New("unsupported request")
)

type Status struct {
	Code    int
	Message string
}

type Volume struct {
	UUID        string
	Name        string
	Description string
	Size        uint64
	NQN         string
}

type CreateVolumeRequest struct {
	Name        string
	Description string
	Size        uint64
}

type CreateVolumeResponse struct {
	Status
	Volume *Volume
}

type GetVolumeRequest struct {
	UUID string
}

type GetVolumeResponse struct {
	Status
	Volume *Volume
}

type DeleteVolumeRequest struct {
	UUID string
}

type DeleteVolumeResponse struct {
	Status
}

type ListVolumeRequest struct {
}

type ListVolumeResponse struct {
	Status  Status
	Volumes []Volume
}
