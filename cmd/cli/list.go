package main

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lg",
		Short: "List available looking glasses",
		Long:  "List all configured looking glasses and their status",
		RunE:  runList,
	}
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Available Looking Glasses")
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Name", "Vendor", "Type", "Protocol", "Country", "URL"})

	for _, lg := range cfg.LookingGlasses {
		t.AppendRow(table.Row{
			lg.Name,
			lg.Vendor,
			lg.Type,
			lg.Protocol,
			lg.Country,
			lg.URL,
		})
	}

	t.Render()
	return nil
}
