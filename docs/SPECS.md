# Specification — `tinywasm/widget`

Strict functional requirements: exact public surface, exact scale mappings, exact
emitted output, exact failure conditions. Structure and reasoning are not repeated
here — see [ARCHITECTURE.md](ARCHITECTURE.md) and [DESIGN.md](DESIGN.md).

Every table below is a test assertion. An implementation is correct when it
matches these tables byte for byte.

> **STATUS (remove this note when the closed-API release lands):** this document
> specifies the target. See [MIGRATION.md](MIGRATION.md) for the published API.

---

## 1. Package `widget` (WASM-compatible)

```go
type Name string
type Part string
type Class string   // no public constructor

func (n Name) Root() Class
func (n Name) Class(p Part) Class
func (c Class) String() string
func (c Class) AsAttr() fmt.KeyValue

type Kind uint8
const (Region, Listbox, Menu, Dialog, Disclosure, Tabs, Toolbar, Grid, Combobox, Form, Alert)
func (k Kind) Role() fmt.KeyValue
func (k Kind) Layer() Layer        // stacking level of the pattern
func (k Kind) Allows(s State) bool

// Layer is the stacking level of a pattern. It is an enum, not a css.Token:
// this package travels inside the WASM binary and must not import css.
// widget/style maps it to the --z-* catalog.
type Layer uint8
const (LayerBase, LayerDropdown, LayerSticky, LayerModal, LayerToast, LayerTooltip)

type State uint8
const (Selected, Disabled, Locked, Invalid, Busy, Open, Current)
func (s State) Attr() fmt.KeyValue

type Cue uint8
const (Hover, Focus, Press, Target)

type Widget interface {
    WidgetName() Name
    WidgetKind() Kind
}

type Selectable  interface{ Select(id string) }
type Dismissible interface{ Dismiss() }
type Expandable  interface{ Expand(open bool) }
```

### 1.1 Class derivation

| Call | Result |
|---|---|
| `Name("targetlist").Root()` | `targetlist` |
| `Name("targetlist").Class("row")` | `targetlist__row` |

### 1.2 State attributes

| `State` | attribute |
|---|---|
| `Selected` | `data-selected="true"` |
| `Disabled` | `data-disabled="true"` |
| `Locked` | `data-locked="true"` |
| `Invalid` | `data-invalid="true"` |
| `Busy` | `data-busy="true"` |
| `Open` | `data-open="true"` |
| `Current` | `data-current="true"` |

### 1.3 `Kind` → role, stacking level, allowed states

| `Kind` | `Role()` | `Layer()` | states beyond the universal set |
|---|---|---|---|
| `Region` | `role="region"` | `LayerBase` | — |
| `Listbox` | `role="listbox"` | `LayerBase` | `Selected`, `Current` |
| `Menu` | `role="menu"` | `LayerDropdown` | `Open`, `Current` |
| `Dialog` | `role="dialog"` | `LayerModal` | `Open` |
| `Disclosure` | `role="group"` | `LayerDropdown` | `Open` |
| `Tabs` | `role="tablist"` | `LayerBase` | `Selected`, `Current` |
| `Toolbar` | `role="toolbar"` | `LayerBase` | — |
| `Grid` | `role="grid"` | `LayerBase` | `Selected`, `Current` |
| `Combobox` | `role="combobox"` | `LayerDropdown` | `Open`, `Selected`, `Invalid` |
| `Form` | `role="form"` | `LayerBase` | `Invalid`, `Busy` |
| `Alert` | `role="alert"` | `LayerToast` | `Open` |

Universal set, allowed for every `Kind`: `Disabled`, `Locked`, `Busy`.

`widget/style` maps `Layer` onto the catalog: `LayerBase`→`--z-base`,
`LayerDropdown`→`--z-dropdown`, `LayerSticky`→`--z-sticky`,
`LayerModal`→`--z-modal`, `LayerToast`→`--z-toast`,
`LayerTooltip`→`--z-tooltip`.

---

## 2. Package `widget/style` — scales

All values are `tinywasm/css` token references. No literal may be emitted for any
of these.

### 2.1 `Space` — 8 steps, 8 distinct tokens

| Constant | Token | Value |
|---|---|---|
| `SpaceNone` | `--space-0` | `0` |
| `SpaceXs` | `--space-1` | `0.25rem` |
| `SpaceSm` | `--space-2` | `0.5rem` |
| `SpaceMd` | `--space-3` | `0.75rem` |
| `SpaceLg` | `--space-4` | `1rem` |
| `SpaceXl` | `--space-6` | `1.5rem` |
| `Space2xl` | `--space-8` | `2rem` |
| `Space3xl` | `--space-12` | `3rem` |

**Invariant:** no two steps resolve to the same token.

### 2.2 Other scales

| Type | Constants | Tokens |
|---|---|---|
| `Radius` | `RadiusNone, RadiusSm, RadiusMd, RadiusLg, RadiusFull` | `0`, `--radius-sm/md/lg/full` |
| `TextSize` | `TextXs, TextSm, TextBase, TextLg, TextXl, Text2xl` | `--text-*` |
| `Weight` | `WeightRegular, WeightMedium, WeightBold` | `--font-weight-*` |
| `Elevation` | `Flat, Raised, Floating, Popover` | `none`, `--shadow-sm/md/lg` |
| `Motion` | `MotionNone, MotionFast, MotionBase, MotionSlow` | `none`, `--duration-*` + `--ease-in-out` |
| `Track` | `TrackSm, TrackMd, TrackLg` | `--track-sm/md/lg` |
| `Size` | `Content, Prose, Third, Half, TwoThirds, Full` | `max-content`, `--max-w-prose`, `33.33%`, `50%`, `66.66%`, `100%` |
| `SplitRatio` | `SplitHalf, SplitTwoThirds, SplitThreeQuarters` | `1fr`, `2fr`, `3fr` (against a trailing `1fr`) |
| `Aspect` | `AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9` | `1/1`, `3/2`, `4/3`, `16/9` |
| `Scope` | `Parent, Viewport` | `position: absolute` / `fixed` |

`Size` percentages and `Aspect` fractions are geometry, not theme, and are the
only literals the drift guard permits. `Track` requires `--track-*` to exist in
`tinywasm/css`; adding them is a prerequisite of the release.

---

## 3. Surfaces

```go
type Surface uint8
const (
    Page, Panel, Sunken, Accent, Secondary, Highlight, Success, Danger, Muted, Dimmed
)
```

A surface resolves background, text, border, radius and padding together.

| Surface | Background | Text | Border | Radius | Padding |
|---|---|---|---|---|---|
| `Page` | `--color-background` | `--color-on-surface` | — | none | none |
| `Panel` | `--color-surface` | `--color-on-surface` | `1px solid --color-outline` | `RadiusMd` | `SpaceMd` |
| `Sunken` | `--color-surface-sunken` | `--color-on-surface` | `1px solid --color-outline` | `RadiusSm` | `SpaceSm` |
| `Accent` | `--color-primary` | `--color-on-primary` | — | `RadiusSm` | `SpaceSm` |
| `Secondary` | `--color-secondary` | `--color-on-secondary` | — | `RadiusSm` | `SpaceSm` |
| `Highlight` | `--color-selection` | `--color-on-selection` | — | `RadiusSm` | `SpaceSm` |
| `Success` | `--color-success` | `--color-on-success` | — | `RadiusSm` | `SpaceSm` |
| `Danger` | `--color-error` | `--color-on-error` | — | `RadiusSm` | `SpaceSm` |
| `Muted` | `transparent` | `--color-muted` | — | none | none |
| `Dimmed` | `--color-disabled` | `--color-on-disabled` | — | `RadiusSm` | `SpaceSm` |

Explicit `Round()`, `Pad()` or `Raise()` on the same rule overrides the surface
default.

### 3.1 Interaction families

`Interactive(s)` emits the base surface plus three state rules. The per-state
tokens are **private** and follow the pattern `--color-<family>-hover|focus|press`,
declared in `tinywasm/css`.

| Cue | Selector suffix | Change |
|---|---|---|
| hover | `:hover` | background → `--color-<family>-hover` |
| focus | `:focus-visible` | background → `--color-<family>-focus`, plus `1px solid --color-primary` |
| press | `:active` | background → `--color-<family>-press` |

`Muted` is the one family whose hover/press must not be a black wash: it resolves
to `--color-hover` so it remains visible in dark mode.

**Invariant:** it is not expressible to combine one family's base with another
family's interaction state.

---

## 4. Flow primitives

| Option | Emits |
|---|---|
| `Stack(gap)` | `display:flex; flex-direction:column; min-height:0`, and `> * + * { margin-block-start: var(--gap) }` |
| `Row(gap)` | `display:flex; flex-wrap:wrap; gap:var(--gap); align-items:center` |
| `Split(r, gap)` | `container-type:inline-size; display:grid; gap:var(--gap); grid-template-columns:var(--ratio) 1fr`, collapsing to `1fr` under `@container (max-width: 40rem)` |
| `Grid(min, gap)` | `display:grid; gap:var(--gap); grid-template-columns:repeat(auto-fit, minmax(min(var(--track),100%),1fr))` |
| `Center(max)` | `margin-inline:auto; width:100%; max-width:var(--max-width)` |
| `Cover()` | `display:grid; place-items:center; min-height:100%; width:100%` |
| `Reel(gap)` | `display:flex; gap:var(--gap); overflow-x:auto; scroll-snap-type:x mandatory`, and `> * { scroll-snap-align:start; flex:0 0 auto }` |
| `Frame(a)` | `aspect-ratio:var(--ratio); overflow:hidden; display:flex; justify-content:center; align-items:center`, and `> img, > video { width:100%; height:100%; object-fit:cover }` |

No emitted selector may begin with `.fl-` or `.exc-`.

---

## 5. Remaining options

```go
func On(s Surface) Opt
func Interactive(s Surface) Opt
func RevealedBy(st widget.State) Opt
func Pad(Space) Opt
func Round(Radius) Opt
func Raise(Elevation) Opt
func Width(Size) Opt
func FontSize(TextSize) Opt
func FontWeight(Weight) Opt
func Animate(Motion) Opt
func Fill() Opt
func Scrolls() Opt
func Fixed() Opt
func Flush() Opt
func Clip() Opt
func Backdrop(Scope) Opt
func Scrim() Opt
```

| Option | Emits |
|---|---|
| `Fill()` | `height:100%; min-height:0; flex-grow:1` |
| `Scrolls()` | `overflow-y:auto` plus everything `Fill()` emits |
| `Fixed()` | `flex-shrink:0; flex-grow:0` |
| `Flush()` | `margin:0; border-radius:0` |
| `Clip()` | `overflow:hidden` |
| `Backdrop(Parent)` | `position:absolute; inset:0; z-index:var(<Kind layer>)` |
| `Backdrop(Viewport)` | `position:fixed; inset:0; z-index:var(<Kind layer>)` |
| `Scrim()` | `background-color: color-mix(in srgb, var(--color-surface,<fallback>) 60%, transparent)` |
| `Animate(m)` | `transition: all var(--duration-*) var(--ease-in-out)` |

`Scrim()` must emit the token **with its catalog fallback**. Every rule carrying
`Animate` is repeated under `@media (prefers-reduced-motion: reduce)` with
`transition: none`.

### 5.1 `RevealedBy` display resolution

Base rule emits `display:none`. The state rule emits:

| Flow declared on the same part | `display` |
|---|---|
| `Stack`, `Row`, `Reel`, `Frame` | `flex` |
| `Split`, `Grid`, `Cover` | `grid` |
| `Center`, or no flow | `block` |

**Invariant:** `display: revert-layer` is never emitted — it resolves to the base
`display:none` and leaves the element hidden.

---

## 6. Sheet API

```go
func Of(n widget.Name) *Sheet
func (s *Sheet) Root(opts ...Opt) *Sheet
func (s *Sheet) Part(p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) Validate() []error
func (s *Sheet) Stylesheet() *css.Stylesheet   // panics if Validate() is non-empty
```

`Rule`, `Triplet` and all `Sheet` fields are unexported. `FlowType` is an enum,
not a string. An empty `Part` argument to `When`/`Cue` targets the root.

### 6.1 Validation conditions

`Validate()` returns **all** matching problems, never just the first.

| Condition | Message shape |
|---|---|
| `When`/`Cue` names a part never declared with `Part()` | `sheet <name>: rule for undeclared part "<part>"` |
| A declared part produces no declarations | `sheet <name>: part "<part>" emits nothing` |
| `Scrim()` without `Backdrop()` on the same rule | `sheet <name>: Scrim() requires Backdrop()` |
| `When` uses a state `Kind.Allows` rejects | `sheet <name>: state <state> is not meaningful for kind <kind>` |

---

## 7. Emission structure

Exact order of the emitted document:

```
@layer tokens, primitives, widgets, states;

@layer primitives { … }    omitted entirely when empty
@layer widgets    { … }    omitted entirely when empty
@layer states     { … }    omitted entirely when empty

@media (prefers-reduced-motion: reduce) { … }   only if some rule carries Animate
```

Within each layer: root rule first, then parts in ascending name order; state
rules ordered by state value then part name; cue rules ordered by cue value then
part name. Declarations within a rule are sorted; selectors within a rule are
sorted.

### 7.1 Global invariants

1. Two emissions of the same sheet are byte-identical.
2. Every `var()` matches a `tinywasm/css` token **including its fallback**.
3. No `#`, no `rgba(`, no `!important`, no `vw`/`vh` anywhere in the output.
4. No selector begins with `.fl-` or `.exc-`.
5. No empty `@layer` block.
6. Every selector is `.class`, `.class[data-*="true"]` or `.class:pseudo`.

---

## 8. SSR provider contract

```go
func (w *T) RenderCSS() *css.Stylesheet
```

Called on a zero value (`&T{}`) by `tinywasm/ssr`. It must not read fields, and it
must be deterministic. See [ARCHITECTURE.md §8](ARCHITECTURE.md#8-contract-with-tinywasmssr).

---

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — structure and constraints.
- [DESIGN.md](DESIGN.md) — why these values and not others.
- [MIGRATION.md](MIGRATION.md) — mapping from the published API.
