package update

import (
	"os"
	"os/exec"
)

// SpawnDetached starts a process that outlives the caller. The update script
// terminates pm-desktop while it runs, so the updater must not share the
// app's session or process group.
func SpawnDetached(dir string, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.SysProcAttr = detachedProcAttr()

	if devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0); err == nil {
		defer devNull.Close()
		command.Stdin = devNull
		command.Stdout = devNull
		command.Stderr = devNull
	}

	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
