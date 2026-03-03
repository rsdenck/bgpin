package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRouteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Show BGP routes",
		Long:  "Show BGP routes from looking glasses",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show [prefix]",
		Short: "Show routes for a prefix",
		Args:  cobra.ExactArgs(1),
		RunE:  runRouteShow,
	})

	return cmd
}

func runRouteShow(cmd *cobra.Command, args []string) error {
	prefix := args[0]
	fmt.Printf("Showing routes for prefix: %s\n", prefix)
	return nil
}
