//go:build !linux

package health

import "errors"

func statfs(string) (uint64, uint64, error) {
	return 0, 0, errors.New("disk space is only checked on Linux")
}
