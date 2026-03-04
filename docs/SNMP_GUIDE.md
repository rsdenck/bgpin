# Guia SNMP - bgpin

Este guia mostra como usar os comandos SNMP do bgpin para monitoramento de dispositivos de rede.

## Visão Geral

O bgpin inclui comandos SNMP completos para monitoramento de:
- **Interfaces de rede**: Tráfego, status, velocidade
- **Sistema**: Uptime, descrição, localização
- **BGP**: Peers, estatísticas de sessão
- **Exploração**: Walk e Get para OIDs customizadas

## Comandos Disponíveis

### 1. Interfaces de Rede

```bash
# Listar todas as interfaces
bgpin snmp interfaces 192.168.1.1

# Com community customizada
bgpin snmp interfaces router.local --community private

# Com timeout maior
bgpin snmp interfaces 10.0.0.1 --timeout 10
```

**Saída:**
```
╭─────────────┬──────────────────────┬────────┬───────────┬────────────┬─────────────╮
│ Interface   │ Descrição            │ Status │ Velocidade│ In (bytes) │ Out (bytes) │
├─────────────┼──────────────────────┼────────┼───────────┼────────────┼─────────────┤
│ 1           │ GigabitEthernet0/0/1 │ up     │ 1.0 Gbps  │ 2.5 GB     │ 1.8 GB      │
│ 2           │ GigabitEthernet0/0/2 │ down   │ 1.0 Gbps  │ 0 B        │ 0 B         │
╰─────────────┴──────────────────────┴────────┴───────────┴────────────┴─────────────╯
```

### 2. Informações do Sistema

```bash
# Informações básicas do sistema
bgpin snmp system 192.168.1.1

# Com community específica
bgpin snmp system switch.local --community monitoring
```

**Saída:**
```
╭─────────────┬─────────────────────────────────────────────╮
│ Campo       │ Valor                                       │
├─────────────┼─────────────────────────────────────────────┤
│ Nome        │ router-core.example.com                     │
│ Descrição   │ Cisco IOS Software, Version 15.1           │
│ Uptime      │ 45 dias, 12 horas, 30 minutos             │
│ Contato     │ admin@example.com                          │
│ Localização │ Datacenter Principal - Rack 42             │
╰─────────────┴─────────────────────────────────────────────╯
```

### 3. Estatísticas BGP

```bash
# Monitorar peers BGP
bgpin snmp bgp 192.168.1.1

# Com versão SNMP específica
bgpin snmp bgp router-bgp.net --version 2c
```

**Saída:**
```
╭─────────────────┬───────────┬─────────────┬─────────────┬──────────────╮
│ Peer IP         │ Remote AS │ Estado      │ In Updates  │ Out Updates  │
├─────────────────┼───────────┼─────────────┼─────────────┼──────────────┤
│ 10.0.0.2        │ 65001     │ Established │ 1250        │ 15           │
│ 10.0.0.3        │ 65002     │ Established │ 980         │ 15           │
│ 192.168.100.1   │ 65003     │ Idle        │ 0           │ 0            │
╰─────────────────┴───────────┴─────────────┴─────────────┴──────────────╯
```

### 4. SNMP Walk

```bash
# Walk na árvore de sistema
bgpin snmp walk 192.168.1.1 1.3.6.1.2.1.1

# Walk em interfaces
bgpin snmp walk router.local 1.3.6.1.2.1.2.2.1.2
```

**Saída:**
```
╭─────────────────────────────────┬─────────────┬──────────────────────────╮
│ OID                             │ Tipo        │ Valor                    │
├─────────────────────────────────┼─────────────┼──────────────────────────┤
│ 1.3.6.1.2.1.1.1.0              │ OctetString │ Cisco IOS Software       │
│ 1.3.6.1.2.1.1.2.0              │ ObjectId    │ 1.3.6.1.4.1.9.1.1       │
│ 1.3.6.1.2.1.1.3.0              │ TimeTicks   │ 45 dias, 12 horas       │
╰─────────────────────────────────┴─────────────┴──────────────────────────╯
```

### 5. SNMP Get

```bash
# Get de OIDs específicas
bgpin snmp get 192.168.1.1 1.3.6.1.2.1.1.1.0

# Múltiplas OIDs
bgpin snmp get router.local 1.3.6.1.2.1.1.3.0 1.3.6.1.2.1.1.5.0
```

## Configuração

### 1. Arquivo bgpin.yaml

```yaml
snmp:
  default_community: "public"
  default_version: "2c"
  default_port: 161
  default_timeout: 5
  default_retries: 3
  
  # Dispositivos pré-configurados
  devices:
    router-core:
      host: "192.168.1.1"
      community: "bgp-monitoring"
      version: "2c"
      description: "Core Router"
    
    switch-access:
      host: "192.168.1.10"
      community: "network-ops"
      version: "2c"
      description: "Access Switch"
```

### 2. Flags Disponíveis

| Flag | Descrição | Padrão |
|------|-----------|--------|
| `--community` | Community string SNMP | `public` |
| `--version` | Versão SNMP (1, 2c, 3) | `2c` |
| `--port` | Porta SNMP | `161` |
| `--timeout` | Timeout em segundos | `5` |
| `--retries` | Número de tentativas | `3` |
| `--output` | Formato de saída | `table` |

## OIDs Importantes

### Sistema (1.3.6.1.2.1.1)
- `1.3.6.1.2.1.1.1.0` - sysDescr (Descrição do sistema)
- `1.3.6.1.2.1.1.3.0` - sysUpTime (Uptime)
- `1.3.6.1.2.1.1.4.0` - sysContact (Contato)
- `1.3.6.1.2.1.1.5.0` - sysName (Nome)
- `1.3.6.1.2.1.1.6.0` - sysLocation (Localização)

### Interfaces (1.3.6.1.2.1.2.2.1)
- `1.3.6.1.2.1.2.2.1.2` - ifDescr (Descrição da interface)
- `1.3.6.1.2.1.2.2.1.8` - ifOperStatus (Status operacional)
- `1.3.6.1.2.1.2.2.1.5` - ifSpeed (Velocidade)
- `1.3.6.1.2.1.2.2.1.10` - ifInOctets (Bytes recebidos)
- `1.3.6.1.2.1.2.2.1.16` - ifOutOctets (Bytes enviados)

### BGP (1.3.6.1.2.1.15.3.1)
- `1.3.6.1.2.1.15.3.1.2` - bgpPeerState (Estado do peer)
- `1.3.6.1.2.1.15.3.1.9` - bgpPeerRemoteAs (AS remoto)
- `1.3.6.1.2.1.15.3.1.10` - bgpPeerInUpdates (Updates recebidos)
- `1.3.6.1.2.1.15.3.1.11` - bgpPeerOutUpdates (Updates enviados)

## Casos de Uso

### 1. Monitoramento de Interfaces

```bash
# Verificar status de todas as interfaces
bgpin snmp interfaces 192.168.1.1

# Monitorar tráfego específico
bgpin snmp walk 192.168.1.1 1.3.6.1.2.1.2.2.1.10  # InOctets
```

### 2. Health Check de Sistema

```bash
# Verificar uptime e informações básicas
bgpin snmp system 192.168.1.1

# Verificar descrição do sistema
bgpin snmp get 192.168.1.1 1.3.6.1.2.1.1.1.0
```

### 3. Monitoramento BGP

```bash
# Status de todos os peers BGP
bgpin snmp bgp 192.168.1.1

# Verificar estado específico de um peer
bgpin snmp get 192.168.1.1 1.3.6.1.2.1.15.3.1.2.10.0.0.2
```

### 4. Troubleshooting

```bash
# Explorar MIB completa
bgpin snmp walk 192.168.1.1 1.3.6.1.2.1

# Verificar OID específica
bgpin snmp get 192.168.1.1 1.3.6.1.2.1.1.3.0
```

## Integração com Outros Comandos

### Combinando SNMP + BGP

```bash
# 1. Verificar peers BGP via SNMP
bgpin snmp bgp 192.168.1.1

# 2. Analisar rotas BGP via RIPE
bgpin asn neighbors 262978

# 3. Correlacionar com dados de flow
bgpin flow top
```

### Monitoramento Completo

```bash
# Script de monitoramento
#!/bin/bash

echo "=== Sistema ==="
bgpin snmp system 192.168.1.1

echo "=== Interfaces ==="
bgpin snmp interfaces 192.168.1.1

echo "=== BGP Peers ==="
bgpin snmp bgp 192.168.1.1

echo "=== Análise BGP ==="
bgpin asn info 262978
```

## Troubleshooting

### Erro de Conexão
```bash
# Verificar conectividade
ping 192.168.1.1

# Testar porta SNMP
telnet 192.168.1.1 161
```

### Community Incorreta
```bash
# Testar diferentes communities
bgpin snmp system 192.168.1.1 --community public
bgpin snmp system 192.168.1.1 --community private
```

### Timeout
```bash
# Aumentar timeout
bgpin snmp system 192.168.1.1 --timeout 10 --retries 5
```

## Próximos Passos

1. **Configurar dispositivos** no `bgpin.yaml`
2. **Automatizar monitoramento** com scripts
3. **Integrar com TUI** para visualização em tempo real
4. **Combinar com Flow data** para análise completa

O SNMP no bgpin complementa perfeitamente os dados BGP e Flow, fornecendo uma visão completa da infraestrutura de rede!