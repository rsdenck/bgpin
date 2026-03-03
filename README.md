<div align="center">

# Platform Border Gateway Protocol

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/rsdenck/bgpin.svg)](https://github.com/rsdenck/bgpin/releases)
[![Tests](https://img.shields.io/badge/tests-passing-success.svg)](https://github.com/rsdenck/bgpin)

CLI profissional + SDK + Telemetria + Flow Analysis

</div>

---

## Características

### CLI Completa
- Consulta informações de ASN (holder, block, status)
- Lista vizinhos BGP (upstream, downstream, peers)
- Mostra prefixos anunciados (IPv4 e IPv6)
- Visualiza RIS peers por RRC
- Análise de prefixos específicos
- Múltiplos formatos de saída (table, JSON, YAML)
- UX profissional com go-pretty/table

### SDK RIPE RIS
- Rate limiting configurável
- Retry com exponential backoff
- Context support completo
- Testes de integração reais (sem mocks)
- Thread-safe para requisições concorrentes

### Telemetria & Observabilidade
- OpenTelemetry integration
- Distributed tracing
- Métricas de performance
- Exportação para OTLP, Jaeger, Prometheus
- Dashboards Grafana

### Flow Analysis
- Coleta NetFlow/sFlow/IPFIX em tempo real
- Análise de tráfego por ASN e prefixo
- Detecção de anomalias (DDoS, spikes, drops)
- Comparação de upstreams
- Correlação BGP + Flow
- Agregação configurável
- Worker pools para alta performance

### Vendor Parsers
- Cisco (IOS, IOS-XE, IOS-XR, NX-OS) - Implementado
- Juniper (JunOS) - Implementado
- Arista, Nokia, MikroTik - Planejado

## Instalação

```bash
# Download binary (Linux)
wget https://github.com/rsdenck/bgpin/releases/latest/download/bgpin-linux-amd64
chmod +x bgpin-linux-amd64
sudo mv bgpin-linux-amd64 /usr/local/bin/bgpin

# Ou compile do código fonte
git clone https://github.com/rsdenck/bgpin
cd bgpin
go build -o bgpin ./cmd/cli/
```

## Uso Rápido

### CLI

```bash
# Informações de um ASN
bgpin asn info 262978

# Vizinhos BGP
bgpin asn neighbors 262978

# Prefixos anunciados
bgpin asn prefixes 262978

# Flow telemetry
bgpin flow top
bgpin flow asn 15169
bgpin flow anomaly

# Formato JSON
bgpin asn info 262978 -o json
```

### SDK

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/bgpin/bgpin/sdk"
)

func main() {
    client := sdk.NewDefaultClient()
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    info, err := client.GetASNInfo(ctx, 262978)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("ASN: %d - %s\n", info.ASN, info.Holder)
}
```

## Documentação

- [Installation Guide](docs/INSTALLATION.md) - Guia completo de instalação
- [Quick Start](docs/QUICK_START.md) - Início rápido
- [CLI Guide](docs/CLI_GUIDE.md) - Todos os comandos
- [Flow Collector](docs/FLOW_COLLECTOR.md) - NetFlow/sFlow/IPFIX
- [Telemetry](docs/TELEMETRY.md) - OpenTelemetry
- [Architecture](docs/ARCHITECTURE.md) - Design do sistema
- [SDK Documentation](sdk/README.md) - SDK completo
- [Vendors Status](docs/vendors/STATUS.md) - Status dos parsers

### Releases
- [v0.2.0](docs/releases/v0.2.0.md) - Flow collector + Cisco/Juniper parsers
- [v0.1.0](docs/releases/v0.1.0.md) - Initial release

## Exemplos de Saída

### Informações de ASN
```
╭───────────────────────────────────────────────────────────╮
│ ASN Information: AS262978                                 │
├───────────┬───────────────────────────────────────────────┤
│ Holder    │ Centro de Tecnologia Armazem Datacenter Ltda. │
│ Announced │ true                                          │
│ Block     │ 262144-263167                                 │
╰───────────┴───────────────────────────────────────────────╯
```

### Flow Top Prefixes
```
╭────────────────────────────────────────────────────────────────────────╮
│ Top Prefixes by Traffic                                                │
├───┬─────────────────┬─────────┬──────────────────┬──────┬──────────────┤
│ # │ PREFIX          │ ASN     │ TRAFFIC          │ PPS  │ TOP PROTOCOL │
│ 1 │ 8.8.8.0/24      │ AS15169 │ 850 Mbps         │ 120k │ TCP (443)    │
│ 2 │ 1.1.1.0/24      │ AS13335 │ 640 Mbps         │ 98k  │ UDP (53)     │
╰───┴─────────────────┴─────────┴──────────────────┴──────┴──────────────╯
```

## Arquitetura

```
bgpin/
├── cmd/cli/              # CLI commands
├── internal/
│   ├── adapters/         # HTTP, SSH, NETCONF
│   ├── parsers/          # Cisco, Juniper, etc
│   ├── flow/             # NetFlow/sFlow/IPFIX
│   └── telemetry/        # OpenTelemetry
├── sdk/                  # RIPE RIS SDK
└── docs/                 # Documentation
```

## Tecnologias

- Go 1.23+
- Cobra (CLI framework)
- Viper (Configuration)
- go-pretty/table (Output formatting)
- OpenTelemetry (Observability)
- golang.org/x/crypto/ssh (SSH client)

## Roadmap

- [x] SDK RIPE RIS completo
- [x] CLI com comandos ASN e Prefix
- [x] Flow collector (NetFlow/sFlow/IPFIX)
- [x] Cisco parser (IOS/IOS-XE/IOS-XR/NX-OS)
- [x] Juniper parser (JunOS/NETCONF)
- [x] OpenTelemetry integration
- [ ] Arista parser (EOS)
- [ ] Nokia parser (SR OS)
- [ ] RPKI validation
- [ ] Machine Learning anomaly detection
- [ ] Time-series database storage
- [ ] Grafana dashboards

## Contribuindo

Contribuições são bem-vindas! Veja [CONTRIBUTING.md](CONTRIBUTING.md)

## Licença

MIT License - veja [LICENSE](LICENSE)

## Suporte

- Issues: https://github.com/rsdenck/bgpin/issues
- Discussions: https://github.com/rsdenck/bgpin/discussions
- Documentation: [docs/](docs/)

---

**bgpin** - Platform Border Gateway Protocol
