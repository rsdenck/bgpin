package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Aplicar configurações e iniciar interfaces",
		Long:  "Comandos para aplicar configurações e iniciar interfaces interativas",
	}

	cmd.AddCommand(newApplyTUICommand())

	return cmd
}

func newApplyTUICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Iniciar interface TUI de monitoramento BGP/ASN",
		Long:  "Interface de monitoramento em tempo real estilo BTOP para BGP e ASN",
		RunE:  runApplyTUI,
		Example: `  bgpin apply tui
  # Inicia interface interativa de monitoramento
  # Use setas para navegar, 'q' para sair
  # Monitora BGP, ASN, flows e anomalias em tempo real`,
	}
}

func runApplyTUI(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Iniciando TUI Monitor BGP/ASN...")
	fmt.Println("📊 Interface de monitoramento em tempo real")
	fmt.Println("⚡ Estilo BTOP para redes BGP")
	fmt.Println("")
	fmt.Println("🔧 TUI Monitor em desenvolvimento!")
	fmt.Println("📋 Funcionalidades planejadas:")
	fmt.Println("   • Dashboard interativo BGP/ASN")
	fmt.Println("   • Monitoramento de flows em tempo real")
	fmt.Println("   • Visualização de anomalias")
	fmt.Println("   • Navegação por teclado")
	fmt.Println("   • Métricas de performance")
	fmt.Println("   • Alertas visuais")
	fmt.Println("")
	fmt.Println("💡 Use 'bgpin ai flow' para análise de flows com IA")
	fmt.Println("💡 Use 'bgpin flow top' para top flows atual")
	
	return nil
}