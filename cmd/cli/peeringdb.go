package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gmazoyer/peeringdb"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newPeeringDBCommand() *cobra.Command {
	var pdbCmd = &cobra.Command{
		Use:   "peeringdb",
		Short: "Consultas PeeringDB para informações de peering",
		Long: `Comandos para consultar a base de dados PeeringDB.
		
Obtenha informações sobre ASNs, redes, facilities, exchanges e pontos de troca.
Ideal para planejamento de peering e análise de conectividade.`,
		Aliases: []string{"pdb"},
		Example: `  bgpin peeringdb asn 262978
  bgpin peeringdb network "Armazem"
  bgpin peeringdb ix "IX.br São Paulo"
  bgpin peeringdb facility "Equinix SP1"`,
	}

	// Subcomandos PeeringDB
	pdbCmd.AddCommand(newPDBASNCommand())
	pdbCmd.AddCommand(newPDBNetworkCommand())
	pdbCmd.AddCommand(newPDBIXCommand())
	pdbCmd.AddCommand(newPDBFacilityCommand())

	return pdbCmd
}

func newPDBASNCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "asn <asn>",
		Short: "Informações de ASN no PeeringDB",
		Long: `Consulta informações detalhadas de um ASN no PeeringDB.
		
Mostra dados da rede, facilities, IXs conectados, políticas de peering.`,
		Args: cobra.ExactArgs(1),
		Example: `  bgpin peeringdb asn 262978
  bgpin peeringdb asn 15169`,
		RunE: runPDBASN,
	}

	addPDBFlags(cmd)
	return cmd
}

func newPDBNetworkCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "network <name>",
		Short: "Buscar redes por nome no PeeringDB", 
		Long: `Busca redes por nome ou parte do nome no PeeringDB.`,
		Args: cobra.ExactArgs(1),
		RunE: runPDBNetwork,
	}

	addPDBFlags(cmd)
	return cmd
}
func newPDBIXCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ix <name>",
		Short: "Informações de Internet Exchange",
		Long: `Consulta informações de Internet Exchanges no PeeringDB.`,
		Args: cobra.ExactArgs(1),
		RunE: runPDBIX,
	}

	addPDBFlags(cmd)
	return cmd
}

func newPDBFacilityCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "facility <name>",
		Short: "Informações de facilities/datacenters",
		Long: `Consulta informações de facilities e datacenters no PeeringDB.`,
		Args: cobra.ExactArgs(1),
		RunE: runPDBFacility,
	}

	addPDBFlags(cmd)
	return cmd
}

// addPDBFlags adiciona flags comuns para comandos PeeringDB
func addPDBFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Formato de saída (table, json, yaml)")
	cmd.Flags().IntP("limit", "l", 10, "Limite de resultados")
}

// createPDBClient cria cliente PeeringDB
func createPDBClient() *peeringdb.API {
	return peeringdb.NewAPI()
}

// runPDBASN executa consulta de ASN
func runPDBASN(cmd *cobra.Command, args []string) error {
	asnStr := args[0]
	asn, err := strconv.Atoi(strings.TrimPrefix(asnStr, "AS"))
	if err != nil {
		return fmt.Errorf("ASN inválido: %s", asnStr)
	}

	client := createPDBClient()
	ctx := context.Background()
	
	// Buscar rede por ASN usando método específico
	network, err := client.GetASN(ctx, asn)
	if err != nil {
		return fmt.Errorf("erro ao consultar PeeringDB: %w", err)
	}

	if network == nil {
		fmt.Printf("ASN %d não encontrado no PeeringDB\n", asn)
		return nil
	}
	
	// Renderizar informações
	outputFormat, _ := cmd.Flags().GetString("output")
	return renderPDBASN(*network, outputFormat)
}

// runPDBNetwork executa busca de rede
func runPDBNetwork(cmd *cobra.Command, args []string) error {
	name := args[0]
	limit, _ := cmd.Flags().GetInt("limit")

	client := createPDBClient()
	ctx := context.Background()
	
	// Criar parâmetros de busca
	search := url.Values{}
	search.Set("name__icontains", name)
	
	networks, err := client.GetNetwork(ctx, search)
	if err != nil {
		return fmt.Errorf("erro ao consultar PeeringDB: %w", err)
	}

	if len(networks) == 0 {
		fmt.Printf("Nenhuma rede encontrada com nome '%s'\n", name)
		return nil
	}

	// Limitar resultados
	if len(networks) > limit {
		networks = networks[:limit]
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	return renderPDBNetworks(networks, outputFormat)
}

// runPDBIX executa consulta de IX
func runPDBIX(cmd *cobra.Command, args []string) error {
	name := args[0]
	limit, _ := cmd.Flags().GetInt("limit")

	client := createPDBClient()
	ctx := context.Background()
	
	// Criar parâmetros de busca
	search := url.Values{}
	search.Set("name__icontains", name)
	
	ixs, err := client.GetInternetExchange(ctx, search)
	if err != nil {
		return fmt.Errorf("erro ao consultar PeeringDB: %w", err)
	}

	if len(ixs) == 0 {
		fmt.Printf("Nenhum IX encontrado com nome '%s'\n", name)
		return nil
	}

	if len(ixs) > limit {
		ixs = ixs[:limit]
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	return renderPDBIXs(ixs, outputFormat)
}

// runPDBFacility executa consulta de facility
func runPDBFacility(cmd *cobra.Command, args []string) error {
	name := args[0]
	limit, _ := cmd.Flags().GetInt("limit")

	client := createPDBClient()
	ctx := context.Background()
	
	// Criar parâmetros de busca
	search := url.Values{}
	search.Set("name__icontains", name)
	
	facilities, err := client.GetFacility(ctx, search)
	if err != nil {
		return fmt.Errorf("erro ao consultar PeeringDB: %w", err)
	}

	if len(facilities) == 0 {
		fmt.Printf("Nenhuma facility encontrada com nome '%s'\n", name)
		return nil
	}

	if len(facilities) > limit {
		facilities = facilities[:limit]
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	return renderPDBFacilities(facilities, outputFormat)
}
// Funções de renderização

func renderPDBASN(network peeringdb.Network, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(network, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false
	
	// Limitar largura das colunas
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 15},
		{Number: 2, WidthMax: 60},
	})

	t.AppendHeader(table.Row{"Campo", "Valor"})
	
	t.AppendRow(table.Row{"ASN", fmt.Sprintf("AS%d", network.ASN)})
	t.AppendRow(table.Row{"Nome", network.Name})
	t.AppendRow(table.Row{"Website", network.Website})
	t.AppendRow(table.Row{"Política", network.PolicyGeneral})
	
	// Informações adicionais
	if network.InfoPrefixes4 > 0 {
		t.AppendRow(table.Row{"Prefixos IPv4", fmt.Sprintf("%d", network.InfoPrefixes4)})
	}
	if network.InfoPrefixes6 > 0 {
		t.AppendRow(table.Row{"Prefixos IPv6", fmt.Sprintf("%d", network.InfoPrefixes6)})
	}
	if network.InternetExchangeCount > 0 {
		t.AppendRow(table.Row{"IXs", fmt.Sprintf("%d", network.InternetExchangeCount)})
	}
	if network.FacilityCount > 0 {
		t.AppendRow(table.Row{"Facilities", fmt.Sprintf("%d", network.FacilityCount)})
	}
	
	// Truncar notas se muito longas e adicionar no final
	notes := network.Notes
	if len(notes) > 80 {
		notes = notes[:80] + "..."
	}
	if notes != "" {
		t.AppendRow(table.Row{"Notas", notes})
	}

	t.Render()
	return nil
}

func renderPDBNetworks(networks []peeringdb.Network, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(networks, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"ASN", "Nome", "Website", "Política"})

	for _, network := range networks {
		t.AppendRow(table.Row{
			fmt.Sprintf("AS%d", network.ASN),
			network.Name,
			network.Website,
			network.PolicyGeneral,
		})
	}

	t.Render()
	return nil
}

func renderPDBIXs(ixs []peeringdb.InternetExchange, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(ixs, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Nome", "Cidade", "País", "Website"})

	for _, ix := range ixs {
		t.AppendRow(table.Row{
			ix.Name,
			ix.City,
			ix.Country,
			ix.Website,
		})
	}

	t.Render()
	return nil
}

func renderPDBFacilities(facilities []peeringdb.Facility, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(facilities, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Nome", "Cidade", "País", "Website"})

	for _, facility := range facilities {
		t.AppendRow(table.Row{
			facility.Name,
			facility.City,
			facility.Country,
			facility.Website,
		})
	}

	t.Render()
	return nil
}