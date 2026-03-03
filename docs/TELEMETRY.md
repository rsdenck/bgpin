# bgpin - Telemetry & Observability

## OpenTelemetry Integration

bgpin integra OpenTelemetry para observabilidade completa de todas as operaÃ§Ãµes BGP.

### CaracterÃ­sticas

- âœ… Distributed tracing
- âœ… MÃ©tricas de performance
- âœ… LatÃªncia de queries
- âœ… Contadores de erros
- âœ… ExportaÃ§Ã£o para OTLP, Jaeger, Prometheus

### ConfiguraÃ§Ã£o

```yaml
# bgpin.yaml
telemetry:
  enabled: true
  export_type: stdout  # stdout, otlp, jaeger
  endpoint: "localhost:4317"
  
  # MÃ©tricas
  metrics:
    enabled: true
    interval: 60s
  
  # Traces
  traces:
    enabled: true
    sample_rate: 1.0  # 100% sampling
```

### Traces DisponÃ­veis

#### 1. Query Traces
```
bgpin.query
â”œâ”€â”€ bgpin.sdk.get_asn_info
â”‚   â”œâ”€â”€ http.request
â”‚   â””â”€â”€ json.unmarshal
â”œâ”€â”€ bgpin.output.format
â””â”€â”€ bgpin.render.table
```

#### 2. Attributes
- `bgp.provider` - Provider do Looking Glass
- `bgp.prefix` - Prefixo consultado
- `bgp.asn` - ASN consultado
- `bgp.as_path` - AS Path
- `cli.command` - Comando executado
- `cli.output_format` - Formato de saÃ­da
- `result.count` - NÃºmero de resultados
- `duration_ms` - DuraÃ§Ã£o em milissegundos

### MÃ©tricas DisponÃ­veis

#### Counters
- `bgpin.queries.total` - Total de queries executadas
- `bgpin.errors.total` - Total de erros
- `bgpin.prefixes.total` - Total de prefixos consultados
- `bgpin.neighbors.total` - Total de vizinhos consultados

#### Histograms
- `bgpin.query.duration` - DuraÃ§Ã£o das queries (ms)

### Exemplo de Uso

```go
import (
    "github.com/bgpin/bgpin/internal/telemetry"
)

func queryASN(ctx context.Context, asn int) error {
    // Start span
    ctx, span := telemetry.StartSpan(ctx, "query.asn",
        telemetry.AttrASN.Int(asn),
        telemetry.AttrCommand.String("asn info"),
    )
    defer span.End()
    
    start := time.Now()
    
    // Execute query
    result, err := client.GetASNInfo(ctx, asn)
    
    // Record metrics
    telemetry.RecordLatency(span, start)
    
    if err != nil {
        telemetry.RecordError(span, err)
        return err
    }
    
    telemetry.RecordSuccess(span)
    return nil
}
```

### Exporters

#### Stdout (Development)
```bash
bgpin asn info 262978
```

Output:
```json
{
  "Name": "bgpin.query",
  "SpanContext": {
    "TraceID": "4bf92f3577b34da6a3ce929d0e0e4736",
    "SpanID": "00f067aa0ba902b7"
  },
  "Attributes": [
    {"Key": "bgp.asn", "Value": 262978},
    {"Key": "cli.command", "Value": "asn info"},
    {"Key": "duration_ms", "Value": 234}
  ]
}
```

#### OTLP (Production)
```yaml
telemetry:
  export_type: otlp
  endpoint: "otel-collector:4317"
```

#### Jaeger (Distributed Tracing)
```yaml
telemetry:
  export_type: jaeger
  endpoint: "jaeger:14268"
```

### VisualizaÃ§Ã£o

#### Jaeger UI
```
http://localhost:16686
```

Visualize:
- Trace completo de cada query
- LatÃªncia por componente
- Erros e exceÃ§Ãµes
- DependÃªncias entre serviÃ§os

#### Prometheus + Grafana
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'bgpin'
    static_configs:
      - targets: ['localhost:9090']
```

Dashboards:
- Query rate
- Error rate
- Latency percentiles (p50, p95, p99)
- Top ASNs queried
- Top prefixes queried

### Alertas

#### Prometheus Alerts
```yaml
groups:
  - name: bgpin
    rules:
      - alert: HighErrorRate
        expr: rate(bgpin_errors_total[5m]) > 0.1
        annotations:
          summary: "High error rate detected"
      
      - alert: SlowQueries
        expr: histogram_quantile(0.95, bgpin_query_duration_ms) > 1000
        annotations:
          summary: "95th percentile latency > 1s"
```

## Flow Telemetry

bgpin pode coletar e analisar dados de flow (NetFlow/sFlow/IPFIX) para correlacionar com BGP.

### Arquitetura

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚   Router    â”‚ NetFlow/sFlow
â”‚  (Exporter) â”‚â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜            â”‚
                           â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”      â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚   Switch    â”‚â”€â”€â”€â”€â”€â–¶â”‚ bgpin Flow   â”‚
â”‚  (Exporter) â”‚      â”‚  Collector   â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜      â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜
                            â”‚
                            â–¼
                     â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
                     â”‚  Aggregator  â”‚
                     â”‚  & Analyzer  â”‚
                     â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜
                            â”‚
                            â–¼
                     â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
                     â”‚   Storage    â”‚
                     â”‚ (Time Series)â”‚
                     â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

### ConfiguraÃ§Ã£o

```yaml
# bgpin.yaml
flow:
  enabled: true
  listen_addr: "0.0.0.0:2055"  # NetFlow port
  protocols:
    - netflow_v5
    - netflow_v9
    - sflow
    - ipfix
  
  aggregation:
    window: 60s
    max_flows: 100000
  
  anomaly_detection:
    enabled: true
    window: 300s
    thresholds:
      bps: 1000000000  # 1 Gbps
      pps: 100000      # 100k pps
```

### Comandos Flow

#### 1. Top Prefixes
```bash
bgpin flow top
bgpin flow top --limit 20
bgpin flow top --prefix 8.8.8.0/24
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ Top Prefixes by Traffic                                                â”‚
â”œâ”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ # â”‚ PREFIX          â”‚ ASN     â”‚ TRAFFIC          â”‚ PPS  â”‚ TOP PROTOCOL â”‚
â”œâ”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ 1 â”‚ 8.8.8.0/24      â”‚ AS15169 â”‚ 850 Mbps         â”‚ 120k â”‚ TCP (443)    â”‚
â”‚ 2 â”‚ 1.1.1.0/24      â”‚ AS13335 â”‚ 640 Mbps         â”‚ 98k  â”‚ UDP (53)     â”‚
â”‚ 3 â”‚ 208.67.222.0/24 â”‚ AS36692 â”‚ 1.2 Gbps         â”‚ 150k â”‚ TCP (80)     â”‚
â•°â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

#### 2. ASN Traffic
```bash
bgpin flow asn 15169
bgpin flow asn AS262978
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ Traffic Statistics for AS15169                 â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ METRIC   â”‚ INBOUND      â”‚ OUTBOUND             â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Traffic  â”‚ 850 Mbps     â”‚ 640 Mbps             â”‚
â”‚ Packets  â”‚ 120k pps     â”‚ 98k pps              â”‚
â”‚ Flows    â”‚ 15,234       â”‚ 12,891               â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

#### 3. Anomaly Detection
```bash
bgpin flow anomaly
bgpin flow anomaly --severity high
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ Detected Traffic Anomalies                                                         â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ TIME     â”‚ TYPE  â”‚ SEVERITY â”‚ PREFIX          â”‚ ASN     â”‚ DESCRIPTION              â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ 11:45:23 â”‚ DDoS  â”‚ CRITICAL â”‚ 8.8.8.0/24      â”‚ AS15169 â”‚ High PPS detected (250k) â”‚
â”‚ 11:42:15 â”‚ Spike â”‚ HIGH     â”‚ 1.1.1.0/24      â”‚ AS13335 â”‚ Traffic spike (2.5 Gbps) â”‚
â”‚ 11:38:07 â”‚ Drop  â”‚ MEDIUM   â”‚ 208.67.222.0/24 â”‚ AS36692 â”‚ Traffic drop (80%)       â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

#### 4. Upstream Comparison
```bash
bgpin flow upstream-compare
bgpin flow upstream-compare --prefix 8.8.8.0/24
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ Upstream Provider Comparison                                              â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ PROVIDER â”‚ ASN    â”‚ AS PATH    â”‚ TRAFFIC   â”‚ PPS  â”‚ LATENCY   â”‚ LOSS      â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Telia    â”‚ AS1299 â”‚ 1299 15169 â”‚ 850 Mbps  â”‚ 120k â”‚ 42ms      â”‚ 0.1%      â”‚
â”‚ Level3   â”‚ AS3356 â”‚ 3356 15169 â”‚ 640 Mbps  â”‚ 98k  â”‚ 38ms      â”‚ 0.2%      â”‚
â”‚ GTT      â”‚ AS3257 â”‚ 3257 15169 â”‚ 1.2 Gbps  â”‚ 150k â”‚ 55ms      â”‚ 0.3%      â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### Casos de Uso

#### 1. DetecÃ§Ã£o de DDoS
```bash
# Monitorar anomalias em tempo real
watch -n 5 'bgpin flow anomaly --severity critical'
```

#### 2. AnÃ¡lise de TrÃ¡fego por ASN
```bash
# Ver trÃ¡fego de um ASN especÃ­fico
bgpin flow asn 262978 -o json | jq '.inbound_bps'
```

#### 3. Engenharia de TrÃ¡fego
```bash
# Comparar upstreams para decisÃµes de roteamento
bgpin flow upstream-compare --prefix 8.8.8.0/24
```

#### 4. ValidaÃ§Ã£o BGP vs TrÃ¡fego Real
```bash
# Verificar se prefixos anunciados tÃªm trÃ¡fego
bgpin asn prefixes 262978 -o json | \
  jq -r '.prefixes[].prefix' | \
  xargs -I {} bgpin flow top --prefix {}
```

### IntegraÃ§Ã£o com BGP

#### CorrelaÃ§Ã£o AutomÃ¡tica
```bash
# bgpin correlaciona automaticamente:
# - Prefixos anunciados (BGP)
# - TrÃ¡fego real (Flow)
# - AS Path (BGP)
# - Performance (Flow)

bgpin flow top --correlate-bgp
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚ BGP + Flow Correlation                                                              â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ PREFIX          â”‚ ASN     â”‚ AS PATH       â”‚ TRAFFIC   â”‚ PPS  â”‚ BGP STATUS           â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ 8.8.8.0/24      â”‚ AS15169 â”‚ 1299 15169    â”‚ 850 Mbps  â”‚ 120k â”‚ âœ“ Announced          â”‚
â”‚ 1.1.1.0/24      â”‚ AS13335 â”‚ 3356 13335    â”‚ 640 Mbps  â”‚ 98k  â”‚ âœ“ Announced          â”‚
â”‚ 192.0.2.0/24    â”‚ AS64512 â”‚ -             â”‚ 1.2 Gbps  â”‚ 150k â”‚ âœ— NOT ANNOUNCED      â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### Alertas

#### Anomalias CrÃ­ticas
```bash
# Configurar alerta para DDoS
bgpin flow anomaly --watch --severity critical --alert-webhook https://slack.com/webhook
```

#### TrÃ¡fego NÃ£o Anunciado
```bash
# Alertar sobre trÃ¡fego em prefixos nÃ£o anunciados
bgpin flow top --check-bgp --alert-unanounced
```

## PrÃ³ximos Passos

- [ ] Implementar coletor NetFlow/sFlow/IPFIX
- [ ] Integrar com goflow
- [ ] Storage em time-series DB (InfluxDB, TimescaleDB)
- [ ] Dashboard Grafana
- [ ] Alertas automÃ¡ticos
- [ ] Machine Learning para detecÃ§Ã£o de anomalias
- [ ] CorrelaÃ§Ã£o BGP + Flow em tempo real

---

**Observabilidade Enterprise para BGP** ðŸ”
