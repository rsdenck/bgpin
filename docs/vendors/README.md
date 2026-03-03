# Vendor Parsers

This directory contains documentation about vendor-specific BGP parsers.

## Documentation

- [STATUS.md](STATUS.md) - Complete implementation status of all vendors

## Implemented Vendors

### âœ… Cisco Systems
- **Status:** 100% Complete
- **Supported:** IOS, IOS-XE, IOS-XR, NX-OS
- **Protocol:** SSH CLI
- **File:** `internal/parsers/cisco/cisco.go`

### âœ… Juniper Networks
- **Status:** 100% Complete
- **Supported:** JunOS
- **Protocol:** NETCONF/XML RPC
- **File:** `internal/parsers/junos/junos.go`

## Planned Vendors

- Arista Networks (EOS)
- Nokia (SR OS)
- MikroTik (RouterOS)
- Huawei (VRP)
- GoBGP (Native BGP)
- Cloud providers (AWS, GCP, Azure)

See [STATUS.md](STATUS.md) for complete details and implementation priorities.
