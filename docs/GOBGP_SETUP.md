# Configuração GoBGP para bgpin TUI

Este guia mostra como configurar o GoBGP para funcionar com a TUI do bgpin usando dados BGP reais.

## Pré-requisitos

1. **Instalar GoBGP**:
```bash
go install github.com/osrg/gobgp/v3/cmd/gobgpd@latest
go install github.com/osrg/gobgp/v3/cmd/gobgp@latest
```

2. **Verificar instalação**:
```bash
gobgpd --version
gobgp --version
```

## Configuração Básica

### 1. Criar arquivo de configuração

Copie o arquivo exemplo:
```bash
cp gobgp.conf.example gobgp.conf
```

### 2. Configuração mínima para ASN 262978

```toml
[global.config]
as = 262978
router-id = "192.168.255.1"

# Habilita gRPC API para TUI
[global.api]
listen-addresses = ["127.0.0.1:50051"]

# Neighbor BGP real
[[neighbors]]
[neighbors.config]
neighbor-address = "10.0.255.1"
peer-as = 65001
description = "Upstream Provider"
```

### 3. Iniciar GoBGP daemon

```bash
gobgpd -f gobgp.conf
```

### 4. Verificar status

```bash
# Verificar neighbors
gobgp neighbor

# Verificar rotas
gobgp global rib

# Verificar API gRPC
netstat -an | grep 50051
```

## Configuração Avançada

### Múltiplos Neighbors

```toml
[[neighbors]]
[neighbors.config]
neighbor-address = "10.0.255.1"
peer-as = 65001
description = "Upstream Provider 1"

[[neighbors]]
[neighbors.config]
neighbor-address = "10.0.255.2"
peer-as = 65002
description = "Upstream Provider 2"

[[neighbors]]
[neighbors.config]
neighbor-address = "10.0.255.3"
peer-as = 262978
description = "Peer Exchange"
```

### Anúncio de Prefixos

```toml
# Definir prefixos para anunciar
[[defined-sets.prefix-sets]]
prefix-set-name = "my-prefixes"
[[defined-sets.prefix-sets.prefix-list]]
ip-prefix = "203.0.113.0/24"
masklength-range = "24..24"

# Política de exportação
[[policy-definitions]]
name = "export-policy"
[[policy-definitions.statements]]
name = "export-my-prefixes"
[policy-definitions.statements.conditions.match-prefix-set]
prefix-set = "my-prefixes"
[policy-definitions.statements.actions]
route-disposition = "accept-route"
```

## Integração com bgpin TUI

### 1. Iniciar TUI

```bash
# TUI com ASN 262978 (padrão)
bgpin tui

# TUI com ASN específico
bgpin tui --asn 262978

# TUI com refresh de 1 segundo
bgpin tui --refresh 1s
```

### 2. Painéis da TUI

- **Graph Panel (1)**: Mostra topologia AS-PATH em tempo real
- **Peers Panel (2)**: Lista neighbors BGP com status
- **Routes Panel (3)**: Tabela de rotas BGP recebidas
- **Flows Panel (4)**: Dados NetFlow/sFlow (opcional)
- **Summary Panel (5)**: Resumo geral do sistema

### 3. Navegação

- `Tab/Shift+Tab`: Alternar entre painéis
- `1-5`: Pular para painel específico
- `r`: Refresh manual
- `q/Ctrl+C`: Sair
- `h`: Ajuda

## Troubleshooting

### TUI mostra "Aguardando conexão GoBGP..."

1. Verificar se GoBGP está rodando:
```bash
ps aux | grep gobgpd
```

2. Verificar se API gRPC está ativa:
```bash
netstat -an | grep 50051
```

3. Testar conexão gRPC:
```bash
gobgp neighbor
```

### Neighbors não aparecem

1. Verificar configuração de rede
2. Verificar se o peer remoto está configurado
3. Verificar logs do GoBGP:
```bash
gobgpd -f gobgp.conf -l debug
```

### Sem rotas no painel Routes

1. Verificar se neighbors estão Established:
```bash
gobgp neighbor
```

2. Verificar RIB:
```bash
gobgp global rib
```

## Dados de Flow (Opcional)

Para dados NetFlow/sFlow no painel Flows:

1. Configure NetFlow/sFlow no router
2. Aponte para bgpin flow collector
3. Configure `bgpin.yaml` com collector settings

O BGP funciona independente dos dados de flow.

## Exemplo Completo

```bash
# 1. Instalar GoBGP
go install github.com/osrg/gobgp/v3/cmd/gobgpd@latest

# 2. Criar configuração
cat > gobgp.conf << EOF
[global.config]
as = 262978
router-id = "192.168.255.1"

[global.api]
listen-addresses = ["127.0.0.1:50051"]

[[neighbors]]
[neighbors.config]
neighbor-address = "10.0.255.1"
peer-as = 65001
description = "Upstream Provider"
EOF

# 3. Iniciar GoBGP
gobgpd -f gobgp.conf &

# 4. Iniciar TUI
bgpin tui
```

A TUI mostrará dados BGP reais assim que os neighbors estiverem estabelecidos.