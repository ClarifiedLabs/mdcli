//go:build !darwin && !linux

// Package term reports the size of the controlling terminal and detects
// whether a file descriptor is a terminal, using only the standard library.
package term

import "os"

// Size always reports no controlling terminal on unsupported platforms.
func Size() (rows, cols int, ok bool) { return 0, 0, false }

// IsTerminal always reports false on unsupported platforms.
func IsTerminal(f *os.File) bool { return false }
