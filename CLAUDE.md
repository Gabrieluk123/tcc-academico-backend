# Contexto do Projeto: TCC - Sistema de Gerenciamento Escolar

## 1. Visão Geral
Sistema single-tenant para a Escola Municipal Dr. Auto de Oliveira Pinto. Foco: Controle de Estoque e Monitoramento de Frequência. 
**NÃO implemente lógicas de multi-tenancy.**

## 2. Stack Tecnológico
- **Backend:** Go (Golang) + Echo Framework.
- **Frontend:** ReactJS (SPA).
- **Banco de Dados:** PostgreSQL (GORM).
- **Observabilidade:** Logs estruturados em JSON (ex: pacote `log/slog` nativo do Go 1.21+).

## 3. Padrão Arquitetural: Go Standard Layout + Clean Architecture
A aplicação segue o padrão oficial do Go, isolando código privado na pasta `internal/`.
- **`cmd/api/main.go`**: Ponto de entrada da aplicação. Inicializa dependências (Wire up).
- **`internal/common/`**: Utilitários globais, formatação de erros padrão, logger e structs base (ex: `TimeOfDay`).
- **`internal/domain/` (FLAT DOMAIN):** Contém TODAS as entidades (structs puras) e Interfaces (Contratos) em um único pacote (`package domain`). Isso evita ciclos de importação (Circular Dependencies). Não crie subpastas aqui.
- **`internal/usecase/` (PACKAGE BY FEATURE):** Lógica de negócio agrupada por módulo. Ex: `internal/usecase/auth` (`package auth`).
- **`internal/http/` (TRANSPORT):** Focada no framework Echo.
  - `/handler`: Agrupado por feature. Ex: `internal/http/handler/auth`.
  - `/middleware`: Middlewares customizados do Echo.
  - `/route.go`: Configuração das rotas.
- **`internal/infrastructure/` (PACKAGE BY FEATURE):** Implementações reais agrupadas por módulo. Ex: `internal/infrastructure/auth` (Hash, JWT) e `internal/infrastructure/database` (GORM).

## 4. Regras de Modelagem de Dados
- **Identificadores (IDs Múltiplos):**
  - Use `UUIDv7` APENAS para entidades principais e sensíveis a enumeração (Ex: `User`, `Aluno`, `Turma`, `Aula`).
  - Use `Auto-increment` (IDs numéricos sequenciais) para entidades de alto volume ou puramente transacionais (Ex: `Produto`, `Movimentacao` de estoque).
- **Enums:** Propriedades restritas devem usar tipos customizados em Go atuando como Enums. 
  - Exemplo para `DayOfWeek`: Crie `type DayOfWeek int` e defina constantes (`Segunda DayOfWeek = 1`, etc). NUNCA use `int` solto ou `string` para representar dias da semana.
- **Horários:** A entidade `Aula` (agenda recorrente) usa o utilitário `TimeOfDay` (HH:MM:SS). A entidade `DiarioDeClasse` (registro histórico) usa `time.Time`.
- **Controle de Acesso (IAM e Perfis):**
  - NUNCA faça hardcode de Roles (Papéis) ou Permissões usando Enums (`iota`). O sistema possui criação dinâmica de perfis para integração com Casbin.
  - O domínio DEVE possuir as entidades `Role` (ID UUIDv7, Nome, Descricao) e `Permission` (ID UUIDv7, Slug, Descricao).
  - A entidade `User` DEVE possuir um campo `RoleID` do tipo `uuid.UUID` apontando para a entidade `Role`.
  - O papel "Admin" é apenas o registro inicial (bootstrap) no banco de dados, e não uma constante solta no código.

## 5. Regras de Observabilidade e Logging (CRÍTICO)
- **Logs Estruturados:** Todos os logs devem ter níveis (INFO, WARN, ERROR) e formato JSON.
- **Rastreabilidade (Traceability):** Todo erro logado deve conter o contexto de *onde* ocorreu e *quais* dados estavam envolvidos.
- **Sanitização (LGPD):** É ESTRITAMENTE PROIBIDO logar senhas em texto claro, CPFs completos, emails não mascarados ou tokens. Os dados do log DEVEM ser retificados (redacted) antes da emissão.

## 6. Princípios de Design (SOLID e DRY)
- **Single Responsibility (SRP):** Cada struct, função ou pacote deve ter apenas um motivo para mudar. (Ex: O Handler só faz parse de HTTP, não contém regras de negócio).
- **Dependency Inversion (DIP):** Dependa de abstrações (Interfaces), não de implementações concretas.
- **DRY (Don't Repeat Yourself):** Se uma lógica (como validação de permissão, extração de token ou formatação de log) se repetir, extraia para um middleware, utility ou serviço de domínio focado.