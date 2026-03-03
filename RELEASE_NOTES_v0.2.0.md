# Release Notes - bgpin v0.2.0

**Release Date:** March 3, 2026

## 🎉 Major Features

### NetFlow/sFlow/IPFIX Collector

bgpin v0.2.0 introduces a complete flow telemetry collector with real-time traffic analysis and BGP correlation.

#### Key Features

- **Multi-Protocol Support**: NetFlow v5/v9/v10 (IPFIX), sFlow v5
- **Real-Time Processing**: Concurrent flow processing with configurable worker pools
- **BGP Correlation**: Match traffic flows with BGP routing data
- **Anomaly Detection**: Automatic detection of DDoS, traffic spikes, and drops
- **ASN Analytics**: Detailed traffic statistics per ASN
- **Performance**: Memory-efficient with configurable limits and aggregation windows

### New CLI Commands

```bash
# View top prefixes by traffic
bgpin flow top

# ASN traffic statistics
bgpin flow asn 15169

# Detect traffic anomalies
bgpin flow anomaly

# Compare upstream providers
bgpin flow upstream-compare

# Collector statistics
bgpin flow stats
```

### GitHub Actions CI/CD

Automated builds for Linux platforms:
- Linux amd64
- Linux arm64
- Linux 386

All builds include checksums and are automatically released on tag push.

## 📦 Installation

### Download Pre-built Binaries

```bash
# Linux amd64
wget https://github.com/rsdenck/bgpin/releases/download/v0.2.0/bgpin-linux-amd64
chmod +x bgpin-linux-amd64
sudo mv bgpin-linux-amd64 /usr/local/bin/bgpin

# Linux arm64
wget https://github.com/rsdenck/bgpin/releases/download/v0.2.0/bgpin-linux-arm64
chmod +x bgpin-linux-arm64
sudo mv bgpin-linux-arm64 /usr/local/bin/bgpin
```

### Build from Source

```bash
git clone https://github.com/rsdenck/bgpin
cd bgpin
git checkout v0.2.0
go build -o bgpin ./cmd/cli/
```

## 🚀 Quick Start - Flow Collector

### 1. Configure bgpin

Copy the example configuration:

```bash
cp bgpin.yaml.example bgpin.yaml
```

Edit `bgpin.yaml` and enable flow collection:

```yaml
flow:
  enabled: true
  
  netflow:
    enabled: true
    addr: "0.0.0.0"
    port: 2055
  
  sflow:
    enabled: true
    addr: "0.0.0.0"
    port: 6343
  
  bgp_correlation: true
  anomaly_detection: true
```

### 2. Configure Your Router/Switch

#### Cisco IOS/IOS-XE

```
flow exporter BGPIN
 destination <bgpin-server-ip> 2055
 transport udp 2055

flow monitor BGPIN-MONITOR
 exporter BGPIN
 record netflow ipv4 original-input

interface GigabitEthernet0/0
 ip flow monitor BGPIN-MONITOR input
```

#### Juniper JunOS

```
set protocols sflow collector <bgpin-server-ip> udp-port 6343
set protocols sflow interfaces ge-0/0/0
set protocols sflow sample-rate ingress 1000
```

### 3. Start bgpin and View Traffic

```bash
# View top prefixes
bgpin flow top

# Monitor specific ASN
bgpin flow asn 15169

# Watch for anomalies
bgpin flow anomaly
```

## 📊 Example Output

### Top Prefixes by Traffic

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

### Traffic Anomalies

```
╭────────────────────────────────────────────────────────────────────────────╮
│                      Detected Traffic Anomalies                            │
├──────────┬────────┬──────────┬──────────────────┬──────────┬──────────────┤
│ Time     │ Type   │ Severity │ Prefix           │ ASN      │ Description  │
│ 11:45:23 │ DDoS   │ CRITICAL │ 8.8.8.0/24       │ AS15169  │ High PPS     │
│ 11:42:15 │ Spike  │ HIGH     │ 1.1.1.0/24       │ AS13335  │ Traffic spike│
╰──────────┴────────┴──────────┴──────────────────┴──────────┴──────────────╯
```

## 🔧 Configuration Options

### Flow Collector Settings

```yaml
flow:
  enabled: true
  
  # Listeners
  netflow:
    enabled: true
    addr: "0.0.0.0"
    port: 2055
  
  sflow:
    enabled: true
    addr: "0.0.0.0"
    port: 6343
  
  ipfix:
    enabled: true
    addr: "0.0.0.0"
    port: 4739
  
  # Performance tuning
  workers: 4              # Flow processing workers
  buffer_size: 10000      # Flow buffer size
  aggregate_window: 60    # Aggregation window (seconds)
  max_flows: 100000       # Max flows in memory
  
  # Features
  bgp_correlation: true   # Enable BGP correlation
  anomaly_detection: true # Enable anomaly detection
  anomaly_window: 300     # Anomaly detection window (seconds)
```

## 📚 Documentation

- [Flow Collector Guide](docs/FLOW_COLLECTOR.md) - Complete setup and usage guide
- [CLI Guide](docs/CLI_GUIDE.md) - All CLI commands
- [Telemetry Guide](docs/TELEMETRY.md) - OpenTelemetry integration
- [Architecture](docs/ARCHITECTURE.md) - System design

## 🔄 What's Changed

### Added
- Complete NetFlow/sFlow/IPFIX collector implementation
- Real-time flow aggregation and processing
- BGP correlation engine
- Anomaly detection algorithms
- GitHub Actions workflow for Linux builds
- Comprehensive flow collector documentation
- New flow CLI commands with real data support

### Changed
- Updated README with flow collector features
- Enhanced CLI output with real flow data when available
- Improved error handling in flow processing

### Technical Improvements
- GoFlowCollector with UDP listeners
- Concurrent flow processing with worker pools
- Thread-safe aggregation
- Memory-efficient flow management
- Configurable thresholds and limits

## 🐛 Known Issues

- Flow collector uses simplified NetFlow/sFlow parsing (full goflow2 integration planned)
- Anomaly detection uses basic threshold-based algorithms (ML planned)
- No persistent storage yet (time-series DB planned)

## 🚧 Roadmap

### v0.3.0 (Planned)
- Time-series database storage (InfluxDB, TimescaleDB)
- Grafana dashboard templates
- Machine learning anomaly detection
- GeoIP enrichment
- Real-time streaming API

### v0.4.0 (Planned)
- Multiple Looking Glass support
- Vendor-specific parsers (Cisco, Juniper, FRR)
- RPKI validation
- Interactive TUI mode
- PeeringDB integration

## 🙏 Acknowledgments

- RIPE NCC for the RIS API
- goflow2 project for flow collection inspiration
- go-pretty for beautiful table formatting
- OpenTelemetry community

## 📝 Full Changelog

See [CHANGELOG.md](CHANGELOG.md) for complete details.

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**Download:** [GitHub Releases](https://github.com/rsdenck/bgpin/releases/tag/v0.2.0)

**Issues:** [GitHub Issues](https://github.com/rsdenck/bgpin/issues)

**Discussions:** [GitHub Discussions](https://github.com/rsdenck/bgpin/discussions)
