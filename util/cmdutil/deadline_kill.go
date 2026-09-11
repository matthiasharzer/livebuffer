package cmdutil

import (
	"os/exec"
	"time"
)

func DeadlineKill(cmd *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		err := cmd.Process.Kill()
		if err != nil {
			return err
		}
		return <-done
	case err := <-done:
		return err
	}
}
