//go:build !windows

package data

// SecureConfDir is a no-op on non-Windows platforms.
func SecureConfDir(dir string) error {
	return nil
}
