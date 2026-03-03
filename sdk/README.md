# RIPE RIS SDK for Go

SDK profissional em Golang para integraÃ§Ã£o com os serviÃ§os da RIPE RIS (Routing Information Service).

## CaracterÃ­sticas

- âœ… Rate limiting configurÃ¡vel
- âœ… Retry com exponential backoff
- âœ… Timeout configurÃ¡vel
- âœ… Context support
- âœ… Sem mocks - testes reais com ASN 262978
- âœ… Arquitetura limpa e extensÃ­vel
- âœ… Tratamento robusto de erros

## InstalaÃ§Ã£o

```bash
go get github.com/bgpin/bgpin/sdk
```

## Uso BÃ¡sico

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
    // Criar cliente com configuraÃ§Ã£o padrÃ£o
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
}
```

## ConfiguraÃ§Ã£o Customizada

```go
config := sdk.Config{
    Timeout:      30 * time.Second,
    RateLimit:    10, // 10 requisiÃ§Ãµes por segundo
    RetryMax:     3,
    RetryWaitMin: 1 * time.Second,
    RetryWaitMax: 10 * time.Second,
    UserAgent:    "my-app/1.0",
    BaseURL:      "https://stat.ripe.net/data",
}

client := sdk.NewClient(config)
```

## Funcionalidades

### 1. InformaÃ§Ãµes do ASN
```go
info, err := client.GetASNInfo(ctx, 262978)
```

### 2. Vizinhos BGP
```go
neighbors, err := client.GetASNNeighbors(ctx, 262978)
```

### 3. Prefixos Anunciados
```go
prefixes, err := client.GetAnnouncedPrefixes(ctx, 262978)
```

### 4. VisÃ£o Geral do Prefixo
```go
overview, err := client.GetPrefixOverview(ctx, "200.160.0.0/20")
```

### 5. RIS Peers
```go
peers, err := client.GetRISPeers(ctx, 262978)
```

## Testes

Todos os testes sÃ£o de integraÃ§Ã£o real usando o ASN 262978:

```bash
# Executar todos os testes
go test -v ./sdk/integration_test/

# Executar teste especÃ­fico
go test -v ./sdk/integration_test/ -run TestGetASNInfo_262978
```

## Tratamento de Erros

```go
info, err := client.GetASNInfo(ctx, asn)
if err != nil {
    switch {
    case errors.Is(err, sdk.ErrInvalidASN):
        // ASN invÃ¡lido
    case errors.Is(err, sdk.ErrTimeout):
        // Timeout
    case errors.Is(err, sdk.ErrRateLimitExceeded):
        // Rate limit excedido
    default:
        // Outro erro
    }
}
```

## Rate Limiting

O SDK implementa rate limiting automÃ¡tico:

```go
config := sdk.DefaultConfig()
config.RateLimit = 5 // 5 requisiÃ§Ãµes por segundo
client := sdk.NewClient(config)
```

## Retry com Exponential Backoff

Retry automÃ¡tico em caso de erros 5xx ou 429:

```go
config := sdk.DefaultConfig()
config.RetryMax = 3
config.RetryWaitMin = 1 * time.Second
config.RetryWaitMax = 10 * time.Second
client := sdk.NewClient(config)
```

## RequisiÃ§Ãµes Concorrentes

O SDK Ã© thread-safe e suporta requisiÃ§Ãµes concorrentes:

```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(asn int) {
        defer wg.Done()
        info, err := client.GetASNInfo(ctx, asn)
        // processar resultado
    }(262978)
}

wg.Wait()
```

## Estrutura do Projeto

```
sdk/
â”œâ”€â”€ client.go          # Cliente principal
â”œâ”€â”€ config.go          # ConfiguraÃ§Ã£o
â”œâ”€â”€ types.go           # Tipos de dados
â”œâ”€â”€ errors.go          # Erros customizados
â”œâ”€â”€ rate_limit.go      # Rate limiting
â”œâ”€â”€ retry.go           # Retry logic
â”œâ”€â”€ integration_test/  # Testes de integraÃ§Ã£o
â”‚   â””â”€â”€ asn_262978_test.go
â””â”€â”€ examples/          # Exemplos de uso
    â””â”€â”€ basic_usage.go
```

## LicenÃ§a

MIT
