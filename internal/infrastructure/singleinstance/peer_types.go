//go:build !windows

package singleinstance

type syscallRawConn interface {
	Control(f func(fd uintptr)) error
}
