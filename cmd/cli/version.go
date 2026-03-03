package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	version   = "0.1.0"
	buildDate = "2026-03-03"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display version, build date, and runtime information",
		Run:   runVersion,
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("bgpin version %s\n", version)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
