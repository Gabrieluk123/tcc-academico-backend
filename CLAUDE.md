# Contexto do Projeto: TCC - Sistema de Gerenciamento Escolar

## 1. Visão Geral
Sistema single-tenant para a Escola Municipal Dr. Auto de Oliveira Pinto. Foco: Controle de Estoque e Monitoramento de Frequência. 
**NÃO implemente lógicas de multi-tenancy.**

## 2. Stack Tecnológico
- **Backend:** Go (Golang) + Echo Framework.
- **Frontend:** ReactJS (SPA).
- **Banco de Dados:** PostgreSQL (GORM).
- **Observabilidade:** Logs estruturados em JSON (ex: pacote `log/slog` nativo do Go 1.21+).

## 3. Padrão Arquitetural: Clean Architecture + DDD
A aplicação DEVE ser estritamente focada em contratos (Interfaces). A dependência deve sempre apontar para o centro (Domínio).
- **Domain Layer:** Contém as Entidades (structs puras) e as Interfaces dos Repositórios (Contratos). ZERO dependências externas ou de banco de dados.
- **Use Case Layer:** Contém a lógica de negócio. Recebe as interfaces dos repositórios via Injeção de Dependência.
- **HTTP Layer (Transport):** Focada no framework Echo.
  - `/http/handler`: Funções que fazem o bind do request, chamam o Use Case e retornam JSON.
  - `/http/middleware`: Middlewares customizados do Echo (JWT, injeção de Contexto, Casbin).
  - `/http/route.go`: Arquivo central de roteamento.
- **Infrastructure Layer:** Implementações reais das interfaces (GORM Repositories, DB connection).

## 4. Regras de Modelagem de Dados
- **Identificadores (IDs Múltiplos):**
  - Use `UUIDv7` APENAS para entidades principais e sensíveis a enumeração (Ex: `User`, `Aluno`, `Turma`, `Aula`).
  - Use `Auto-increment` (IDs numéricos sequenciais) para entidades de alto volume ou puramente transacionais (Ex: `Produto`, `Movimentacao` de estoque).
- **Enums:** Propriedades restritas devem usar tipos customizados em Go atuando como Enums. 
  - Exemplo para `DayOfWeek`: Crie `type DayOfWeek int` e defina constantes (`Segunda DayOfWeek = 1`, etc). NUNCA use `int` solto ou `string` para representar dias da semana.
- **Horários:** A entidade `Aula` (agenda recorrente) usa o utilitário `TimeOfDay` (HH:MM:SS). A entidade `DiarioDeClasse` (registro histórico) usa `time.Time`.

## 5. Regras de Observabilidade e Logging (CRÍTICO)
- **Logs Estruturados:** Todos os logs devem ter níveis (INFO, WARN, ERROR) e formato JSON.
- **Rastreabilidade (Traceability):** Todo erro logado deve conter o contexto de *onde* ocorreu e *quais* dados estavam envolvidos.
- **Sanitização (LGPD):** É ESTRITAMENTE PROIBIDO logar senhas em texto claro, CPFs completos, emails não mascarados ou tokens. Os dados do log DEVEM ser retificados (redacted) antes da emissão.

## 6. Princípios de Design (SOLID e DRY)
- **Single Responsibility (SRP):** Cada struct, função ou pacote deve ter apenas um motivo para mudar. (Ex: O Handler só faz parse de HTTP, não contém regras de negócio).
- **Dependency Inversion (DIP):** Dependa de abstrações (Interfaces), não de implementações concretas.
- **DRY (Don't Repeat Yourself):** Se uma lógica (como validação de permissão, extração de token ou formatação de log) se repetir, extraia para um middleware, utility ou serviço de domínio focado.