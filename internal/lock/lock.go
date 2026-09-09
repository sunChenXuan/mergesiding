// Package lock provides an exclusive file lock for serial integrate.
package lock

import (
	"errors"
	"os"
	"time"

	"github.com/sunChenXuan/mergesiding/internal/paths"
)

// ErrTimeout is returned when the integrate lock cannot be acquired.
var ErrTimeout = errors.New("integrate lock timeout")

// IntegrateLock is an exclusive lock around integrate operations.
type IntegrateLock struct {
	Paths paths.Paths
	file  *os.File
}

// Acquire tries to hold the integrate lock. If blocking is false, fails immediately when held.
// timeout is ignored when blocking is false; when blocking, zero means wait forever.
func (l *IntegrateLock) Acquire(blocking bool, timeout time.Duration) error {
	if err := l.Paths.Ensure(); err != nil {
		return err
	}
	path := l.Paths.IntegrateLock()
	start := time.Now()
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return err
		}
		if err := tryExclusive(f); err != nil {
			_ = f.Close()
			if !blocking {
				return ErrTimeout
			}
			if timeout > 0 && time.Since(start) > timeout {
				return ErrTimeout
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		l.file = f
		return nil
	}
}

// Release unlocks and closes the lock file.
func (l *IntegrateLock) Release() error {
	if l.file == nil {
		return nil
	}
	err := unlock(l.file)
	cerr := l.file.Close()
	l.file = nil
	if err != nil {
		return err
	}
	return cerr
}
