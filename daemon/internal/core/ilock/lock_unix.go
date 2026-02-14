//go:build linux || darwin

package ilock

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const LOCK_NAME = "backerd.lock"

type InstanceLock struct {
	f *os.File // The file on which acquires the lock
}

func getUnixLockPath(name string) string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, name)
	}
	return filepath.Join(os.TempDir(), name)
}

func NewInstanceLock(name string) (*InstanceLock, error) {
	f, err := os.OpenFile(getUnixLockPath(name), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// Non-blocking exclusive lock
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("another instance is running: %w", err)
	}

	return &InstanceLock{f: f}, nil
}

func (il *InstanceLock) Close() error {
	if il == nil || il.f == nil {
		return nil
	}

	return il.f.Close()
}
