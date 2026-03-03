# bgpin - Exemplos de Output

## Formato Table (Padrão)

Todos os outputs usam a biblioteca `go-pretty/table` com estilo `StyleRounded` para uma aparência profissional.

### ASN Information

```bash
bgpin asn info 262978
```

```
╭───────────────────────────────────────────────────────────╮
│ ASN Information: AS262978                                 │
├───────────┬───────────────────────────────────────────────┤
│ Holder    │ Centro de Tecnologia Armazem Datacenter Ltda. │
│ Announced │ true                                          │
│ Block     │ 262144-263167                                 │
╰───────────┴───────────────────────────────────────────────╯
```

### BGP Neighbors

```bash
bgpin asn neighbors 262978
```

```
╭────────────────────────────────────────────────╮
│ BGP Neighbors for AS262978 (Total: 34)        │
├─────────┬────────────┬───────┤
│ ASN     │ TYPE       │ POWER │
├─────────┼────────────┼───────┤
│ AS13786 │ left       │ 127   │
│ AS14840 │ left       │ 446   │
│ AS16735 │ left       │ 18    │
│ AS263998│ left       │ 8     │
│ AS28220 │ left       │ 3     │
│ AS28343 │ left       │ 19    │
│ AS35280 │ left       │ 8     │
│ AS39120 │ left       │ 2     │
│ AS4230  │ left       │ 135   │
│ AS52674 │ left       │ 1     │
│ ...     │ and 24 more neighbors │
╰─────────┴────────────┴───────╯
```

### Announced Prefixes

```bash
bgpin asn prefixes 262978
```

```
╭────────────────────────────────────────────────╮
│ Announced Prefixes for AS262978 (Total: 19)   │
├────┬────────────────────┬──────┤
│  # │ PREFIX             │ TYPE │
├────┼────────────────────┼──────┤
│  1 │ 2804:4d44:10::/48  │ IPv6 │
│  2 │ 143.0.121.0/24     │ IPv4 │
│  3 │ 186.250.187.0/24   │ IPv4 │
│  4 │ 186.250.184.0/24   │ IPv4 │
│  5 │ 2804:4d44:c::/48   │ IPv6 │
│  6 │ 132.255.220.0/24   │ IPv4 │
│  7 │ 132.255.220.0/22   │ IPv4 │
│  8 │ 143.0.122.0/24     │ IPv4 │
│  9 │ 143.0.123.0/24     │ IPv4 │
│ 10 │ 2804:4d44:ada::/48 │ IPv6 │
│ 11 │ 186.250.184.0/22   │ IPv4 │
│ 12 │ 132.255.221.0/24   │ IPv4 │
│ 13 │ 132.255.223.0/24   │ IPv4 │
│ 14 │ 132.255.222.0/24   │ IPv4 │
│ 15 │ 143.0.120.0/22     │ IPv4 │
│ 16 │ 2804:4d44::/32     │ IPv6 │
│ 17 │ 143.0.120.0/24     │ IPv4 │
│ 18 │ 186.250.186.0/24   │ IPv4 │
│ 19 │ 186.250.185.0/24   │ IPv4 │
╰────┴────────────────────┴──────╯
```

### RIS Peers

```bash
bgpin asn peers 262978
```

```
╭────────────────────────────────────────────────────────────────────╮
│ RIS Peers (Total: 1449)                                            │
├────┬───────────────────────────────────────────────────────────────┤
│  # │ PEER INFORMATION                                              │
├────┼───────────────────────────────────────────────────────────────┤
│  1 │ [rrc06] AS20473 - 2001:de8:8::2:473:1 (v4: 0, v6: 497)       │
│  2 │ [rrc06] AS20473 - 210.171.225.160 (v4: 264, v6: 0)           │
│  3 │ [rrc06] AS23815 - 2001:de8:8::2:3815:11 (v4: 0, v6: 8092)    │
│  4 │ [rrc06] AS23815 - 210.171.224.1 (v4: 133509, v6: 0)          │
│  5 │ [rrc06] AS2497 - 2001:200:0:fe00::9c1:0 (v4: 0, v6: 232042)  │
│  6 │ [rrc06] AS2497 - 202.249.2.169 (v4: 1031043, v6: 0)          │
│  7 │ [rrc06] AS2500 - 2001:200:0:fe00::9c4:11 (v4: 0, v6: 229853) │
│  8 │ [rrc06] AS2500 - 202.249.2.83 (v4: 1032644, v6: 0)           │
│  9 │ [rrc06] AS25152 - 2001:200:0:fe00::6249:0 (v4: 0, v6: 242287)│
│ 10 │ [rrc06] AS25152 - 202.249.2.185 (v4: 1053992, v6: 0)         │
│ ...│ and 1439 more peers                                           │
╰────┴───────────────────────────────────────────────────────────────╯
```

### Prefix Overview

```bash
bgpin prefix overview 186.250.184.0/24
```

```
╭──────────────────────────────────────────────╮
│ Prefix Overview: 186.250.184.0/24            │
├──────────────────┬───────────────────────────┤
│ Is Less Specific │ false                     │
│ Announcing ASNs  │ AS262978                  │
│ Query Time       │ 2026-03-03T11:52:20-03:00 │
╰──────────────────┴───────────────────────────╯
```

### Looking Glasses List

```bash
bgpin lg
```

```
╭───────────────────────────────────────────────────────────────────────────╮
│ Available Looking Glasses                                                 │
├────────────────────┬─────────┬────────┬──────────┬─────────┬──────────────┤
│ NAME               │ VENDOR  │ TYPE   │ PROTOCOL │ COUNTRY │ URL          │
├────────────────────┼─────────┼────────┼──────────┼─────────┼──────────────┤
│ Hurricane Electric │ cisco   │ public │ http     │ US      │ lg.he.net    │
│ NTT America        │ cisco   │ public │ http     │ US      │ lg.ntt.net   │
│ Telia Carrier      │ juniper │ public │ http     │ SE      │ lg.telia.net │
╰────────────────────┴─────────┴────────┴──────────┴─────────┴──────────────╯
```

## Formato JSON

```bash
bgpin asn info 262978 -o json
```

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

## Formato YAML

```bash
bgpin asn info 262978 -o yaml
```

```yaml
asn: 262978
holder: Centro de Tecnologia Armazem Datacenter Ltda.
announced: true
block: 262144-263167
description: ""
country: ""
```

## Características do Output

### Estilo Visual
- **Bordas arredondadas**: Usando caracteres Unicode (╭╮╰╯)
- **Linhas limpas**: Sem separadores entre linhas de dados
- **Título centralizado**: Informação principal destacada
- **Cores**: Suporte a cores no terminal (futuro)

### Configuração go-pretty
```go
t := table.NewWriter()
t.SetOutputMirror(os.Stdout)
t.SetTitle("Título da Tabela")
t.Style().Title.Align = text.AlignCenter
t.SetStyle(table.StyleRounded)
t.Style().Options.SeparateRows = false
```

### Vantagens
- ✅ Visual profissional e limpo
- ✅ Fácil leitura
- ✅ Compatível com Unicode
- ✅ Consistente em todos os comandos
- ✅ Suporta largura dinâmica
- ✅ Footer para indicar dados truncados

## Comparação com Outros Estilos

### StyleRounded (Usado)
```
╭─────────────╮
│ Title       │
├─────┬───────┤
│ A   │ B     │
│ 1   │ 2     │
╰─────┴───────╯
```

### StyleLight (Alternativo)
```
┌─────────────┐
│ Title       │
├─────┬───────┤
│ A   │ B     │
│ 1   │ 2     │
└─────┴───────┘
```

### StyleBold (Alternativo)
```
┏━━━━━━━━━━━━━┓
┃ Title       ┃
┣━━━━━┳━━━━━━━┫
┃ A   ┃ B     ┃
┃ 1   ┃ 2     ┃
┗━━━━━┻━━━━━━━┛
```

## Truncamento Inteligente

Quando há muitos dados, o output é truncado com footer informativo:

```
│ 28 │ 186.250.185.0/24   │ IPv4 │
│ 29 │ 186.250.186.0/24   │ IPv4 │
│ 30 │ 186.250.187.0/24   │ IPv4 │
├────┴────────────────────┴──────┤
│ ... and 150 more prefixes      │
╰────────────────────────────────╯
```

## Detecção Automática

### IPv4 vs IPv6
O sistema detecta automaticamente o tipo de prefixo:
- IPv4: Não contém `:` 
- IPv6: Contém `:`

### Largura da Tabela
A largura se ajusta automaticamente ao conteúdo.

## Pipeline com Outras Ferramentas

### Com grep
```bash
bgpin asn prefixes 262978 | grep IPv6
```

### Com less
```bash
bgpin asn peers 262978 | less
```

### Com tee
```bash
bgpin asn info 262978 | tee asn_info.txt
```

## Exportar para Arquivo

### Formato Table
```bash
bgpin asn info 262978 > asn_info.txt
```

### Formato JSON
```bash
bgpin asn info 262978 -o json > asn_info.json
```

### Formato YAML
```bash
bgpin asn info 262978 -o yaml > asn_info.yaml
```

## Dicas de UX

1. **Use table para visualização**: Melhor para leitura humana
2. **Use JSON para automação**: Melhor para parsing com jq
3. **Use YAML para configuração**: Melhor para documentação

## Futuras Melhorias

- [ ] Cores customizáveis
- [ ] Temas (dark/light)
- [ ] Paginação interativa
- [ ] Exportar para CSV
- [ ] Exportar para Markdown
- [ ] Gráficos ASCII
