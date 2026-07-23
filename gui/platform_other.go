//go:build !linux && !windows && !darwin

package main

import (
	"fmt"
	"runtime"
)

func openExternalURL(string) error {
	return fmt.Errorf("opening links is not supported on %s", runtime.GOOS)
}
