# bgpin SDK Examples

This directory contains example programs demonstrating how to use the bgpin SDK.

## Available Examples

### Basic Usage
**Location:** `basic_usage/main.go`

Simple example showing basic SDK usage:
- Creating a client
- Querying ASN information
- Error handling

```bash
cd basic_usage
go run main.go
```

### Demo
**Location:** `demo/main.go`

Comprehensive demo showing:
- Multiple API calls
- Context usage
- Rate limiting
- Error handling
- Output formatting

```bash
cd demo
go run main.go
```

## Running Examples

Each example is in its own directory with a `main.go` file:

```bash
# Run basic usage example
cd basic_usage && go run main.go

# Run demo example
cd demo && go run main.go
```

## Building Examples

You can build standalone binaries:

```bash
# Build basic usage
cd basic_usage
go build -o basic_usage

# Build demo
cd demo
go build -o demo
```

## SDK Documentation

For complete SDK documentation, see [../README.md](../README.md)

## Creating Your Own

To create your own program using the bgpin SDK:

1. Initialize a Go module:
```bash
mkdir my-bgp-tool
cd my-bgp-tool
go mod init my-bgp-tool
```

2. Add bgpin SDK dependency:
```bash
go get github.com/bgpin/bgpin/sdk
```

3. Create your program:
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
    client := sdk.NewDefaultClient()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    info, err := client.GetASNInfo(ctx, 15169)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("ASN: %d\n", info.ASN)
    fmt.Printf("Holder: %s\n", info.Holder)
}
```

4. Run it:
```bash
go run main.go
```

## Common Use Cases

### Query Multiple ASNs
```go
asns := []int{15169, 13335, 8075}
for _, asn := range asns {
    info, err := client.GetASNInfo(ctx, asn)
    if err != nil {
        log.Printf("Error querying ASN %d: %v", asn, err)
        continue
    }
    fmt.Printf("%d: %s\n", info.ASN, info.Holder)
}
```

### Get BGP Neighbors
```go
neighbors, err := client.GetASNNeighbors(ctx, 15169)
if err != nil {
    log.Fatal(err)
}

for _, neighbor := range neighbors {
    fmt.Printf("Neighbor: AS%d (%s)\n", neighbor.ASN, neighbor.Type)
}
```

### Get Announced Prefixes
```go
prefixes, err := client.GetAnnouncedPrefixes(ctx, 15169)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total prefixes: %d\n", len(prefixes))
for _, prefix := range prefixes {
    fmt.Printf("  %s\n", prefix.Prefix)
}
```

## Error Handling

Always handle errors properly:

```go
info, err := client.GetASNInfo(ctx, asn)
if err != nil {
    switch {
    case errors.Is(err, sdk.ErrNotFound):
        fmt.Println("ASN not found")
    case errors.Is(err, sdk.ErrRateLimited):
        fmt.Println("Rate limited, try again later")
    case errors.Is(err, context.DeadlineExceeded):
        fmt.Println("Request timeout")
    default:
        log.Fatal(err)
    }
    return
}
```

## Rate Limiting

The SDK includes built-in rate limiting (10 req/s by default):

```go
config := sdk.Config{
    RateLimit: 5, // 5 requests per second
}
client := sdk.NewClient(config)
```

## Context and Timeouts

Always use context with timeouts:

```go
// 30 second timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

info, err := client.GetASNInfo(ctx, asn)
```

## More Information

- [SDK Documentation](../README.md)
- [API Reference](https://pkg.go.dev/github.com/bgpin/bgpin/sdk)
- [Main Documentation](../../docs/README.md)
