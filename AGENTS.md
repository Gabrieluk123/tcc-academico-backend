# Diretrizes de Comportamento e Desenvolvimento (AI Agents)

## 1. Regra de Ouro: Contratos Primeiro (Interface-First)
- O domínio é o chefe. Sempre defina a `interface` no pacote de domínio primeiro.
- Ao escrever a implementação (ex: no pacote `repository/postgres`), você DEVE garantir que a struct implementa a interface do domínio.
  ```go
  // Garante em tempo de compilação que a implementação satisfaz o contrato
  var _ domain.ProdutoRepository = (*produtoGormRepository)(nil)