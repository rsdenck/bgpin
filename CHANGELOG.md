# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-03-03

### Added
- Complete NetFlow/sFlow/IPFIX collector implementation
- Real-time flow aggregation and processing
- BGP correlation for traffic analysis
- Anomaly detection (DDoS, traffic spikes, drops)
- GitHub Actions workflow for Linux builds (amd64, arm64, 386)
- Comprehensive flow collector documentation

#### Flow Collector Features
- Multi-protocol support (NetFlow v5/v9/v10, sFlow v5, IPFIX)
- Concurrent flow processing with worker pools
- Configurable aggregation windows
- ASN-level traffic statistics
- Memory-efficient with flow limits
- Thread-safe operations

#### New CLI Commands
- `bgpin flow top` - Show top prefixes by traffic (real data)
- `bgpin flow asn <asn>` - ASN traffic statistics (real data)
- `bgpin flow anomaly` - Detect traffic anomalies (real data)
- `bgpin flow upstream-compare` - Compare upstream providers
- `bgpin flow stats` - Show collector statistics

#### Configuration
- Flow collection settings in bgpin.yaml
- NetFlow, sFlow, IPFIX listener configuration
- Worker pool and buffer size tuning
- BGP correlation toggle
- Anomaly detection thresholds

#### Documentation
- docs/FLOW_COLLECTOR.md - Complete flow collector guide
- Exporter configuration examples (Cisco, Juniper, Linux)
- Performance tuning guidelines
- Troubleshooting section
- Architecture diagrams

### Changed
- Updated README with flow collector features
- Enhanced CLI output with real flow data when available
- Improved error handling in flow processing

### Technical Details
- GoFlowCollector with UDP listeners
- Flow aggregation with time windows
- BGP correlation engine
- Anomaly detection algorithms
- Configurable thresholds and limits

### CI/CD
- GitHub Actions for automated builds
- Multi-architecture Linux support (amd64, arm64, 386)
- Automated releases with checksums
- Test coverage reporting

## [0.1.0] - 2026-03-03

### Added
- Initial release of bgpin
- Complete RIPE RIS SDK with 5 API methods
- CLI with ASN and Prefix commands
- Professional UX with go-pretty/table
- OpenTelemetry integration for observability
- Flow analysis structure (NetFlow/sFlow/IPFIX)
- Multiple output formats (table, JSON, YAML)
- Rate limiting and retry with exponential backoff
- Context support for all operations
- 9 integration tests using real data (ASN 262978)
- Complete documentation

#### CLI Commands
- `bgpin asn info` - Get ASN information
- `bgpin asn neighbors` - List BGP neighbors
- `bgpin asn prefixes` - Show announced prefixes
- `bgpin asn peers` - Show RIS peers
- `bgpin prefix overview` - Get prefix overview
- `bgpin lg` - List looking glasses
- `bgpin version` - Show version information

#### SDK Features
- GetASNInfo() - ASN information
- GetASNNeighbors() - BGP neighbors
- GetAnnouncedPrefixes() - Announced prefixes
- GetPrefixOverview() - Prefix details
- GetRISPeers() - RIS peers by RRC

#### Telemetry
- Distributed tracing with OpenTelemetry
- Performance metrics
- Query latency tracking
- Error counters
- Export to stdout, OTLP, Jaeger

#### Documentation
- README.md - Main documentation
- docs/CLI_GUIDE.md - Complete CLI guide
- docs/ARCHITECTURE.md - Architecture details
- docs/OUTPUT_EXAMPLES.md - Output examples
- docs/TELEMETRY.md - Telemetry guide
- QUICK_START.md - Quick start guide
- PROJECT_SUMMARY.md - Project summary
- UX_IMPLEMENTATION.md - UX implementation details

### Technical Details
- Go 1.25+
- Clean Architecture + Hexagonal
- Thread-safe operations
- Professional error handling
- IPv4 and IPv6 support
- Unicode table borders
- Automatic type detection

### Dependencies
- github.com/spf13/cobra - CLI framework
- github.com/spf13/viper - Configuration
- github.com/jedib0t/go-pretty/v6 - Table formatting
- go.opentelemetry.io/otel - Observability
- golang.org/x/time/rate - Rate limiting
- gopkg.in/yaml.v3 - YAML support

## [Unreleased]

### Planned
- Time-series database storage (InfluxDB, TimescaleDB)
- Grafana dashboard templates
- Machine learning for advanced anomaly detection
- Multiple Looking Glass support
- Vendor-specific parsers (Cisco, Juniper, FRR)
- RPKI validation
- Intelligent caching
- Interactive TUI mode
- PeeringDB integration
- RouteViews integration
- Real-time streaming API
- GeoIP enrichment
- Export to Kafka/NATS

---

[0.2.0]: https://github.com/rsdenck/bgpin/releases/tag/v0.2.0
[0.1.0]: https://github.com/rsdenck/bgpin/releases/tag/v0.1.0
