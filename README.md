# bgpin - Plataforma BGP Completa

**CLI profissional + SDK + Telemetria + Flow Analysis**

Plataforma completa em Golang para análise BGP, consulta de Looking Glass, telemetria de rede e correlação com flow data (NetFlow/sFlow/IPFIX).

## 🚀 Características

### ✅ CLI Completa
- Consulta informações de ASN (holder, block, status)
- Lista vizinhos BGP (upstream, downstream, peers)
- Mostra prefixos anunciados (IPv4 e IPv6)
- Visualiza RIS peers por RRC
- Análise de prefixos específicos
- Múltiplos formatos de saída (table, JSON, YAML)
- UX profissional com go-pretty/table

### ✅ SDK RIPE RIS
- Rate limiting configurável
- Retry com exponential backoff
- Context support completo
- Testes de integração reais (sem mocks)
- Thread-safe para requisições concorrentes

### ✅ Telemetria & Observabilidade
- OpenTelemetry integration
- Distributed tracing
- Métricas de performance
- Exportação para OTLP, Jaeger, Prometheus
- Dashboards Grafana

### ✅ Flow Analysis (Preparado)
- Coleta NetFlow/sFlow/IPFIX
- Análise de tráfego por ASN
- Detecção de anomalias (DDoS, spikes)
- Comparação de upstreams
- Correlação BGP + Flow

### 🔄 Em Desenvolvimento
- Consulta múltiplos Looking Glass
- Parsing estruturado de múltiplos vendors (Cisco, Juniper, FRR)
- Validação RPKI
- Cache inteligente
- Machine Learning para anomalias

## 📦 Instalação

```bash
# Clone o repositório
git clone https://github.com/bgpin/bgpin
cd bgpin

# Compile a CLI
go build -o bgpin ./cmd/cli/

# Ou instale globalmente
go install ./cmd/cli/
```

## 🎯 Uso Rápido

### CLI

```bash
# Informações de um ASN
bgpin asn info 262978

# Vizinhos BGP
bgpin asn neighbors 262978

# Prefixos anunciados
bgpin asn prefixes 262978

# RIS peers
bgpin asn peers 262978

# Análise de prefixo
bgpin prefix overview 186.250.184.0/24

# Formato JSON
bgpin asn info 262978 -o json

# Listar Looking Glasses
bgpin lg

# Versão
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
    
    // Obter informações do ASN
    info, err := client.GetASNInfo(ctx, 262978)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("ASN: %d\n", info.ASN)
    fmt.Printf("Holder: %s\n", info.Holder)
    fmt.Printf("Announced: %v\n", info.Announced)
}
```

## 📖 Documentação

- [Guia Completo da CLI](docs/CLI_GUIDE.md) - Todos os comandos e exemplos
- [Arquitetura](docs/ARCHITECTURE.md) - Design e estrutura do projeto
- [SDK README](sdk/README.md) - Documentação do SDK

## 🎨 Exemplos de Saída

### Informações de ASN (Formato Tabela)
```
╭───────────────────────────────────────────────────────────╮
│ ASN Information: AS262978                                 │
├───────────┬───────────────────────────────────────────────┤
│ Holder    │ Centro de Tecnologia Armazem Datacenter Ltda. │
│ Announced │ true                                          │
│ Block     │ 262144-263167                                 │
╰───────────┴───────────────────────────────────────────────╯
```

### Prefixos Anunciados
```
╭────────────────────────────────────────────────╮
│ Announced Prefixes for AS262978 (Total: 19)   │
├────┬────────────────────┬──────┤
│  # │ PREFIX             │ TYPE │
├────┼────────────────────┼──────┤
│  1 │ 2804:4d44:10::/48  │ IPv6 │
│  2 │ 143.0.121.0/24     │ IPv4 │
│  3 │ 186.250.187.0/24   │ IPv4 │
│  4 │ 186.250.184.0/24   │ IPv4 │
│  5 │ 2804:4d44:c::/48   │ IPv6 │
...
╰────┴────────────────────┴──────╯
```

### Looking Glasses
```
╭───────────────────────────────────────────────────────────────────────────╮
│ Available Looking Glasses                                                 │
├────────────────────┬─────────┬────────┬──────────┬─────────┬──────────────┤
│ NAME               │ VENDOR  │ TYPE   │ PROTOCOL │ COUNTRY │ URL          │
├────────────────────┼─────────┼────────┼──────────┼─────────┼──────────────┤
│ Hurricane Electric │ cisco   │ public │ http     │ US      │ lg.he.net    │
│ NTT America        │ cisco   │ public │ http     │ US      │ lg.ntt.net   │
│ Telia Carrier      │ juniper │ public │ http     │ SE      │ lg.telia.net │
╰────────────────────┴─────────┴────────┴──────────┴─────────┴──────────────╯
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

## 🧪 Testes

### SDK - Testes de Integração
Todos os testes usam dados reais do ASN 262978 (sem mocks):

```bash
# Executar todos os testes
go test -v ./sdk/integration_test/

# Teste específico
go test -v ./sdk/integration_test/ -run TestGetASNInfo_262978

# Executar exemplo
go run sdk/examples/demo.go
```

### Resultados dos Testes
```
✅ TestGetASNInfo_262978
✅ TestGetASNNeighbors_262978 (34 vizinhos)
✅ TestGetAnnouncedPrefixes_262978 (19 prefixos)
✅ TestGetPrefixOverview_262978
✅ TestGetRISPeers_262978 (1449 peers)
✅ TestRateLimiting
✅ TestRetryOnError
✅ TestContextTimeout
✅ TestConcurrentRequests
```

## Arquitetura

```
bgpin/
├── cmd/cli/              # CLI commands
├── internal/
│   ├── adapters/         # HTTP, SSH, Telnet adapters
│   ├── core/             # Domain logic (BGP, AS Path, RPKI)
│   ├── parsers/          # Vendor-specific parsers
│   └── services/         # Business logic
├── pkg/
│   ├── config/           # Configuration
│   └── telemetry/        # OpenTelemetry (futuro)
└── sdk/                  # RIPE RIS SDK
    ├── client.go         # Cliente principal
    ├── config.go         # Configuração
    ├── types.go          # Tipos de dados
    ├── errors.go         # Erros customizados
    ├── rate_limit.go     # Rate limiting
    ├── retry.go          # Retry logic
    ├── integration_test/ # Testes de integração
    └── examples/         # Exemplos de uso
```

## Tecnologias

- **Linguagem**: Go 1.25+
- **CLI Framework**: Cobra
- **Config**: Viper
- **Rate Limiting**: golang.org/x/time/rate
- **HTTP Client**: net/http (stdlib)
- **Testing**: Testes de integração reais (sem mocks)

## Roadmap

- [x] SDK RIPE RIS completo
- [x] Rate limiting e retry
- [x] Testes de integração
- [ ] CLI completa
- [ ] Suporte a múltiplos LG (HTTP, Telnet, SSH)
- [ ] Parsers para Cisco, Juniper, FRR
- [ ] Análise de anomalias BGP
- [ ] Validação RPKI
- [ ] OpenTelemetry
- [ ] Cache inteligente
- [ ] Modo interativo (TUI)

## Exemplos de Uso

### Obter informações completas de um ASN
```bash
go run sdk/examples/demo.go
```

### Executar testes de integração
```bash
go test -v ./sdk/integration_test/
```

## Contribuindo

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## Licença

MIT License - veja o arquivo LICENSE para detalhes

## Autor

bgpin team

## Agradecimentos

- RIPE NCC pela API RIS
- Comunidade BGP
- ASN 262978 (Centro de Tecnologia Armazem Datacenter Ltda.) usado nos testes


## 🏗️ Arquitetura

```
bgpin/
├── cmd/cli/              # CLI commands
│   ├── root.go          # Root command
│   ├── asn.go           # ASN commands
│   ├── prefix.go        # Prefix commands
│   └── version.go       # Version info
├── internal/
│   ├── adapters/        # HTTP, SSH, Telnet adapters
│   ├── core/            # Domain logic (BGP, AS Path, RPKI)
│   ├── parsers/         # Vendor-specific parsers
│   └── services/        # Business logic
├── pkg/
│   ├── config/          # Configuration
│   └── telemetry/       # OpenTelemetry (futuro)
└── sdk/                 # RIPE RIS SDK
    ├── client.go        # Main client
    ├── config.go        # Configuration
    ├── types.go         # Data types
    ├── errors.go        # Custom errors
    ├── rate_limit.go    # Rate limiting
    ├── retry.go         # Retry logic
    ├── integration_test/
    └── examples/
```

## 🛠️ Tecnologias

- **Linguagem**: Go 1.25+
- **CLI Framework**: Cobra
- **Config**: Viper
- **Rate Limiting**: golang.org/x/time/rate
- **HTTP Client**: net/http (stdlib)
- **Testing**: Testes de integração reais (sem mocks)

## 📋 Comandos CLI

### ASN Commands
```bash
bgpin asn info [asn]         # Informações do ASN
bgpin asn neighbors [asn]    # Vizinhos BGP
bgpin asn prefixes [asn]     # Prefixos anunciados
bgpin asn peers [asn]        # RIS peers
```

### Prefix Commands
```bash
bgpin prefix overview [prefix]  # Visão geral do prefixo
```

### Flow Commands (Preparado)
```bash
bgpin flow top                  # Top prefixes por tráfego
bgpin flow asn [asn]           # Estatísticas de tráfego do ASN
bgpin flow anomaly             # Detectar anomalias de tráfego
bgpin flow upstream-compare    # Comparar upstreams
```

### Utility Commands
```bash
bgpin lg          # Listar Looking Glasses
bgpin version     # Informações de versão
```

### Flags Globais
```bash
-o, --output string   # Formato: table, json, yaml (default: table)
-t, --timeout int     # Timeout em segundos (default: 30)
-v, --verbose         # Modo verbose
--config string       # Arquivo de configuração
```

## 🔧 Configuração

Crie um arquivo `bgpin.yaml`:

```yaml
# Configurações gerais
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

## 💡 Exemplos Práticos

### Investigar um ASN
```bash
# Informações básicas
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

### Automação
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

## 🗺️ Roadmap

- [x] SDK RIPE RIS completo
- [x] CLI básica com comandos ASN e Prefix
- [x] Múltiplos formatos de saída (table, JSON, YAML)
- [x] Rate limiting e retry
- [x] Testes de integração
- [ ] Suporte a múltiplos LG (HTTP, Telnet, SSH)
- [ ] Parsers para Cisco, Juniper, FRR
- [ ] Análise de anomalias BGP
- [ ] Validação RPKI
- [ ] OpenTelemetry
- [ ] Cache inteligente
- [ ] Modo interativo (TUI)

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## 📝 Licença

MIT License - veja o arquivo LICENSE para detalhes

## 👥 Autor

bgpin team

## 🙏 Agradecimentos

- RIPE NCC pela API RIS
- Comunidade BGP
- ASN 262978 (Centro de Tecnologia Armazem Datacenter Ltda.) usado nos testes

## 📞 Suporte

- GitHub Issues: https://github.com/bgpin/bgpin/issues
- Documentação: [docs/](docs/)
- Email: support@bgpin.dev

---

**bgpin** - BGP Intelligence Tool 🌐


## 🎨 UX Profissional

A CLI usa a biblioteca `go-pretty/table` para outputs visuais profissionais:

- ✅ Bordas arredondadas Unicode (╭╮╰╯)
- ✅ Títulos centralizados
- ✅ Linhas limpas sem separadores extras
- ✅ Detecção automática de IPv4/IPv6
- ✅ Truncamento inteligente com footer
- ✅ Largura dinâmica

Veja mais exemplos em [docs/OUTPUT_EXAMPLES.md](docs/OUTPUT_EXAMPLES.md)
