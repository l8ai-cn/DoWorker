# Do Worker Design System

Design-as-code using [`pastel`](https://github.com/sharkdp/pastel). Source `.pastel` files are the source of truth; rendered PNGs under `out/` are artifacts.

## Philosophy

**Notion + GitHub hybrid**: warm neutrals, contained primary, clear borders, controlled motion, tight typography (14px body).

- No hard black/white — use muted neutrals
- No rainbow — primary is GitHub blue, status has one clear color each
- No noisy hover — subtle background shifts, not `opacity/90`
- Dense but breathable — 14px body, 16/20/24 spacing, generous page padding

## Layout

```
design/
├── tokens/               # 跨端共享：颜色/间距/圆角/字体/阴影
└── out/                  # 渲染输出（git-ignored）
```

## Quick commands

```bash
pastel check  <file>            # validate
pastel build  <file> -o out.png # render
pastel fmt    <file>            # format
pastel serve  <file>            # live preview
```

Batch render tokens:
```bash
# Shared tokens
for f in design/tokens/*.pastel; do
  out="design/out/tokens/$(basename ${f%.pastel}).png"
  pastel build "$f" -o "$out"
done
```

## Tokens reference

### Colors (Light)

| Token | Value | Use |
|-------|-------|-----|
| `background` | `#FFFFFF` | page background |
| `card` | `#FFFFFF` | elevated surface |
| `muted` | `#F6F8FA` | hover/inset background, subtle fills |
| `subtle` | `#F0F3F6` | row stripes, ambient |
| `foreground` | `#1F2328` | primary text |
| `muted_foreground` | `#656D76` | secondary text, meta |
| `primary` | `#0969DA` | CTAs, active state, links |
| `primary_hover` | `#0860C8` | primary hover (explicit, not `opacity/90`) |
| `accent` | `#DDF4FF` | blue-tinted hover background |
| `success` | `#1A7F37` | success indicator |
| `warning` | `#9A6700` | warning indicator |
| `danger` | `#CF222E` | destructive, error |
| `border` | `#D0D7DE` | default border (contrast ≥ 3:1) |
| `border_strong` | `#AFB8C1` | emphasized border |

### Colors (Dark)

Same keys, tuned for `#0D1117` background: `foreground=#C9D1D9`, `primary=#2F81F7`, `border=#30363D`, `danger=#F85149`, etc.

### Spacing — 8pt ladder

```
xs=4  sm=8  md=12  lg=16  xl=24  2xl=32  3xl=48
```

**Forbidden**: gap-1 (4 alone), gap-5 (20), gap-7 (28). Use semantic aliases:
- `space.inline=8` — icon+label
- `space.stack=12` — stacked blocks
- `space.card_p=16` — card inner padding
- `space.section=24` — between sections
- `space.page_y=24`, `space.page_x=24` — page edges

### Radius

```
sm=4  md=6  lg=8  full=999
```

No `xl` / `2xl` — too bubbly for an IDE-like app.

### Typography

- Sizes: `12 / 13 / 14 / 16 / 20` (only 5 tiers)
- Weights: `normal (400) / medium (500) / semibold (600)` — no bold
- Line height: body `20`, heading `26`

Semantic: `caption`, `body`, `body_m`, `label`, `title`, `page` → see `tokens/typography.pastel`.

### Shadows

```
xs = [0, 1, 2, #0000000D]    # subtle elevation
sm = [0, 2, 4, #00000014]    # hover cards, dropdowns
md = [0, 4, 8, #0000001F]    # dialogs, popovers
```

No `lg` / `xl` shadows in-app.

## Icons

- Default: `16×16` (`w-4 h-4`)
- Emphasis: `20×20` (`w-5 h-5`)
- Status dot: `8×8` (`w-2 h-2`)

Use `lucide-react` exclusively. No custom SVG unless logo.

## Phase roadmap

- **0.1** ✅ Tokens + 5 base components (this delivery)
- **0.2** Shell, key pages (Tickets/Workspace/Settings)
- **0.3** Loading/empty/error/disconnected states
- **A–D** Land into code: CSS variables → React components → page refactor → interaction polish

## Decisions worth recording

- **Primary = `#0969DA` not `#0ea5e9`** — GitHub blue sits more neutrally against grays; sky-500 felt cold and marketing-y.
- **Body size = 14px not 16px** — lets dashboards fit more, matches Notion/GitHub density. Still accessible (line-height 20px).
- **No `bold` weight** — semibold (600) is enough for hierarchy; bold (700) looks heavy at 14px body.
- **Border `#D0D7DE` not `#E5E5E5`** — previous border was too light, cards blended into background. New border passes 3:1 contrast.
- **Hover uses explicit color not opacity** — `primary_hover` token provides a concrete value; `opacity/90` at `#0969DA` is visually ambiguous.
- **No count badges on tabs / filters unless actionable** — `[Mine 5]` and `[Completed 23]` look informative but the number alone doesn't help the user decide anything; it's noise. Show counts only when they drive a decision (e.g. "Inbox (3 unread)" where unread count matters). Pod list counts do not.
- **No description text on list-page PageHeaders** — phrases like "Track work items, assign agents, and ship." are onboarding material, not everyday UI. Users who reach `/tickets` already know what it is. Keep title + actions only. Detail-page PageHeaders keep their subtitle **only when it's ticket-specific meta** (e.g. "Opened 3 days ago by Dev User · 2 pods linked").
