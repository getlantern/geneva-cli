//go:build !windows

package main

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
)

func NewInterceptor(_ string) (Interceptor, error) {
	return nil, fmt.Errorf("intercept not supported on %s", runtime.GOOS)
}

func (p *interceptor) Intercept() error {
	return fmt.Errorf("intercept not supported on %s", runtime.GOOS)
}

func listAdaptersWrapper(c *cli.Context) error {
	return fmt.Errorf("intercept not supported on %s", runtime.GOOS)
}
