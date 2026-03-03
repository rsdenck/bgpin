# bgpin Documentation

Complete documentation for the Platform Border Gateway Protocol CLI and SDK.

## 📚 Table of Contents

### Getting Started
- [Installation Guide](INSTALLATION.md) - Complete installation instructions
- [Quick Start](QUICK_START.md) - Get started in 5 minutes
- [CLI Guide](CLI_GUIDE.md) - All CLI commands and examples

### Features
- [Flow Collector](FLOW_COLLECTOR.md) - NetFlow/sFlow/IPFIX setup and usage
- [Telemetry](TELEMETRY.md) - OpenTelemetry integration
- [Output Examples](OUTPUT_EXAMPLES.md) - Visual examples of CLI output

### Architecture & Design
- [Architecture](ARCHITECTURE.md) - System design and structure
- [UX Implementation](UX_IMPLEMENTATION.md) - User experience design
- [Project Summary](PROJECT_SUMMARY.md) - High-level project overview

### Vendors
- [Vendor Status](vendors/STATUS.md) - Implementation status of all vendors
- [Vendor Documentation](vendors/README.md) - Vendor-specific guides

### Releases
- [Release Notes](releases/README.md) - All release notes
- [v0.2.0](releases/v0.2.0.md) - Latest release
- [v0.1.0](releases/v0.1.0.md) - Initial release

### SDK
- [SDK Documentation](../sdk/README.md) - Go SDK documentation
- [SDK Examples](../sdk/examples/) - Code examples

## 🚀 Quick Links

### For Users
- [Installation](INSTALLATION.md) → [Quick Start](QUICK_START.md) → [CLI Guide](CLI_GUIDE.md)

### For Developers
- [Architecture](ARCHITECTURE.md) → [SDK Documentation](../sdk/README.md) → [Contributing](../CONTRIBUTING.md)

### For Network Engineers
- [Flow Collector](FLOW_COLLECTOR.md) → [Vendor Status](vendors/STATUS.md)

## 📖 Documentation by Topic

### BGP Operations
- Query ASN information
- List BGP neighbors
- View announced prefixes
- Analyze AS_PATH
- RPKI validation (planned)

### Flow Analysis
- NetFlow/sFlow/IPFIX collection
- Real-time traffic aggregation
- BGP correlation
- Anomaly detection
- Upstream comparison

### Vendor Integration
- Cisco (IOS, IOS-XE, IOS-XR, NX-OS) ✅
- Juniper (JunOS) ✅
- Arista (EOS) - Planned
- Nokia (SR OS) - Planned
- MikroTik (RouterOS) - Planned
- GoBGP - Planned

### Observability
- OpenTelemetry tracing
- Metrics collection
- Performance monitoring
- Error tracking

## 🔗 External Resources

- [GitHub Repository](https://github.com/rsdenck/bgpin)
- [Issue Tracker](https://github.com/rsdenck/bgpin/issues)
- [Discussions](https://github.com/rsdenck/bgpin/discussions)
- [Releases](https://github.com/rsdenck/bgpin/releases)

## 📝 Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on how to contribute to bgpin.

## 📄 License

bgpin is released under the MIT License. See [LICENSE](../LICENSE) for details.
