# Contributing to bgpin

Obrigado por considerar contribuir com o bgpin! 🎉

## Como Contribuir

### Reportar Bugs

Se você encontrou um bug, por favor abra uma issue com:
- Descrição clara do problema
- Passos para reproduzir
- Comportamento esperado vs atual
- Versão do bgpin (`bgpin version`)
- Sistema operacional

### Sugerir Features

Adoramos novas ideias! Abra uma issue com:
- Descrição da feature
- Caso de uso
- Exemplos de como seria usado

### Pull Requests

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

### Padrões de Código

- Use `gofmt` para formatar o código
- Execute `go vet` antes de commitar
- Adicione testes para novas funcionalidades
- Mantenha a cobertura de testes acima de 70%
- Siga os padrões do projeto

### Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com cobertura
go test -cover ./...

# Executar testes de integração
go test -v ./sdk/integration_test/
```

### Documentação

- Atualize o README.md se necessário
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
- `fix`: Correção de bug
- `docs`: Documentação
- `test`: Testes
- `refactor`: Refatoração
- `perf`: Performance
- `chore`: Manutenção

### Code Review

Todos os PRs passam por code review. Esperamos:
- Código limpo e legível
- Testes adequados
- Documentação atualizada
- Sem breaking changes (ou bem documentados)

### Licença

Ao contribuir, você concorda que suas contribuições serão licenciadas sob a MIT License.

## Desenvolvimento

### Setup

```bash
# Clone o repositório
git clone https://github.com/rsdenck/bgpin
cd bgpin

# Instale dependências
go mod download

# Compile
go build -o bgpin ./cmd/cli/

# Execute testes
go test ./...
```

### Estrutura do Projeto

```
bgpin/
├── cmd/cli/              # CLI commands
├── internal/             # Internal packages
│   ├── adapters/        # External adapters
│   ├── core/            # Core domain
│   ├── flow/            # Flow analysis
│   ├── parsers/         # Vendor parsers
│   └── telemetry/       # Observability
├── pkg/                 # Public packages
├── sdk/                 # RIPE RIS SDK
└── docs/                # Documentation
```

### Áreas para Contribuir

- [ ] Parsers para novos vendors (Arista, Bird, etc)
- [ ] Novos adapters (SSH, Telnet)
- [ ] Análise de anomalias BGP
- [ ] Validação RPKI
- [ ] Integração com mais APIs (PeeringDB, RouteViews)
- [ ] Dashboard web
- [ ] Testes adicionais
- [ ] Documentação
- [ ] Exemplos de uso

## Comunidade

- GitHub Issues: https://github.com/rsdenck/bgpin/issues
- Discussions: https://github.com/rsdenck/bgpin/discussions

## Código de Conduta

Seja respeitoso e profissional. Não toleramos:
- Linguagem ofensiva
- Assédio
- Discriminação
- Spam

Queremos uma comunidade acolhedora para todos!

---

Obrigado por contribuir! 🚀
