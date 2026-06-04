//go:build !darwin

package project

import (
	"context"
	"fmt"
)

func InstallDesktopDarwin(ctx context.Context, options DesktopInstallOptions, runner Runner) error {
	return fmt.Errorf("desktop install for macOS requires building on darwin")
}
