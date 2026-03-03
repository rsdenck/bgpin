# RIPE RIS SDK for Go

SDK profissional em Golang para integração com os serviços da RIPE RIS (Routing Information Service).

## Características

- ✅ Rate limiting configurável
- ✅ Retry com exponential backoff
- ✅ Timeout configurável
- ✅ Context support
- ✅ Sem mocks - testes reais com ASN 262978
- ✅ Arquitetura limpa e extensível
- ✅ Tratamento robusto de erros

## Instalação

```bash
go get github.com/bgpin/bgpin/sdk
```

## Uso Básico

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
    // Criar cliente com configuração padrão
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
}
```

## Configuração Customizada

```go
config := sdk.Config{
    Timeout:      30 * time.Second,
    RateLimit:    10, // 10 requisições por segundo
    RetryMax:     3,
    RetryWaitMin: 1 * time.Second,
    RetryWaitMax: 10 * time.Second,
    UserAgent:    "my-app/1.0",
    BaseURL:      "https://stat.ripe.net/data",
}

client := sdk.NewClient(config)
```

## Funcionalidades

### 1. Informações do ASN
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

### 4. Visão Geral do Prefixo
```go
overview, err := client.GetPrefixOverview(ctx, "200.160.0.0/20")
```

### 5. RIS Peers
```go
peers, err := client.GetRISPeers(ctx, 262978)
```

## Testes

Todos os testes são de integração real usando o ASN 262978:

```bash
# Executar todos os testes
go test -v ./sdk/integration_test/

# Executar teste específico
go test -v ./sdk/integration_test/ -run TestGetASNInfo_262978
```

## Tratamento de Erros

```go
info, err := client.GetASNInfo(ctx, asn)
if err != nil {
    switch {
    case errors.Is(err, sdk.ErrInvalidASN):
        // ASN inválido
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

O SDK implementa rate limiting automático:

```go
config := sdk.DefaultConfig()
config.RateLimit = 5 // 5 requisições por segundo
client := sdk.NewClient(config)
```

## Retry com Exponential Backoff

Retry automático em caso de erros 5xx ou 429:

```go
config := sdk.DefaultConfig()
config.RetryMax = 3
config.RetryWaitMin = 1 * time.Second
config.RetryWaitMax = 10 * time.Second
client := sdk.NewClient(config)
```

## Requisições Concorrentes

O SDK é thread-safe e suporta requisições concorrentes:

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
├── client.go          # Cliente principal
├── config.go          # Configuração
├── types.go           # Tipos de dados
├── errors.go          # Erros customizados
├── rate_limit.go      # Rate limiting
├── retry.go           # Retry logic
├── integration_test/  # Testes de integração
│   └── asn_262978_test.go
└── examples/          # Exemplos de uso
    └── basic_usage.go
```

## Licença

MIT
