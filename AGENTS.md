# Diretrizes de Comportamento e Desenvolvimento (AI Agents)

## 1. Regra de Ouro: Contratos Primeiro (Interface-First)
- O domínio é o chefe. Sempre defina a `interface` no pacote de domínio primeiro.
- Ao escrever a implementação (ex: no pacote `repository/postgres`), você DEVE garantir que a struct implementa a interface do domínio.
  ```go
  // Garante em tempo de compilação que a implementação satisfaz o contrato
  var _ domain.ProdutoRepository = (*produtoGormRepository)(nil)
  ```

## 2. Padrão de Logging Incremental (Context Propagation)
- TODAS as funções da camada de Use Case, Repository e Adapters DEVEM receber ctx context.Context como o primeiro parâmetro.
- Acréscimo de Contexto: Use o context.Context para propagar variáveis de rastreabilidade (ex: request_id, user_id, turma_id).

- Quando um erro ocorrer no meio de um processo longo (ex: salvar diário), o logger deve extrair esse contexto incremental acumulado no ctx para gerar um log rico sem a necessidade de passar o logger explicitamente por todos os parâmetros da função.
  ```go
  logger.ErrorContext(ctx, "falha ao realizar operacao", 
      slog.String("entidade_id", entidade.ID.String()),
      slog.String("error", err.Error()),
  )
  ```

3. Fluxo de TDD (Test-Driven Development)
- Ao ser solicitado para criar uma nova funcionalidade de Use Case, escreva os testes primeiro.
- Testes Unitários: Isole a camada de Use Case utilizando Mocks gerados a partir das interfaces de Domínio (utilize a biblioteca testify/mock).
- Testes de Integração: Ao testar a camada de Infraestrutura (GORM Repositories), utilize o banco de dados real.

4. Tratamento de Erros e Early Returns
- Se uma função falhar, adicione contexto ao erro antes de subir na pilha usando fmt.Errorf("contexto da falha: %w", err). Nunca "engula" erros omitindo o erro original.

- Use Early Returns para evitar aninhamento visual profundo (if/else hell). Retorne erros o mais rápido possível na função.