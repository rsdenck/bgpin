# bgpin - Arquitetura

## VisÃ£o Geral

bgpin segue uma arquitetura Clean + Hexagonal (Ports & Adapters), separando claramente as responsabilidades e permitindo extensibilidade.

## Estrutura de DiretÃ³rios

```
bgpin/
â”œâ”€â”€ cmd/cli/              # CLI Layer - Entrada do usuÃ¡rio
â”‚   â”œâ”€â”€ root.go          # Comando raiz e configuraÃ§Ã£o
â”‚   â”œâ”€â”€ asn.go           # Comandos ASN
â”‚   â”œâ”€â”€ prefix.go        # Comandos de prefixo
â”‚   â”œâ”€â”€ lookup.go        # Lookup em LGs
â”‚   â”œâ”€â”€ route.go         # Comandos de rota
â”‚   â”œâ”€â”€ neighbors.go     # Comandos de vizinhos
â”‚   â”œâ”€â”€ analyze.go       # AnÃ¡lise de anomalias
â”‚   â”œâ”€â”€ list.go          # Listar LGs
â”‚   â””â”€â”€ version.go       # VersÃ£o
â”‚
â”œâ”€â”€ internal/            # CÃ³digo interno (nÃ£o exportÃ¡vel)
â”‚   â”œâ”€â”€ core/           # Domain Layer - LÃ³gica de negÃ³cio
â”‚   â”‚   â”œâ”€â”€ bgp/        # Tipos e lÃ³gica BGP
â”‚   â”‚   â”œâ”€â”€ aspath/     # AnÃ¡lise de AS-PATH
â”‚   â”‚   â””â”€â”€ rpki/       # ValidaÃ§Ã£o RPKI
â”‚   â”‚
â”‚   â”œâ”€â”€ adapters/       # Adapters Layer - ComunicaÃ§Ã£o externa
â”‚   â”‚   â”œâ”€â”€ http/       # HTTP Looking Glass
â”‚   â”‚   â”œâ”€â”€ ssh/        # SSH Looking Glass
â”‚   â”‚   â””â”€â”€ telnet/     # Telnet Looking Glass
â”‚   â”‚
â”‚   â”œâ”€â”€ parsers/        # Parsing Layer - Vendor-specific
â”‚   â”‚   â”œâ”€â”€ cisco/      # Parser Cisco IOS
â”‚   â”‚   â”œâ”€â”€ junos/      # Parser Juniper JunOS
â”‚   â”‚   â””â”€â”€ frr/        # Parser FRRouting
â”‚   â”‚
â”‚   â”œâ”€â”€ services/       # Service Layer - OrquestraÃ§Ã£o
â”‚   â”‚   â”œâ”€â”€ lg/         # ServiÃ§o de Looking Glass
â”‚   â”‚   â””â”€â”€ analyzer/   # ServiÃ§o de anÃ¡lise
â”‚   â”‚
â”‚   â””â”€â”€ output/         # Output Layer - FormataÃ§Ã£o
â”‚       â”œâ”€â”€ json/       # Formatador JSON
â”‚       â”œâ”€â”€ yaml/       # Formatador YAML
â”‚       â””â”€â”€ table/      # Formatador tabela
â”‚
â”œâ”€â”€ pkg/                # CÃ³digo pÃºblico (exportÃ¡vel)
â”‚   â”œâ”€â”€ config/         # ConfiguraÃ§Ã£o
â”‚   â””â”€â”€ telemetry/      # OpenTelemetry (futuro)
â”‚
â”œâ”€â”€ sdk/                # RIPE RIS SDK
â”‚   â”œâ”€â”€ client.go       # Cliente HTTP
â”‚   â”œâ”€â”€ config.go       # ConfiguraÃ§Ã£o SDK
â”‚   â”œâ”€â”€ types.go        # Tipos de dados
â”‚   â”œâ”€â”€ errors.go       # Erros customizados
â”‚   â”œâ”€â”€ rate_limit.go   # Rate limiting
â”‚   â”œâ”€â”€ retry.go        # Retry logic
â”‚   â”œâ”€â”€ integration_test/
â”‚   â””â”€â”€ examples/
â”‚
â””â”€â”€ docs/               # DocumentaÃ§Ã£o
    â”œâ”€â”€ CLI_GUIDE.md
    â””â”€â”€ ARCHITECTURE.md
```

## Camadas

### 1. CLI Layer (cmd/cli/)

**Responsabilidade**: Interface com o usuÃ¡rio

- Parsing de argumentos (Cobra)
- ValidaÃ§Ã£o de entrada
- Chamada aos serviÃ§os
- FormataÃ§Ã£o de saÃ­da

**DependÃªncias**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - ConfiguraÃ§Ã£o
- `internal/services/*` - ServiÃ§os de negÃ³cio

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

**Responsabilidade**: LÃ³gica de negÃ³cio pura

- Tipos de dados BGP
- ValidaÃ§Ãµes de domÃ­nio
- AnÃ¡lise de AS-PATH
- DetecÃ§Ã£o de anomalias

**Regras**:
- âŒ NÃƒO pode importar adapters
- âŒ NÃƒO pode importar HTTP/SSH/Telnet
- âœ… Apenas lÃ³gica de negÃ³cio pura

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

**Responsabilidade**: ComunicaÃ§Ã£o com sistemas externos

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

**ImplementaÃ§Ã£o**:
```go
type Parser interface {
    ParseRoutes(output string) ([]bgp.Route, error)
}

type CiscoParser struct {
    routeRegex *regexp.Regexp
}

func (p *CiscoParser) ParseRoutes(output string) ([]bgp.Route, error) {
    // Regex parsing especÃ­fico para Cisco
}
```

### 5. Services Layer (internal/services/)

**Responsabilidade**: OrquestraÃ§Ã£o de lÃ³gica de negÃ³cio

#### Looking Glass Service
```go
type LGService struct {
    adapters map[string]Adapter
    parsers  map[string]Parser
}

func (s *LGService) QueryMultipleLGs(ctx context.Context, prefix string) ([]Result, error) {
    // Consulta mÃºltiplos LGs em paralelo
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

**Responsabilidade**: FormataÃ§Ã£o de saÃ­da

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

**CaracterÃ­sticas**:
- Thread-safe
- ConfigurÃ¡vel
- Testado com dados reais (ASN 262978)

## Fluxo de Dados

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚   UsuÃ¡rio   â”‚
â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”˜
       â”‚
       â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚         CLI Layer (cmd/cli/)            â”‚
â”‚  â€¢ Parse argumentos                     â”‚
â”‚  â€¢ Valida entrada                       â”‚
â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
       â”‚
       â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚      Service Layer (services/)          â”‚
â”‚  â€¢ Orquestra lÃ³gica                     â”‚
â”‚  â€¢ Coordena adapters                    â”‚
â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
       â”‚
       â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
       â–¼              â–¼              â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚   HTTP   â”‚   â”‚   SSH    â”‚   â”‚  Telnet  â”‚
â”‚ Adapter  â”‚   â”‚ Adapter  â”‚   â”‚ Adapter  â”‚
â””â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”˜
     â”‚              â”‚              â”‚
     â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
                    â–¼
         â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
         â”‚  Looking Glass      â”‚
         â”‚  (Cisco/Juniper)    â”‚
         â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
                    â–¼
         â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
         â”‚   Parser Layer      â”‚
         â”‚  â€¢ Cisco Parser     â”‚
         â”‚  â€¢ Junos Parser     â”‚
         â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
                    â–¼
         â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
         â”‚   Core Domain       â”‚
         â”‚  â€¢ Route structs    â”‚
         â”‚  â€¢ Validations      â”‚
         â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
                    â–¼
         â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
         â”‚  Analyzer Service   â”‚
         â”‚  â€¢ Anomaly detect   â”‚
         â”‚  â€¢ RPKI validation  â”‚
         â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
                    â–¼
         â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
         â”‚   Output Layer      â”‚
         â”‚  â€¢ JSON/YAML/Table  â”‚
         â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
                    â–¼
              â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”
              â”‚ UsuÃ¡rio â”‚
              â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

## PrincÃ­pios de Design

### 1. Separation of Concerns
Cada camada tem uma responsabilidade Ãºnica e bem definida.

### 2. Dependency Inversion
Camadas superiores dependem de abstraÃ§Ãµes, nÃ£o de implementaÃ§Ãµes concretas.

### 3. Interface Segregation
Interfaces pequenas e especÃ­ficas ao invÃ©s de grandes e genÃ©ricas.

### 4. Open/Closed Principle
Aberto para extensÃ£o (novos vendors), fechado para modificaÃ§Ã£o.

### 5. Single Responsibility
Cada mÃ³dulo tem uma Ãºnica razÃ£o para mudar.

## Extensibilidade

### Adicionar novo Looking Glass

1. Criar adapter em `internal/adapters/`
2. Implementar interface `Adapter`
3. Registrar no service layer

### Adicionar novo vendor parser

1. Criar parser em `internal/parsers/`
2. Implementar interface `Parser`
3. Adicionar regex patterns especÃ­ficos

### Adicionar nova anÃ¡lise

1. Adicionar funÃ§Ã£o em `internal/services/analyzer/`
2. Implementar lÃ³gica de detecÃ§Ã£o
3. Retornar tipo `Anomaly`

## Testes

### Unit Tests
- Core domain (lÃ³gica pura)
- Parsers (regex patterns)
- Analyzers (detecÃ§Ã£o)

### Integration Tests
- SDK (RIPE RIS API real)
- Adapters (LGs reais)

### E2E Tests
- CLI completa
- Fluxo end-to-end

## Performance

### ConcorrÃªncia
- Consultas paralelas a mÃºltiplos LGs
- Goroutines + WaitGroup
- Context para cancelamento

### Cache
- In-memory cache com TTL
- Ristretto (futuro)

### Rate Limiting
- Token bucket algorithm
- ConfigurÃ¡vel por fonte

## SeguranÃ§a

### Input Validation
- SanitizaÃ§Ã£o de comandos
- Whitelist de comandos permitidos
- ValidaÃ§Ã£o de prefixos/ASNs

### Rate Limiting
- ProteÃ§Ã£o contra abuse
- Limites configurÃ¡veis

### Timeout
- Todas as operaÃ§Ãµes com timeout
- Context-based cancellation

## Observabilidade (Futuro)

### OpenTelemetry
- Traces de requisiÃ§Ãµes
- MÃ©tricas de performance
- Logs estruturados

### Exporters
- Jaeger
- Prometheus
- OTLP

## Roadmap

- [x] SDK RIPE RIS
- [x] CLI bÃ¡sica
- [ ] MÃºltiplos LG adapters
- [ ] Parsers completos
- [ ] AnÃ¡lise de anomalias
- [ ] ValidaÃ§Ã£o RPKI
- [ ] Cache inteligente
- [ ] OpenTelemetry
- [ ] TUI interativo
