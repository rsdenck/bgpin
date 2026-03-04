package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bgpin/bgpin/internal/ai/providers"
	"github.com/bgpin/bgpin/internal/ai/schema"
	"github.com/bgpin/bgpin/internal/parsers/http"
	"github.com/bgpin/bgpin/internal/parsers/rpki"
	"github.com/spf13/cobra"
)

var (
	aiProvider    string
	aiModel       string
	aiPromptFile  string
	aiInteractive bool
)

func newAICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered BGP analysis",
		Long:  "Use LLM providers to analyze BGP data with AI",
	}

	cmd.PersistentFlags().StringVarP(&aiProvider, "provider", "p", "ollama", "LLM provider: openai, claude, gemini, ollama")
	cmd.PersistentFlags().StringVarP(&aiModel, "model", "m", "", "Model to use (provider-specific)")
	cmd.PersistentFlags().StringVarP(&aiPromptFile, "file", "f", "", "Input JSON file with BGP data")
	cmd.PersistentFlags().BoolVarP(&aiInteractive, "interactive", "i", false, "Interactive copilot mode")

	cmd.AddCommand(newAIAnalyzeCommand())
	cmd.AddCommand(newAIExplainCommand())
	cmd.AddCommand(newAICopilotCommand())

	return cmd
}

func newAIAnalyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [prefix]",
		Short: "Analyze BGP prefix with AI",
		Long:  "Query BGP data and send to LLM for analysis",
		Args:  cobra.ExactArgs(1),
		RunE:  runAIAnalyze,
		Example: `  bgpin ai analyze 8.8.8.0/24 --provider openai
  bgpin ai analyze 1.1.1.0/24 --provider ollama`,
	}
}

func runAIAnalyze(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 60)
	defer cancel()

	ripeParser := http.NewRIPEParser()
	result, err := ripeParser.QueryBGP(ctx, prefix)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	asn := 0
	if len(result.Routes) > 0 && len(result.Routes[0].ASPath) > 0 {
		asn = result.Routes[0].ASPath[len(result.Routes[0].ASPath)-1]
	}

	rpkiValidator := rpki.NewRIPKIValidator(rpki.Config{Timeout: 30})
	rpkiStatus := "unknown"

	if asn > 0 {
		if rpkiResult, err := rpkiValidator.ValidateRoute(ctx, asn, prefix); err == nil {
			rpkiStatus = rpkiResult.State
		}
	}

	bgpData := schema.NormalizePrefixAnalysis(prefix, asn, result.Routes[0].ASPath, result.Routes[0].Community, rpkiStatus)

	provider, err := providers.GetProvider(aiProvider)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	systemPrompt := `You are a senior BGP network engineer. Analyze the following BGP route data and provide technical insights.
Identify:
1. Hijack risk assessment
2. Route leak indicators  
3. Path anomalies
4. Optimization suggestions
5. Security concerns

Return a concise, structured technical analysis.`

	analysis, err := provider.Analyze(ctx, systemPrompt, bgpData)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	fmt.Println("=== AI Analysis ===")
	fmt.Println(analysis)
	fmt.Println("\n=== Raw Data ===")
	fmt.Printf("Prefix: %s\n", prefix)
	fmt.Printf("Origin ASN: AS%d\n", asn)
	fmt.Printf("RPKI Status: %s\n", rpkiStatus)

	return nil
}

func newAIExplainCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain [prefix]",
		Short: "Explain BGP route in plain English",
		Long:  "Get human-readable explanation of BGP route attributes",
		Args:  cobra.ExactArgs(1),
		RunE:  runAIExplain,
		Example: `  bgpin ai explain 8.8.8.0/24
  bgpin ai explain 1.1.1.0/24 --provider claude`,
	}
}

func runAIExplain(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 60)
	defer cancel()

	ripeParser := http.NewRIPEParser()
	result, err := ripeParser.QueryBGP(ctx, prefix)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	provider, err := providers.GetProvider(aiProvider)
	if err != nil {
		return err
	}

	var dataJSON string
	if aiPromptFile != "" {
		content, err := os.ReadFile(aiPromptFile)
		if err != nil {
			return err
		}
		dataJSON = string(content)
	} else {
		asPath := result.Routes[0].ASPath
		dataJSON = fmt.Sprintf("Prefix: %s, AS Path: %v, Routes count: %d", prefix, asPath, len(result.Routes))
	}

	systemPrompt := `You are a BGP expert. Explain the following route information in simple, clear English. 
Focus on what the AS path means, why communities matter, and what the RPKI status indicates.`

	explanation, err := provider.Analyze(ctx, systemPrompt, dataJSON)
	if err != nil {
		return fmt.Errorf("AI explanation failed: %w", err)
	}

	fmt.Println("=== BGP Route Explanation ===")
	fmt.Println(explanation)

	return nil
}

func newAICopilotCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "copilot",
		Short: "Interactive BGP Copilot mode",
		Long:  "Start an interactive session with AI assistant",
		RunE:  runAICopilot,
		Example: `  bgpin ai copilot
  bgpin ai copilot --provider openai`,
	}
}

func runAICopilot(cmd *cobra.Command, args []string) error {
	provider, err := providers.GetProvider(aiProvider)
	if err != nil {
		return err
	}

	fmt.Println("=== BGP Copilot ===")
	fmt.Println("Provider:", provider.Name())
	fmt.Println("Type 'exit' to quit, 'help' for commands\n")

	ctx := context.Background()

	commands := map[string]string{
		"help":    "Available commands: prefix <prefix>, asn <asn>, rpki <asn> <prefix>, analyze <prefix>",
		"prefix":  "Usage: prefix 8.8.8.0/24",
		"asn":     "Usage: asn 13335",
		"rpki":    "Usage: rpki 15169 8.8.8.0/24",
		"analyze": "Usage: analyze 8.8.8.0/24",
	}

	questions := map[string]string{
		"why is this prefix flapping?": "This could indicate route instability, network issues, or BGP convergence problems. Check for frequent updates in the AS path.",
		"show path anomaly risk":       "Analyze AS path for loops, private AS usage, and unusual AS sequences.",
		"compare yesterday":            "Historical comparison requires archived data. Use MRT files for historical analysis.",
	}

	for {
		fmt.Print("> ")
		var input string
		fmt.Scanln(&input)

		input = strings.TrimSpace(input)
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "help" {
			fmt.Println(commands["help"])
			continue
		}

		if input == "" {
			continue
		}

		lowerInput := strings.ToLower(input)

		if q, ok := questions[lowerInput]; ok {
			analysis, _ := provider.Analyze(ctx, "You are a BGP engineer. Answer this question concisely.", q)
			fmt.Println(analysis)
			continue
		}

		analysis, err := provider.Analyze(ctx, "You are a BGP expert. Answer the user's question about BGP networking.", input)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(analysis)
		}
	}

	return nil
}
