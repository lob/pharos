package signals

import (
	"syscall"
	"testing"
	"time"
)

func TestSetup(t *testing.T) {
	ch := Setup()

	syscall.Kill(syscall.Getpid(), syscall.SIGINT) //nolint:gosec

	select {
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for signal")
	case <-ch:
	}
}
