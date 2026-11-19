//go:build unix

package store

import "syscall"

func syscallUmask(m int) int { return syscall.Umask(m) }
