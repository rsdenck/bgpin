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
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚                     Top Prefixes by Traffic                          â”‚
â”œâ”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ # â”‚ Prefix           â”‚ ASN      â”‚ Traffic  â”‚ PPS    â”‚ Top Protocol  â”‚
â”‚ 1 â”‚ 8.8.8.0/24       â”‚ AS15169  â”‚ 850 Mbps â”‚ 120k   â”‚ TCP (443)     â”‚
â”‚ 2 â”‚ 1.1.1.0/24       â”‚ AS13335  â”‚ 640 Mbps â”‚ 98k    â”‚ UDP (53)      â”‚
â”‚ 3 â”‚ 208.67.222.0/24  â”‚ AS36692  â”‚ 1.2 Gbps â”‚ 150k   â”‚ TCP (80)      â”‚
â•°â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### View ASN Traffic Statistics

```bash
bgpin flow asn 15169
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚       Traffic Statistics for AS15169             â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Metric   â”‚ Inbound      â”‚ Outbound             â”‚
â”‚ Traffic  â”‚ 850 Mbps     â”‚ 640 Mbps             â”‚
â”‚ Packets  â”‚ 120k pps     â”‚ 98k pps              â”‚
â”‚ Flows    â”‚ 15,234       â”‚ 12,891               â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### Detect Traffic Anomalies

```bash
bgpin flow anomaly
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚                      Detected Traffic Anomalies                            â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Time     â”‚ Type   â”‚ Severity â”‚ Prefix           â”‚ ASN      â”‚ Description  â”‚
â”‚ 11:45:23 â”‚ DDoS   â”‚ CRITICAL â”‚ 8.8.8.0/24       â”‚ AS15169  â”‚ High PPS     â”‚
â”‚ 11:42:15 â”‚ Spike  â”‚ HIGH     â”‚ 1.1.1.0/24       â”‚ AS13335  â”‚ Traffic spikeâ”‚
â”‚ 11:38:07 â”‚ Drop   â”‚ MEDIUM   â”‚ 208.67.222.0/24  â”‚ AS36692  â”‚ Traffic drop â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### Compare Upstream Providers

```bash
bgpin flow upstream-compare
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚                      Upstream Provider Comparison                            â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Provider â”‚ ASN      â”‚ AS Path      â”‚ Traffic  â”‚ PPS    â”‚ Latency â”‚ Loss    â”‚
â”‚ Telia    â”‚ AS1299   â”‚ 1299 15169   â”‚ 850 Mbps â”‚ 120k   â”‚ 42ms    â”‚ 0.1%    â”‚
â”‚ Level3   â”‚ AS3356   â”‚ 3356 15169   â”‚ 640 Mbps â”‚ 98k    â”‚ 38ms    â”‚ 0.2%    â”‚
â”‚ GTT      â”‚ AS3257   â”‚ 3257 15169   â”‚ 1.2 Gbps â”‚ 150k   â”‚ 55ms    â”‚ 0.3%    â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
```

### View Collector Statistics

```bash
bgpin flow stats
```

Output:
```
â•­â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•®
â”‚      Flow Collector Statistics             â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚ Metric             â”‚ Value                 â”‚
â”‚ NetFlow Packets    â”‚ 125,432               â”‚
â”‚ sFlow Packets      â”‚ 89,234                â”‚
â”‚ IPFIX Packets      â”‚ 45,123                â”‚
â”‚ Total Flows        â”‚ 259,789               â”‚
â”‚ Dropped Flows      â”‚ 12                    â”‚
â”‚ Processing Errors  â”‚ 3                     â”‚
â”‚ Last Update        â”‚ 2026-03-03 12:35:42   â”‚
â•°â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â•¯
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
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  Router/Switch  â”‚
â”‚  (NetFlow/sFlow)â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â”‚ UDP
         â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  bgpin Listener â”‚
â”‚  (Port 2055)    â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â”‚
         â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  Flow Decoder   â”‚
â”‚  (goflow2)      â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â”‚
         â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  Aggregator     â”‚
â”‚  (60s window)   â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â”‚
         â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  BGP Correlator â”‚
â”‚  (RIPE RIS SDK) â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â”‚
         â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  Anomaly Detect â”‚
â”‚  (ML/Threshold) â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â”‚
         â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  CLI Output     â”‚
â”‚  (go-pretty)    â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
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
