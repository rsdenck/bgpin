package main

import (
	"fmt"
	"os"

	"github.com/bgpin/bgpin/internal/tui"
	"github.com/spf13/cobra"
)

func newTUICommand() *cobra.Command {
	var tuiCmd = &cobra.Command{
		Use:   "tui",
		Short: "Start interactive BGP TUI (bgptop)",
		Long: `Start the interactive Terminal User Interface for BGP monitoring.
		
bgptop provides a real-time, BTOP-like interface for monitoring:
- BGP routes and announcements
- ASN information and neighbors  
- NetFlow/sFlow/IPFIX traffic analysis
- Network anomalies and security alerts
- Performance metrics and statistics

Navigation:
- Tab/Shift+Tab: Switch between panels
- q/Ctrl+C: Quit
- r: Refresh data
- h: Help`,
		Example: `  bgpin tui
  bgpin tui --refresh 1s
  bgpin tui --asn 262978
  bgpin tui --flows`,
		RunE: runTUI,
	}

	tuiCmd.Flags().StringP("refresh", "r", "1s", "Refresh interval (e.g., 1s, 5s, 30s)")
	tuiCmd.Flags().IntP("asn", "a", 262978, "Focus on specific ASN (default: 262978)")
	tuiCmd.Flags().BoolP("flows", "f", false, "Start with flows panel active")

	return tuiCmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	refresh, _ := cmd.Flags().GetString("refresh")
	asn, _ := cmd.Flags().GetInt("asn")
	flows, _ := cmd.Flags().GetBool("flows")

	config := tui.Config{
		RefreshInterval: refresh,
		FocusASN:        asn,
		StartWithFlows:  flows,
	}

	if err := tui.Start(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting TUI: %v\n", err)
		return err
	}

	return nil
}