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
- **Domain Layer:** Define o núcleo. Contém as Entidades e as **Interfaces** dos Repositórios (Contratos). ZERO dependências externas.
- **Infrastructure Layer:** A implementação REAL DEVE obrigatoriamente satisfazer as interfaces criadas no domínio.
- A regra é estrita: a camada de *Use Case* conhece apenas as interfaces. A injeção de dependência define qual implementação de infraestrutura será usada em tempo de execução.

## 4. Regras de Modelagem de Dados
- **Identificadores (IDs Múltiplos):**
  - Use `UUIDv7` APENAS para entidades principais e sensíveis a enumeração (Ex: `User`, `Aluno`, `Turma`, `Aula`).
  - Use `Auto-increment` (IDs numéricos sequenciais) para entidades de alto volume ou puramente transacionais (Ex: `Produto`, `Movimentacao` de estoque).
- **Enums:** Propriedades restritas devem usar tipos customizados em Go atuando como Enums. 
  - Exemplo para `DayOfWeek`: Crie `type DayOfWeek int` e defina constantes (`Segunda DayOfWeek = 1`, etc). NUNCA use `int` solto ou `string` para representar dias da semana.
- **Horários:** A entidade `Aula` usa o utilitário `TimeOfDay` (HH:MM:SS). A entidade `DiarioDeClasse` usa `time.Time`.

## 5. Regras de Observabilidade e Logging (CRÍTICO)
- **Logs Estruturados:** Todos os logs devem ter níveis (INFO, WARN, ERROR) e formato JSON.
- **Rastreabilidade (Traceability):** Todo erro logado deve conter o contexto de *onde* ocorreu e *quais* dados estavam envolvidos.
- **Sanitização (LGPD):** É ESTRITAMENTE PROIBIDO logar senhas em texto claro, CPFs completos, emails não mascarados ou tokens. Os dados do log DEVEM ser retificados (redacted) antes da emissão.