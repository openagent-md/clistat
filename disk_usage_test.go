package clistat_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/coder/clistat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskUsage(t *testing.T) {
	t.Parallel()

	s, err := clistat.New()
	require.NoError(t, err)

	t.Run("EmptyDirectory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		result, err := s.DiskUsage(clistat.PrefixDefault, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(0), result.Used)
		assert.Nil(t, result.Total)
		assert.Equal(t, "B", result.Unit)
	})

	t.Run("DirectoryWithFiles", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		// Create some test files with known sizes
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		subDir := filepath.Join(tmpDir, "subdir")
		file3 := filepath.Join(subDir, "file3.txt")

		require.NoError(t, os.WriteFile(file1, make([]byte, 1024), 0o644))
		require.NoError(t, os.WriteFile(file2, make([]byte, 2048), 0o644))
		require.NoError(t, os.MkdirAll(subDir, 0o755))
		require.NoError(t, os.WriteFile(file3, make([]byte, 4096), 0o644))

		result, err := s.DiskUsage(clistat.PrefixDefault, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// The exact size might vary due to block allocation,
		// but it should be at least the sum of file sizes
		assert.GreaterOrEqual(t, result.Used, float64(1024+2048+4096))
		assert.Nil(t, result.Total)
		assert.Equal(t, "B", result.Unit)
	})

	t.Run("WithPrefix", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "file.txt")
		require.NoError(t, os.WriteFile(file, make([]byte, 1024), 0o644))

		result, err := s.DiskUsage(clistat.PrefixKibi, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, clistat.Prefix(clistat.PrefixKibi), result.Prefix)
		// String representation should use KiB
		str := result.String()
		assert.Contains(t, str, "KiB")
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		t.Parallel()

		result, err := s.DiskUsage(clistat.PrefixDefault, "/nonexistent/path/that/does/not/exist")
		// WalkDir returns an error if the root path doesn't exist
		// The behavior depends on the OS - it may return error or empty result
		if err == nil {
			// If no error, used should be 0 for non-existent path
			assert.Equal(t, float64(0), result.Used)
		}
	})
}

func TestDiskUsageWithTotal(t *testing.T) {
	t.Parallel()

	s, err := clistat.New()
	require.NoError(t, err)

	t.Run("IncludesFilesystemTotal", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "file.txt")
		require.NoError(t, os.WriteFile(file, make([]byte, 1024), 0o644))

		result, err := s.DiskUsageWithTotal(clistat.PrefixDefault, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.Used, float64(1024))
		// Should have a Total from the filesystem
		assert.NotNil(t, result.Total)
		assert.Greater(t, *result.Total, float64(0))
		assert.Equal(t, "B", result.Unit)
	})
}

func TestDiskUsageSimple(t *testing.T) {
	t.Parallel()

	s, err := clistat.New()
	require.NoError(t, err)

	t.Run("UsesFileSizes", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "file.txt")
		// Write exactly 1000 bytes to test that Simple uses file size, not blocks
		require.NoError(t, os.WriteFile(file, make([]byte, 1000), 0o644))

		result, err := s.DiskUsageSimple(clistat.PrefixDefault, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Simple version should report exactly the file size
		assert.Equal(t, float64(1000), result.Used)
	})
}
