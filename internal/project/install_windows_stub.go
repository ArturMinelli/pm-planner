//go:build !windows

package project

import (
	"context"
	"fmt"
)

func InstallDesktopWindows(ctx context.Context, options DesktopInstallOptions, runner Runner) error {
	return fmt.Errorf("desktop install for Windows requires building on windows")
}
