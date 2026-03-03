# bgpin CLI - Guia de Uso Completo

## InstalaÃ§Ã£o

```bash
# Clone o repositÃ³rio
git clone https://github.com/bgpin/bgpin
cd bgpin

# Compile
go build -o bgpin ./cmd/cli/

# Ou instale globalmente
go install ./cmd/cli/
```

## Comandos DisponÃ­veis

### 1. InformaÃ§Ãµes de ASN

#### Obter informaÃ§Ãµes gerais de um ASN
```bash
bgpin asn info 262978
bgpin asn info AS262978  # TambÃ©m aceita com prefixo AS
```

SaÃ­da:
```
â•”â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•—
â•‘ ASN Information: AS262978
â• â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•£
â•‘ Holder:      Centro de Tecnologia Armazem Datacenter Ltda.
â•‘ Announced:   true
â•‘ Block:       262144-263167
â•šâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
```

#### Obter vizinhos BGP
```bash
bgpin asn neighbors 262978
```

Mostra todos os vizinhos BGP (upstream, downstream, peers) com tipo e poder da relaÃ§Ã£o.

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

### 2. InformaÃ§Ãµes de Prefixo

#### Obter visÃ£o geral de um prefixo
```bash
bgpin prefix overview 186.250.184.0/24
bgpin prefix overview 2804:4d44::/32  # IPv6 tambÃ©m suportado
```

Mostra quais ASNs estÃ£o anunciando o prefixo e se Ã© less-specific.

### 3. Formatos de SaÃ­da

Todos os comandos suportam mÃºltiplos formatos de saÃ­da:

#### Formato Tabela (padrÃ£o)
```bash
bgpin asn info 262978
bgpin asn info 262978 -o table
```

#### Formato JSON
```bash
bgpin asn info 262978 -o json
```

SaÃ­da:
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

SaÃ­da:
```yaml
asn: 262978
holder: Centro de Tecnologia Armazem Datacenter Ltda.
announced: true
block: 262144-263167
description: ""
country: ""
```

### 4. ConfiguraÃ§Ã£o de Timeout

Todos os comandos aceitam timeout customizado:

```bash
bgpin asn info 262978 --timeout 60
bgpin asn info 262978 -t 60
```

### 5. Looking Glasses

#### Listar Looking Glasses disponÃ­veis
```bash
bgpin lg
```

Mostra todos os Looking Glasses configurados.

### 6. VersÃ£o

```bash
bgpin version
```

Mostra informaÃ§Ãµes de versÃ£o, build e runtime.

## Exemplos PrÃ¡ticos

### Investigar um ASN completo
```bash
# InformaÃ§Ãµes bÃ¡sicas
bgpin asn info 262978

# Ver vizinhos
bgpin asn neighbors 262978

# Ver prefixos anunciados
bgpin asn prefixes 262978

# Ver peers RIS
bgpin asn peers 262978
```

### Exportar dados para anÃ¡lise
```bash
# Exportar para JSON
bgpin asn info 262978 -o json > asn_262978.json

# Exportar prefixos para YAML
bgpin asn prefixes 262978 -o yaml > prefixes_262978.yaml
```

### Verificar um prefixo especÃ­fico
```bash
# Ver quem estÃ¡ anunciando
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

## Arquivo de ConfiguraÃ§Ã£o

Crie um arquivo `bgpin.yaml` no diretÃ³rio atual:

```yaml
# ConfiguraÃ§Ãµes gerais
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

Use configuraÃ§Ã£o customizada:
```bash
bgpin --config /path/to/config.yaml asn info 262978
```

## Modo Verbose

Para debug e troubleshooting:

```bash
bgpin -v asn info 262978
bgpin --verbose asn neighbors 262978
```

## AutomaÃ§Ã£o e CI/CD

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

### 1. Alias Ãºteis
```bash
alias bgp-info='bgpin asn info'
alias bgp-prefixes='bgpin asn prefixes'
alias bgp-neighbors='bgpin asn neighbors'
```

### 2. FunÃ§Ã£o para verificar mÃºltiplos ASNs
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
O SDK jÃ¡ implementa rate limiting automÃ¡tico (10 req/s padrÃ£o).

### Erros de rede
```bash
# Usar modo verbose para debug
bgpin -v asn info 262978
```

## PrÃ³ximos Passos

- [ ] AnÃ¡lise de anomalias BGP
- [ ] ValidaÃ§Ã£o RPKI
- [ ] Suporte a mÃºltiplos Looking Glasses
- [ ] Modo interativo (TUI)
- [ ] Cache local
- [ ] HistÃ³rico de mudanÃ§as

## Suporte

Para reportar bugs ou sugerir features:
- GitHub Issues: https://github.com/bgpin/bgpin/issues
- Email: support@bgpin.dev
