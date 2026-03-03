# Contributing to bgpin

Obrigado por considerar contribuir com o bgpin! ðŸŽ‰

## Como Contribuir

### Reportar Bugs

Se vocÃª encontrou um bug, por favor abra uma issue com:
- DescriÃ§Ã£o clara do problema
- Passos para reproduzir
- Comportamento esperado vs atual
- VersÃ£o do bgpin (`bgpin version`)
- Sistema operacional

### Sugerir Features

Adoramos novas ideias! Abra uma issue com:
- DescriÃ§Ã£o da feature
- Caso de uso
- Exemplos de como seria usado

### Pull Requests

1. Fork o repositÃ³rio
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanÃ§as (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

### PadrÃµes de CÃ³digo

- Use `gofmt` para formatar o cÃ³digo
- Execute `go vet` antes de commitar
- Adicione testes para novas funcionalidades
- Mantenha a cobertura de testes acima de 70%
- Siga os padrÃµes do projeto

### Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com cobertura
go test -cover ./...

# Executar testes de integraÃ§Ã£o
go test -v ./sdk/integration_test/
```

### DocumentaÃ§Ã£o

- Atualize o README.md se necessÃ¡rio
- Adicione exemplos de uso
- Documente novas flags e comandos
- Mantenha os docs/ atualizados

### Commit Messages

Use mensagens claras e descritivas:

```
feat: add flow anomaly detection
fix: correct ASN parsing with AS prefix
docs: update CLI guide with flow commands
test: add integration tests for flow collector
refactor: improve telemetry initialization
```

Prefixos:
- `feat`: Nova funcionalidade
- `fix`: CorreÃ§Ã£o de bug
- `docs`: DocumentaÃ§Ã£o
- `test`: Testes
- `refactor`: RefatoraÃ§Ã£o
- `perf`: Performance
- `chore`: ManutenÃ§Ã£o

### Code Review

Todos os PRs passam por code review. Esperamos:
- CÃ³digo limpo e legÃ­vel
- Testes adequados
- DocumentaÃ§Ã£o atualizada
- Sem breaking changes (ou bem documentados)

### LicenÃ§a

Ao contribuir, vocÃª concorda que suas contribuiÃ§Ãµes serÃ£o licenciadas sob a MIT License.

## Desenvolvimento

### Setup

```bash
# Clone o repositÃ³rio
git clone https://github.com/rsdenck/bgpin
cd bgpin

# Instale dependÃªncias
go mod download

# Compile
go build -o bgpin ./cmd/cli/

# Execute testes
go test ./...
```

### Estrutura do Projeto

```
bgpin/
â”œâ”€â”€ cmd/cli/              # CLI commands
â”œâ”€â”€ internal/             # Internal packages
â”‚   â”œâ”€â”€ adapters/        # External adapters
â”‚   â”œâ”€â”€ core/            # Core domain
â”‚   â”œâ”€â”€ flow/            # Flow analysis
â”‚   â”œâ”€â”€ parsers/         # Vendor parsers
â”‚   â””â”€â”€ telemetry/       # Observability
â”œâ”€â”€ pkg/                 # Public packages
â”œâ”€â”€ sdk/                 # RIPE RIS SDK
â””â”€â”€ docs/                # Documentation
```

### Ãreas para Contribuir

- [ ] Parsers para novos vendors (Arista, Bird, etc)
- [ ] Novos adapters (SSH, Telnet)
- [ ] AnÃ¡lise de anomalias BGP
- [ ] ValidaÃ§Ã£o RPKI
- [ ] IntegraÃ§Ã£o com mais APIs (PeeringDB, RouteViews)
- [ ] Dashboard web
- [ ] Testes adicionais
- [ ] DocumentaÃ§Ã£o
- [ ] Exemplos de uso

## Comunidade

- GitHub Issues: https://github.com/rsdenck/bgpin/issues
- Discussions: https://github.com/rsdenck/bgpin/discussions

## CÃ³digo de Conduta

Seja respeitoso e profissional. NÃ£o toleramos:
- Linguagem ofensiva
- AssÃ©dio
- DiscriminaÃ§Ã£o
- Spam

Queremos uma comunidade acolhedora para todos!

---

Obrigado por contribuir! ðŸš€
