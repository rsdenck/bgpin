# bgpin - Arquitetura

## Visão Geral

bgpin segue uma arquitetura Clean + Hexagonal (Ports & Adapters), separando claramente as responsabilidades e permitindo extensibilidade.

## Estrutura de Diretórios

```
bgpin/
├── cmd/cli/              # CLI Layer - Entrada do usuário
│   ├── root.go          # Comando raiz e configuração
│   ├── asn.go           # Comandos ASN
│   ├── prefix.go        # Comandos de prefixo
│   ├── lookup.go        # Lookup em LGs
│   ├── route.go         # Comandos de rota
│   ├── neighbors.go     # Comandos de vizinhos
│   ├── analyze.go       # Análise de anomalias
│   ├── list.go          # Listar LGs
│   └── version.go       # Versão
│
├── internal/            # Código interno (não exportável)
│   ├── core/           # Domain Layer - Lógica de negócio
│   │   ├── bgp/        # Tipos e lógica BGP
│   │   ├── aspath/     # Análise de AS-PATH
│   │   └── rpki/       # Validação RPKI
│   │
│   ├── adapters/       # Adapters Layer - Comunicação externa
│   │   ├── http/       # HTTP Looking Glass
│   │   ├── ssh/        # SSH Looking Glass
│   │   └── telnet/     # Telnet Looking Glass
│   │
│   ├── parsers/        # Parsing Layer - Vendor-specific
│   │   ├── cisco/      # Parser Cisco IOS
│   │   ├── junos/      # Parser Juniper JunOS
│   │   └── frr/        # Parser FRRouting
│   │
│   ├── services/       # Service Layer - Orquestração
│   │   ├── lg/         # Serviço de Looking Glass
│   │   └── analyzer/   # Serviço de análise
│   │
│   └── output/         # Output Layer - Formatação
│       ├── json/       # Formatador JSON
│       ├── yaml/       # Formatador YAML
│       └── table/      # Formatador tabela
│
├── pkg/                # Código público (exportável)
│   ├── config/         # Configuração
│   └── telemetry/      # OpenTelemetry (futuro)
│
├── sdk/                # RIPE RIS SDK
│   ├── client.go       # Cliente HTTP
│   ├── config.go       # Configuração SDK
│   ├── types.go        # Tipos de dados
│   ├── errors.go       # Erros customizados
│   ├── rate_limit.go   # Rate limiting
│   ├── retry.go        # Retry logic
│   ├── integration_test/
│   └── examples/
│
└── docs/               # Documentação
    ├── CLI_GUIDE.md
    └── ARCHITECTURE.md
```

## Camadas

### 1. CLI Layer (cmd/cli/)

**Responsabilidade**: Interface com o usuário

- Parsing de argumentos (Cobra)
- Validação de entrada
- Chamada aos serviços
- Formatação de saída

**Dependências**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuração
- `internal/services/*` - Serviços de negócio

**Exemplo**:
```go
func runASNInfo(cmd *cobra.Command, args []string) error {
    asn, err := parseASN(args[0])
    client := sdk.NewDefaultClient()
    info, err := client.GetASNInfo(ctx, asn)
    return outputASNInfo(info, format)
}
```

### 2. Core Domain Layer (internal/core/)

**Responsabilidade**: Lógica de negócio pura

- Tipos de dados BGP
- Validações de domínio
- Análise de AS-PATH
- Detecção de anomalias

**Regras**:
- ❌ NÃO pode importar adapters
- ❌ NÃO pode importar HTTP/SSH/Telnet
- ✅ Apenas lógica de negócio pura

**Exemplo**:
```go
type Route struct {
    Prefix    string
    ASPath    []int
    NextHop   string
    LocalPref int
    MED       int
    Valid     bool
    Best      bool
}

func (r *Route) HasLoop() bool {
    seen := make(map[int]bool)
    for _, asn := range r.ASPath {
        if seen[asn] {
            return true
        }
        seen[asn] = true
    }
    return false
}
```

### 3. Adapters Layer (internal/adapters/)

**Responsabilidade**: Comunicação com sistemas externos

#### HTTP Adapter
```go
type HTTPAdapter struct {
    baseURL string
    client  *http.Client
}

func (a *HTTPAdapter) QueryBGP(ctx context.Context, prefix string) (string, error)
```

#### SSH Adapter (futuro)
```go
type SSHAdapter struct {
    host   string
    config *ssh.ClientConfig
}
```

#### Telnet Adapter (futuro)
```go
type TelnetAdapter struct {
    host string
    port int
}
```

### 4. Parsers Layer (internal/parsers/)

**Responsabilidade**: Parsing de output vendor-specific

Cada vendor tem formato diferente:

#### Cisco IOS
```
*> 186.250.184.0/24  200.160.0.1    0    100      0 262978 i
```

#### Juniper JunOS
```
186.250.184.0/24 *[BGP/170] 00:01:23, localpref 100
                    AS path: 262978 I
```

#### FRRouting
```
*> 186.250.184.0/24  200.160.0.1    0    100      0 262978 i
```

**Implementação**:
```go
type Parser interface {
    ParseRoutes(output string) ([]bgp.Route, error)
}

type CiscoParser struct {
    routeRegex *regexp.Regexp
}

func (p *CiscoParser) ParseRoutes(output string) ([]bgp.Route, error) {
    // Regex parsing específico para Cisco
}
```

### 5. Services Layer (internal/services/)

**Responsabilidade**: Orquestração de lógica de negócio

#### Looking Glass Service
```go
type LGService struct {
    adapters map[string]Adapter
    parsers  map[string]Parser
}

func (s *LGService) QueryMultipleLGs(ctx context.Context, prefix string) ([]Result, error) {
    // Consulta múltiplos LGs em paralelo
    // Usa goroutines + WaitGroup
}
```

#### Analyzer Service
```go
type AnalyzerService struct {
    rpkiValidator RPKIValidator
}

func (s *AnalyzerService) DetectAnomalies(route *bgp.Route) []Anomaly {
    // Detecta prepend excessivo
    // Detecta AS-PATH loops
    // Valida RPKI
}
```

### 6. Output Layer (internal/output/)

**Responsabilidade**: Formatação de saída

```go
type Formatter interface {
    Format(data interface{}) (string, error)
}

type JSONFormatter struct{}
type YAMLFormatter struct{}
type TableFormatter struct{}
```

### 7. SDK Layer (sdk/)

**Responsabilidade**: Cliente para RIPE RIS API

- Rate limiting
- Retry com exponential backoff
- Context support
- Tipos estruturados

**Características**:
- Thread-safe
- Configurável
- Testado com dados reais (ASN 262978)

## Fluxo de Dados

```
┌─────────────┐
│   Usuário   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────┐
│         CLI Layer (cmd/cli/)            │
│  • Parse argumentos                     │
│  • Valida entrada                       │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│      Service Layer (services/)          │
│  • Orquestra lógica                     │
│  • Coordena adapters                    │
└──────┬──────────────────────────────────┘
       │
       ├──────────────┬──────────────┐
       ▼              ▼              ▼
┌──────────┐   ┌──────────┐   ┌──────────┐
│   HTTP   │   │   SSH    │   │  Telnet  │
│ Adapter  │   │ Adapter  │   │ Adapter  │
└────┬─────┘   └────┬─────┘   └────┬─────┘
     │              │              │
     └──────────────┴──────────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │  Looking Glass      │
         │  (Cisco/Juniper)    │
         └──────────┬──────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │   Parser Layer      │
         │  • Cisco Parser     │
         │  • Junos Parser     │
         └──────────┬──────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │   Core Domain       │
         │  • Route structs    │
         │  • Validations      │
         └──────────┬──────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │  Analyzer Service   │
         │  • Anomaly detect   │
         │  • RPKI validation  │
         └──────────┬──────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │   Output Layer      │
         │  • JSON/YAML/Table  │
         └──────────┬──────────┘
                    │
                    ▼
              ┌─────────┐
              │ Usuário │
              └─────────┘
```

## Princípios de Design

### 1. Separation of Concerns
Cada camada tem uma responsabilidade única e bem definida.

### 2. Dependency Inversion
Camadas superiores dependem de abstrações, não de implementações concretas.

### 3. Interface Segregation
Interfaces pequenas e específicas ao invés de grandes e genéricas.

### 4. Open/Closed Principle
Aberto para extensão (novos vendors), fechado para modificação.

### 5. Single Responsibility
Cada módulo tem uma única razão para mudar.

## Extensibilidade

### Adicionar novo Looking Glass

1. Criar adapter em `internal/adapters/`
2. Implementar interface `Adapter`
3. Registrar no service layer

### Adicionar novo vendor parser

1. Criar parser em `internal/parsers/`
2. Implementar interface `Parser`
3. Adicionar regex patterns específicos

### Adicionar nova análise

1. Adicionar função em `internal/services/analyzer/`
2. Implementar lógica de detecção
3. Retornar tipo `Anomaly`

## Testes

### Unit Tests
- Core domain (lógica pura)
- Parsers (regex patterns)
- Analyzers (detecção)

### Integration Tests
- SDK (RIPE RIS API real)
- Adapters (LGs reais)

### E2E Tests
- CLI completa
- Fluxo end-to-end

## Performance

### Concorrência
- Consultas paralelas a múltiplos LGs
- Goroutines + WaitGroup
- Context para cancelamento

### Cache
- In-memory cache com TTL
- Ristretto (futuro)

### Rate Limiting
- Token bucket algorithm
- Configurável por fonte

## Segurança

### Input Validation
- Sanitização de comandos
- Whitelist de comandos permitidos
- Validação de prefixos/ASNs

### Rate Limiting
- Proteção contra abuse
- Limites configuráveis

### Timeout
- Todas as operações com timeout
- Context-based cancellation

## Observabilidade (Futuro)

### OpenTelemetry
- Traces de requisições
- Métricas de performance
- Logs estruturados

### Exporters
- Jaeger
- Prometheus
- OTLP

## Roadmap

- [x] SDK RIPE RIS
- [x] CLI básica
- [ ] Múltiplos LG adapters
- [ ] Parsers completos
- [ ] Análise de anomalias
- [ ] Validação RPKI
- [ ] Cache inteligente
- [ ] OpenTelemetry
- [ ] TUI interativo
