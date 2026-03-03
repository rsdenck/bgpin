# bgpin CLI - Guia de Uso Completo

## Instalação

```bash
# Clone o repositório
git clone https://github.com/bgpin/bgpin
cd bgpin

# Compile
go build -o bgpin ./cmd/cli/

# Ou instale globalmente
go install ./cmd/cli/
```

## Comandos Disponíveis

### 1. Informações de ASN

#### Obter informações gerais de um ASN
```bash
bgpin asn info 262978
bgpin asn info AS262978  # Também aceita com prefixo AS
```

Saída:
```
╔═══════════════════════════════════════════════════════════════╗
║ ASN Information: AS262978
╠═══════════════════════════════════════════════════════════════╣
║ Holder:      Centro de Tecnologia Armazem Datacenter Ltda.
║ Announced:   true
║ Block:       262144-263167
╚═══════════════════════════════════════════════════════════════╝
```

#### Obter vizinhos BGP
```bash
bgpin asn neighbors 262978
```

Mostra todos os vizinhos BGP (upstream, downstream, peers) com tipo e poder da relação.

#### Obter prefixos anunciados
```bash
bgpin asn prefixes 262978
```

Lista todos os prefixos IPv4 e IPv6 anunciados pelo ASN.

#### Obter RIS peers
```bash
bgpin asn peers 262978
```

Lista todos os peers RIPE RIS que veem o ASN, incluindo RRC, ASN do peer, IP e contadores de prefixos.

### 2. Informações de Prefixo

#### Obter visão geral de um prefixo
```bash
bgpin prefix overview 186.250.184.0/24
bgpin prefix overview 2804:4d44::/32  # IPv6 também suportado
```

Mostra quais ASNs estão anunciando o prefixo e se é less-specific.

### 3. Formatos de Saída

Todos os comandos suportam múltiplos formatos de saída:

#### Formato Tabela (padrão)
```bash
bgpin asn info 262978
bgpin asn info 262978 -o table
```

#### Formato JSON
```bash
bgpin asn info 262978 -o json
```

Saída:
```json
{
  "asn": 262978,
  "holder": "Centro de Tecnologia Armazem Datacenter Ltda.",
  "announced": true,
  "block": "262144-263167",
  "description": "",
  "country": ""
}
```

#### Formato YAML
```bash
bgpin asn info 262978 -o yaml
```

Saída:
```yaml
asn: 262978
holder: Centro de Tecnologia Armazem Datacenter Ltda.
announced: true
block: 262144-263167
description: ""
country: ""
```

### 4. Configuração de Timeout

Todos os comandos aceitam timeout customizado:

```bash
bgpin asn info 262978 --timeout 60
bgpin asn info 262978 -t 60
```

### 5. Looking Glasses

#### Listar Looking Glasses disponíveis
```bash
bgpin lg
```

Mostra todos os Looking Glasses configurados.

### 6. Versão

```bash
bgpin version
```

Mostra informações de versão, build e runtime.

## Exemplos Práticos

### Investigar um ASN completo
```bash
# Informações básicas
bgpin asn info 262978

# Ver vizinhos
bgpin asn neighbors 262978

# Ver prefixos anunciados
bgpin asn prefixes 262978

# Ver peers RIS
bgpin asn peers 262978
```

### Exportar dados para análise
```bash
# Exportar para JSON
bgpin asn info 262978 -o json > asn_262978.json

# Exportar prefixos para YAML
bgpin asn prefixes 262978 -o yaml > prefixes_262978.yaml
```

### Verificar um prefixo específico
```bash
# Ver quem está anunciando
bgpin prefix overview 186.250.184.0/24

# Exportar para JSON
bgpin prefix overview 186.250.184.0/24 -o json
```

### Pipeline com jq (JSON)
```bash
# Extrair apenas o holder do ASN
bgpin asn info 262978 -o json | jq -r '.holder'

# Contar quantos prefixes um ASN anuncia
bgpin asn prefixes 262978 -o json | jq '.prefixes | length'

# Listar apenas prefixes IPv4
bgpin asn prefixes 262978 -o json | jq -r '.prefixes[].prefix' | grep -v ':'
```

## Arquivo de Configuração

Crie um arquivo `bgpin.yaml` no diretório atual:

```yaml
# Configurações gerais
timeout: 30
output: table

# Cache
cache:
  enabled: true
  ttl: 300

# Looking Glasses
looking_glasses:
  - name: "RIPE RIS"
    url: "https://stat.ripe.net"
    vendor: "ripe"
    type: "http"
    enabled: true

# RIPE RIS SDK
ripe:
  rate_limit: 10
  retry_max: 3
  retry_wait_min: 1
  retry_wait_max: 10
```

Use configuração customizada:
```bash
bgpin --config /path/to/config.yaml asn info 262978
```

## Modo Verbose

Para debug e troubleshooting:

```bash
bgpin -v asn info 262978
bgpin --verbose asn neighbors 262978
```

## Automação e CI/CD

### Script Bash
```bash
#!/bin/bash
ASN=262978

echo "Checking ASN $ASN..."
bgpin asn info $ASN -o json > asn_info.json

if [ $? -eq 0 ]; then
    echo "ASN check successful"
    bgpin asn prefixes $ASN -o json > prefixes.json
else
    echo "ASN check failed"
    exit 1
fi
```

### GitHub Actions
```yaml
name: BGP Check
on: [push]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Install bgpin
        run: go install ./cmd/cli/
      - name: Check ASN
        run: bgpin asn info 262978 -o json
```

## Dicas e Truques

### 1. Alias úteis
```bash
alias bgp-info='bgpin asn info'
alias bgp-prefixes='bgpin asn prefixes'
alias bgp-neighbors='bgpin asn neighbors'
```

### 2. Função para verificar múltiplos ASNs
```bash
check_asns() {
    for asn in "$@"; do
        echo "Checking AS$asn..."
        bgpin asn info $asn -o json
    done
}

check_asns 262978 13335 15169
```

### 3. Monitoramento de prefixos
```bash
# Verificar se um prefixo mudou de ASN
watch -n 300 'bgpin prefix overview 186.250.184.0/24'
```

## Troubleshooting

### Timeout errors
```bash
# Aumentar timeout
bgpin asn info 262978 --timeout 60
```

### Rate limiting
O SDK já implementa rate limiting automático (10 req/s padrão).

### Erros de rede
```bash
# Usar modo verbose para debug
bgpin -v asn info 262978
```

## Próximos Passos

- [ ] Análise de anomalias BGP
- [ ] Validação RPKI
- [ ] Suporte a múltiplos Looking Glasses
- [ ] Modo interativo (TUI)
- [ ] Cache local
- [ ] Histórico de mudanças

## Suporte

Para reportar bugs ou sugerir features:
- GitHub Issues: https://github.com/bgpin/bgpin/issues
- Email: support@bgpin.dev
