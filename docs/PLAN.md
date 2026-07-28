# PLAN — one breaking release: a closed, self-describing API a junior cannot misuse

## Development Rules

Copied from the house documentation standard, plus the constraints this module
already commits to. Every step below is bound by these.

- **Documentation first.** Update the docs before writing code.
- **WASM boundary.** `github.com/tinywasm/widget` travels inside the WASM binary:
  it stays identity-only, with zero style logic. `widget/style` keeps its
  `//go:build !wasm` tag and must never appear in a WASM dependency graph.
  `consumer_test.go` already asserts this — keep that assertion.
- **Zero escape.** No free strings, no `vw`/`vh`, no arbitrary values, no
  `!important`. If it is not on a closed scale, it does not exist.
- **Never invent a value.** Every emitted value is a `tinywasm/css` token
  reference. `widget` decides *which* token applies; it never decides *what the
  token is worth*.
- **Deterministic output.** Emitting twice is byte-for-byte identical.
- **SSR contract.** Style builders are called by `tinywasm/ssr` on a **zero-value
  receiver**, through a method literally named `RenderCSS()`. Nothing in a style
  builder may depend on runtime data.
- **Ephemeral document.** This file is not indexed by `README.md`. Rationale and
  rejected alternatives belong in `DESIGN.md`; contracts belong in
  `ARCHITECTURE.md`.

---

## 1. Goal

Someone who does **not** know design must be able to build a correct, accessible
widget without reading the library's source. The design principle driving every
decision below:

> **Less is more.** CSS is untyped and easy to get wrong precisely because its API
> is enormous, when almost nothing is needed to actually build. The Go API in
> front of it must be small, semantic, self-describing, and closed — so the wrong
> thing cannot be expressed, and the right thing takes one call.

Two measurable targets:

| Metric | Today | Target |
|---|---|---|
| Public identifiers in `widget/style` | ~150 | ~90 |
| `Opt` calls to style the reference widget (`consumer_test.go`) | 13 | 9 |
| `Opt` calls to style one interactive part | 4 | 1 |

This lands as **one breaking release**. The module has no published tag, so
breaking now costs the least it ever will, and phasing it would mean designing
the same signatures twice.

---

## 2. Findings

All verified by running the library at commit `b291d31`, not assumed.

### 2.1 `Shown()` destroys the layout of the element it reveals — a defect

```go
Part("actions", Row(Space2), Hidden()).
When(widget.Open, "actions", Shown())
```
```css
@layer primitives { .bar__actions, .fl-row { display: flex; gap: var(--gap); … } }
@layer widgets    { .bar__actions { display: none; } }
@layer states     { .bar__actions[data-open="true"] { display: block; } }  /* wins */
```

Once open, the row is no longer a row: children stack vertically and `gap` stops
applying. This hits **every** flow primitive (`Stack`, `Row`, `Split`, `Grid`,
`Cover`, `Reel`, `Frame`). The Go reads perfectly and the symptom exists only in
the open state, so a junior has nowhere to look.

The `Hidden()`/`Shown()` pair also carries an ordering rule that lives only in a
comment ("`Shown` goes in a state rule, never in the base") and nothing checks it.

### 2.2 Hardcoded z-index breaks stacking against the token scale

`emit.go` emits `z-index: 100` for `Backdrop()` and `101` for `Above()`. Meanwhile
`tinywasm/css` publishes a full scale that goes unused:

```
--z-base 0   --z-dropdown 100   --z-sticky 200   --z-modal 300   --z-toast 400   --z-tooltip 500
```

So a modal's backdrop sits at 100 and its panel at 101, while a sticky header sits
at 200: **the header renders on top of an open modal**. That is a live bug, and it
is also the boundary breach — `widget` chose a value instead of referencing a
token.

### 2.3 User mistakes produce no error

A typo in a part name compiles, runs, and emits dead CSS:

```go
Part("item", On(Muted)).
When(widget.Selected, "itm", On(Selected))   // typo
```
```css
@layer states { .list__itm[data-selected="true"] { … } }   /* matches nothing */
```

There is no validation of any kind. `Scrim()` without `Backdrop()` is silently
inert; so is a declared part that emits nothing.

### 2.4 The API still demands design judgement

`Surface` exposes 37 constants in families of four, and pairing them is manual and
unchecked:

```go
Part("cta", On(Accent)).
Cue(widget.Hover, "cta", On(DangerHover))   // button turns red on hover
```

That compiles and emits exactly that — the very decision the library claimed to
have taken away from the junior.

Alongside it, two accessibility gaps: the focus cue emits `:focus` instead of
`:focus-visible` (drawing the ring on mouse click, which is what everyone then
removes incorrectly), and `Kind` is mandatory on the `Widget` interface but
**nothing reads it** — no `role`, no validation, nothing. Same for
`Selectable`/`Dismissible`/`Expandable`: declared, never consumed.

### 2.5 Scales that lie

`Space` advertises 13 steps and delivers 6 distinct values:

```
Space4 → --space-4        Space7, Space8   → --space-8
Space5, Space6 → --space-6    Space9…Space12 → --space-12
```

`Ratio` means two different things depending on where it is used. In `Split` it is
a column share and the names fit; in `Frame` it is an aspect ratio and they do not:

```go
Frame(RatioHalf) → aspect-ratio: 1/1    // a square, not a half
```

And `Width()` changes meaning based on what accompanies it:

```go
Root(Width(Half))            → width: 50%
Root(Center(), Width(Half))  → --max-width: 50%
```

### 2.6 Name collisions and a hole in "zero escape"

From the authors' own consumer test:

```go
When(widget.Selected, "item", style.On(style.Selected))
```

`widget.Selected` is a `State`; `style.Selected` is a `Surface`. Same for
`widget.Disabled` vs `style.Disabled`. And `style.Overlay` is an `Elevation` step
unrelated to `overlay.go` — the source comment already explains `Backdrop` is
named that way only because `Overlay` was taken.

`Sheet` and `Rule` expose every field, including `FlowType string`. This compiles:

```go
sh.PartRules["p"] = Rule{HasFlow: true, FlowType: "wobble", HasPad: true, Pad: Space2}
```

Editor autocomplete offers it to the junior as a legitimate way to use the library.

### 2.7 Every sheet ships dead CSS, once per widget

`emitPrimitive` always prepends the utility selector:

```css
.alpha, .fl-stack { display: flex; flex-direction: column; min-height: 0; }
```

But `Class` has no public constructor — a consumer **cannot** put `fl-stack` in
markup. Those `.fl-*` and `.exc-*` selectors are unreachable, and `ssr`
concatenates sheets, so they repeat once per widget. Two trivial widgets already
produce four `.fl-stack` occurrences. Empty `@layer` blocks are emitted too.

### 2.8 No entry point

8-line README with no code block. `example/main.go` prints `"button"` to the
console: it builds no sheet, generates no CSS, shows no markup. No `doc.go` in
either package, no `Example` functions, and nowhere is it explained how the
`*css.Stylesheet` reaches the browser. Comments are Spanish in `widget.go`,
`kind.go`, `state.go` and all of `style/`, English in `field.go` and
`ARCHITECTURE.md`.

---

## 3. Relationship with `tinywasm/css` — and whether it is still needed

**Short answer: yes, and the dependency should get *stronger*, not weaker.**

### What widget actually uses from css

Three things, and only three:

1. **The token catalog** — `css.ColorPrimary`, `css.Space4`, `css.RadiusMd`,
   `css.DurationFast`… `widget` calls `.Var()` on them to emit
   `var(--color-primary,#1E6B9E)`.
2. **`css.Token`** — the type, reused to declare 18 tokens of widget's own.
3. **`css.Stylesheet` as an opaque container.** Look at the last line of
   `emit.go`: widget builds the entire stylesheet as a **string** with
   `fmt.Builder`, then wraps it in `css.NewStylesheet(css.Raw(…))`. The typed DSL
   inside `css` (`rule`, `decl`, `media`, `keyframes`) is unexported anyway, so
   widget could not use it even if it wanted to.

So point 3 contributes nothing but a return type, and point 2 is a boundary
breach (see below). **The dependency's whole value is the token catalog** — and
that value is real and non-negotiable.

### Why the token catalog cannot be dropped

Widget emits *references*: `var(--color-primary, …)`. Something must **declare**
`--color-primary` in `:root`, and switch it for dark mode. That is exactly what
`css.RootCSS()`, `css.RenderCSS()` and `css.Theme(css.Set(…))` do. If widget
dropped `css`, it would have to inline literal colors and would lose theming and
dark mode outright.

There is a second reason that goes to the heart of "you should not need to know
design": `css` carries `contrast_test.go`, which verifies contrast ratios across
the palette. That test is what makes `On(Accent)` *safe* to hand to a junior. It
belongs in the library that owns the values, and it can only live there.

So the division of responsibility is:

| Library | Owns | Never does |
|---|---|---|
| `tinywasm/css` | **Values** — what a colour, space, duration, z-level *is*; light/dark switching; contrast guarantees | Know anything about components |
| `tinywasm/widget` | **Decisions** — which token applies to which part in which state | Invent a value |
| `tinywasm/ssr` | **Delivery** — collect the sheets actually used, order and dedupe them | Know what a widget is |

### Where widget breaks that boundary today

Four places, all verified:

1. **18 tokens with hardcoded hex live in `surface.go`** —
   `ColorPrimaryHover = Token{"--color-primary-hover", "#0096bc"}` and friends.
   These are colour decisions sitting in the wrong library, outside the reach of
   `css`'s contrast test.
2. **`MutedHover` returns the literal `rgba(0,0,0,0.05)`** (and `MutedPress`
   `rgba(0,0,0,0.1)`). Not a token at all — and a 5% black wash is invisible on a
   dark surface, so it is broken in dark mode. The drift guard cannot catch it
   because it is not a `var()` call.
3. **`Scrim` emits `var(--color-surface)` with no fallback**, drifting from the
   catalog. `consumer_test.go` checks every `var()` against the catalog including
   its fallback — but its fixture widget never calls `Scrim()`, so the check has a
   hole and something is already slipping through it.
4. **`z-index: 100` / `101` hardcoded** while the `--z-*` scale goes unused
   (finding 2.2).

Plus `trackValue` returns hardcoded `15rem`/`25rem`/`40rem`, for which no token
exists at all.

### Decisions

- **D1.** Keep the `tinywasm/css` dependency. Narrow widget's use of it to one
  rule: *reference tokens, never invent values.*
- **D2.** Move the 18 interaction tokens from `style/surface.go` into
  `tinywasm/css`, where they get contrast-tested with the rest of the palette.
  This requires a `css` release; it is the only cross-repo prerequisite.
- **D3.** Replace the `rgba(0,0,0,…)` literals with real tokens (D2 covers them).
- **D4.** Derive z-index from the `--z-*` scale, keyed by `Kind` (see §4.6).
- **D5.** Add `--track-sm/md/lg` to `css`, or accept these three as geometry
  rather than theme and document the exception. Recommend the former for
  consistency.
- **D6.** Keep returning `*css.Stylesheet`. It is what `ssr` expects from
  `RenderCSS()`, so the container is the integration contract even though the DSL
  is unused.

---

## 4. Target API

### 4.1 `widget` — `Kind` starts earning its place

```go
func (k Kind) Role() fmt.KeyValue   // the ARIA role of the pattern
func (k Kind) Layer() Elevation     // the stacking level of the pattern
func (k Kind) Allows(s State) bool  // which states are meaningful here
```

Same shape as the existing `State.Attr()`. This is what lets the junior declare
"this is a Dialog" **once** and get the role, the stacking level and the
validation for free — instead of choosing three things they have no basis to
choose.

If the team decides not to implement these, then `Kind` and the three capability
interfaces must be **deleted**. An API that demands data it ignores teaches
people not to trust it.

### 4.2 Scales — semantic, no repeated steps, no double meanings

```go
type Space uint8
const (
    SpaceNone Space = iota // --space-0   0
    SpaceXs                // --space-1   0.25rem
    SpaceSm                // --space-2   0.5rem
    SpaceMd                // --space-3   0.75rem
    SpaceLg                // --space-4   1rem
    SpaceXl                // --space-6   1.5rem
    Space2xl               // --space-8   2rem
    Space3xl               // --space-12  3rem
)
```

Eight steps, eight distinct values, same naming convention as `Radius`,
`TextSize` and `Elevation`. Not a number picked by feel: these are exactly the
eight `--space-*` tokens `tinywasm/css@v0.3.0` publishes, so the scale stops
inventing steps the token system does not have.

`Ratio` splits, because it is already two things:

```go
type SplitRatio uint8   // column share, for Split
const (SplitHalf, SplitTwoThirds, SplitThreeQuarters)

type Aspect uint8       // aspect ratio, for Frame
const (AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9)
```

`Elevation.Overlay` → `Elevation.Popover`, freeing the name and dropping the false
kinship with `Backdrop`. `Flat`, `Raised` and `Floating` stay: they are semantic
and say more to a junior than an `ElevationMd` would.

### 4.3 `Surface` — 37 constants down to 10, and it carries shape

```go
type Surface uint8
const (
    Page      // application background
    Panel     // card, panel
    Sunken    // well: input field
    Accent    // primary action
    Secondary // secondary action
    Highlight // selected item        (was Selected)
    Success
    Danger
    Muted     // secondary text, no background
    Dimmed    // disabled             (was Disabled)
)
```

The 27 `*Hover`/`*Focus`/`*Press` variants become **private**. `Highlight` and
`Dimmed` clear the collision with `widget.Selected` and `widget.Disabled`.

**A surface now carries its shape, not just its colours.** A "Panel" is a complete
visual decision — background, text, border, radius, padding — not a colour triplet
that leaves the junior to guess a radius. This is the single biggest boilerplate
cut: `Round()` and `Pad()` disappear from the common path and survive only as
overrides for the rare case.

Styling something clickable becomes one call:

```go
// On applies a static surface.
func On(s Surface) Opt

// Interactive applies s and derives its hover, focus and press states.
func Interactive(s Surface) Opt
```

```go
Part("cta", Interactive(Accent))   // replaces four calls, and cannot be mispaired
```

`Sheet.Cue()` survives as an escape hatch for `Target`, not as a normal tool.

### 4.4 Flow — no hidden meanings

```go
func Stack(gap Space) Opt
func Row(gap Space) Opt
func Split(r SplitRatio, gap Space) Opt
func Grid(min Track, gap Space) Opt
func Center(max Size) Opt      // takes its cap; no longer reads it from Width()
func Cover() Opt
func Reel(gap Space) Opt
func Frame(a Aspect) Opt       // takes a proportion, not a column share
```

With `Center(Size)`, `Width(Size)` recovers a single meaning.

### 4.5 Visibility — `Hidden`/`Shown` are deleted

```go
// RevealedBy hides the element and shows it when the widget owns st,
// restoring the display its own flow requires.
func RevealedBy(st widget.State) Opt
```

```go
Part("actions", Row(SpaceSm), RevealedBy(widget.Open))
```

One call instead of a pair split across two rules. The ordering rule disappears,
and so does finding 2.1: the sheet knows the part's `FlowType` and emits the right
`display`.

| part's flow | `display` the state rule emits |
|---|---|
| `Stack`, `Row`, `Reel`, `Frame` | `flex` |
| `Split`, `Grid`, `Cover` | `grid` |
| `Center` or no flow | `block` |

> Implementation warning: the obvious fix — emitting `display: revert-layer` —
> **does not work**. The base `display: none` lives in the `widgets` layer, so
> `revert-layer` from `states` rolls back exactly to that `none`. Resolve the
> value from `FlowType`.

### 4.6 Stacking — the junior stops choosing z-index

`Above()` is deleted. `Backdrop(Scope)` and the panel above it take their level
from the widget's `Kind`, mapped onto the `--z-*` token scale:

| `Kind` | token |
|---|---|
| `Menu`, `Combobox`, `Disclosure` | `--z-dropdown` |
| `Dialog` | `--z-modal` |
| `Alert` | `--z-toast` |
| everything else | `--z-base` |

A junior who declared `WidgetKind() = Dialog` gets correct stacking with no
further decision, and the sticky-header-over-modal bug becomes unexpressible.

### 4.7 Remaining options

```go
func Pad(Space) Opt          // override; rarely needed once Surface carries shape
func Round(Radius) Opt       // override
func Raise(Elevation) Opt    // override
func Width(Size) Opt
func FontSize(TextSize) Opt  // was Text(): a matched pair with FontWeight
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

### 4.8 `Sheet` — closed and validated

```go
func Of(n widget.Name) *Sheet
func (s *Sheet) Root(opts ...Opt) *Sheet
func (s *Sheet) Part(p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Opt) *Sheet

// Validate returns ALL problems with the sheet, not just the first.
func (s *Sheet) Validate() []error

// Stylesheet emits the CSS. It panics if Validate() finds anything: a
// malformed sheet is a programming error, not a runtime condition — the
// same call the standard library makes with regexp.MustCompile.
func (s *Sheet) Stylesheet() *css.Stylesheet
```

`Rule`, `Triplet`, `stateKey`, `cueKey` and every `Sheet` field become private, and
`FlowType` becomes an enum. The only way to build a sheet is `Of` + `Opt`, which is
what the architecture already claimed.

`Validate()` catches at minimum:

- `When`/`Cue` on a part never declared with `Part()` — the typo in 2.3
- a declared part that produces no declarations
- `Scrim()` without `Backdrop()`
- a `State` that `Kind.Allows` rejects for this widget

The panic is deliberate. A junior handed a returned `error` ignores it; a junior
handed a panic naming the misspelled part fixes it in ten seconds.

### 4.9 Focus and dead CSS

- `Cue(widget.Focus, …)` emits `:focus-visible` instead of `:focus`.
- `emitPrimitive` stops prepending the unreachable `.fl-*` / `.exc-*` selectors
  (finding 2.7), and empty `@layer` blocks are omitted.

### 4.10 What this looks like in practice

The reference widget from `consumer_test.go`, before:

```go
style.Of(m.WidgetName()).
    Root(style.Grid(style.TrackSm, style.Space2), style.On(style.Page), style.Scrolls()).
    Part("master", style.Stack(style.Space1), style.On(style.Panel), style.Round(style.RadiusMd)).
    Part("detail", style.Stack(style.Space2), style.On(style.Panel), style.Round(style.RadiusMd)).
    Part("item",   style.Row(style.Space1),   style.On(style.Muted)).
    When(widget.Selected, "item", style.On(style.Selected)).
    Cue(widget.Hover,     "item", style.On(style.MutedHover))
```

and after — 13 `Opt` calls down to 9, with the radius and the hover pairing gone:

```go
style.Of(m.WidgetName()).
    Root(style.Grid(style.TrackSm, style.SpaceSm), style.On(style.Page), style.Scrolls()).
    Part("master", style.Stack(style.SpaceXs), style.On(style.Panel)).
    Part("detail", style.Stack(style.SpaceSm), style.On(style.Panel)).
    Part("item",   style.Row(style.SpaceXs),   style.Interactive(style.Muted)).
    When(widget.Selected, "item", style.On(style.Highlight))
```

---

## 5. The contract `tinywasm/ssr` consumes

`ssr` compiles a generated `main.go` that instantiates each package's type as a
**zero value** and calls a method matched by name via regex. For CSS that method
is `RenderCSS()`, discovered only if the package has a `css.go`. A `css.go` with
no provider is a hard error — good, and it should stay.

**Minimum boilerplate for a widget author** is therefore one file and one method.
Do not introduce a separate `Style()` method; it is pure ceremony:

```go
// css.go
package masterdetail

func (m *MasterDetail) RenderCSS() *css.Stylesheet {
    return style.Of(m.WidgetName()).
        Root(style.Grid(style.TrackSm, style.SpaceSm), style.On(style.Page)).
        Part("item", style.Row(style.SpaceXs), style.Interactive(style.Muted)).
        Stylesheet()
}
```

Two consequences bind this plan:

- **Zero-value safety.** `RenderCSS()` runs on `&T{}`. No style builder may read a
  field. Worth an explicit `Example` and a test.
- **Sheets are concatenated.** `ssr` does `merged.Render += out.Render`, so
  anything a sheet repeats is repeated once per widget in the shipped CSS. That is
  what makes finding 2.7 worth fixing here rather than in `ssr`.

`tinywasm/ssr` gets its own `docs/PLAN.md` covering the delivery side: layer
statement placement, deduplication, and ordering guarantees.

---

## 6. Migration

| Before | After |
|---|---|
| `Space0…Space12` | `SpaceNone, SpaceXs, SpaceSm, SpaceMd, SpaceLg, SpaceXl, Space2xl, Space3xl` |
| `Split(RatioTwoThirds, g)` | `Split(SplitTwoThirds, g)` |
| `Frame(RatioHalf)` | `Frame(AspectSquare)` |
| `Center()` + `Width(s)` | `Center(s)` |
| `Center()` alone | `Center(Prose)` |
| `Raise(Overlay)` | `Raise(Popover)` |
| `On(Selected)` | `On(Highlight)` |
| `On(Disabled)` | `On(Dimmed)` |
| `On(X)` + three `Cue(…, On(X*))` | `Interactive(X)` |
| `On(Panel)` + `Round(RadiusMd)` + `Pad(s)` | `On(Panel)` |
| `Text(TextSm)` | `FontSize(TextSm)` |
| `Hidden()` + `When(st, p, Shown())` | `RevealedBy(st)` |
| `Above()` | deleted — derived from `Kind` |
| `Rule{…}` literal | does not exist: `Of` + `Opt` only |

Known consumers to notify: `tinywasm/form` and `tinywasm/components/fieldset`.
`NameField` and its parts do not change, but any `RenderCSS()` they have must be
rewritten. Also worth checking before starting: `tinywasm/layout` exists as a
separate module and may overlap with the flow primitives here — if it does, that
overlap should be resolved in this same release rather than after it.

---

## 7. Implementation order

One deliverable; the order is dependency, not risk.

1. **`tinywasm/css` release** carrying the 18 interaction tokens (D2), real tokens
   for the `rgba()` literals (D3), and optionally `--track-*` (D5). Cross-repo
   prerequisite — start here.
2. **Scales and renames**: `Space`, `SplitRatio`/`Aspect`, `Popover`,
   `Highlight`/`Dimmed`, `FontSize`. Mechanical, and unblocks everything else.
3. **`Surface` to 10 constants**, carrying shape, with the interaction variants
   private and derived; `Interactive()` added.
4. **`RevealedBy()`**, deleting `Hidden`/`Shown` and resolving `display` from
   `FlowType`. Closes finding 2.1.
5. **`Center(Size)`**, and `Width` with one meaning.
6. **`:focus-visible`** in `cuePseudo`; drop the dead `.fl-*`/`.exc-*` selectors
   and empty layer blocks.
7. **`Kind.Role()`, `Kind.Layer()`, `Kind.Allows()`** in the root package; wire
   `Backdrop` to the `--z-*` scale and delete `Above()`.
8. **Close the API**: `Rule` and the `Sheet` fields private, `FlowType` an enum.
   After 2–7, so the emitter is not rewritten twice.
9. **`Validate()`** and the panic in `Stylesheet()`, resting on 7 and 8.
10. **Documentation**: `GUIDE.md`, `doc.go` ×2, `Example` functions,
    `example/main.go`, README, and one language throughout.
11. **Update `ARCHITECTURE.md`**, which today describes scales that cease to exist.

---

## 8. Test strategy

Each of these maps to a finding, so a regression is a named failure.

| Test | Asserts |
|---|---|
| `TestRevealedByKeepsFlow` | a `Row` with `RevealedBy(Open)` still emits `display: flex` in the state rule — finding 2.1 |
| `TestStackingFromKind` | a `Dialog` backdrop emits `var(--z-modal)`, a `Menu` emits `var(--z-dropdown)`; no literal `z-index` integer appears — 2.2 |
| `TestValidateUndeclaredPart` | `When` on an unknown part is reported by `Validate()` and panics in `Stylesheet()` — 2.3 |
| `TestValidateScrimWithoutBackdrop` | reported — 2.3 |
| `TestInteractiveDerivesFamily` | `Interactive(Accent)` emits hover/focus/press from the `Accent` family and nothing else — 2.4 |
| `TestFocusVisible` | the focus cue emits `:focus-visible`, never bare `:focus` — 2.4 |
| `TestSpaceStepsDistinct` | every `Space` step resolves to a different token — 2.5 |
| `TestNoInventedValues` | **extend the existing drift guard**: every `var()` matches the catalog *including its fallback*, and the sheet contains no `rgba(`, no `#`, and no bare `rem`/`px` outside the known geometry exceptions. Run it over a fixture exercising **every** `Opt`, closing the hole that let `Scrim` drift — 2.6, §3 |
| `TestNoUnreachableSelectors` | no emitted selector starts with `.fl-` or `.exc-`; no empty `@layer` block — 2.7 |
| `TestZeroValueReceiver` | `(&T{}).RenderCSS()` works without touching a field — §5 |
| existing WASM guard | `GOOS=js go list -deps` on the example still excludes `widget/style` — keep as is |
| existing determinism check | two emissions byte-identical — keep as is |

The drift guard extension is the highest-value item in this table: it is the test
that mechanically enforces "never invent a value", which is the whole of the
`css`↔`widget` boundary.

---

## 9. Acceptance criterion

"Is it intuitive for a junior?" has an objective answer. Hand `GUIDE.md` to
someone who does not know design and ask for a collapsible panel containing a
selectable list, with no help and without opening the library's source.

Today they fail: the layout breaks on open (2.1), a typo tells them nothing
(2.3), and they have no basis for pairing surfaces (2.4).

After this release they should succeed — and if they make the typo, the library
tells them which part, by name.
