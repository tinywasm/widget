# Specification — `tinywasm/widget`

Strict functional requirements: exact public surface, exact scale mappings, exact
emitted output, exact failure conditions. Structure and reasoning are not repeated
here — see [ARCHITECTURE.md](ARCHITECTURE.md) and [DESIGN.md](DESIGN.md).

Every table below is a test assertion. An implementation is correct when it
matches these tables byte for byte.

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
| `Space1` | `--space-1` | `0.25rem` |
| `Space2` | `--space-2` | `0.5rem` |
| `Space3` | `--space-3` | `0.75rem` |
| `Space4` | `--space-4` | `1rem` |
| `Space6` | `--space-6` | `1.5rem` |
| `Space8` | `--space-8` | `2rem` |
| `Space12` | `--space-12` | `3rem` |

**Invariant:** no two steps resolve to the same token.

### 2.2 Other scales

| Type | Constants | Tokens |
|---|---|---|
| `Radius` | `RadiusNone, RadiusSm, RadiusMd, RadiusLg, RadiusFull` | `0`, `--radius-sm/md/lg/full` |
| `TextSize` | `TextXs, TextSm, TextBase, TextLg, TextXl, Text2xl` | `--text-*` |
| `Weight` | `WeightRegular, WeightMedium, WeightBold` | `--font-weight-*` |
| `Elevation` | `Flat, Raised, Floating, Popover` | `none`, `--shadow-sm/md/lg` |
| `Motion` | `MotionNone, MotionFast, MotionBase, MotionSlow` | `none`, `--duration-*` + `--ease-in-out` |
| `ColumnWidth` | `ColumnNarrow, ColumnMedium, ColumnWide` | `--column-narrow/medium/wide` |
| `Size` | `Content, Readable, Third, Half, TwoThirds, Full` | `max-content`, `--max-w-readable`, `33.33%`, `50%`, `66.66%`, `100%` |
| `SplitRatio` | `SplitHalf, SplitTwoThirds, SplitThreeQuarters` | `1fr`, `2fr`, `3fr` (against a trailing `1fr`) |
| `Aspect` | `AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9` | `1/1`, `3/2`, `4/3`, `16/9` |
| `Scope` | `Parent, Viewport` | `position: absolute` / `fixed` |

`Size` percentages and `Aspect` fractions are geometry, not theme, and are the
only literals the drift guard permits. `ColumnWidth` requires `--column-*`, and `Readable` requires `--max-w-readable`
(today `--max-w-prose`); both are prerequisites of the release, because a scale
step is named after the token it emits and the two must not drift apart.

---

## 3. Surfaces

```go
type Surface uint8
const (
    Page, Panel, Inset, Primary, Secondary, Highlight, Success, Danger, Subtle, Inactive
)
```

A surface resolves background, text, border and **radius** together. It does
**not** resolve padding — reasoning in
[DESIGN.md §5](DESIGN.md#5-why-a-surface-carries-shape).

| Surface | Background | Text | Border | Radius |
|---|---|---|---|---|
| `Page` | `--color-background` | `--color-on-surface` | — | none |
| `Panel` | `--color-surface` | `--color-on-surface` | `1px solid --color-outline` | `RadiusMd` |
| `Inset` | `--color-surface-sunken` | `--color-on-surface` | `1px solid --color-outline` | `RadiusSm` |
| `Primary` | `--color-primary` | `--color-on-primary` | — | `RadiusSm` |
| `Secondary` | `--color-surface` | `--color-on-surface` | — | `RadiusSm` |
| `Highlight` | `--color-selection` | `--color-on-selection` | — | `RadiusSm` |
| `Success` | `--color-success` | `--color-on-success` | — | `RadiusSm` |
| `Danger` | `--color-danger` | `--color-on-danger` | — | `RadiusSm` |
| `Subtle` | `transparent` | `--color-muted` | — | none |
| `Inactive` | `--color-surface` | `--color-muted` | — | `RadiusSm` |

Explicit `Round()` or `Raise()` on the same rule overrides the surface default.
`Pad()` is always explicit — there is no default to override.

### 3.1 Interaction families

`Interactive(s)` emits the base surface plus three state rules. The interactive states
are derived programmatically from the base token of the surface family using functions
`css.Hover(base)`, `css.Focus(base)`, and `css.Press(base)` defined in `tinywasm/css`.

| Cue | Selector suffix | Change |
|---|---|---|
| hover | `:hover` | background → `css.Hover(base)` |
| focus | `:focus-visible` | background → `css.Focus(base)` |
| press | `:active` | background → `css.Press(base)` |

### 3.2 Which surfaces accept `Interactive`

| Surface | `Interactive` | Why |
|---|---|---|
| `Panel`, `Inset`, `Primary`, `Secondary`, `Highlight`, `Success`, `Danger`, `Subtle` | yes | can be a control or a selectable row |
| `Page` | no | the page background is not a control |
| `Inactive` | no | it *is* the non-interactive state; deriving a hover from it is a contradiction |

`Interactive(Page)` and `Interactive(Inactive)` are reported by `Validate()`
(§6.1). This is the eight-family set the private `--color-<family>-*` tokens must
cover — 24 tokens, not 27.

**Invariant:** it is not expressible to combine one family's base with another
family's interaction state.

---

## 4. Flow primitives

| Option | Emits |
|---|---|
| `Stack(gap)` | `display:flex; flex-direction:column; min-height:0`, and `> * + * { margin-block-start: var(--gap) }` |
| `Row(gap)` | `display:flex; flex-wrap:wrap; gap:var(--gap); align-items:center` |
| `Split(r, gap)` | `display:flex; flex-wrap:wrap; gap:var(--gap)`, and `> * { flex-grow:1; flex-basis:calc((40rem - 100%) * 999) }` plus `> :first-child { flex-grow:var(--ratio) }` — stacks below ~40rem of **its own** width, no query and no wrapper element |
| `Grid(min, gap)` | `display:grid; gap:var(--gap); grid-template-columns:repeat(auto-fit, minmax(min(var(--column),100%),1fr))` |
| `Center(max)` | `margin-inline:auto; width:100%; max-width:var(--max-width)` |
| `FillCentered()` | `display:grid; place-items:center; min-height:100%; width:100%` |
| `ScrollRow(gap)` | `display:flex; gap:var(--gap); overflow-x:auto; scroll-snap-type:x mandatory`, and `> * { scroll-snap-align:start; flex:0 0 auto }` |
| `MediaBox(a)` | `aspect-ratio:var(--ratio); overflow:hidden; display:flex; justify-content:center; align-items:center`, and `> img, > video { width:100%; height:100%; object-fit:cover }` |

No emitted selector may begin with `.fl-` or `.exc-`.

### 4.1 `Split` uses no container query, deliberately

The published `Split` sets `container-type: inline-size` on the same selector the
`@container` rule targets. **An element is never its own query container**, so the
rule never applies. Measured in Chromium at a 320px viewport, well below the
40rem breakpoint:

```
self-query   (published)  grid-template-columns: 213.328px 106.656px   two columns
ancestor-query (correct)  grid-template-columns: 320px                 collapsed
```

`Split` therefore has no responsive behaviour at all today. The correct
query-based form needs a wrapper element to act as the container, which would
couple the engine to DOM structure — something §7.1 forbids.

The specified form uses intrinsic sizing instead, and needs neither. Measured
across viewports:

| viewport | result | ratio |
|---|---|---|
| 320px | stacked | — |
| 500px | stacked | — |
| 800px | side by side | 2.00 : 1 |
| 1200px | side by side | 2.00 : 1 |

**Invariant:** the emitted sheet contains no `@container` rule and no
`container-type` declaration. Nothing else in the API produces one; `Grid` is
intrinsically responsive through `auto-fit`/`minmax`.

**Correction on record.** An earlier revision claimed `container-type` makes an
element a containing block for `position: fixed` descendants, and specified a
validation rule against `Backdrop(Viewport)` inside `Split`. That is false, and
was measured false: a fixed child inside a `container-type` element covers the
full viewport (320×600 of a 320×600 viewport), identical to no containment. Only
full `contain: layout` contains it (320×0). The rule was removed rather than
shipped.

---

## 5. Remaining options

```go
func As(s Surface) Option
func Interactive(s Surface) Option
func RevealedBy(st widget.State) Option
func Pad(Space) Option
func Round(Radius) Option
func Raise(Elevation) Option
func Width(Size) Option
func FontSize(TextSize) Option
func FontWeight(Weight) Option
func Animate(Motion) Option
func Fill() Option
func Scroll() Option
func KeepSize() Option
func EdgeToEdge() Option
func HideOverflow() Option
func Backdrop(Scope) Option
func Veil() Option
```

| Option | Emits |
|---|---|
| `Fill()` | `height:100%; min-height:0; flex-grow:1` |
| `Scroll()` | `overflow-y:auto` plus everything `Fill()` emits |
| `KeepSize()` | `flex-shrink:0; flex-grow:0` |
| `EdgeToEdge()` | `margin:0; border-radius:0` |
| `HideOverflow()` | `overflow:hidden` |
| `Backdrop(Parent)` | `position:absolute; inset:0; z-index:var(<Kind layer>)` |
| `Backdrop(Viewport)` | `position:fixed; inset:0; z-index:var(<Kind layer>)` |
| `Veil()` | `background-color: color-mix(in srgb, var(--color-surface,<fallback>) 60%, transparent)` |
| `Animate(m)` | `transition: all var(--duration-*) var(--ease-in-out)` |

`Veil()` must emit the token **with its catalog fallback**. Every rule carrying
`Animate` is repeated under `@media (prefers-reduced-motion: reduce)` with
`transition: none`.

### 5.1 `RevealedBy` display resolution

Base rule emits `display:none`. The state rule emits:

| Flow declared on the same part | `display` |
|---|---|
| `Stack`, `Row`, `Split`, `ScrollRow`, `MediaBox` | `flex` |
| `Grid`, `FillCentered` | `grid` |
| `Center`, or no flow | `block` |

**Invariant:** `display: revert-layer` is never emitted — it resolves to the base
`display:none` and leaves the element hidden.

---

## 6. Sheet API

```go
func For(w widget.Widget) *Sheet   // needs Kind, not just Name — see §1.3, §5, §6.1
func (s *Sheet) Root(opts ...Option) *Sheet
func (s *Sheet) Part(p widget.Part, opts ...Option) *Sheet
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) Parts() []widget.Part   // declared parts, sorted — for component tests
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
| `Veil()` without `Backdrop()` on the same rule | `sheet <name>: Veil() requires Backdrop()` |
| `When` uses a state `Kind.Allows` rejects | `sheet <name>: state <state> is not meaningful for kind <kind>` |
| `Interactive()` on `Page` or `Inactive` (§3.2) | `sheet <name>: surface <surface> has no interaction states` |

Every message names the sheet and the part. The panic in `Stylesheet()` is the
only signal the author gets, and it surfaces from inside `ssr`'s generated
program — far from the source that caused it.

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
