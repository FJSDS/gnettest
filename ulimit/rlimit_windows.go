//go:build windows
// +build windows

package ulimit

func SetRLimit() error {
	return nil
}
