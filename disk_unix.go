//go:build !windows

package clistat

import (
	"io/fs"
	"path/filepath"
	"syscall"

	"tailscale.com/types/ptr"
)

// Disk returns the disk usage of the given path at the filesystem level.
// If path is empty, it returns the usage of the root directory.
func (*Statter) Disk(p Prefix, path string) (*Result, error) {
	if path == "" {
		path = "/"
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, err
	}
	var r Result
	r.Total = ptr.To(float64(stat.Blocks * uint64(stat.Bsize)))
	r.Used = float64(stat.Blocks-stat.Bfree) * float64(stat.Bsize)
	r.Unit = "B"
	r.Prefix = p
	return &r, nil
}

// DiskUsage returns the actual disk usage of a directory tree,
// similar to "du -sh". This is useful in containerized environments
// where you want to track usage of specific directories rather than
// the entire filesystem.
//
// Unlike Disk(), which uses statfs to get filesystem-level usage,
// DiskUsage walks the directory tree and sums up file sizes.
//
// Note: This operation can be expensive for large directory trees
// with many small files. Consider using appropriate refresh intervals.
func (*Statter) DiskUsage(p Prefix, path string) (*Result, error) {
	if path == "" {
		path = "/"
	}

	var totalSize int64
	// Track visited inodes to avoid double-counting hard links
	visited := make(map[uint64]struct{})

	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip files/directories we can't access
			return nil
		}

		// Skip directories themselves, we only count file sizes
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Get the underlying syscall.Stat_t to check for hard links
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			// Skip if we've already counted this inode (hard link)
			if _, seen := visited[stat.Ino]; seen {
				return nil
			}
			visited[stat.Ino] = struct{}{}
			// Use actual disk blocks allocated (accounts for sparse files)
			totalSize += stat.Blocks * 512 // Blocks are always 512-byte units
		} else {
			// Fallback to reported size if we can't get block info
			totalSize += info.Size()
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		Used:   float64(totalSize),
		Total:  nil, // Directory usage doesn't have a "total" concept
		Unit:   "B",
		Prefix: p,
	}, nil
}

// DiskUsageWithTotal returns the actual disk usage of a directory tree
// along with the total filesystem capacity. This combines DiskUsage
// with filesystem-level total from Disk.
func (s *Statter) DiskUsageWithTotal(p Prefix, path string) (*Result, error) {
	if path == "" {
		path = "/"
	}

	usage, err := s.DiskUsage(p, path)
	if err != nil {
		return nil, err
	}

	// Get the filesystem total for the path
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		// Return usage without total if we can't get fs stats
		return usage, nil
	}

	usage.Total = ptr.To(float64(stat.Blocks * uint64(stat.Bsize)))
	return usage, nil
}

// DiskUsageSimple returns the actual disk usage of a directory tree
// using only file sizes (not disk blocks). This is faster but less
// accurate for sparse files.
func (*Statter) DiskUsageSimple(p Prefix, path string) (*Result, error) {
	if path == "" {
		path = "/"
	}

	var totalSize int64

	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		totalSize += info.Size()
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		Used:   float64(totalSize),
		Total:  nil,
		Unit:   "B",
		Prefix: p,
	}, nil
}
