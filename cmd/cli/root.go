package main

import (
	"fmt"
	"os"

	"github.com/bgpin/bgpin/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd := &cobra.Command{
		Use:   "bgpin",
		Short: "BGP Looking Glass CLI - Query and analyze BGP data",
		Long: `bgpin is a CLI tool for querying multiple BGP Looking Glasses,
executing standardized BGP commands, and analyzing route data.

Supports multiple vendors (Cisco, Juniper, FRR) and provides
structured output in JSON, YAML, or table format.`,
		Version: "0.1.0",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./bgpin.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(
		newLookupCommand(),
		newRouteCommand(),
		newNeighborsCommand(),
		newAnalyzeCommand(),
		newListCommand(),
		newASNCommand(),
		newPrefixCommand(),
		newFlowCommand(),
		newRPKICommand(),
		newConfigCommand(),
		newAICommand(),
		newApplyCommand(),
		newTUICommand(),
		newSNMPCommand(),
		newPeeringDBCommand(), // Novo comando PeeringDB
		newVersionCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("bgpin")
	}

	viper.AutomaticEnv()
	viper.SetDefault("timeout", 30)
	viper.SetDefault("output", "table")
	viper.SetDefault("cache.enabled", true)
	viper.SetDefault("cache.ttl", 300)

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

func GetConfig() *config.Config {
	return &config.Config{
		Timeout: viper.GetInt("timeout"),
		Output:  viper.GetString("output"),
		Cache: config.CacheConfig{
			Enabled: viper.GetBool("cache.enabled"),
			TTL:     viper.GetInt("cache.ttl"),
		},
		LookingGlasses: config.GetDefaultLGs(),
	}
}
