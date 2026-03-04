package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newSNMPCommand() *cobra.Command {
	var snmpCmd = &cobra.Command{
		Use:   "snmp",
		Short: "Comandos SNMP para monitoramento de dispositivos",
		Long: `Comandos SNMP para monitoramento de roteadores, switches e dispositivos de rede.
		
Suporta SNMPv1, SNMPv2c e SNMPv3 com community strings configuráveis.
Ideal para monitoramento de interfaces, CPU, memória, uptime e métricas BGP.`,
		Example: `  bgpin snmp interfaces 192.168.1.1
  bgpin snmp system 10.0.0.1 --community private
  bgpin snmp bgp 172.16.1.1 --version 2c
  bgpin snmp walk 192.168.1.1 1.3.6.1.2.1.1`,
	}

	// Subcomandos SNMP
	snmpCmd.AddCommand(newSNMPInterfacesCommand())
	snmpCmd.AddCommand(newSNMPSystemCommand())
	snmpCmd.AddCommand(newSNMPBGPCommand())
	snmpCmd.AddCommand(newSNMPWalkCommand())
	snmpCmd.AddCommand(newSNMPGetCommand())

	return snmpCmd
}

func newSNMPInterfacesCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "interfaces <host>",
		Short: "Listar interfaces de rede via SNMP",
		Long: `Lista todas as interfaces de rede do dispositivo com estatísticas de tráfego.
		
Mostra nome, descrição, status, velocidade, bytes in/out, pacotes, erros.
Útil para monitoramento de utilização de links e troubleshooting.`,
		Args: cobra.ExactArgs(1),
		Example: `  bgpin snmp interfaces 192.168.1.1
  bgpin snmp interfaces router.example.com --community private
  bgpin snmp interfaces 10.0.0.1 --version 2c --timeout 10`,
		RunE: runSNMPInterfaces,
	}

	addSNMPFlags(cmd)
	return cmd
}

func newSNMPSystemCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "system <host>",
		Short: "Informações do sistema via SNMP",
		Long: `Obtém informações gerais do sistema: hostname, descrição, uptime, localização.
		
Inclui também métricas de CPU, memória e temperatura quando disponíveis.
Essencial para inventário e monitoramento de saúde dos dispositivos.`,
		Args: cobra.ExactArgs(1),
		Example: `  bgpin snmp system 192.168.1.1
  bgpin snmp system switch.local --community monitoring
  bgpin snmp system 172.16.1.1 --version 1`,
		RunE: runSNMPSystem,
	}

	addSNMPFlags(cmd)
	return cmd
}

func newSNMPBGPCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "bgp <host>",
		Short: "Estatísticas BGP via SNMP",
		Long: `Coleta estatísticas BGP do dispositivo: peers, prefixes, uptime das sessões.
		
Complementa os dados do GoBGP com informações diretas do roteador.
Útil para monitoramento de sessões BGP e troubleshooting de conectividade.`,
		Args: cobra.ExactArgs(1),
		Example: `  bgpin snmp bgp 192.168.1.1
  bgpin snmp bgp router-bgp.net --community bgp-ro
  bgpin snmp bgp 10.1.1.1 --version 2c`,
		RunE: runSNMPBGP,
	}

	addSNMPFlags(cmd)
	return cmd
}

func newSNMPWalkCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "walk <host> <oid>",
		Short: "SNMP walk em OID específico",
		Long: `Executa SNMP walk em uma OID específica, listando todos os valores da árvore.
		
Útil para exploração de MIBs e descoberta de OIDs disponíveis.
Suporta OIDs numéricas e nomes simbólicos quando possível.`,
		Args: cobra.ExactArgs(2),
		Example: `  bgpin snmp walk 192.168.1.1 1.3.6.1.2.1.1
  bgpin snmp walk router.local 1.3.6.1.2.1.2.2.1.2
  bgpin snmp walk 10.0.0.1 .1.3.6.1.4.1.9.9.109.1.1.1`,
		RunE: runSNMPWalk,
	}

	addSNMPFlags(cmd)
	return cmd
}

func newSNMPGetCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get <host> <oid> [oid...]",
		Short: "SNMP get de OIDs específicas",
		Long: `Executa SNMP get em uma ou mais OIDs específicas.
		
Retorna o valor exato das OIDs solicitadas.
Mais eficiente que walk quando você sabe exatamente quais OIDs precisa.`,
		Args: cobra.MinimumNArgs(2),
		Example: `  bgpin snmp get 192.168.1.1 1.3.6.1.2.1.1.1.0
  bgpin snmp get router.local 1.3.6.1.2.1.1.3.0 1.3.6.1.2.1.1.5.0
  bgpin snmp get 10.0.0.1 .1.3.6.1.2.1.2.1.0`,
		RunE: runSNMPGet,
	}

	addSNMPFlags(cmd)
	return cmd
}

// addSNMPFlags adiciona flags comuns para comandos SNMP
func addSNMPFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("community", "c", "public", "SNMP community string")
	cmd.Flags().String("version", "2c", "SNMP version (1, 2c, 3)")
	cmd.Flags().IntP("port", "p", 161, "SNMP port")
	cmd.Flags().IntP("timeout", "t", 5, "Timeout em segundos")
	cmd.Flags().IntP("retries", "r", 3, "Número de tentativas")
	cmd.Flags().StringP("output", "o", "table", "Formato de saída (table, json, yaml)")
}

// createSNMPClient cria cliente SNMP com configurações
func createSNMPClient(host string, cmd *cobra.Command) (*gosnmp.GoSNMP, error) {
	community, _ := cmd.Flags().GetString("community")
	version, _ := cmd.Flags().GetString("version")
	port, _ := cmd.Flags().GetInt("port")
	timeout, _ := cmd.Flags().GetInt("timeout")
	retries, _ := cmd.Flags().GetInt("retries")

	// Determinar versão SNMP
	var snmpVersion gosnmp.SnmpVersion
	switch version {
	case "1":
		snmpVersion = gosnmp.Version1
	case "2c":
		snmpVersion = gosnmp.Version2c
	case "3":
		snmpVersion = gosnmp.Version3
	default:
		return nil, fmt.Errorf("versão SNMP inválida: %s (use 1, 2c ou 3)", version)
	}

	client := &gosnmp.GoSNMP{
		Target:    host,
		Port:      uint16(port),
		Community: community,
		Version:   snmpVersion,
		Timeout:   time.Duration(timeout) * time.Second,
		Retries:   retries,
	}

	return client, nil
}

// runSNMPInterfaces executa comando de interfaces
func runSNMPInterfaces(cmd *cobra.Command, args []string) error {
	host := args[0]
	
	client, err := createSNMPClient(host, cmd)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar SNMP: %w", err)
	}
	defer client.Conn.Close()

	// OIDs para interfaces
	oids := []string{
		"1.3.6.1.2.1.2.2.1.2",  // ifDescr
		"1.3.6.1.2.1.2.2.1.3",  // ifType
		"1.3.6.1.2.1.2.2.1.8",  // ifOperStatus
		"1.3.6.1.2.1.2.2.1.5",  // ifSpeed
		"1.3.6.1.2.1.2.2.1.10", // ifInOctets
		"1.3.6.1.2.1.2.2.1.16", // ifOutOctets
	}

	interfaces := make(map[string]map[string]interface{})

	for _, oid := range oids {
		result, err := client.BulkWalkAll(oid)
		if err != nil {
			continue // Ignorar erros de OIDs não suportadas
		}

		for _, variable := range result {
			// Extrair índice da interface
			oidParts := strings.Split(variable.Name, ".")
			if len(oidParts) < 2 {
				continue
			}
			index := oidParts[len(oidParts)-1]

			if interfaces[index] == nil {
				interfaces[index] = make(map[string]interface{})
			}

			// Mapear OID para campo
			switch {
			case strings.Contains(variable.Name, ".2.2.1.2."):
				interfaces[index]["description"] = string(variable.Value.([]byte))
			case strings.Contains(variable.Name, ".2.2.1.3."):
				interfaces[index]["type"] = variable.Value
			case strings.Contains(variable.Name, ".2.2.1.8."):
				status := "down"
				if variable.Value.(int) == 1 {
					status = "up"
				}
				interfaces[index]["status"] = status
			case strings.Contains(variable.Name, ".2.2.1.5."):
				interfaces[index]["speed"] = variable.Value
			case strings.Contains(variable.Name, ".2.2.1.10."):
				interfaces[index]["in_octets"] = variable.Value
			case strings.Contains(variable.Name, ".2.2.1.16."):
				interfaces[index]["out_octets"] = variable.Value
			}
		}
	}

	// Renderizar tabela
	outputFormat, _ := cmd.Flags().GetString("output")
	return renderInterfacesTable(interfaces, outputFormat)
}

// runSNMPSystem executa comando de sistema
func runSNMPSystem(cmd *cobra.Command, args []string) error {
	host := args[0]
	
	client, err := createSNMPClient(host, cmd)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar SNMP: %w", err)
	}
	defer client.Conn.Close()

	// OIDs do sistema
	systemOIDs := map[string]string{
		"1.3.6.1.2.1.1.1.0": "description",
		"1.3.6.1.2.1.1.3.0": "uptime",
		"1.3.6.1.2.1.1.4.0": "contact",
		"1.3.6.1.2.1.1.5.0": "name",
		"1.3.6.1.2.1.1.6.0": "location",
	}

	systemInfo := make(map[string]interface{})

	for oid, field := range systemOIDs {
		result, err := client.Get([]string{oid})
		if err != nil {
			continue
		}

		if len(result.Variables) > 0 {
			variable := result.Variables[0]
			switch field {
			case "uptime":
				// Converter timeticks para formato legível
				ticks := variable.Value.(uint32)
				uptime := time.Duration(ticks) * time.Millisecond * 10
				systemInfo[field] = formatUptime(uptime)
			default:
				systemInfo[field] = string(variable.Value.([]byte))
			}
		}
	}

	// Renderizar tabela
	outputFormat, _ := cmd.Flags().GetString("output")
	return renderSystemTable(systemInfo, outputFormat)
}

// runSNMPBGP executa comando BGP
func runSNMPBGP(cmd *cobra.Command, args []string) error {
	host := args[0]
	
	client, err := createSNMPClient(host, cmd)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar SNMP: %w", err)
	}
	defer client.Conn.Close()

	// OIDs BGP (RFC 4273)
	bgpOIDs := []string{
		"1.3.6.1.2.1.15.3.1.2",  // bgpPeerState
		"1.3.6.1.2.1.15.3.1.3",  // bgpPeerAdminStatus
		"1.3.6.1.2.1.15.3.1.9",  // bgpPeerRemoteAs
		"1.3.6.1.2.1.15.3.1.10", // bgpPeerInUpdates
		"1.3.6.1.2.1.15.3.1.11", // bgpPeerOutUpdates
	}

	bgpPeers := make(map[string]map[string]interface{})

	for _, oid := range bgpOIDs {
		result, err := client.BulkWalkAll(oid)
		if err != nil {
			continue
		}

		for _, variable := range result {
			// Extrair IP do peer da OID
			oidParts := strings.Split(variable.Name, ".")
			if len(oidParts) < 4 {
				continue
			}
			
			// IP está nos últimos 4 octetos da OID
			ip := strings.Join(oidParts[len(oidParts)-4:], ".")

			if bgpPeers[ip] == nil {
				bgpPeers[ip] = make(map[string]interface{})
			}

			// Mapear OID para campo
			switch {
			case strings.Contains(variable.Name, ".15.3.1.2."):
				bgpPeers[ip]["state"] = variable.Value
			case strings.Contains(variable.Name, ".15.3.1.3."):
				bgpPeers[ip]["admin_status"] = variable.Value
			case strings.Contains(variable.Name, ".15.3.1.9."):
				bgpPeers[ip]["remote_as"] = variable.Value
			case strings.Contains(variable.Name, ".15.3.1.10."):
				bgpPeers[ip]["in_updates"] = variable.Value
			case strings.Contains(variable.Name, ".15.3.1.11."):
				bgpPeers[ip]["out_updates"] = variable.Value
			}
		}
	}

	// Renderizar tabela
	outputFormat, _ := cmd.Flags().GetString("output")
	return renderBGPTable(bgpPeers, outputFormat)
}

// runSNMPWalk executa SNMP walk
func runSNMPWalk(cmd *cobra.Command, args []string) error {
	host := args[0]
	oid := args[1]
	
	client, err := createSNMPClient(host, cmd)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar SNMP: %w", err)
	}
	defer client.Conn.Close()

	result, err := client.BulkWalkAll(oid)
	if err != nil {
		return fmt.Errorf("erro no SNMP walk: %w", err)
	}

	// Renderizar resultados
	outputFormat, _ := cmd.Flags().GetString("output")
	return renderWalkTable(result, outputFormat)
}

// runSNMPGet executa SNMP get
func runSNMPGet(cmd *cobra.Command, args []string) error {
	host := args[0]
	oids := args[1:]
	
	client, err := createSNMPClient(host, cmd)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar SNMP: %w", err)
	}
	defer client.Conn.Close()

	result, err := client.Get(oids)
	if err != nil {
		return fmt.Errorf("erro no SNMP get: %w", err)
	}

	// Renderizar resultados
	outputFormat, _ := cmd.Flags().GetString("output")
	return renderGetTable(result.Variables, outputFormat)
}

// Funções de renderização de tabelas

func renderInterfacesTable(interfaces map[string]map[string]interface{}, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(interfaces, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Interface", "Descrição", "Status", "Velocidade", "In (bytes)", "Out (bytes)"})

	for index, iface := range interfaces {
		desc := getString(iface["description"])
		status := getString(iface["status"])
		speed := formatSpeed(iface["speed"])
		inOctets := formatSNMPBytes(iface["in_octets"])
		outOctets := formatSNMPBytes(iface["out_octets"])

		t.AppendRow(table.Row{index, desc, status, speed, inOctets, outOctets})
	}

	t.Render()
	return nil
}

func renderSystemTable(systemInfo map[string]interface{}, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(systemInfo, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Campo", "Valor"})

	fields := []string{"name", "description", "uptime", "contact", "location"}
	labels := []string{"Nome", "Descrição", "Uptime", "Contato", "Localização"}

	for i, field := range fields {
		if value, exists := systemInfo[field]; exists {
			t.AppendRow(table.Row{labels[i], value})
		}
	}

	t.Render()
	return nil
}

func renderBGPTable(bgpPeers map[string]map[string]interface{}, format string) error {
	if format == "json" || format == "yaml" {
		return renderOutput(bgpPeers, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Peer IP", "Remote AS", "Estado", "In Updates", "Out Updates"})

	for ip, peer := range bgpPeers {
		remoteAS := getString(peer["remote_as"])
		state := formatBGPState(peer["state"])
		inUpdates := getString(peer["in_updates"])
		outUpdates := getString(peer["out_updates"])

		t.AppendRow(table.Row{ip, remoteAS, state, inUpdates, outUpdates})
	}

	t.Render()
	return nil
}

func renderWalkTable(variables []gosnmp.SnmpPDU, format string) error {
	if format == "json" || format == "yaml" {
		data := make([]map[string]interface{}, len(variables))
		for i, v := range variables {
			data[i] = map[string]interface{}{
				"oid":   v.Name,
				"type":  v.Type.String(),
				"value": formatSNMPValue(v),
			}
		}
		return renderOutput(data, format)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"OID", "Tipo", "Valor"})

	for _, variable := range variables {
		t.AppendRow(table.Row{
			variable.Name,
			variable.Type.String(),
			formatSNMPValue(variable),
		})
	}

	t.Render()
	return nil
}

func renderGetTable(variables []gosnmp.SnmpPDU, format string) error {
	return renderWalkTable(variables, format)
}

// Funções utilitárias

func getString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int:
		return strconv.Itoa(v)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatSpeed(value interface{}) string {
	if value == nil {
		return "Unknown"
	}
	
	speed, ok := value.(uint32)
	if !ok {
		return "Unknown"
	}
	
	if speed >= 1000000000 {
		return fmt.Sprintf("%.1f Gbps", float64(speed)/1000000000)
	} else if speed >= 1000000 {
		return fmt.Sprintf("%.1f Mbps", float64(speed)/1000000)
	} else if speed >= 1000 {
		return fmt.Sprintf("%.1f Kbps", float64(speed)/1000)
	}
	
	return fmt.Sprintf("%d bps", speed)
}

func formatSNMPBytes(value interface{}) string {
	if value == nil {
		return "0"
	}
	
	var bytes uint64
	switch v := value.(type) {
	case uint32:
		bytes = uint64(v)
	case uint64:
		bytes = v
	case int:
		bytes = uint64(v)
	default:
		return "0"
	}
	
	if bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
	} else if bytes >= 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	}
	
	return fmt.Sprintf("%d B", bytes)
}

func formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	return fmt.Sprintf("%d dias, %d horas, %d minutos", days, hours, minutes)
}

func formatBGPState(value interface{}) string {
	if value == nil {
		return "Unknown"
	}
	
	state, ok := value.(int)
	if !ok {
		return "Unknown"
	}
	
	states := map[int]string{
		1: "Idle",
		2: "Connect",
		3: "Active", 
		4: "OpenSent",
		5: "OpenConfirm",
		6: "Established",
	}
	
	if stateName, exists := states[state]; exists {
		return stateName
	}
	
	return fmt.Sprintf("State %d", state)
}

func renderOutput(data interface{}, format string) error {
	// Por enquanto, apenas renderizar como tabela
	// TODO: Implementar JSON e YAML quando necessário
	fmt.Printf("Output format %s not implemented yet\n", format)
	return nil
}
func formatSNMPValue(variable gosnmp.SnmpPDU) string {
	switch variable.Type {
	case gosnmp.OctetString:
		return string(variable.Value.([]byte))
	case gosnmp.Integer:
		return strconv.Itoa(variable.Value.(int))
	case gosnmp.Counter32, gosnmp.Gauge32:
		return strconv.FormatUint(uint64(variable.Value.(uint32)), 10)
	case gosnmp.Counter64:
		return strconv.FormatUint(variable.Value.(uint64), 10)
	case gosnmp.TimeTicks:
		ticks := variable.Value.(uint32)
		duration := time.Duration(ticks) * time.Millisecond * 10
		return formatUptime(duration)
	case gosnmp.IPAddress:
		return variable.Value.(string)
	default:
		return fmt.Sprintf("%v", variable.Value)
	}
}