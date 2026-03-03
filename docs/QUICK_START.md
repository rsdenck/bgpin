# bgpin - Quick Start Guide

## 🚀 Instalação Rápida

```bash
# Clone o repositório
git clone https://github.com/bgpin/bgpin
cd bgpin

# Compile
go build -o bgpin ./cmd/cli/

# Teste
./bgpin version
```

## 📝 Comandos Essenciais

### 1. Informações de ASN
```bash
# Informações básicas
bgpin asn info 262978

# Formato JSON
bgpin asn info 262978 -o json

# Formato YAML
bgpin asn info 262978 -o yaml
```

### 2. Vizinhos BGP
```bash
# Ver todos os vizinhos
bgpin asn neighbors 262978

# Exportar para JSON
bgpin asn neighbors 262978 -o json > neighbors.json
```

### 3. Prefixos Anunciados
```bash
# Listar prefixos
bgpin asn prefixes 262978

# Salvar em arquivo
bgpin asn prefixes 262978 -o yaml > prefixes.yaml
```

### 4. RIS Peers
```bash
# Ver peers RIPE RIS
bgpin asn peers 262978

# Formato JSON
bgpin asn peers 262978 -o json
```

### 5. Análise de Prefixo
```bash
# Ver quem anuncia um prefixo
bgpin prefix overview 186.250.184.0/24

# IPv6 também funciona
bgpin prefix overview 2804:4d44::/32
```

## 🎯 Exemplos Práticos

### Investigar um ASN Completo
```bash
#!/bin/bash
ASN=262978

echo "=== ASN Information ==="
bgpin asn info $ASN

echo -e "\n=== BGP Neighbors ==="
bgpin asn neighbors $ASN

echo -e "\n=== Announced Prefixes ==="
bgpin asn prefixes $ASN

echo -e "\n=== RIS Peers ==="
bgpin asn peers $ASN
```

### Exportar Tudo para JSON
```bash
#!/bin/bash
ASN=262978
DIR="reports/AS${ASN}"

mkdir -p $DIR

bgpin asn info $ASN -o json > $DIR/info.json
bgpin asn neighbors $ASN -o json > $DIR/neighbors.json
bgpin asn prefixes $ASN -o json > $DIR/prefixes.json
bgpin asn peers $ASN -o json > $DIR/peers.json

echo "Reports saved to $DIR/"
```

### Pipeline com jq
```bash
# Extrair apenas o holder
bgpin asn info 262978 -o json | jq -r '.holder'

# Contar prefixos
bgpin asn prefixes 262978 -o json | jq '.prefixes | length'

# Listar apenas prefixos IPv4
bgpin asn prefixes 262978 -o json | jq -r '.prefixes[].prefix' | grep -v ':'

# Listar apenas prefixos IPv6
bgpin asn prefixes 262978 -o json | jq -r '.prefixes[].prefix' | grep ':'

# Contar vizinhos por tipo
bgpin asn neighbors 262978 -o json | jq '.neighbors | group_by(.type) | map({type: .[0].type, count: length})'
```

## 🧪 Testar o SDK

```bash
# Executar todos os testes
go test -v ./sdk/integration_test/

# Teste específico
go test -v ./sdk/integration_test/ -run TestGetASNInfo_262978

# Executar exemplo
go run sdk/examples/demo.go
```

## ⚙️ Configuração

### Criar arquivo de configuração
```bash
cp bgpin.yaml.example bgpin.yaml
```

### Editar configuração
```yaml
# bgpin.yaml
timeout: 30
output: table

cache:
  enabled: true
  ttl: 300

ripe:
  rate_limit: 10
  retry_max: 3
```

### Usar configuração customizada
```bash
bgpin --config /path/to/config.yaml asn info 262978
```

## 🔍 Troubleshooting

### Timeout
```bash
# Aumentar timeout para 60 segundos
bgpin asn info 262978 --timeout 60
```

### Modo Verbose
```bash
# Ver detalhes de execução
bgpin -v asn info 262978
```

### Verificar versão
```bash
bgpin version
```

## 📚 Mais Informações

- [Guia Completo da CLI](docs/CLI_GUIDE.md)
- [Arquitetura](docs/ARCHITECTURE.md)
- [SDK README](sdk/README.md)
- [README Principal](README.md)

## 🎓 Exemplos de Uso Real

### 1. Monitorar mudanças em prefixos
```bash
# Verificar a cada 5 minutos
watch -n 300 'bgpin prefix overview 186.250.184.0/24'
```

### 2. Comparar ASNs
```bash
#!/bin/bash
for asn in 262978 13335 15169; do
    echo "=== AS$asn ==="
    bgpin asn info $asn
    echo ""
done
```

### 3. Alertar se ASN não está anunciando
```bash
#!/bin/bash
ASN=262978

ANNOUNCED=$(bgpin asn info $ASN -o json | jq -r '.announced')

if [ "$ANNOUNCED" != "true" ]; then
    echo "ALERT: AS$ASN is not announcing!"
    # Enviar notificação
fi
```

### 4. Gerar relatório diário
```bash
#!/bin/bash
DATE=$(date +%Y-%m-%d)
ASN=262978

bgpin asn info $ASN -o json > "report_${DATE}_info.json"
bgpin asn prefixes $ASN -o json > "report_${DATE}_prefixes.json"

echo "Daily report generated for $DATE"
```

## 🚀 Deploy

### Docker (futuro)
```dockerfile
FROM golang:1.25-alpine
WORKDIR /app
COPY . .
RUN go build -o bgpin ./cmd/cli/
ENTRYPOINT ["./bgpin"]
```

### Systemd Service (futuro)
```ini
[Unit]
Description=bgpin monitoring service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/bgpin asn info 262978
Restart=always

[Install]
WantedBy=multi-user.target
```

## 💡 Dicas

1. **Use aliases**: `alias bgp='bgpin asn'`
2. **Exporte para JSON**: Facilita parsing com jq
3. **Use timeout**: Para operações longas
4. **Modo verbose**: Para debug
5. **Cache**: Habilite para melhor performance

## 🎯 Casos de Uso

- ✅ Investigação de ASN
- ✅ Monitoramento de prefixos
- ✅ Análise de vizinhos BGP
- ✅ Auditoria de anúncios
- ✅ Troubleshooting de rotas
- ✅ Automação de relatórios
- ✅ CI/CD pipelines
- ✅ SOC/Blue Team operations

---

**Pronto para começar? Execute: `bgpin asn info 262978`** 🚀
