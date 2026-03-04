package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bgpin/bgpin/internal/ai/providers"
	"github.com/bgpin/bgpin/internal/ai/schema"
	"github.com/bgpin/bgpin/internal/parsers/http"
	"github.com/bgpin/bgpin/internal/parsers/rpki"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
		Short: "Análise BGP com inteligência artificial",
		Long:  "Use provedores LLM para analisar dados BGP com IA",
	}

	cmd.PersistentFlags().StringVarP(&aiProvider, "provider", "p", "ollama", "Provedor LLM: openai, claude, gemini, ollama")
	cmd.PersistentFlags().StringVarP(&aiModel, "model", "m", "", "Modelo a usar (específico do provedor)")
	cmd.PersistentFlags().StringVarP(&aiPromptFile, "file", "f", "", "Arquivo JSON de entrada com dados BGP")
	cmd.PersistentFlags().BoolVarP(&aiInteractive, "interactive", "i", false, "Modo copiloto interativo")

	cmd.AddCommand(newAIAnalyzeCommand())
	cmd.AddCommand(newAIExplainCommand())
	cmd.AddCommand(newAICopilotCommand())
	cmd.AddCommand(newAIFlowCommand())

	return cmd
}

func newAIAnalyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [prefix]",
		Short: "Analisar prefixo BGP com IA",
		Long:  "Consultar dados BGP e enviar para LLM para análise",
		Args:  cobra.ExactArgs(1),
		RunE:  runAIAnalyze,
		Example: `  bgpin ai analyze 8.8.8.0/24 --provider openai
  bgpin ai analyze 1.1.1.0/24 --provider ollama`,
	}
}

func runAIAnalyze(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	systemPrompt := `Você é um engenheiro de rede BGP sênior. Analise os seguintes dados de rota BGP e forneça insights técnicos.
Identifique:
1. Avaliação de risco de sequestro
2. Indicadores de vazamento de rota
3. Anomalias no caminho
4. Sugestões de otimização
5. Preocupações de segurança

Retorne uma análise técnica estruturada e concisa em português.`

	analysis, err := provider.Analyze(ctx, systemPrompt, bgpData)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	fmt.Println("=== Análise IA ===")
	fmt.Println(analysis)
	fmt.Println("\n=== Dados Brutos ===")
	fmt.Printf("Prefixo: %s\n", prefix)
	fmt.Printf("ASN Origem: AS%d\n", asn)
	fmt.Printf("Status RPKI: %s\n", rpkiStatus)

	return nil
}

func newAIExplainCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain [prefix]",
		Short: "Explicar rota BGP em português claro",
		Long:  "Obter explicação legível de atributos de rota BGP",
		Args:  cobra.ExactArgs(1),
		RunE:  runAIExplain,
		Example: `  bgpin ai explain 8.8.8.0/24
  bgpin ai explain 1.1.1.0/24 --provider claude`,
	}
}

func runAIExplain(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	systemPrompt := `Você é um especialista em BGP. Explique as seguintes informações de rota em português simples e claro.
Foque no que o caminho AS significa, por que as comunidades importam e o que o status RPKI indica.`

	explanation, err := provider.Analyze(ctx, systemPrompt, dataJSON)
	if err != nil {
		return fmt.Errorf("AI explanation failed: %w", err)
	}

	fmt.Println("=== Explicação da Rota BGP ===")
	fmt.Println(explanation)

	return nil
}

func newAICopilotCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "copilot",
		Short: "Modo Copiloto BGP interativo",
		Long:  "Iniciar sessão interativa com assistente IA",
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

	fmt.Println("=== Copiloto BGP ===")
	fmt.Println("Provedor:", provider.Name())
	fmt.Println("Digite 'exit' para sair, 'help' para comandos\n")

	ctx := context.Background()

	commands := map[string]string{
		"help":    "Comandos disponíveis: prefix <prefix>, asn <asn>, rpki <asn> <prefix>, analyze <prefix>",
		"prefix":  "Uso: prefix 8.8.8.0/24",
		"asn":     "Uso: asn 13335",
		"rpki":    "Uso: rpki 15169 8.8.8.0/24",
		"analyze": "Uso: analyze 8.8.8.0/24",
	}

	questions := map[string]string{
		"por que este prefixo está oscilando?": "Isso pode indicar instabilidade de rota, problemas de rede ou problemas de convergência BGP. Verifique atualizações frequentes no caminho AS.",
		"mostrar risco de anomalia no caminho": "Analise o caminho AS para loops, uso de AS privado e sequências AS incomuns.",
		"comparar com ontem":                   "Comparação histórica requer dados arquivados. Use arquivos MRT para análise histórica.",
	}

	for {
		fmt.Print("> ")
		var input string
		fmt.Scanln(&input)

		input = strings.TrimSpace(input)
		if input == "exit" || input == "quit" {
			fmt.Println("Até logo!")
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

		analysis, err := provider.Analyze(ctx, "Você é um especialista em BGP. Responda a pergunta do usuário sobre redes BGP.", input)
		if err != nil {
			fmt.Println("Erro:", err)
		} else {
			fmt.Println(analysis)
		}
	}

	return nil
}

func newAIFlowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "flow",
		Short: "Análise de flows com IA para detecção de anomalias",
		Long:  "Exportar dados de flow e analisar com IA para detectar padrões e anomalias",
		RunE:  runAIFlow,
		Example: `  bgpin ai flow --provider ollama
  bgpin ai flow --provider openai --limit 50`,
	}
}

func runAIFlow(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Simular dados de flow para análise (em produção, viria do collector)
	flowData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"flows": []map[string]interface{}{
			{
				"src_ip":     "192.168.1.100",
				"dst_ip":     "8.8.8.8",
				"src_port":   12345,
				"dst_port":   53,
				"protocol":   "UDP",
				"bytes":      1024,
				"packets":    8,
				"src_asn":    65001,
				"dst_asn":    15169,
				"duration":   5.2,
				"flags":      "normal",
			},
			{
				"src_ip":     "10.0.0.50",
				"dst_ip":     "1.1.1.1",
				"src_port":   443,
				"dst_port":   443,
				"protocol":   "TCP",
				"bytes":      15360,
				"packets":    120,
				"src_asn":    65002,
				"dst_asn":    13335,
				"duration":   30.1,
				"flags":      "suspicious_volume",
			},
			{
				"src_ip":     "172.16.0.200",
				"dst_ip":     "203.0.113.10",
				"src_port":   80,
				"dst_port":   80,
				"protocol":   "TCP",
				"bytes":      512000,
				"packets":    4000,
				"src_asn":    65003,
				"dst_asn":    64512,
				"duration":   2.5,
				"flags":      "ddos_pattern",
			},
		},
		"summary": map[string]interface{}{
			"total_flows":      3,
			"total_bytes":      528384,
			"total_packets":    4128,
			"unique_src_ips":   3,
			"unique_dst_ips":   3,
			"protocols":        []string{"UDP", "TCP"},
			"suspicious_flows": 2,
		},
	}

	provider, err := providers.GetProvider(aiProvider)
	if err != nil {
		return fmt.Errorf("falha ao obter provedor: %w", err)
	}

	systemPrompt := `Você é um especialista em segurança de rede e análise de tráfego. Analise os seguintes dados de flow NetFlow/sFlow e identifique:

1. ANOMALIAS DE TRÁFEGO:
   - Padrões de DDoS
   - Volumes suspeitos
   - Comportamentos anômalos

2. ANÁLISE DE SEGURANÇA:
   - Possíveis ataques
   - Tráfego malicioso
   - Indicadores de comprometimento

3. RECOMENDAÇÕES:
   - Ações de mitigação
   - Regras de firewall
   - Monitoramento adicional

4. CLASSIFICAÇÃO DE RISCO:
   - Alto, Médio, Baixo
   - Justificativa técnica

Forneça uma análise detalhada em português com recomendações práticas.`

	analysis, err := provider.Analyze(ctx, systemPrompt, flowData)
	if err != nil {
		return fmt.Errorf("análise de IA falhou: %w", err)
	}

	fmt.Println("=== Análise de Flows com IA ===")
	fmt.Println(analysis)
	
	fmt.Println("\n=== Resumo dos Dados ===")
	
	// Criar tabela para resumo dos dados
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Resumo da Análise de Flows")
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Métrica", "Valor"})
	t.AppendRow(table.Row{"Total de Flows", flowData["summary"].(map[string]interface{})["total_flows"]})
	t.AppendRow(table.Row{"Total de Bytes", fmt.Sprintf("%v", flowData["summary"].(map[string]interface{})["total_bytes"])})
	t.AppendRow(table.Row{"Total de Pacotes", flowData["summary"].(map[string]interface{})["total_packets"]})
	t.AppendRow(table.Row{"Flows Suspeitos", flowData["summary"].(map[string]interface{})["suspicious_flows"]})
	t.AppendRow(table.Row{"IPs Únicos (Origem)", flowData["summary"].(map[string]interface{})["unique_src_ips"]})
	t.AppendRow(table.Row{"IPs Únicos (Destino)", flowData["summary"].(map[string]interface{})["unique_dst_ips"]})
	t.AppendRow(table.Row{"Protocolos", fmt.Sprintf("%v", flowData["summary"].(map[string]interface{})["protocols"])})

	t.Render()

	// Tabela detalhada dos flows analisados
	fmt.Println("\n=== Detalhes dos Flows Analisados ===")
	
	t2 := table.NewWriter()
	t2.SetOutputMirror(os.Stdout)
	t2.SetTitle("Flows Detectados")
	t2.Style().Title.Align = text.AlignCenter
	t2.SetStyle(table.StyleRounded)
	t2.Style().Options.SeparateRows = false

	t2.AppendHeader(table.Row{"IP Origem", "IP Destino", "Protocolo", "Bytes", "Pacotes", "Status"})
	
	flows := flowData["flows"].([]map[string]interface{})
	for _, flow := range flows {
		status := flow["flags"].(string)
		if status == "normal" {
			status = "Normal"
		} else if status == "suspicious_volume" {
			status = "Volume Suspeito"
		} else if status == "ddos_pattern" {
			status = "Padrão DDoS"
		}
		
		t2.AppendRow(table.Row{
			flow["src_ip"],
			flow["dst_ip"],
			flow["protocol"],
			fmt.Sprintf("%v", flow["bytes"]),
			fmt.Sprintf("%v", flow["packets"]),
			status,
		})
	}
	
	t2.Render()

	return nil
}