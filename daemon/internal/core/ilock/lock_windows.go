//go:build windows

package ilock

import (
	"fmt"

	"golang.org/x/sys/windows"
)

const LOCK_NAME = "Global\\backerd-lock"

type InstanceLock struct {
	h windows.Handle // The file on which acquires the lock
}

// On windows this function will create a named mutex
func NewInstanceLock(name string) (*InstanceLock, error) {
	n, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	h, err := windows.CreateMutex(nil, true, n)
	if err != nil {
		return nil, err
	}

	// If it already existed, someone else holds/created it
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("another instance is running")
	}

	return &InstanceLock{h: h}, nil
}

func (l *InstanceLock) Close() error {
	if l == nil || l.h == 0 {
		return nil
	}
	return windows.CloseHandle(l.h)
}
