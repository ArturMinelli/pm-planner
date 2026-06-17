//go:build !linux

package location

import (
	"context"
	"fmt"
	"runtime"
)

func getDeviceLocation(ctx context.Context) (*DeviceLocation, error) {
	_ = ctx
	return nil, fmt.Errorf("%w: localização do sistema ainda não suportada em %s", ErrUnavailable, runtime.GOOS)
}
