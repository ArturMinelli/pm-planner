package location

import (
	"context"
	"errors"
)

// DeviceLocation is a coordinate fix from the host OS.
type DeviceLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"`
}

// ErrUnavailable is returned when the platform cannot provide a location fix.
var ErrUnavailable = errors.New("localização indisponível neste sistema")

// GetDeviceLocation returns the current device coordinates from the OS.
func GetDeviceLocation(ctx context.Context) (*DeviceLocation, error) {
	return getDeviceLocation(ctx)
}
