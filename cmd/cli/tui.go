package main

import (
	"fmt"

	"github.com/bgpin/bgpin/internal/tui"
	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
)

func newTUICommand() *cobra.Command {
	var tuiCmd = &cobra.Command{
		Use:   "tui",
		Short: "Interface TUI moderna para monitoramento BGP em tempo real",
		Long: `Interface TUI moderna para monitoramento de BGP, rotas e sistema.
		
Conecta diretamente ao router via SSH para obter dados em tempo real.
Exibe informações de peers BGP, rotas, interfaces e sistema em uma interface moderna.

Recursos:
- Monitoramento BGP em tempo real
- Estatísticas de sistema e interfaces
- Dados obtidos diretamente do router via SSH
- Interface moderna com múltiplos painéis
- Auto-refresh configurável

Navegação:
- Tab/Shift+Tab: Alternar entre painéis
- q/Ctrl+C: Sair
- r: Refresh manual
- h/?: Ajuda`,
		Example: `  bgpin tui
  bgpin tui --router 192.168.1.1 --user admin --pass secret
  bgpin tui --refresh 5s
  bgpin tui --demo`,
		RunE: runTUI,
	}

	// Flags para conexão SSH
	tuiCmd.Flags().String("router", "192.168.0.1", "IP do router para conectar")
	tuiCmd.Flags().String("user", "adcoperador", "Usuário SSH")
	tuiCmd.Flags().String("pass", "1515qwd", "Senha SSH")
	tuiCmd.Flags().String("refresh", "2s", "Intervalo de refresh dos dados")
	tuiCmd.Flags().Bool("demo", false, "Modo demonstração com dados simulados")

	return tuiCmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Obter configurações
	routerIP, _ := cmd.Flags().GetString("router")
	username, _ := cmd.Flags().GetString("user")
	password, _ := cmd.Flags().GetString("pass")
	demo, _ := cmd.Flags().GetBool("demo")

	if demo {
		fmt.Println("Iniciando TUI em modo demonstração...")
		// Usar dados simulados
		routerIP = "demo"
		username = "demo"
		password = "demo"
	}

	// Criar nova TUI moderna
	modernTUI := tui.NewModernTUI(routerIP, username, password)

	// Iniciar programa
	p := tea.NewProgram(modernTUI, tea.WithAltScreen(), tea.WithMouseCellMotion())

	fmt.Printf("🌐 BGPIN MONITOR - Conectando ao router %s...\n", routerIP)
	fmt.Println("Pressione 'q' para sair, 'r' para refresh manual, Tab para navegar")

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar TUI: %w", err)
	}

	return nil
}