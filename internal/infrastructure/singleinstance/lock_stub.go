//go:build !windows

package singleinstance

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"vrchat-tweaker/internal/platform/paths"
)

const activateMessage = "activate"

type stubGuard struct {
	name     string
	lockPath string
	sockPath string
	lockFile *os.File
	listener net.Listener
	wg       sync.WaitGroup
}

func newPlatformGuard(name string) platformGuard {
	return &stubGuard{name: name}
}

func (s *stubGuard) paths() error {
	if s.lockPath != "" {
		return nil
	}
	dir, err := paths.AppDataDir()
	if err != nil {
		return err
	}
	s.lockPath = filepath.Join(dir, s.name+".lock")
	s.sockPath = filepath.Join(dir, s.name+".sock")
	return nil
}

func (s *stubGuard) acquire() (bool, error) {
	if err := s.paths(); err != nil {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(s.lockPath), 0700); err != nil {
		return false, err
	}
	file, err := os.OpenFile(s.lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return false, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errno, ok := err.(syscall.Errno); ok && (errno == syscall.EWOULDBLOCK || errno == syscall.EAGAIN) {
			return false, nil
		}
		return false, err
	}
	s.lockFile = file
	return true, nil
}

func (s *stubGuard) notifyExisting() error {
	if err := s.paths(); err != nil {
		return err
	}
	conn, err := net.Dial("unix", s.sockPath)
	if err != nil {
		return fmt.Errorf("singleinstance: existing instance could not be activated: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := io.WriteString(conn, activateMessage); err != nil {
		return err
	}
	return nil
}

func (s *stubGuard) start(stop <-chan struct{}, onActivate func()) {
	if onActivate == nil {
		return
	}
	if err := s.paths(); err != nil {
		return
	}
	_ = os.Remove(s.sockPath)
	ln, err := net.Listen("unix", s.sockPath)
	if err != nil {
		return
	}
	s.listener = ln
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			unixLn, ok := ln.(*net.UnixListener)
			if !ok {
				return
			}
			if err := unixLn.SetDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
				return
			}
			conn, err := ln.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				return
			}
			buf := make([]byte, len(activateMessage))
			_, _ = io.ReadFull(conn, buf)
			_ = conn.Close()
			if string(buf) == activateMessage {
				onActivate()
			}
		}
	}()
}

func (s *stubGuard) release() {
	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}
	s.wg.Wait()
	_ = os.Remove(s.sockPath)
	if s.lockFile != nil {
		_ = syscall.Flock(int(s.lockFile.Fd()), syscall.LOCK_UN)
		_ = s.lockFile.Close()
		s.lockFile = nil
	}
}
