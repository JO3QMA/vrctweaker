//go:build !windows

package singleinstance

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"vrchat-tweaker/internal/platform/paths"
)

const (
	activateMessage = "activate"
	// sunPathMax is the maximum Unix domain socket path length (excluding NUL).
	sunPathMax = 104
)

type stubGuard struct {
	name        string
	lockPath    string
	sockPath    string
	sockNetwork string
	sockAddr    string
	lockFile    *os.File
	listener    net.Listener
	wg          sync.WaitGroup
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

func abstractActivateAddr(appName string) string {
	return fmt.Sprintf("@%s_%d_activate", appName, os.Getuid())
}

func (s *stubGuard) resolveSocket() (network, addr string, err error) {
	if err := s.paths(); err != nil {
		return "", "", err
	}
	if len(s.sockPath) < sunPathMax {
		return "unix", s.sockPath, nil
	}
	if runtime.GOOS == "linux" {
		return "unix", abstractActivateAddr(s.name), nil
	}
	tmpAddr := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d.sock", s.name, os.Getuid()))
	if len(tmpAddr) >= sunPathMax {
		return "", "", fmt.Errorf("singleinstance: activation socket path too long (%d bytes): %q", len(s.sockPath), s.sockPath)
	}
	return "unix", tmpAddr, nil
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
	network, addr, err := s.resolveSocket()
	if err != nil {
		return err
	}
	if s.sockNetwork == "" {
		s.sockNetwork = network
		s.sockAddr = addr
	}
	conn, err := net.Dial(network, addr)
	if err != nil {
		return fmt.Errorf("singleinstance: existing instance could not be activated: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := io.WriteString(conn, activateMessage); err != nil {
		return err
	}
	return nil
}

func (s *stubGuard) start(stop <-chan struct{}, dispatch func()) error {
	network, addr, err := s.resolveSocket()
	if err != nil {
		return err
	}
	s.sockNetwork = network
	s.sockAddr = addr

	if !strings.HasPrefix(addr, "@") {
		if rmErr := os.Remove(addr); rmErr != nil && !os.IsNotExist(rmErr) {
			log.Printf("singleinstance: failed to remove stale socket %q: %v", addr, rmErr)
		}
	}
	ln, err := net.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("singleinstance: failed to bind activation socket at %q: %w", addr, err)
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
			if !authorizedPeer(conn) {
				_ = conn.Close()
				continue
			}
			buf := make([]byte, len(activateMessage))
			_, _ = io.ReadFull(conn, buf)
			_ = conn.Close()
			if string(buf) == activateMessage {
				dispatch()
			}
		}
	}()
	return nil
}

func (s *stubGuard) release() {
	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}
	s.wg.Wait()
	if s.sockAddr != "" && !strings.HasPrefix(s.sockAddr, "@") {
		if err := os.Remove(s.sockAddr); err != nil && !os.IsNotExist(err) {
			log.Printf("singleinstance: failed to remove activation socket %q: %v", s.sockAddr, err)
		}
	}
	if s.lockFile != nil {
		_ = syscall.Flock(int(s.lockFile.Fd()), syscall.LOCK_UN)
		_ = s.lockFile.Close()
		s.lockFile = nil
	}
}
