# Plan — Fix frontend function coverage

## Tasks

- [ ] Write `NoteCard.test.tsx` covering all handlers
- [ ] Write `NoteForm.test.tsx` covering all handlers
- [ ] Write `NoteGrid.test.tsx` covering all render states
- [ ] Expand `App.test.tsx` with missing function paths
- [ ] Verify `pnpm --dir frontend run coverage` passes (≥60% functions)

## Detail

### Problem

`pnpm --dir frontend run coverage` fails:

```
Coverage for functions (51.02%) does not meet global threshold (60%)
```

The three Note components have zero test files. The overall function coverage
is dragged down by:

| File         | % Funcs |
|--------------|---------|
| NoteCard.tsx | 7.14%   |
| NoteForm.tsx | 12.5%   |
| App.tsx      | 33.33%  |
| main.tsx     | 0%      |

`main.tsx` is intentionally excluded from coverage targets (bootstrapping only).
`NoteGrid.tsx` is already at 100% functions. Focus is NoteCard + NoteForm + App.

---

### Test infrastructure

- Vitest + jsdom, config in `vite.config.ts`
- Globals enabled (no need to import `describe`/`it`/`expect`)
- Setup: `./src/test/setup.ts` → `@testing-library/jest-dom`
- Only `fireEvent` (no userEvent) — follow existing pattern
- `vi.fn()` for callbacks; `vi.fn().mockResolvedValue(undefined)` for async handlers

---

### NoteCard.test.tsx

Location: `frontend/src/components/NoteCard.test.tsx`

Fixture:
```ts
const note: Todo = {
  id: 1, title: 'Test title', body: 'Test body',
  color: 'default', pinned: false, done: false,
  created_at: '2026-06-21T01:00:00Z',
  updated_at: '2026-06-21T01:00:00Z',
};
```

Functions to cover (each one at minimum one test):

1. **Render** — title and body are visible
2. **Pin button** — click calls `onTogglePin(1)`
3. **Done button** — click calls `onEdit(1, { done: true })`
4. **Delete button** — click calls `onDelete(1)`
5. **Edit title** — click title → textarea appears → blur with new value → `onEdit(1, { title: 'New' })`
6. **Revert title on empty** — blur empty textarea → `onEdit` not called, reverts to original
7. **Revert title on Escape** — Escape key → exits editing without saving
8. **Edit body** — click body → textarea appears → blur with new value → `onEdit(1, { body: 'New' })`
9. **Revert body on Escape** — Escape key → exits without saving
10. **Color picker open** — click 🎨 → picker appears
11. **Color swatch select** — click swatch → `onEdit(1, { color: 'red' })` called, picker closes
12. **done class** — when `todo.done=true`, card has `note-card--done` class

---

### NoteForm.test.tsx

Location: `frontend/src/components/NoteForm.test.tsx`

Functions to cover:

1. **Initial render** — placeholder "Tomar una nota..." is visible
2. **Expand on click** — click placeholder → title input + body textarea appear
3. **Commit with title** — type title, type body, click "Cerrar" → `onAdd('title', 'body', 'default')` called
4. **Commit with empty title** — expand, type body only, click "Cerrar" → `onAdd` NOT called
5. **Reset on Escape** — typing then Escape → form collapses, `onAdd` not called
6. **Color picker toggle** — click 🎨 → picker opens; click again → closes
7. **Color swatch select** — open picker, click swatch → color set, picker closes
8. **Click outside** — `fireEvent.mouseDown(document.body)` → `commit()` triggered

---

### NoteGrid.test.tsx

Location: `frontend/src/components/NoteGrid.test.tsx`

Functions to cover:

1. **Loading state** — renders `aria-label="Cargando notas"` spinner
2. **Error state** — renders error string
3. **Empty state** — renders empty message
4. **Pinned + unpinned** — shows "FIJADAS" and "OTRAS" section headers
5. **Only pinned** — shows "FIJADAS", no "OTRAS"
6. **Only unpinned** — no section headers, cards render
7. **Props passthrough** — `onEdit`, `onDelete`, `onTogglePin` forwarded to NoteCard

Fixtures: build `Todo[]` with `pinned: true` and `pinned: false` variants.

---

### App.test.tsx — additions

The existing file has 2 tests. Add:

1. **Search filtering** — mock returns 2 todos; type in search → only matching card visible
2. **Clear search button** — type then click ✕ → query clears, all todos visible
3. **addTodo wired** — mock `addTodo`; expand NoteForm, fill title, close → `addTodo` called
4. **Error state** — mock `useTodos` with `error: 'fail'` → error message renders

---

### Verification

```
pnpm --dir frontend run coverage
```

Expected: all thresholds pass (≥60% on functions, statements, branches, lines).
