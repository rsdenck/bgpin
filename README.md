<div align="center">

# Platform Border Gateway Protocol

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/rsdenck/bgpin.svg)](https://github.com/rsdenck/bgpin/releases)
[![Tests](https://img.shields.io/badge/tests-passing-success.svg)](https://github.com/rsdenck/bgpin)

CLI profissional + SDK + Telemetria + Flow Analysis

</div>

---

## ðŸš€ CaracterÃ­sticas

### âœ… CLI Completa
- Consulta informaÃ§Ãµes de ASN (holder, block, status)
- Lista vizinhos BGP (upstream, downstream, peers)
- Mostra prefixos anunciados (IPv4 e IPv6)
- Visualiza RIS peers por RRC
- AnÃ¡lise de prefixos especÃ­ficos
- MÃºltiplos formatos de saÃ­da (table, JSON, YAML)
- UX profissional com go-pretty/table

### âœ… SDK RIPE RIS
- Rate limiting configurÃ¡vel
- Retry com exponential backoff
- Context support completo
- Testes de integraÃ§Ã£o reais (sem mocks)
- Thread-safe para requisiÃ§Ãµes concorrentes

### âœ… Telemetria & Observabilidade
- OpenTelemetry integration
- Distributed tracing
- MÃ©tricas de performance
- ExportaÃ§Ã£o para OTLP, Jaeger, Prometheus
- Dashboards Grafana

### âœ… Flow Analysis
- Coleta NetFlow/sFlow/IPFIX em tempo real
- AnÃ¡lise de trÃ¡fego por ASN e prefixo
- DetecÃ§Ã£o de anomalias (DDoS, spikes, drops)
- ComparaÃ§Ã£o de upstreams
- CorrelaÃ§Ã£o BGP + Flow
- AgregaÃ§Ã£o configurÃ¡vel
- Worker pools para alta performance

### ðŸ”„ Em Desenvolvimento
- Consulta mÃºltiplos Looking Glass
- Parsing estruturado de mÃºltiplos vendors (Cisco, Juniper, FRR)
- ValidaÃ§Ã£o RPKI
- Cache inteligente
- Machine Learning para anomalias

## ðŸ“¦ InstalaÃ§Ã£o

```bash
# Clone o repositÃ³rio
git clone https://github.com/bgpin/bgpin
cd bgpin

# Compile a CLI
go build -o bgpin ./cmd/cli/

# Ou instale globalmente
go install ./cmd/cli/
```

## ðŸŽ¯ Uso RÃ¡pido

### CLI

```bash
# InformaÃ§Ãµes de um ASN
bgpin asn info 262978

# Vizinhos BGP
bgpin asn neighbors 262978

# Prefixos anunciados
bgpin asn prefixes 262978

# RIS peers
bgpin asn peers 262978

# AnÃ¡lise de prefixo
bgpin prefix overview 186.250.184.0/24

# Flow telemetry (requer configuraÃ§Ã£o)
bgpin flow top                    # Top prefixes por trÃ¡fego
bgpin flow asn 15169              # EstatÃ­sticas de trÃ¡fego do ASN
bgpin flow anomaly                # Detectar anomalias de trÃ¡fego
bgpin flow upstream-compare       # Comparar upstreams
bgpin flow stats                  # EstatÃ­sticas do coletor

# Formato JSON
bgpin asn info 262978 -o json

# Listar Looking Glasses
bgpin lg

# VersÃ£o
bgpin version
```

### SDK

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/bgpin/bgpin/sdk"
)

func main() {
    // Criar cliente
    client := sdk.NewDefaultClient()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Obter informaÃ§Ãµes do ASN
    info, err := client.GetASNInfo(ctx, 262978)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("ASN: %d\n", info.ASN)
    fmt.Printf("Holder: %s\n", info.Holder)
    fmt.Printf("Announced: %v\n", info.Announced)
}
```

## ðŸ“– DocumentaÃ§Ã£o

- [Installation Guide](docs/INSTALLATION.md) - Guia completo de instalaÃ§Ã£o
- [Quick Start](docs/QUICK_START.md) - InÃ­cio rÃ¡pido
- [CLI Guide](docs/CLI_GUIDE.md) - Todos os comandos e exemplos
- [Flow Collector Guide](docs/FLOW_COLLECTOR.md) - NetFlow/sFlow/IPFIX setup e uso
- [Telemetria](docs/TELEMETRY.md) - OpenTelemetry integration
- [Arquitetura](docs/ARCHITECTURE.md) - Design e estrutura do projeto
- [SDK README](sdk/README.md) - DocumentaÃ§Ã£o do SDK
- [Exemplos de Output](docs/OUTPUT_EXAMPLES.md) - Exemplos visuais
- [Vendors Status](docs/vendors/STATUS.md) - Status de implementaÃ§Ã£o dos vendors
- [Project Summary](docs/PROJECT_SUMMARY.md) - Resumo do projeto

### Releases
- [v0.2.0](docs/releases/v0.2.0.md) - Flow collector + Cisco/Juniper parsers
- [v0.1.0](docs/releases/v0.1.0.md) - Initial release

## ðŸŽ¨ Exemplos de SaÃ­da

### InformaÃ§Ãµes de ASN (Formato Tabela)
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ ASN Information: AS262978                                 â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Holder    â”‚ Centro de Tecnologia Armazem Datacenter Ltda. â”‚
â”‚ Announced â”‚ true                                          â”‚
â”‚ Block     â”‚ 262144-263167                                 â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### Prefixos Anunciados
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ Announced Prefixes for AS262978 (Total: 19)   â”‚
â”œâ”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”¤
â”‚  # â”‚ PREFIX             â”‚ TYPE â”‚
â”œâ”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”¤
â”‚  1 â”‚ 2804:4d44:10::/48  â”‚ IPv6 â”‚
â”‚  2 â”‚ 143.0.121.0/24     â”‚ IPv4 â”‚
â”‚  3 â”‚ 186.250.187.0/24   â”‚ IPv4 â”‚
â”‚  4 â”‚ 186.250.184.0/24   â”‚ IPv4 â”‚
â”‚  5 â”‚ 2804:4d44:c::/48   â”‚ IPv6 â”‚
...
â•°â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â•¯
```

### Looking Glasses
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ Available Looking Glasses                                                 â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ NAME               â”‚ VENDOR  â”‚ TYPE   â”‚ PROTOCOL â”‚ COUNTRY â”‚ URL          â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Hurricane Electric â”‚ cisco   â”‚ public â”‚ http     â”‚ US      â”‚ lg.he.net    â”‚
â”‚ NTT America        â”‚ cisco   â”‚ public â”‚ http     â”‚ US      â”‚ lg.ntt.net   â”‚
â”‚ Telia Carrier      â”‚ juniper â”‚ public â”‚ http     â”‚ SE      â”‚ lg.telia.net â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### Formato JSON
```json
{
  "asn": 262978,
  "holder": "Centro de Tecnologia Armazem Datacenter Ltda.",
  "announced": true,
  "block": "262144-263167"
}
```

## ðŸ§ª Testes

### SDK - Testes de IntegraÃ§Ã£o
Todos os testes usam dados reais do ASN 262978 (sem mocks):

```bash
# Executar todos os testes
go test -v ./sdk/integration_test/

# Teste especÃ­fico
go test -v ./sdk/integration_test/ -run TestGetASNInfo_262978

# Executar exemplo
go run sdk/examples/demo.go
```

### Resultados dos Testes
```
âœ… TestGetASNInfo_262978
âœ… TestGetASNNeighbors_262978 (34 vizinhos)
âœ… TestGetAnnouncedPrefixes_262978 (19 prefixos)
âœ… TestGetPrefixOverview_262978
âœ… TestGetRISPeers_262978 (1449 peers)
âœ… TestRateLimiting
âœ… TestRetryOnError
âœ… TestContextTimeout
âœ… TestConcurrentRequests
```

## Arquitetura

```
bgpin/
â”œâ”€â”€ cmd/cli/              # CLI commands
â”œâ”€â”€ internal/
â”‚   â”œâ”€â”€ adapters/         # HTTP, SSH, Telnet adapters
â”‚   â”œâ”€â”€ core/             # Domain logic (BGP, AS Path, RPKI)
â”‚   â”œâ”€â”€ parsers/          # Vendor-specific parsers
â”‚   â””â”€â”€ services/         # Business logic
â”œâ”€â”€ pkg/
â”‚   â”œâ”€â”€ config/           # Configuration
â”‚   â””â”€â”€ telemetry/        # OpenTelemetry (futuro)
â””â”€â”€ sdk/                  # RIPE RIS SDK
    â”œâ”€â”€ client.go         # Cliente principal
    â”œâ”€â”€ config.go         # ConfiguraÃ§Ã£o
    â”œâ”€â”€ types.go          # Tipos de dados
    â”œâ”€â”€ errors.go         # Erros customizados
    â”œâ”€â”€ rate_limit.go     # Rate limiting
    â”œâ”€â”€ retry.go          # Retry logic
    â”œâ”€â”€ integration_test/ # Testes de integraÃ§Ã£o
    â””â”€â”€ examples/         # Exemplos de uso
```

## Tecnologias

- **Linguagem**: Go 1.25+
- **CLI Framework**: Cobra
- **Config**: Viper
- **Rate Limiting**: golang.org/x/time/rate
- **HTTP Client**: net/http (stdlib)
- **Testing**: Testes de integraÃ§Ã£o reais (sem mocks)

## Roadmap

- [x] SDK RIPE RIS completo
- [x] Rate limiting e retry
- [x] Testes de integraÃ§Ã£o
- [ ] CLI completa
- [ ] Suporte a mÃºltiplos LG (HTTP, Telnet, SSH)
- [ ] Parsers para Cisco, Juniper, FRR
- [ ] AnÃ¡lise de anomalias BGP
- [ ] ValidaÃ§Ã£o RPKI
- [ ] OpenTelemetry
- [ ] Cache inteligente
- [ ] Modo interativo (TUI)

## Exemplos de Uso

### Obter informaÃ§Ãµes completas de um ASN
```bash
go run sdk/examples/demo.go
```

### Executar testes de integraÃ§Ã£o
```bash
go test -v ./sdk/integration_test/
```

## Contribuindo

ContribuiÃ§Ãµes sÃ£o bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanÃ§as (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## LicenÃ§a

MIT License - veja o arquivo LICENSE para detalhes

## Autor

bgpin team

## Agradecimentos

- RIPE NCC pela API RIS
- Comunidade BGP
- ASN 262978 (Centro de Tecnologia Armazem Datacenter Ltda.) usado nos testes


## ðŸ—ï¸ Arquitetura

```
bgpin/
â”œâ”€â”€ cmd/cli/              # CLI commands
â”‚   â”œâ”€â”€ root.go          # Root command
â”‚   â”œâ”€â”€ asn.go           # ASN commands
â”‚   â”œâ”€â”€ prefix.go        # Prefix commands
â”‚   â””â”€â”€ version.go       # Version info
â”œâ”€â”€ internal/
â”‚   â”œâ”€â”€ adapters/        # HTTP, SSH, Telnet adapters
â”‚   â”œâ”€â”€ core/            # Domain logic (BGP, AS Path, RPKI)
â”‚   â”œâ”€â”€ parsers/         # Vendor-specific parsers
â”‚   â””â”€â”€ services/        # Business logic
â”œâ”€â”€ pkg/
â”‚   â”œâ”€â”€ config/          # Configuration
â”‚   â””â”€â”€ telemetry/       # OpenTelemetry (futuro)
â””â”€â”€ sdk/                 # RIPE RIS SDK
    â”œâ”€â”€ client.go        # Main client
    â”œâ”€â”€ config.go        # Configuration
    â”œâ”€â”€ types.go         # Data types
    â”œâ”€â”€ errors.go        # Custom errors
    â”œâ”€â”€ rate_limit.go    # Rate limiting
    â”œâ”€â”€ retry.go         # Retry logic
    â”œâ”€â”€ integration_test/
    â””â”€â”€ examples/
```

## ðŸ› ï¸ Tecnologias

- **Linguagem**: Go 1.25+
- **CLI Framework**: Cobra
- **Config**: Viper
- **Rate Limiting**: golang.org/x/time/rate
- **HTTP Client**: net/http (stdlib)
- **Testing**: Testes de integraÃ§Ã£o reais (sem mocks)

## ðŸ“‹ Comandos CLI

### ASN Commands
```bash
bgpin asn info [asn]         # InformaÃ§Ãµes do ASN
bgpin asn neighbors [asn]    # Vizinhos BGP
bgpin asn prefixes [asn]     # Prefixos anunciados
bgpin asn peers [asn]        # RIS peers
```

### Prefix Commands
```bash
bgpin prefix overview [prefix]  # VisÃ£o geral do prefixo
```

### Flow Commands (Preparado)
```bash
bgpin flow top                  # Top prefixes por trÃ¡fego
bgpin flow asn [asn]           # EstatÃ­sticas de trÃ¡fego do ASN
bgpin flow anomaly             # Detectar anomalias de trÃ¡fego
bgpin flow upstream-compare    # Comparar upstreams
```

### Utility Commands
```bash
bgpin lg          # Listar Looking Glasses
bgpin version     # InformaÃ§Ãµes de versÃ£o
```

### Flags Globais
```bash
-o, --output string   # Formato: table, json, yaml (default: table)
-t, --timeout int     # Timeout em segundos (default: 30)
-v, --verbose         # Modo verbose
--config string       # Arquivo de configuraÃ§Ã£o
```

## ðŸ”§ ConfiguraÃ§Ã£o

Crie um arquivo `bgpin.yaml`:

```yaml
# ConfiguraÃ§Ãµes gerais
timeout: 30
output: table

# Cache
cache:
  enabled: true
  ttl: 300

# RIPE RIS SDK
ripe:
  rate_limit: 10
  retry_max: 3
  retry_wait_min: 1
  retry_wait_max: 10
```

## ðŸ’¡ Exemplos PrÃ¡ticos

### Investigar um ASN
```bash
# InformaÃ§Ãµes bÃ¡sicas
bgpin asn info 262978

# Ver todos os vizinhos
bgpin asn neighbors 262978

# Exportar prefixos para JSON
bgpin asn prefixes 262978 -o json > prefixes.json
```

### Pipeline com jq
```bash
# Extrair apenas o holder
bgpin asn info 262978 -o json | jq -r '.holder'

# Contar prefixos
bgpin asn prefixes 262978 -o json | jq '.prefixes | length'

# Filtrar apenas IPv4
bgpin asn prefixes 262978 -o json | jq -r '.prefixes[].prefix' | grep -v ':'
```

### AutomaÃ§Ã£o
```bash
#!/bin/bash
# Script para monitorar ASN

ASN=262978
OUTPUT_DIR="./reports"

mkdir -p $OUTPUT_DIR

echo "Collecting data for AS$ASN..."
bgpin asn info $ASN -o json > $OUTPUT_DIR/asn_info.json
bgpin asn prefixes $ASN -o json > $OUTPUT_DIR/prefixes.json
bgpin asn neighbors $ASN -o json > $OUTPUT_DIR/neighbors.json

echo "Data collected successfully!"
```

## ðŸ—ºï¸ Roadmap

- [x] SDK RIPE RIS completo
- [x] CLI bÃ¡sica com comandos ASN e Prefix
- [x] MÃºltiplos formatos de saÃ­da (table, JSON, YAML)
- [x] Rate limiting e retry
- [x] Testes de integraÃ§Ã£o
- [ ] Suporte a mÃºltiplos LG (HTTP, Telnet, SSH)
- [ ] Parsers para Cisco, Juniper, FRR
- [ ] AnÃ¡lise de anomalias BGP
- [ ] ValidaÃ§Ã£o RPKI
- [ ] OpenTelemetry
- [ ] Cache inteligente
- [ ] Modo interativo (TUI)

## ðŸ¤ Contribuindo

ContribuiÃ§Ãµes sÃ£o bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanÃ§as (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## ðŸ“ LicenÃ§a

MIT License - veja o arquivo LICENSE para detalhes

## ðŸ‘¥ Autor

bgpin team

## ðŸ™ Agradecimentos

- RIPE NCC pela API RIS
- Comunidade BGP
- ASN 262978 (Centro de Tecnologia Armazem Datacenter Ltda.) usado nos testes

## ðŸ“ž Suporte

- GitHub Issues: https://github.com/bgpin/bgpin/issues
- DocumentaÃ§Ã£o: [docs/](docs/)
- Email: support@bgpin.dev

---

**bgpin** - BGP Intelligence Tool ðŸŒ


## ðŸŽ¨ UX Profissional

A CLI usa a biblioteca `go-pretty/table` para outputs visuais profissionais:

- âœ… Bordas arredondadas Unicode (â•­â•®â•°â•¯)
- âœ… TÃ­tulos centralizados
- âœ… Linhas limpas sem separadores extras
- âœ… DetecÃ§Ã£o automÃ¡tica de IPv4/IPv6
- âœ… Truncamento inteligente com footer
- âœ… Largura dinÃ¢mica

Veja mais exemplos em [docs/OUTPUT_EXAMPLES.md](docs/OUTPUT_EXAMPLES.md)


## Overview

Plataforma completa em Golang para anÃ¡lise BGP, consulta de Looking Glass, telemetria de rede e correlaÃ§Ã£o com flow data (NetFlow/sFlow/IPFIX).
