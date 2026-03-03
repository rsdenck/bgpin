package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNeighborsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "neighbors",
		Short: "Show BGP neighbors",
		Long:  "Show BGP neighbor information",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all BGP neighbors",
		RunE:  runNeighborsList,
	})

	return cmd
}

func runNeighborsList(cmd *cobra.Command, args []string) error {
	fmt.Println("Listing BGP neighbors...")
	return nil
}
