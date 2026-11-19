//go:build !unix

package store

func syscallUmask(m int) int { return m }
