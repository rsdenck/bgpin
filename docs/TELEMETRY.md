# bgpin - Telemetry & Observability

## OpenTelemetry Integration

bgpin integra OpenTelemetry para observabilidade completa de todas as operações BGP.

### Características

- ✅ Distributed tracing
- ✅ Métricas de performance
- ✅ Latência de queries
- ✅ Contadores de erros
- ✅ Exportação para OTLP, Jaeger, Prometheus

### Configuração

```yaml
# bgpin.yaml
telemetry:
  enabled: true
  export_type: stdout  # stdout, otlp, jaeger
  endpoint: "localhost:4317"
  
  # Métricas
  metrics:
    enabled: true
    interval: 60s
  
  # Traces
  traces:
    enabled: true
    sample_rate: 1.0  # 100% sampling
```

### Traces Disponíveis

#### 1. Query Traces
```
bgpin.query
├── bgpin.sdk.get_asn_info
│   ├── http.request
│   └── json.unmarshal
├── bgpin.output.format
└── bgpin.render.table
```

#### 2. Attributes
- `bgp.provider` - Provider do Looking Glass
- `bgp.prefix` - Prefixo consultado
- `bgp.asn` - ASN consultado
- `bgp.as_path` - AS Path
- `cli.command` - Comando executado
- `cli.output_format` - Formato de saída
- `result.count` - Número de resultados
- `duration_ms` - Duração em milissegundos

### Métricas Disponíveis

#### Counters
- `bgpin.queries.total` - Total de queries executadas
- `bgpin.errors.total` - Total de erros
- `bgpin.prefixes.total` - Total de prefixos consultados
- `bgpin.neighbors.total` - Total de vizinhos consultados

#### Histograms
- `bgpin.query.duration` - Duração das queries (ms)

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

### Visualização

#### Jaeger UI
```
http://localhost:16686
```

Visualize:
- Trace completo de cada query
- Latência por componente
- Erros e exceções
- Dependências entre serviços

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
┌─────────────┐
│   Router    │ NetFlow/sFlow
│  (Exporter) │────────────┐
└─────────────┘            │
                           ▼
┌─────────────┐      ┌──────────────┐
│   Switch    │─────▶│ bgpin Flow   │
│  (Exporter) │      │  Collector   │
└─────────────┘      └──────┬───────┘
                            │
                            ▼
                     ┌──────────────┐
                     │  Aggregator  │
                     │  & Analyzer  │
                     └──────┬───────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   Storage    │
                     │ (Time Series)│
                     └──────────────┘
```

### Configuração

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
╭────────────────────────────────────────────────────────────────────────╮
│ Top Prefixes by Traffic                                                │
├───┬─────────────────┬─────────┬──────────────────┬──────┬──────────────┤
│ # │ PREFIX          │ ASN     │ TRAFFIC          │ PPS  │ TOP PROTOCOL │
├───┼─────────────────┼─────────┼──────────────────┼──────┼──────────────┤
│ 1 │ 8.8.8.0/24      │ AS15169 │ 850 Mbps         │ 120k │ TCP (443)    │
│ 2 │ 1.1.1.0/24      │ AS13335 │ 640 Mbps         │ 98k  │ UDP (53)     │
│ 3 │ 208.67.222.0/24 │ AS36692 │ 1.2 Gbps         │ 150k │ TCP (80)     │
╰───┴─────────────────┴─────────┴──────────────────┴──────┴──────────────╯
```

#### 2. ASN Traffic
```bash
bgpin flow asn 15169
bgpin flow asn AS262978
```

Output:
```
╭────────────────────────────────────────────────╮
│ Traffic Statistics for AS15169                 │
├──────────┬──────────────┬──────────────────────┤
│ METRIC   │ INBOUND      │ OUTBOUND             │
├──────────┼──────────────┼──────────────────────┤
│ Traffic  │ 850 Mbps     │ 640 Mbps             │
│ Packets  │ 120k pps     │ 98k pps              │
│ Flows    │ 15,234       │ 12,891               │
╰──────────┴──────────────┴──────────────────────╯
```

#### 3. Anomaly Detection
```bash
bgpin flow anomaly
bgpin flow anomaly --severity high
```

Output:
```
╭────────────────────────────────────────────────────────────────────────────────────╮
│ Detected Traffic Anomalies                                                         │
├──────────┬───────┬──────────┬─────────────────┬─────────┬──────────────────────────┤
│ TIME     │ TYPE  │ SEVERITY │ PREFIX          │ ASN     │ DESCRIPTION              │
├──────────┼───────┼──────────┼─────────────────┼─────────┼──────────────────────────┤
│ 11:45:23 │ DDoS  │ CRITICAL │ 8.8.8.0/24      │ AS15169 │ High PPS detected (250k) │
│ 11:42:15 │ Spike │ HIGH     │ 1.1.1.0/24      │ AS13335 │ Traffic spike (2.5 Gbps) │
│ 11:38:07 │ Drop  │ MEDIUM   │ 208.67.222.0/24 │ AS36692 │ Traffic drop (80%)       │
╰──────────┴───────┴──────────┴─────────────────┴─────────┴──────────────────────────╯
```

#### 4. Upstream Comparison
```bash
bgpin flow upstream-compare
bgpin flow upstream-compare --prefix 8.8.8.0/24
```

Output:
```
╭───────────────────────────────────────────────────────────────────────────╮
│ Upstream Provider Comparison                                              │
├──────────┬────────┬────────────┬───────────┬──────┬───────────┬───────────┤
│ PROVIDER │ ASN    │ AS PATH    │ TRAFFIC   │ PPS  │ LATENCY   │ LOSS      │
├──────────┼────────┼────────────┼───────────┼──────┼───────────┼───────────┤
│ Telia    │ AS1299 │ 1299 15169 │ 850 Mbps  │ 120k │ 42ms      │ 0.1%      │
│ Level3   │ AS3356 │ 3356 15169 │ 640 Mbps  │ 98k  │ 38ms      │ 0.2%      │
│ GTT      │ AS3257 │ 3257 15169 │ 1.2 Gbps  │ 150k │ 55ms      │ 0.3%      │
╰──────────┴────────┴────────────┴───────────┴──────┴───────────┴───────────╯
```

### Casos de Uso

#### 1. Detecção de DDoS
```bash
# Monitorar anomalias em tempo real
watch -n 5 'bgpin flow anomaly --severity critical'
```

#### 2. Análise de Tráfego por ASN
```bash
# Ver tráfego de um ASN específico
bgpin flow asn 262978 -o json | jq '.inbound_bps'
```

#### 3. Engenharia de Tráfego
```bash
# Comparar upstreams para decisões de roteamento
bgpin flow upstream-compare --prefix 8.8.8.0/24
```

#### 4. Validação BGP vs Tráfego Real
```bash
# Verificar se prefixos anunciados têm tráfego
bgpin asn prefixes 262978 -o json | \
  jq -r '.prefixes[].prefix' | \
  xargs -I {} bgpin flow top --prefix {}
```

### Integração com BGP

#### Correlação Automática
```bash
# bgpin correlaciona automaticamente:
# - Prefixos anunciados (BGP)
# - Tráfego real (Flow)
# - AS Path (BGP)
# - Performance (Flow)

bgpin flow top --correlate-bgp
```

Output:
```
╭─────────────────────────────────────────────────────────────────────────────────────╮
│ BGP + Flow Correlation                                                              │
├─────────────────┬─────────┬───────────────┬───────────┬──────┬──────────────────────┤
│ PREFIX          │ ASN     │ AS PATH       │ TRAFFIC   │ PPS  │ BGP STATUS           │
├─────────────────┼─────────┼───────────────┼───────────┼──────┼──────────────────────┤
│ 8.8.8.0/24      │ AS15169 │ 1299 15169    │ 850 Mbps  │ 120k │ ✓ Announced          │
│ 1.1.1.0/24      │ AS13335 │ 3356 13335    │ 640 Mbps  │ 98k  │ ✓ Announced          │
│ 192.0.2.0/24    │ AS64512 │ -             │ 1.2 Gbps  │ 150k │ ✗ NOT ANNOUNCED      │
╰─────────────────┴─────────┴───────────────┴───────────┴──────┴──────────────────────╯
```

### Alertas

#### Anomalias Críticas
```bash
# Configurar alerta para DDoS
bgpin flow anomaly --watch --severity critical --alert-webhook https://slack.com/webhook
```

#### Tráfego Não Anunciado
```bash
# Alertar sobre tráfego em prefixos não anunciados
bgpin flow top --check-bgp --alert-unanounced
```

## Próximos Passos

- [ ] Implementar coletor NetFlow/sFlow/IPFIX
- [ ] Integrar com goflow
- [ ] Storage em time-series DB (InfluxDB, TimescaleDB)
- [ ] Dashboard Grafana
- [ ] Alertas automáticos
- [ ] Machine Learning para detecção de anomalias
- [ ] Correlação BGP + Flow em tempo real

---

**Observabilidade Enterprise para BGP** 🔍
