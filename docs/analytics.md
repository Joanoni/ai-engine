# Analytics BI Panel

## Status

🟢 **Implementado** — v0.0.18. Bug de crash corrigido em v0.0.22.

---

## Overview

O Analytics BI Panel é um dashboard de observabilidade completo integrado ao frontend "Mission Control". Permite inspecionar sessões passadas em 3 níveis de detalhe: projeto → sessão → agente.

Ativado pelo botão **📊 Analytics** na sidebar (ou `Ctrl+Shift+A`). Quando ativo, substitui o `CockpitArea` (grafo + terminal) pelo painel de analytics — o `MissionPanel` direito permanece visível.

---

## Arquitetura

### Backend — Endpoints

| Endpoint | Método | Descrição |
|---|---|---|
| `GET /tokens` | GET | Retorna tokens e custo acumulado do projeto inteiro (`ProjectTokens`) |
| `GET /sessions/{id}/tokens` | GET | Retorna tokens e custo de uma sessão específica (`SessionTokens`) |
| `GET /sessions/{id}/logs` | GET | Retorna os logs JSONL de todos os agentes da sessão |

#### `GET /tokens` — Resposta

```json
{
  "input_tokens": 945154,
  "output_tokens": 40437,
  "total_tokens": 985591,
  "estimated_cost_usd": 3.442,
  "session_count": 2,
  "last_updated_at": "2026-05-18T15:22:34Z"
}
```

> **Atenção:** o campo de custo se chama `estimated_cost_usd` (não `cost_usd`). Mismatch nesse nome causou o crash de v0.0.18 a v0.0.21.

#### `GET /sessions/{id}/tokens` — Resposta

```json
{
  "session_id": "abc123",
  "input_tokens": 423284,
  "output_tokens": 35534,
  "total_tokens": 458818,
  "estimated_cost_usd": 1.8029,
  "updated_at": "2026-05-18T13:47:09Z"
}
```

#### `GET /sessions/{id}/logs` — Resposta

```json
{
  "agents": {
    "swarmito": [ ...LogEntry[] ],
    "backend-leader": [ ...LogEntry[] ],
    "backend-executor": [ ...LogEntry[] ]
  }
}
```

Lê os arquivos `.ai-engine/logs/{sessionId}/{agentName}/chat.jsonl` do workspace. Retorna `{ "agents": {} }` se o diretório não existir.

### Backend — Token Store

Implementado em [`internal/tokenstore/store.go`](../internal/tokenstore/store.go).

- **`ProjectTokens`** — acumulado do workspace inteiro, persistido em `.ai-engine/tokens.json`.
- **`SessionTokens`** — por sessão, persistido em `.ai-engine/sessions/{id}/tokens.json`.
- `StartSession(id)` — incrementa `session_count` no projeto.
- `AddUsage(id, model, input, output)` — calcula custo via `internal/pricing`, atualiza sessão e projeto.
- `ReadProject()` / `ReadSession(id)` — leitura thread-safe.

### Frontend — Estrutura de Arquivos

```
frontend/src/
├── components/analytics/
│   ├── AnalyticsPanel.tsx     # Container — gerencia navegação 3 níveis
│   ├── ProjectView.tsx        # Nível 1 — visão macro do projeto
│   ├── SessionView.tsx        # Nível 2 — drill-down de sessão
│   └── AgentView.tsx          # Nível 3 — drill-down de agente
├── hooks/
│   └── useAnalytics.ts        # Hook — fetch de logs por sessão
└── types/
    └── logs.ts                # Tipos TypeScript para o formato JSONL
```

---

## Navegação — 3 Níveis

```
Project View
  └─▶ Session View  (clique em sessão na tabela ou no bar chart)
        └─▶ Agent View  (clique em agente na tabela)
```

Breadcrumb no topo: `Analytics / Project` → `Analytics / Session` → `Analytics / Session / {agent-name}`

Botão **⌂ Project** (visível em Session e Agent View) retorna direto ao nível 1.

---

## Componentes

### `AnalyticsPanel.tsx`

Container top-level. Gerencia o estado de navegação (`AnalyticsView`) com 3 variantes:

```ts
type AnalyticsView =
  | { type: 'project' }
  | { type: 'session'; sessionId: string }
  | { type: 'agent'; sessionId: string; agentName: string };
```

Ao selecionar uma sessão, chama `loadSessionLogs(id)` (via `useAnalytics`) antes de navegar para `SessionView`. Passa os logs já carregados para `AgentView` via `sessionLogs.agents[agentName]`.

---

### `ProjectView.tsx`

Visão macro do workspace. Dados carregados de `GET /tokens` e `GET /sessions/{id}/tokens` (para cada sessão).

**Seções:**

1. **6 Stat Cards** — Total Missions, Done, Errors, Input Tokens, Output Tokens, Total Cost.
2. **Cost per Mission (USD)** — `BarChart` horizontal (recharts), uma barra por sessão ordenada por data. Barras clicáveis navegam para `SessionView`. Cor: azul (`#58a6ff`) para done, vermelho (`#f85149`) para error.
3. **Status Distribution** — `PieChart` donut (recharts) com distribuição done/error/running. Legenda manual abaixo.
4. **All Sessions** — tabela com colunas: Prompt (80 chars), Status, Started At, Duration, Cost. Linhas clicáveis navegam para `SessionView`. Ordenada por data decrescente.

**Funções auxiliares:**
- `formatDuration(startedAt, finishedAt)` — retorna `"Xm Ys"` ou `"Xs"` ou `"running..."`.
- `formatDate(iso)` — `toLocaleString()`.

---

### `SessionView.tsx`

Drill-down de uma sessão. Dados de tokens carregados de `GET /sessions/{id}/tokens`. Logs recebidos via prop (já carregados pelo `AnalyticsPanel`).

**Seções:**

1. **Header** — prompt completo, status badge, duração, custo total (`$X.XXXXXX total`).
2. **Agent Timeline** — swimlane horizontal div-based. Cada agente é uma barra colorida posicionada proporcionalmente ao tempo total da sessão (`startTs` / `endTs` dos logs `agent_init` e `finish`).
3. **Agent Summary** — tabela com colunas: Agent, Type (L/E badge), LLM Turns, Tool Calls, Input Tok, Output Tok, Cost, Avg Tool ms. Linhas clicáveis navegam para `AgentView`.
4. **Tool Usage** — `BarChart` horizontal (recharts) com contagem de chamadas por ferramenta. Cor por taxa de sucesso: verde ≥90%, amarelo ≥50%, vermelho <50%.
5. **Cost Breakdown by Agent** — barra horizontal stacked mostrando contribuição percentual de cada agente no custo total. Legenda com valores absolutos.

**Função `computeAgentStats(agentName, entries)`** — deriva estatísticas de um agente a partir dos `LogEntry[]`:
- `turns` = contagem de `llm_request`
- `toolCalls` = contagem de `tool_result`
- `inputTokens` / `outputTokens` = soma dos `llm_response`
- `cost` = calculado localmente: `input * 3/1M + output * 15/1M` (estimativa hardcoded)
- `avgToolDuration` = média dos `duration_ms` dos `tool_result`
- `startTs` / `endTs` = timestamps do `agent_init` e `finish`

**Função `computeToolUsage(logs)`** — agrega chamadas de ferramenta de todos os agentes da sessão.

---

### `AgentView.tsx`

Drill-down de um agente específico. Recebe `LogEntry[]` diretamente.

**Seções:**

1. **Header** — nome do agente, badge LEADER/EXECUTOR, modelo, resumo (turns, tool calls, tokens, custo).
2. **Tokens per Turn** — `LineChart` sparkline (recharts). Linha azul = input tokens, linha roxa = output tokens, por turno.
3. **Avg Tool Duration (ms)** — `BarChart` horizontal (recharts) com duração média por ferramenta.
4. **Turn-by-Turn Detail** — accordion expansível por turno. Botão "Expand All / Collapse All".

**Cada turno no accordion contém:**

| Seção | Conteúdo |
|---|---|
| Header | Número do turno, `stop_reason`, tokens in/out, contagem de tool calls, erros consecutivos |
| 📋 System Prompt | 4 abas: Engine Context / Workspace Tree (L4) / Agent Role / Task Context |
| 💬 Message History | Lista colapsável de mensagens com role + preview. Cada mensagem expansível mostra conteúdo completo |
| 🔧 Tools Available | Lista de badges com nomes das ferramentas disponíveis naquele turno |
| 🤖 LLM Response | stop_reason, tokens, texto da resposta |
| ⚡ Tool Executions | Por tool call: nome, ✓ OK / ✗ ERR, duração ms. Expansível: input JSON + output |

**Função `groupByTurn(logs)`** — agrupa `LogEntry[]` por número de turno em `TurnGroup[]`:
```ts
interface TurnGroup {
  turn: number;
  llmRequest: LogEntry | null;   // role=llm_request
  llmResponse: LogEntry | null;  // role=llm_response
  toolResults: LogEntry[];       // role=tool_result
}
```

**Sub-componentes internos:**
- `SystemPromptTabs` — 4 abas para as camadas do system prompt.
- `MessageHistorySection` — lista colapsável de mensagens com expansão individual.
- `ToolExecutionSection` — lista de tool calls com input/output expansíveis.
- `TurnAccordion` — wrapper de um turno com todas as seções acima.

---

### `useAnalytics.ts`

Hook simples para fetch de logs.

```ts
function useAnalytics(): {
  sessionLogs: SessionLogs | null;
  loadingLogs: boolean;
  loadSessionLogs: (sessionId: string) => Promise<SessionLogs | null>;
  clearSessionLogs: () => void;
}
```

`loadSessionLogs` faz `GET /sessions/{id}/logs`, armazena em estado, retorna o resultado. Erros retornam `null` silenciosamente.

---

## Tipos — `types/logs.ts`

```ts
interface LogEntry {
  ts: string;
  turn: number;
  role: 'agent_init' | 'user' | 'llm_request' | 'llm_response' | 'tool_result' | 'error' | 'finish';

  // agent_init
  agent_name?: string;
  agent_type?: string;
  session_id?: string;
  model?: string;

  // llm_request
  system_prompt?: string;
  system_layers?: SystemLayers;   // { engine_context, dynamic_context, agent_role, task_context }
  messages?: MessageLog[];
  tools?: ToolLog[];
  message_count?: number;
  total_tool_calls_so_far?: number;
  consecutive_errors?: number;

  // llm_response
  text?: string;
  tool_calls?: ToolCallEntry[];
  stop_reason?: string;
  input_tokens?: number;
  output_tokens?: number;

  // tool_result
  tool_use_id?: string;
  tool?: string;
  success?: boolean;
  output?: string;
  duration_ms?: number;

  // error / finish
  message?: string;
  result?: string;
}

interface SessionLogs {
  agents: Record<string, LogEntry[]>;
}
```

---

## Integração com `App.tsx`

```tsx
// App.tsx
const [showAnalytics, setShowAnalytics] = useState(false);
const { sessionLogs, loadingLogs, loadSessionLogs, clearSessionLogs } = useAnalytics();

// Ctrl+Shift+A toggle
// Quando showAnalytics=true: renderiza <AnalyticsPanel> no lugar de <CockpitArea>
```

```tsx
// Sidebar.tsx — botão Analytics
<button onClick={onToggleAnalytics}>📊 Analytics</button>
// Estilo: accent roxo quando showAnalytics=true
```

---

## Dependências

| Pacote | Versão | Uso |
|---|---|---|
| `recharts` | (via npm) | BarChart, PieChart, LineChart nos 3 views |

Bundle com recharts: **841 KB** (721 módulos, +33 pacotes recharts).

---

## Bugs Conhecidos e Histórico

### v0.0.18–v0.0.21 — Crash ao abrir Analytics

**Erro:** `TypeError: Cannot read properties of undefined (reading 'toFixed')`

**Root cause:** mismatch de nome de campo. O backend serializa `estimated_cost_usd` mas as interfaces `TokenData` em `ProjectView` e `SessionView` declaravam `cost_usd`. O campo chegava como `undefined` → `undefined.toFixed()` → crash do React.

**Corrigido em v0.0.22:**
- [`frontend/src/components/analytics/ProjectView.tsx`](../frontend/src/components/analytics/ProjectView.tsx) — `TokenData.cost_usd` → `estimated_cost_usd`; guard `?.estimated_cost_usd != null` adicionado.
- [`frontend/src/components/analytics/SessionView.tsx`](../frontend/src/components/analytics/SessionView.tsx) — `TokenData.cost_usd` → `estimated_cost_usd`; guard `?.estimated_cost_usd != null` adicionado.

---

## Arquivos Relevantes

| Arquivo | Descrição |
|---|---|
| [`frontend/src/components/analytics/AnalyticsPanel.tsx`](../frontend/src/components/analytics/AnalyticsPanel.tsx) | Container — navegação 3 níveis |
| [`frontend/src/components/analytics/ProjectView.tsx`](../frontend/src/components/analytics/ProjectView.tsx) | Nível 1 — visão macro |
| [`frontend/src/components/analytics/SessionView.tsx`](../frontend/src/components/analytics/SessionView.tsx) | Nível 2 — drill-down de sessão |
| [`frontend/src/components/analytics/AgentView.tsx`](../frontend/src/components/analytics/AgentView.tsx) | Nível 3 — drill-down de agente |
| [`frontend/src/hooks/useAnalytics.ts`](../frontend/src/hooks/useAnalytics.ts) | Hook de fetch de logs |
| [`frontend/src/types/logs.ts`](../frontend/src/types/logs.ts) | Tipos TypeScript para LogEntry |
| [`internal/tokenstore/store.go`](../internal/tokenstore/store.go) | Persistência de tokens |
| [`internal/server/server.go`](../internal/server/server.go) | Endpoints `/tokens`, `/sessions/{id}/tokens`, `/sessions/{id}/logs` |
| [`internal/chatlog/logger.go`](../internal/chatlog/logger.go) | Geração dos logs JSONL consumidos pelo Analytics |
