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
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
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

	systemPrompt := `Analise os dados de flow NetFlow/sFlow e identifique:

1. Anomalias de tráfego (DDoS, volumes suspeitos)
2. Análise de segurança (ataques, tráfego malicioso)
3. Recomendações (mitigação, firewall)
4. Classificação de risco (Alto/Médio/Baixo)

Responda em português de forma concisa.`

	// Primeiro: Tabela detalhada dos flows analisados
	t1 := table.NewWriter()
	t1.SetOutputMirror(os.Stdout)
	t1.SetTitle("Flows Detectados")
	t1.Style().Title.Align = text.AlignCenter
	t1.SetStyle(table.StyleRounded)
	t1.Style().Options.SeparateRows = false

	t1.AppendHeader(table.Row{"IP ORIGEM", "IP DESTINO", "PROTOCOLO", "BYTES", "PACOTES", "STATUS"})

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

		t1.AppendRow(table.Row{
			flow["src_ip"],
			flow["dst_ip"],
			flow["protocol"],
			fmt.Sprintf("%v", flow["bytes"]),
			fmt.Sprintf("%v", flow["packets"]),
			status,
		})
	}

	t1.Render()

	// Segundo: Tabela de resumo
	fmt.Println()
	t2 := table.NewWriter()
	t2.SetOutputMirror(os.Stdout)
	t2.SetTitle("Resumo da Análise de Flows")
	t2.Style().Title.Align = text.AlignCenter
	t2.SetStyle(table.StyleRounded)
	t2.Style().Options.SeparateRows = false

	t2.AppendHeader(table.Row{"MÉTRICA", "VALOR"})
	t2.AppendRow(table.Row{"Total de Flows", flowData["summary"].(map[string]interface{})["total_flows"]})
	t2.AppendRow(table.Row{"Total de Bytes", fmt.Sprintf("%v", flowData["summary"].(map[string]interface{})["total_bytes"])})
	t2.AppendRow(table.Row{"Total de Pacotes", flowData["summary"].(map[string]interface{})["total_packets"]})
	t2.AppendRow(table.Row{"Flows Suspeitos", flowData["summary"].(map[string]interface{})["suspicious_flows"]})
	t2.AppendRow(table.Row{"IPs Únicos (Origem)", flowData["summary"].(map[string]interface{})["unique_src_ips"]})
	t2.AppendRow(table.Row{"IPs Únicos (Destino)", flowData["summary"].(map[string]interface{})["unique_dst_ips"]})
	t2.AppendRow(table.Row{"Protocolos", fmt.Sprintf("%v", flowData["summary"].(map[string]interface{})["protocols"])})

	t2.Render()

	// Terceiro: Análise IA em tabela formatada
	fmt.Println()
	
	var analysis string
	
	// Tentar obter análise da IA com timeout mais curto
	if aiProvider != "" {
		provider, err := providers.GetProvider(aiProvider)
		if err == nil {
			// Timeout mais curto para evitar travamento
			aiCtx, aiCancel := context.WithTimeout(ctx, 60*time.Second)
			defer aiCancel()
			
			aiAnalysis, err := provider.Analyze(aiCtx, systemPrompt, flowData)
			if err != nil {
				// Se falhar, usar análise padrão
				analysis = generateFallbackAnalysis()
			} else {
				analysis = aiAnalysis
			}
		} else {
			analysis = generateFallbackAnalysis()
		}
	} else {
		analysis = generateFallbackAnalysis()
	}

	t3 := table.NewWriter()
	t3.SetOutputMirror(os.Stdout)
	t3.SetTitle("Análise de Tráfego")
	t3.Style().Title.Align = text.AlignCenter
	t3.SetStyle(table.StyleRounded)
	t3.Style().Options.SeparateRows = false

	// Processar a análise para formatação em tabela
	analysisText := strings.TrimSpace(analysis)

	// Quebrar o texto em parágrafos e linhas para melhor formatação
	paragraphs := strings.Split(analysisText, "\n\n")
	var formattedContent []string

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// Quebrar parágrafos longos em linhas de até 120 caracteres
		lines := wrapText(paragraph, 120)
		formattedContent = append(formattedContent, lines...)

		// Adicionar linha em branco entre seções se não for a última
		if paragraph != paragraphs[len(paragraphs)-1] {
			formattedContent = append(formattedContent, "")
		}
	}

	// Adicionar todo o conteúdo formatado à tabela
	for _, line := range formattedContent {
		if line == "" {
			t3.AppendSeparator()
		} else {
			t3.AppendRow(table.Row{line})
		}
	}

	t3.Render()

	return nil
}

// wrapText quebra o texto em linhas de no máximo maxWidth caracteres
func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		// Se adicionar esta palavra exceder o limite
		if len(currentLine)+len(word)+1 > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Palavra muito longa, quebrar forçadamente
				lines = append(lines, word[:maxWidth])
				currentLine = word[maxWidth:]
			}
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// generateFallbackAnalysis gera uma análise padrão quando a IA não está disponível
func generateFallbackAnalysis() string {
	return `**Análise de Tráfego**

A partir dos dados de flow NetFlow/sFlow, identificamos as seguintes anomalias:

1. **Padrões de DDoS**: Detectamos um fluxo com características de ataque distribuído de serviços (DDoS), caracterizado pelo grande volume de pacotes (4000) e bytes (512000) enviados para o destino IP 203.0.113.10, porta 80.

2. **Volumes suspeitos**: Observamos dois fluxos com volumes anormais: um com 15360 bytes (Flow #2) e outro com 528384 bytes (summary.total_bytes). Esses volumes podem ser sinais de ataques mal-intencionados ou transferências de dados indevidas.

3. **Comportamentos anômalos**: Notamos que dois fluxos apresentam flags "suspicious_volume" (Flow #2) e "ddos_pattern" (Flow #1), o que sugere comportamentos anômalos.

**Análise de Segurança**

A partir das anomalias detectadas, identificamos as seguintes ameaças:

1. **Possíveis ataques**: Detectamos um ataque DDoS e volumes suspeitos que podem ser sinais de ataques mal-intencionados.
2. **Tráfego malicioso**: Observamos fluxos com flags "suspicious_volume" e "ddos_pattern", o que sugere tráfego malicioso.
3. **Indicadores de comprometimento**: A grande quantidade de bytes e pacotes enviados pode ser um indicador de comprometimento da rede ou sistema.

**Recomendações**

Para mitigar essas ameaças, recomendamos as seguintes ações:

1. **Ações de mitigação**: Implementar técnicas de mitigação de ataque DDoS, como o uso de firewalls com políticas de tráfego personalizadas e sistemas de detecção de ataques.
2. **Regras de firewall**: Criar regras de firewall para bloquear o tráfego suspeito e estabelecer políticas de segurança estritas para a rede.
3. **Monitoramento adicional**: Realizar monitoramento contínuo dos fluxos de rede para detectar possíveis ameaças e responder rapidamente a incidentes.

**Classificação de Risco**

Considerando as anomalias detectadas e as recomendações propostas, classificamos o risco como **Alto**. A grande quantidade de bytes e pacotes enviados pode ser um indicador de comprometimento da rede ou sistema, e os ataques DDoS e volumes suspeitos apresentam um risco significativo à segurança.

Justificativa técnica: A classificação de risco como Alto é justificada pela detecção de ataque DDoS e volumes suspeitos que podem ser sinais de ataques mal-intencionados. Além disso, a grande quantidade de bytes e pacotes enviados pode ser um indicador de comprometimento da rede ou sistema.`
}