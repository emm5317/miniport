# MiniPort Frontend Redesign Plan

## Context

The current frontend uses hardcoded hex colors, a cramped single-line host strip, a 9-column `<table>` that overflows below 1100px, 6 action buttons per row, plain-text logs, and no responsive breakpoint. The redesign addresses all of these with CSS custom properties, a 5-card host strip, a 7-column CSS grid table, compact icon-based actions with a dropdown, server-side log colorization, and a 900px responsive breakpoint. No new dependencies. No JS build step.

---

## Phase 1: CSS Custom Properties + Kill Utility Bloat

**File:** `web/static/style.css`

- Add `:root` block at top with all semantic tokens using the **mockup's color palette** (not current hex values)
- Adopt the spec's refined palette: `--bg-page: #0d0f12`, `--bg-surface: #111418`, `--bg-raised: #161a20`, `--border: #1f2228`, etc.
- Switch to monospace-forward typography: `'SF Mono', ui-monospace, 'Cascadia Code', 'Fira Code', monospace`
- Replace hardcoded hex in existing component classes (`.toast-*`, `.search-input`, `.stat-bar`, etc.) with `var()` references
- Keep utility classes functional for now — they get removed as each template is rewritten in later phases

**Color tokens (from spec):**
```
--bg-page:       #0d0f12
--bg-surface:    #111418
--bg-raised:     #161a20
--bg-raised-hover: #13161c
--border:        #1f2228
--border-hover:  #2a2d35
--text-primary:  #e2e6ed
--text-secondary: #c8cdd6
--text-muted:    #8891a5
--text-dim:      #555b6b
--text-faint:    #3a3f4b
--color-running: #3ecf8e
--color-stopped: #3a3f4b
--color-warning: #f59e0b
--color-danger:  #ef4444
--color-info:    #4ea8de
--bg-success-subtle: #0a1a10
--bg-warning-subtle: #161008
--bg-danger-subtle:  #160808
--bg-info-subtle:    #0e1929
--border-success-subtle: #1a3a1a
--border-warning-subtle: #2a1f0a
--border-danger-subtle:  #2a1010
--border-info:           #1a3148
```

---

## Phase 2: Nav Bar

**File:** `web/templates/layouts/base.html`

- Split "MiniPort" into `Mini` + `<span class="nav-accent">Port</span>` with green accent
- Restyle Prune dropdown to use new var-based colors
- Reuse `.action-dropdown` pattern for Prune dropdown (unify dropdown CSS)

---

## Phase 3: Host Stats Strip

**Files:** `web/templates/partials/host-strip.html`, `web/static/style.css`

- **Delete** entire current content (28 lines of flex layout)
- **Replace** with 5-card CSS grid: `.host-strip`, `.hstat`, `.hstat-label`, `.hstat-val`, `.hstat-bar`, `.hstat-fill`
- Cards: CPU (green bar), Memory (blue bar), Disk (amber bar), Network (no bar), Uptime (no bar)
- Use existing `.Host.MemPercent` for memory bar width — no `memPct` helper needed

---

## Phase 4: Summary Strip + Filter Bar Merge

**Files:** `web/templates/partials/container-table.html`, `web/templates/partials/summary-strip.html`

- **Inline** summary-strip content into container-table.html's control bar
- Create single `.summary` div: running/stopped badge counts, filter tabs, search input (one row)
- **Important:** prefix all fields with `.Summary.` after inlining (data path change)
- **Delete** `web/templates/partials/summary-strip.html`
- Replace `.filter-tab` with `.tab` / `.tab.active`

---

## Phase 5: Container Table — `<table>` to CSS Grid

**Files:** `web/templates/partials/container-table.html`, `web/static/style.css`, `web/templates/layouts/base.html` (JS), `cmd/miniport/main.go`, `internal/handler/prune.go`

### Template structure
- 7 columns: Container | Image | Status | CPU | Memory | Port/Trend | Actions
- Grid: `.table-head, .container-row { display: grid; grid-template-columns: 160px 180px 85px 85px 85px 115px 130px; }`
- Container cell: `.cname` + `.cage` (Docker status string)
- Memory cell: raw MB via `formatMB` helper
- Port/Trend cell: `.port-badge` above sparkline SVG
- Stopped rows: `.container-row.stopped { opacity: 0.55; }`

### New Go template helper
- Add `FormatMB(b uint64) string` to `internal/handler/prune.go`
- Register as `"formatMB": handler.FormatMB` in `cmd/miniport/main.go`

### JS sort/filter (simpler path)
- Keep `id="container-table"` on wrapper div
- `sortTable()`: change to `container.querySelectorAll('.container-row')` and `.col-h` headers
- Three line changes total, not a full rewrite
- Update `sortTable(N)` indices for 7-column layout

---

## Phase 6: Action Buttons

**Files:** `web/templates/partials/container-table.html`, `web/templates/partials/service-table.html`, `web/static/style.css`

- 6 text buttons → Stop/Start icon (■/▶), Restart icon (↺), Logs text, `<details>` ⋯ dropdown
- Dropdown: Stats, Inspect, `<hr>`, Remove (danger)
- CSS-only via `<details>`/`<summary>`
- Service table: same icon pattern (no dropdown needed, fewer actions)
- Unify Prune nav dropdown with same `.action-dropdown` CSS

---

## Phase 7: Log Syntax Coloring

**Files:** `internal/handler/logs.go`, `internal/handler/service_logs.go`, `web/templates/partials/logs-panel.html`, `web/static/style.css`

### Go-side colorizer
- **Package-level compiled regexes** (not per-call): `var reTimestamp = regexp.MustCompile(...)`
- `colorizeLogLine(line string) string` — HTML-escape FIRST, then regex wrap
- Patterns: timestamp → `.log-ts`, key= → `.log-key`, value → `.log-val`, highlights → `.log-hl`, error lines → `.log-err-line`
- Both `Logs()` and `ServiceLogs()` handlers: split, colorize, pass as `template.HTML`

### JS change (critical)
- `el.insertAdjacentText('beforeend', text)` → `el.insertAdjacentHTML('beforeend', html)`
- Safe because colorizer escapes before wrapping in spans

### CSS classes
- `.log-ts`, `.log-key`, `.log-val`, `.log-hl`, `.log-err-line`

---

## Phase 8: Responsive Breakpoint

**File:** `web/static/style.css`

- Single `@media (max-width: 900px)` block
- Host strip: 3 columns
- Table head: hidden, rows become stacked flex cards
- Hide sparklines on mobile
- Mobile pseudo-labels via `[data-label]::before`

---

## Phase 9: Service Table Consistency

**File:** `web/templates/partials/service-table.html`

- Keep `<table>` structure
- Update colors to CSS vars
- Match visual style to container grid
- Apply stopped opacity for inactive services
- Icon action buttons matching container pattern

---

## Phase 10: Cleanup / Dead Code Removal

### Delete
- `web/templates/partials/summary-strip.html`
- All orphaned utility classes from style.css (`.px-2`, `.gap-4`, `.flex`, `.items-center`, etc.)
- `.filter-tab`, `.filter-tab-active` (replaced by `.tab`)

### Keep
- `.stat-bar`, `.stat-fill` if still used by service table inline metrics
- `.toast-*`, `.inspect-tab-*`, `.log-content`, `.sparkline`
- `.search-input`, `.log-select`, `.log-highlight`

### CSS Size Budget: ~250 lines max
- `:root` vars: ~35 lines
- Reset + base: ~10 lines
- Layout components: ~60 lines
- Action buttons + dropdown: ~40 lines
- Log panel + coloring: ~20 lines
- Toast + inspect + sparkline: ~40 lines
- Responsive: ~20 lines

---

## Execution Order

```
1. CSS custom properties (Phase 1)     — foundation, zero breakage
2. Nav bar (Phase 2)                   — tiny, independent
3. Host strip (Phase 3)                — independent
4. Summary + filter merge (Phase 4)    — depends on 1
5. Container table grid (Phase 5)      — biggest change, depends on 1+4
6. Action buttons (Phase 6)            — same template as 5
7. Log colorizer (Phase 7)             — independent of 2-6
8. Responsive breakpoint (Phase 8)     — depends on 3+5
9. Service table (Phase 9)             — depends on 1
10. Cleanup (Phase 10)                 — last, after verification
```

---

## Critical Files

| File | Changes |
|------|---------|
| `web/static/style.css` | CSS vars, component classes, responsive breakpoint |
| `web/templates/partials/container-table.html` | Complete rewrite: grid, merged columns, actions |
| `web/templates/partials/host-strip.html` | Complete rewrite: 5-card grid |
| `web/templates/partials/summary-strip.html` | **DELETE** |
| `web/templates/layouts/base.html` | Nav redesign, JS sort/filter updates |
| `web/templates/partials/logs-panel.html` | `insertAdjacentText` → `insertAdjacentHTML` |
| `web/templates/partials/service-table.html` | Visual consistency |
| `internal/handler/logs.go` | `colorizeLogLine`, colorize before render |
| `internal/handler/service_logs.go` | Same colorize treatment |
| `internal/handler/prune.go` | Add `FormatMB` helper |
| `cmd/miniport/main.go` | Register `formatMB` in funcMap |

---

## Verification

1. `go build ./cmd/miniport` — compiles
2. `go test ./...` — passes
3. Visual check with `make dev`:
   - Host strip: 5 metric cards with progress bars
   - Container table: 7 columns, grid layout, stopped rows muted
   - Actions: icon buttons + ⋯ dropdown
   - Memory column: raw MB values
   - Port badge + sparkline merged
   - Log panel: colorized key=value
   - Streaming logs: colorized HTML
   - Search/filter/sort working
   - Below 900px: stacked cards, hidden sparklines
   - Service table: matching visual style
   - Toasts + Prune dropdown working
