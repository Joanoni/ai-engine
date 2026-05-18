# Analytics Panel — Bug Debug

## Status

🟢 **Resolvido** — v0.0.22. Bug corrigido em `ProjectView.tsx` e `SessionView.tsx`.

---

## Erro

```
TypeError: Cannot read properties of undefined (reading 'toFixed')
    at PZ (index-Cz60Qz4A.js:49:109766)
    at Co (...)
    at gc (...)
    at Mc (...)
    at Iu (...)
    ...
    at MessagePort.D (...)
```

- A função `PZ` é interna ao recharts (provavelmente `getTicksOfAxis` ou similar).
- O erro ocorre **sempre** ao clicar no botão Analytics na sidebar.
- O React crasha completamente — tela fica preta, DOM quase vazio.
- O erro persiste em todos os builds testados (v0.0.18 → v0.0.21).

---

## Contexto

- Servidor: `projects/budget-tracker/` rodando na porta `8080`
- Há **1 sessão** salva: "Build a Personal Budget Tracker..." (status: done ou error)
- O `ProjectView` é o primeiro componente renderizado ao abrir Analytics
- O `ProjectView` usa recharts: `BarChart` (custo por sessão) + `PieChart` (distribuição de status)

---

## O que já foi tentado (sem sucesso)

### Tentativa 1 — Corrigir `renderPieLabel`
- **Hipótese:** `props.value` undefined no label customizado do `Pie`
- **Ação:** Adicionado `Number(props.value ?? 0)` no `renderPieLabel`
- **Resultado:** Erro persiste

### Tentativa 2 — Remover `label` e `labelLine` do `Pie`
- **Hipótese:** recharts calcula posição de label mesmo sem dados válidos
- **Ação:** Removido `label={renderPieLabel}` e `labelLine={false}` do `<Pie>`
- **Resultado:** Erro persiste

### Tentativa 3 — Corrigir `tickFormatter` do `XAxis`
- **Hipótese:** `v` undefined no `tickFormatter` do `BarChart`
- **Ação:** Mudado `(v: number) => v.toFixed(4)` para `(v: unknown) => Number(v ?? 0).toFixed(4)`
- **Resultado:** Erro persiste

### Tentativa 4 — Adicionar `domain` explícito no `XAxis`
- **Hipótese:** recharts crashando ao calcular domínio com todos os valores zero
- **Ação:** Adicionado `domain={[0, 0.001]}` quando `hasAnyCost === false`, `domain={[0, 'auto']}` quando há custos
- **Resultado:** Erro persiste

### Tentativa 5 — Substituir CSS variables por hex nos `Cell` e `Line`
- **Hipótese:** recharts não suporta CSS variables como `fill`/`stroke` em componentes de dados
- **Ação:** Substituídos `var(--success)`, `var(--error)`, `var(--warning)`, `var(--accent)`, `var(--purple)` por valores hex (`#3fb950`, `#f85149`, `#d29922`, `#58a6ff`, `#bc8cff`) em:
  - `ProjectView.tsx` — `pieData.color`, `Cell fill`, `STATUS_COLORS`
  - `SessionView.tsx` — `toolBarColor()`, `agentColors`, `STATUS_COLORS`
  - `AgentView.tsx` — `Line stroke`, `Cell fill`
- **Resultado:** Erro persiste

---

## Observações importantes

### Offset do erro é estável mas muda com cada build
| Build | Bundle | Offset linha 49 |
|---|---|---|
| v0.0.18 | `index-1HinGc4q.js` | 109791 |
| v0.0.19 | `index-CDKoEuht.js` | 109754 |
| v0.0.20 | `index-BUTZFZ1G.js` | 109804 |
| v0.0.21 | `index-Cz60Qz4A.js` | 109766 |

O offset muda a cada build (código adicionado/removido desloca o bundle), mas a função `PZ` é sempre a mesma do recharts.

### Inspeção do bundle na linha 49, offset 109766 (v0.0.21)
```js
// Trecho ao redor do offset:
...AZ,{label:`Total Cost`,value:n?`$${n.cost_usd.toFixed(4)}`:`—`})...
```

**Isso indica que o erro pode estar em `n.cost_usd.toFixed(4)` onde `n` é `projectTokens`** — mas o código tem guard `n ? ... : '—'`. Porém, o `n` pode ser um objeto com `cost_usd: undefined` (não null/undefined o objeto em si).

### Hipótese mais provável ainda não testada
O endpoint `GET /tokens` retorna um objeto JSON onde `cost_usd` pode ser `null` ou ausente. O guard `n ? ... : '—'` verifica se `n` é truthy (objeto existe), mas não verifica se `n.cost_usd` é um número válido. Se `cost_usd` for `null` ou `undefined`, `null.toFixed(4)` ou `undefined.toFixed(4)` causaria exatamente esse erro.

**Verificar:** O que o endpoint `GET /tokens` retorna quando não há dados de token ainda?

---

## Próximos passos sugeridos

1. **Verificar resposta do endpoint `/tokens`:**
   ```bash
   curl http://localhost:8080/tokens
   ```
   Se retornar `{"cost_usd": null}` ou `{"cost_usd": 0}` mas sem o campo, o guard `n ?` não protege.

2. **Corrigir o guard em `ProjectView.tsx` linha 127:**
   ```tsx
   // Atual (inseguro se cost_usd for null/undefined):
   value={projectTokens ? `$${projectTokens.cost_usd.toFixed(4)}` : '—'}
   
   // Correto:
   value={projectTokens?.cost_usd != null ? `$${projectTokens.cost_usd.toFixed(4)}` : '—'}
   ```

3. **Verificar também `SessionView.tsx` linha 185:**
   ```tsx
   // Atual:
   ${sessionCost.cost_usd.toFixed(6)} total
   // Correto:
   ${(sessionCost.cost_usd ?? 0).toFixed(6)} total
   ```

4. **Verificar `tokenstore`** — ver o que `internal/tokenstore/store.go` retorna quando não há dados.

---

## Arquivos relevantes

- [`frontend/src/components/analytics/ProjectView.tsx`](../frontend/src/components/analytics/ProjectView.tsx) — componente principal, linha 127 é suspeita
- [`frontend/src/components/analytics/SessionView.tsx`](../frontend/src/components/analytics/SessionView.tsx) — linha 185 suspeita
- [`internal/server/server.go`](../internal/server/server.go) — endpoints `/tokens` e `/sessions/{id}/tokens`
- [`internal/tokenstore/store.go`](../internal/tokenstore/store.go) — lógica de armazenamento de tokens

---

## Estado atual do código (v0.0.21)

As seguintes mudanças foram aplicadas mas **não resolveram** o problema:
- `renderPieLabel` removido do `Pie`
- `tickFormatter` com guard `Number(v ?? 0)`
- `domain` explícito no `XAxis` do `BarChart`
- Todas as CSS variables substituídas por hex nos componentes recharts
- `pieData.color` usando hex direto
