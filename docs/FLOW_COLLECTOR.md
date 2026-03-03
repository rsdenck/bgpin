# Flow Collector Guide

## Overview

bgpin includes a built-in NetFlow/sFlow/IPFIX collector that enables real-time traffic analysis and correlation with BGP routing data. This provides visibility into actual traffic patterns alongside routing information.

## Features

- NetFlow v5/v9/v10 (IPFIX) collection
- sFlow v5 collection
- Real-time traffic aggregation
- BGP correlation (match traffic to announced prefixes)
- Anomaly detection (DDoS, traffic spikes, drops)
- Multi-upstream comparison
- ASN-level traffic statistics

## Configuration

### Enable Flow Collection

Edit `bgpin.yaml`:

```yaml
flow:
  enabled: true
  
  # NetFlow configuration
  netflow:
    enabled: true
    addr: "0.0.0.0"
    port: 2055
  
  # sFlow configuration
  sflow:
    enabled: true
    addr: "0.0.0.0"
    port: 6343
  
  # IPFIX configuration
  ipfix:
    enabled: true
    addr: "0.0.0.0"
    port: 4739
  
  # Processing configuration
  workers: 4
  buffer_size: 10000
  bgp_correlation: true
  
  # Aggregation settings
  aggregate_window: 60
  max_flows: 100000
  
  # Anomaly detection
  anomaly_detection: true
  anomaly_window: 300
```

### Configure Exporters

#### Cisco IOS/IOS-XE

```
! NetFlow v9
flow exporter BGPIN
 destination <bgpin-server-ip> 2055
 transport udp 2055
 template data timeout 60

flow monitor BGPIN-MONITOR
 exporter BGPIN
 record netflow ipv4 original-input

interface GigabitEthernet0/0
 ip flow monitor BGPIN-MONITOR input
```

#### Juniper JunOS

```
# sFlow
set protocols sflow collector <bgpin-server-ip> udp-port 6343
set protocols sflow interfaces ge-0/0/0
set protocols sflow sample-rate ingress 1000
```

#### Linux (softflowd)

```bash
# Install softflowd
apt-get install softflowd

# Configure
softflowd -i eth0 -n <bgpin-server-ip>:2055 -v 9
```

## Usage

### View Top Prefixes by Traffic

```bash
bgpin flow top
```

Output:
```
╭──────────────────────────────────────────────────────────────────────╮
│                     Top Prefixes by Traffic                          │
├───┬──────────────────┬──────────┬──────────┬────────┬───────────────┤
│ # │ Prefix           │ ASN      │ Traffic  │ PPS    │ Top Protocol  │
│ 1 │ 8.8.8.0/24       │ AS15169  │ 850 Mbps │ 120k   │ TCP (443)     │
│ 2 │ 1.1.1.0/24       │ AS13335  │ 640 Mbps │ 98k    │ UDP (53)      │
│ 3 │ 208.67.222.0/24  │ AS36692  │ 1.2 Gbps │ 150k   │ TCP (80)      │
╰───┴──────────────────┴──────────┴──────────┴────────┴───────────────╯
```

### View ASN Traffic Statistics

```bash
bgpin flow asn 15169
```

Output:
```
╭──────────────────────────────────────────────────╮
│       Traffic Statistics for AS15169             │
├──────────┬──────────────┬──────────────────────┤
│ Metric   │ Inbound      │ Outbound             │
│ Traffic  │ 850 Mbps     │ 640 Mbps             │
│ Packets  │ 120k pps     │ 98k pps              │
│ Flows    │ 15,234       │ 12,891               │
╰──────────┴──────────────┴──────────────────────╯
```

### Detect Traffic Anomalies

```bash
bgpin flow anomaly
```

Output:
```
╭────────────────────────────────────────────────────────────────────────────╮
│                      Detected Traffic Anomalies                            │
├──────────┬────────┬──────────┬──────────────────┬──────────┬──────────────┤
│ Time     │ Type   │ Severity │ Prefix           │ ASN      │ Description  │
│ 11:45:23 │ DDoS   │ CRITICAL │ 8.8.8.0/24       │ AS15169  │ High PPS     │
│ 11:42:15 │ Spike  │ HIGH     │ 1.1.1.0/24       │ AS13335  │ Traffic spike│
│ 11:38:07 │ Drop   │ MEDIUM   │ 208.67.222.0/24  │ AS36692  │ Traffic drop │
╰──────────┴────────┴──────────┴──────────────────┴──────────┴──────────────╯
```

### Compare Upstream Providers

```bash
bgpin flow upstream-compare
```

Output:
```
╭──────────────────────────────────────────────────────────────────────────────╮
│                      Upstream Provider Comparison                            │
├──────────┬──────────┬──────────────┬──────────┬────────┬─────────┬─────────┤
│ Provider │ ASN      │ AS Path      │ Traffic  │ PPS    │ Latency │ Loss    │
│ Telia    │ AS1299   │ 1299 15169   │ 850 Mbps │ 120k   │ 42ms    │ 0.1%    │
│ Level3   │ AS3356   │ 3356 15169   │ 640 Mbps │ 98k    │ 38ms    │ 0.2%    │
│ GTT      │ AS3257   │ 3257 15169   │ 1.2 Gbps │ 150k   │ 55ms    │ 0.3%    │
╰──────────┴──────────┴──────────────┴──────────┴────────┴─────────┴─────────╯
```

### View Collector Statistics

```bash
bgpin flow stats
```

Output:
```
╭────────────────────────────────────────────╮
│      Flow Collector Statistics             │
├────────────────────┬───────────────────────┤
│ Metric             │ Value                 │
│ NetFlow Packets    │ 125,432               │
│ sFlow Packets      │ 89,234                │
│ IPFIX Packets      │ 45,123                │
│ Total Flows        │ 259,789               │
│ Dropped Flows      │ 12                    │
│ Processing Errors  │ 3                     │
│ Last Update        │ 2026-03-03 12:35:42   │
╰────────────────────┴───────────────────────╯
```

## BGP Correlation

When `bgp_correlation` is enabled, the flow collector correlates traffic data with BGP routing information:

- Match destination IPs to announced prefixes
- Identify which ASN is receiving traffic
- Detect traffic to prefixes not in BGP table (potential hijack)
- Compare announced routes vs actual traffic patterns

Example workflow:

```bash
# Query BGP data for an ASN
bgpin asn prefixes 15169

# View actual traffic to that ASN
bgpin flow asn 15169

# Compare announced prefixes with traffic
bgpin flow top --asn 15169
```

## Anomaly Detection

The collector includes built-in anomaly detection:

### DDoS Detection
- High packet rate (>100k pps)
- Unusual protocol distribution
- Single source flooding

### Traffic Spikes
- Sudden increase in bandwidth (>2x baseline)
- Rapid flow count increase

### Traffic Drops
- Sudden decrease in traffic (>80% drop)
- Complete loss of flows

### Configuration

Adjust thresholds in code or future config:

```yaml
flow:
  anomaly_detection: true
  anomaly_thresholds:
    ddos_pps: 100000
    spike_multiplier: 2.0
    drop_percentage: 80
```

## Performance Considerations

### Memory Usage
- Each flow record: ~200 bytes
- 100k flows: ~20 MB
- Adjust `max_flows` based on available memory

### CPU Usage
- Workers: 4 (default)
- Increase for high-traffic environments
- Monitor with `bgpin flow stats`

### Network
- NetFlow/sFlow adds ~1-5% overhead on exporters
- Use sampling on high-traffic interfaces
- Recommended sample rates:
  - <1 Gbps: 1:100
  - 1-10 Gbps: 1:1000
  - >10 Gbps: 1:10000

## Troubleshooting

### No flows received

1. Check firewall rules:
```bash
# Allow NetFlow
iptables -A INPUT -p udp --dport 2055 -j ACCEPT

# Allow sFlow
iptables -A INPUT -p udp --dport 6343 -j ACCEPT
```

2. Verify exporter configuration:
```bash
# Cisco
show flow exporter statistics

# Juniper
show sflow
```

3. Check bgpin logs:
```bash
bgpin flow stats
```

### High dropped flows

Increase buffer size:
```yaml
flow:
  buffer_size: 50000  # Increase from 10000
  workers: 8          # Increase workers
```

### Memory issues

Reduce max flows:
```yaml
flow:
  max_flows: 50000    # Reduce from 100000
```

## Integration Examples

### Grafana Dashboard

Export metrics to Prometheus/InfluxDB (future feature):

```yaml
flow:
  export:
    enabled: true
    type: prometheus
    endpoint: "http://localhost:9090"
```

### Alerting

Webhook notifications for anomalies (future feature):

```yaml
flow:
  alerts:
    enabled: true
    webhook: "https://alerts.example.com/webhook"
    severity: ["high", "critical"]
```

## Architecture

```
┌─────────────────┐
│  Router/Switch  │
│  (NetFlow/sFlow)│
└────────┬────────┘
         │ UDP
         ▼
┌─────────────────┐
│  bgpin Listener │
│  (Port 2055)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Flow Decoder   │
│  (goflow2)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Aggregator     │
│  (60s window)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  BGP Correlator │
│  (RIPE RIS SDK) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Anomaly Detect │
│  (ML/Threshold) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  CLI Output     │
│  (go-pretty)    │
└─────────────────┘
```

## Future Enhancements

- [ ] Time-series database storage (InfluxDB, TimescaleDB)
- [ ] Machine learning anomaly detection
- [ ] Grafana dashboard templates
- [ ] Real-time streaming API
- [ ] GeoIP enrichment
- [ ] AS-PATH correlation
- [ ] RPKI validation integration
- [ ] Export to Kafka/NATS
- [ ] Web UI for visualization

## References

- [NetFlow v9 RFC](https://www.rfc-editor.org/rfc/rfc3954)
- [IPFIX RFC](https://www.rfc-editor.org/rfc/rfc7011)
- [sFlow v5 Specification](https://sflow.org/sflow_version_5.txt)
- [goflow2 Documentation](https://github.com/netsampler/goflow2)
