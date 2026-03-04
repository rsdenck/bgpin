package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bgpin/bgpin/internal/parsers/rpki"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var (
	rpkiOutputFormat string
	rpkiTimeout      int
)

func newRPKICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rpki",
		Short: "RPKI validation for BGP routes",
		Long:  "Query RPKI validation status using RIPE RPKI Validator",
	}

	cmd.PersistentFlags().StringVarP(&rpkiOutputFormat, "output", "o", "table", "Output format: table, json, yaml")
	cmd.PersistentFlags().IntVarP(&rpkiTimeout, "timeout", "t", 30, "Timeout in seconds")

	cmd.AddCommand(newRPKIValidateCommand())

	return cmd
}

func newRPKIValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [asn] [prefix]",
		Short: "Validate a BGP route prefix against RPKI",
		Long:  "Check if a prefix and ASN pair is valid according to RPKI ROAs",
		Args:  cobra.ExactArgs(2),
		RunE:  runRPKIValidate,
		Example: `  bgpin rpki validate 15169 8.8.8.0/24
  bgpin rpki validate 13335 1.1.1.0/24 -o json`,
	}
}

func runRPKIValidate(cmd *cobra.Command, args []string) error {
	asn, err := parseASN(args[0])
	if err != nil {
		return err
	}

	prefix := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rpkiTimeout)*time.Second)
	defer cancel()

	validator := rpki.NewRIPKIValidator(rpki.Config{
		Timeout: time.Duration(rpkiTimeout) * time.Second,
	})

	result, err := validator.ValidateRoute(ctx, asn, prefix)
	if err != nil {
		return fmt.Errorf("RPKI validation failed: %w", err)
	}

	return outputRPKIValidation(result, rpkiOutputFormat)
}

func outputRPKIValidation(result *rpki.ValidationResult, format string) error {
	switch format {
	case "json":
		fmt.Printf("{\n")
		fmt.Printf("  \"prefix\": \"%s\",\n", result.Prefix)
		fmt.Printf("  \"asn\": %d,\n", result.ASN)
		fmt.Printf("  \"state\": \"%s\",\n", result.State)
		fmt.Printf("  \"description\": \"%s\",\n", result.Description)
		fmt.Printf("  \"matched_vrps\": %d,\n", len(result.MatchedVRPs))
		fmt.Printf("  \"timestamp\": \"%s\"\n", result.Timestamp.Format(time.RFC3339))
		fmt.Printf("}\n")
	case "yaml":
		fmt.Printf("prefix: %s\n", result.Prefix)
		fmt.Printf("asn: %d\n", result.ASN)
		fmt.Printf("state: %s\n", result.State)
		fmt.Printf("description: %s\n", result.Description)
		fmt.Printf("matched_vrps: %d\n", len(result.MatchedVRPs))
		fmt.Printf("timestamp: %s\n", result.Timestamp.Format(time.RFC3339))
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("RPKI Validation: %s AS%d", result.Prefix, result.ASN))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		stateDisplay := result.State
		if result.State == "valid" {
			stateDisplay = "VALID"
		} else if result.State == "invalid" {
			stateDisplay = "INVALID"
		} else {
			stateDisplay = "UNKNOWN"
		}

		t.AppendHeader(table.Row{"Prefix", "ASN", "State", "Description"})
		t.AppendRow(table.Row{
			result.Prefix,
			fmt.Sprintf("AS%d", result.ASN),
			stateDisplay,
			result.Description,
		})

		t.Render()

		if len(result.MatchedVRPs) > 0 {
			fmt.Println()
			t2 := table.NewWriter()
			t2.SetOutputMirror(os.Stdout)
			t2.SetTitle("Matched VRPs (Valid Route Origin Authorizations)")
			t2.Style().Title.Align = text.AlignCenter
			t2.SetStyle(table.StyleRounded)
			t2.Style().Options.SeparateRows = false

			t2.AppendHeader(table.Row{"ASN", "Prefix", "Max Length"})
			for _, vrp := range result.MatchedVRPs {
				t2.AppendRow(table.Row{vrp.ASN, vrp.Prefix, vrp.MaxLength})
			}
			t2.Render()
		}

		fmt.Printf("\nValidated at: %s\n", result.Timestamp.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Source: RIPE RPKI Validator\n")
	}

	return nil
}
