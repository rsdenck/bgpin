# bgpin CLI Command Reference

This document provides a complete reference of all available commands in the bgpin CLI.

## Table of Contents

- [General Commands](#general-commands)
- [BGP Lookup Commands](#bgp-lookup-commands)
- [ASN Query Commands](#asn-query-commands)
- [Prefix Query Commands](#prefix-query-commands)
- [Flow Analysis Commands](#flow-analysis-commands)
- [Route Analysis Commands](#route-analysis-commands)
- [Looking Glass Commands](#looking-glass-commands)

---

## General Commands

### version

Display version, build date, and runtime information.

```bash
bgpin version
```

**Output:**
- bgpin version
- Build date
- Go version
- OS/Arch

---

### help

Display help information for any command.

```bash
bgpin --help
bgpin [command] --help
```

---

### completion

Generate the autocompletion script for the specified shell.

```bash
bgpin completion [shell]
```

**Supported Shells:**
- bash
- zsh
- fish
- powershell

---

## BGP Lookup Commands

### lookup

Look up a prefix in BGP tables.

```bash
bgpin lookup [prefix]
```

**Arguments:**
- `prefix` - The IP prefix to look up (e.g., 8.8.8.0/24)

**Options:**
- `-l, --lg string` - Looking glass name
- `-o, --output string` - Output format: table, json, yaml (default: table)

**Examples:**
```bash
bgpin lookup 8.8.8.0/24
bgpin lookup 8.8.8.0/24 --lg "Hurricane Electric" -o json
```

---

### route show

Show routes for a specific prefix.

```bash
bgpin route show [prefix]
```

**Arguments:**
- `prefix` - The IP prefix to show routes for

---

### neighbors list

List all BGP neighbors.

```bash
bgpin neighbors list
```

---

## ASN Query Commands

Query Autonomous System Number information from RIPE RIS.

```bash
bgpin asn [command]
```

**Global Options:**
- `-o, --output string` - Output format: table, json, yaml (default: table)
- `-t, --timeout int` - Timeout in seconds (default: 30)

---

### asn info

Get ASN information.

```bash
bgpin asn info [asn]
```

**Arguments:**
- `asn` - The Autonomous System Number (with or without AS prefix)

**Examples:**
```bash
bgpin asn info 262978
bgpin asn info AS262978 -o json
bgpin asn info 262978 --output yaml
```

**Output Fields:**
- Holder
- Announced
- Block
- Description (if available)
- Country (if available)

---

### asn neighbors

Get BGP neighbors of an Autonomous System.

```bash
bgpin asn neighbors [asn]
```

**Arguments:**
- `asn` - The Autonomous System Number

**Examples:**
```bash
bgpin asn neighbors 262978
bgpin asn neighbors 262978 -o json
```

**Output Fields:**
- ASN
- Type
- Power

---

### asn prefixes

Get all prefixes announced by an Autonomous System.

```bash
bgpin asn prefixes [asn]
```

**Arguments:**
- `asn` - The Autonomous System Number

**Examples:**
```bash
bgpin asn prefixes 262978
bgpin asn prefixes 262978 -o json
```

**Output Fields:**
- Prefix Number
- Prefix
- Type (IPv4/IPv6)

---

### asn peers

Get RIPE RIS peers for an Autonomous System.

```bash
bgpin asn peers [asn]
```

**Arguments:**
- `asn` - The Autonomous System Number

**Examples:**
```bash
bgpin asn peers 262978
bgpin asn peers 262978 -o json
```

---

## Prefix Query Commands

Query IP prefix information from RIPE RIS.

```bash
bgpin prefix [command]
```

**Global Options:**
- `-o, --output string` - Output format: table, json, yaml (default: table)
- `-t, --timeout int` - Timeout in seconds (default: 30)

---

### prefix overview

Get detailed overview of an IP prefix.

```bash
bgpin prefix overview [prefix]
```

**Arguments:**
- `prefix` - The IP prefix (e.g., 200.160.0.0/20)

**Examples:**
```bash
bgpin prefix overview 200.160.0.0/20
bgpin prefix overview 2804:4d44::/32 -o json
bgpin prefix overview 186.250.184.0/24 --output yaml
```

**Output Fields:**
- Prefix
- Actual Prefix (if different)
- Is Less Specific
- Announcing ASNs
- Query Time

---

## Flow Analysis Commands

Analyze network flow data (NetFlow/sFlow/IPFIX) and correlate with BGP.

```bash
bgpin flow [command]
```

**Global Options:**
- `-o, --output string` - Output format: table, json, yaml (default: table)
- `-l, --limit int` - Limit number of results (default: 10)

**Note:** Flow collection requires proper configuration in bgpin.yaml with flow.enabled=true.

---

### flow top

Show top prefixes by traffic volume.

```bash
bgpin flow top
```

**Options:**
- `-l, --limit int` - Limit number of results (default: 10)

**Examples:**
```bash
bgpin flow top
bgpin flow top --limit 20
```

**Output Fields:**
- Rank
- Prefix
- ASN
- Traffic
- PPS
- Top Protocol

---

### flow asn

Show traffic statistics for a specific ASN.

```bash
bgpin flow asn [asn]
```

**Arguments:**
- `asn` - The Autonomous System Number

**Examples:**
```bash
bgpin flow asn 15169
bgpin flow asn AS262978
```

**Output Fields:**
- Metric (Traffic, Packets, Flows)
- Inbound
- Outbound

---

### flow anomaly

Detect and display traffic anomalies.

```bash
bgpin flow anomaly
```

**Detected Anomaly Types:**
- DDoS
- Spike
- Drop

**Output Fields:**
- Time
- Type
- Severity
- Prefix
- ASN
- Description

---

### flow upstream-compare

Compare traffic patterns across multiple upstream providers.

```bash
bgpin flow upstream-compare
```

**Output Fields:**
- Provider
- ASN
- AS Path
- Traffic
- PPS
- Latency
- Loss

---

### flow stats

Show flow collector statistics.

```bash
bgpin flow stats
```

**Output Fields:**
- NetFlow Packets
- sFlow Packets
- IPFIX Packets
- Total Flows
- Dropped Flows
- Processing Errors
- Last Update

---

## Route Analysis Commands

Analyze BGP routes for anomalies and security issues.

```bash
bgpin analyze [command]
```

---

### analyze route

Analyze routes for a prefix.

```bash
bgpin analyze route [prefix]
```

**Arguments:**
- `prefix` - The IP prefix to analyze

---

### analyze asn

Analyze routes from an AS.

```bash
bgpin analyze asn [asn]
```

**Arguments:**
- `asn` - The Autonomous System Number

---

## Looking Glass Commands

### lg

List all configured looking glasses and their status.

```bash
bgpin lg
```

**Output Fields:**
- Name
- Vendor
- Type
- Protocol
- Country
- URL

---

## Global Flags

The following flags are available for all commands:

| Flag | Shortcut | Description | Default |
|------|----------|-------------|---------|
| `--config` | - | Config file path | ./bgpin.yaml |
| `--verbose` | `-v` | Verbose output | false |
| `--help` | `-h` | Help for command | - |
| `--version` | - | Version for CLI | - |

---

## Output Formats

All commands support multiple output formats:

- `table` - Human-readable table format (default)
- `json` - JSON format
- `yaml` - YAML format

Specify using: `-o, --output [format]`

---

## Configuration

The CLI uses a configuration file (default: `./bgpin.yaml`). Key configuration options:

- `timeout` - Request timeout in seconds (default: 30)
- `output` - Default output format (default: table)
- `cache.enabled` - Enable caching (default: true)
- `cache.ttl` - Cache TTL in seconds (default: 300)
- `flow.enabled` - Enable flow collection (default: false)

---

## Examples

### Query ASN Information

```bash
bgpin asn info 262978
bgpin asn neighbors 262978 -o json
bgpin asn prefixes 262978
```

### Query Prefix Information

```bash
bgpin prefix overview 8.8.8.0/24
bgpin prefix overview 2001:db8::/32 -o yaml
```

### Analyze Network Traffic

```bash
bgpin flow top
bgpin flow asn 15169
bgpin flow anomaly
```

### Lookup BGP Routes

```bash
bgpin lookup 8.8.8.0/24
bgpin lookup 8.8.8.0/24 -o json
```

### List Available Resources

```bash
bgpin lg
bgpin version
```
