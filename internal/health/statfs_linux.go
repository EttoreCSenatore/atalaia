//go:build linux

package health

import "golang.org/x/sys/unix"

func statfs(path string) (free, total uint64, err error) {
	var s unix.Statfs_t
	if err := unix.Statfs(path, &s); err != nil {
		return 0, 0, err
	}
	bs := uint64(s.Bsize) // #nosec G115 -- block size is a small positive number
	return s.Bavail * bs, s.Blocks * bs, nil
}
