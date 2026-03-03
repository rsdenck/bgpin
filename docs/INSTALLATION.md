# Installation Guide

## Quick Install (Linux)

### Download Pre-built Binary

```bash
# Linux amd64 (most common)
wget https://github.com/rsdenck/bgpin/releases/latest/download/bgpin-linux-amd64
chmod +x bgpin-linux-amd64
sudo mv bgpin-linux-amd64 /usr/local/bin/bgpin

# Linux arm64 (Raspberry Pi, ARM servers)
wget https://github.com/rsdenck/bgpin/releases/latest/download/bgpin-linux-arm64
chmod +x bgpin-linux-arm64
sudo mv bgpin-linux-arm64 /usr/local/bin/bgpin

# Linux 386 (32-bit)
wget https://github.com/rsdenck/bgpin/releases/latest/download/bgpin-linux-386
chmod +x bgpin-linux-386
sudo mv bgpin-linux-386 /usr/local/bin/bgpin
```

### Verify Installation

```bash
bgpin version
```

## Build from Source

### Prerequisites

- Go 1.25 or later
- Git

### Steps

```bash
# Clone repository
git clone https://github.com/rsdenck/bgpin.git
cd bgpin

# Build
go build -o bgpin ./cmd/cli/

# Install globally (optional)
sudo mv bgpin /usr/local/bin/

# Or install with go install
go install ./cmd/cli/
```

## Configuration

### Create Configuration File

```bash
# Copy example configuration
cp bgpin.yaml.example bgpin.yaml

# Edit configuration
nano bgpin.yaml
```

### Basic Configuration

```yaml
# General settings
timeout: 30
output: table

# RIPE RIS SDK
ripe:
  rate_limit: 10
  retry_max: 3
```

### Flow Collector Configuration (Optional)

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

## First Steps

### Test Basic Commands

```bash
# Get ASN information
bgpin asn info 15169

# List BGP neighbors
bgpin asn neighbors 15169

# Show announced prefixes
bgpin asn prefixes 15169

# Prefix overview
bgpin prefix overview 8.8.8.0/24
```

### Enable Flow Collection (Optional)

1. Configure flow settings in `bgpin.yaml`
2. Configure your router/switch to export flows
3. Start bgpin and verify:

```bash
bgpin flow stats
bgpin flow top
```

## System Requirements

### Minimum
- CPU: 1 core
- RAM: 512 MB
- Disk: 50 MB

### Recommended (with Flow Collector)
- CPU: 2+ cores
- RAM: 2+ GB
- Disk: 100 MB
- Network: 100 Mbps+

## Firewall Configuration

### For Flow Collection

```bash
# Allow NetFlow (UDP 2055)
sudo ufw allow 2055/udp

# Allow sFlow (UDP 6343)
sudo ufw allow 6343/udp

# Allow IPFIX (UDP 4739)
sudo ufw allow 4739/udp
```

Or with iptables:

```bash
sudo iptables -A INPUT -p udp --dport 2055 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 6343 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 4739 -j ACCEPT
```

## Running as Service (systemd)

### Create Service File

```bash
sudo nano /etc/systemd/system/bgpin-flow.service
```

```ini
[Unit]
Description=bgpin Flow Collector
After=network.target

[Service]
Type=simple
User=bgpin
WorkingDirectory=/opt/bgpin
ExecStart=/usr/local/bin/bgpin flow stats
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Enable and Start

```bash
sudo systemctl daemon-reload
sudo systemctl enable bgpin-flow
sudo systemctl start bgpin-flow
sudo systemctl status bgpin-flow
```

## Docker Installation (Coming Soon)

```bash
# Pull image
docker pull ghcr.io/rsdenck/bgpin:latest

# Run
docker run -p 2055:2055/udp -p 6343:6343/udp ghcr.io/rsdenck/bgpin:latest
```

## Troubleshooting

### Command not found

```bash
# Check if binary is in PATH
which bgpin

# Add to PATH if needed
export PATH=$PATH:/usr/local/bin
```

### Permission denied

```bash
# Make binary executable
chmod +x bgpin

# Or run with sudo for privileged ports
sudo bgpin flow stats
```

### Flow collector not receiving data

1. Check firewall rules
2. Verify exporter configuration
3. Check bgpin logs
4. Test with tcpdump:

```bash
sudo tcpdump -i any -n udp port 2055
```

## Upgrade

### From Binary

```bash
# Download new version
wget https://github.com/rsdenck/bgpin/releases/latest/download/bgpin-linux-amd64

# Replace old binary
sudo mv bgpin-linux-amd64 /usr/local/bin/bgpin
chmod +x /usr/local/bin/bgpin

# Verify
bgpin version
```

### From Source

```bash
cd bgpin
git pull
go build -o bgpin ./cmd/cli/
sudo mv bgpin /usr/local/bin/
```

## Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/bgpin

# Remove configuration
rm ~/.config/bgpin/bgpin.yaml

# Remove service (if installed)
sudo systemctl stop bgpin-flow
sudo systemctl disable bgpin-flow
sudo rm /etc/systemd/system/bgpin-flow.service
```

## Next Steps

- Read [CLI Guide](docs/CLI_GUIDE.md) for all commands
- Configure [Flow Collector](docs/FLOW_COLLECTOR.md) for traffic analysis
- Set up [Telemetry](docs/TELEMETRY.md) for observability
- Check [Architecture](docs/ARCHITECTURE.md) to understand the design

## Support

- Issues: https://github.com/rsdenck/bgpin/issues
- Discussions: https://github.com/rsdenck/bgpin/discussions
- Documentation: https://github.com/rsdenck/bgpin/tree/main/docs
