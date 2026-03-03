# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- `bgpin flow top` - Top prefixes by traffic (prepared)
- `bgpin flow asn` - ASN traffic statistics (prepared)
- `bgpin flow anomaly` - Detect anomalies (prepared)
- `bgpin flow upstream-compare` - Compare upstreams (prepared)
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
- NetFlow/sFlow/IPFIX collector implementation
- goflow integration
- Time-series database storage
- Grafana dashboards
- Machine learning for anomaly detection
- Multiple Looking Glass support
- Vendor-specific parsers (Cisco, Juniper, FRR)
- RPKI validation
- Intelligent caching
- Interactive TUI mode
- PeeringDB integration
- RouteViews integration

---

[0.1.0]: https://github.com/rsdenck/bgpin/releases/tag/v0.1.0
