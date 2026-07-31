# Migration — closed-API release

Mapping from the published API to the one specified in [SPECS.md](SPECS.md). One
breaking release; there is no compatibility period and no aliases. Reasoning for
that choice is in [DESIGN.md §11](DESIGN.md#11-why-one-breaking-release).

---

## 1. Prerequisite: `tinywasm/css`

This release depends on a `css` release first. Nothing here compiles until it
lands.

| Addition to `css` | Why |
|---|---|
| `--color-<family>-hover|focus|press` for every surface family | These tokens have been replaced by dynamic, programmatic state derivations (`css.Hover()`, `css.Focus()`, and `css.Press()`) to ensure contrast-tested safety. |
| A real token behind the `Subtle` interaction states | They are literal `rgba(0,0,0,0.05)` washes today, invisible on a dark surface |
| `--column-narrow`, `--column-medium`, `--column-wide` | `Grid` column minimums are hardcoded `rem` values with no token |
| `--max-w-readable`, replacing `--max-w-prose` | `Size.Readable` must mirror the token it emits; leaving the token named `prose` reintroduces the translation step |
| `--color-focus-ring` | The focus ring currently reuses `--color-primary`, which is invisible on the `Primary` surface itself |

---

## 2. Renames

Reasoning for the naming pass is in [DESIGN.md §12](DESIGN.md#12-naming). It is
grouped below by *why* the name changed, because the four groups carry very
different risk: the first two are mechanical, the third is a correctness fix, and
the fourth changes which constant you should be reaching for.

### 2.1 Truncations — names now mirror the token they emit

| Before | After | Emits |
|---|---|---|
| `Opt` | `Option` | — |
| `Space0 … Space12` | `SpaceNone, Space1, Space2, Space3, Space4, Space6, Space8, Space12` | `--space-N` |
| `Track` | `ColumnWidth` | — |
| `TrackSm, TrackMd, TrackLg` | `ColumnNarrow, ColumnMedium, ColumnWide` | `--column-narrow/medium/wide` |

`Radius*` and `Text*` are unchanged: they already mirror `--radius-*` and
`--text-*`. If those token names are themselves unreadable, that is a
`tinywasm/css` question — renaming only on the Go side would reintroduce the
translation step this rule removes.

**Spacing is not a 1:1 remap.** The old scale advertised 13 steps and resolved to
6 values, so adjacent constants were identical. Pick by intent:

| Before | Resolved to | Now |
|---|---|---|
| `Space0` | `0` | `SpaceNone` |
| `Space1` | `--space-1` | `Space1` |
| `Space2` | `--space-2` | `Space2` |
| `Space3` | `--space-3` | `Space3` |
| `Space4` | `--space-4` | `Space4` |
| `Space5`, `Space6` | `--space-6` | `Space6` |
| `Space7`, `Space8` | `--space-8` | `Space8` |
| `Space9` … `Space12` | `--space-12` | `Space12` |

The gaps are deliberate: there is no `Space5` because there is no `--space-5`.

### 2.2 Specialist vocabulary — replaced with what the thing does

| Before | After | Why the old name was opaque |
|---|---|---|
| `Reel(gap)` | `ScrollRow(gap)` | *Every Layout* term |
| `Cover()` | `FillCentered()` | *Every Layout* term |
| `Frame(r)` | `MediaBox(a)` | *Every Layout* term; its child rules target `img`/`video` |
| `Scrim()` | `Veil()` | stage-lighting term |
| `Size.Prose` | `Size.Readable` | typography term for a readable line length |
| `Surface.Sunken` | `Surface.Inset` | design term |
| `Flush()` | `EdgeToEdge()` | design term |
| `Clip()` | `HideOverflow()` | states the mechanism instead of a metaphor |
| `Of(name)` | `For(widget)` | a preposition stating no relation |
| `On(surface)` | `As(surface)` | a preposition stating no relation |

`Of` → `For` and `On` → `As` are the weakest entries; `On` is also the most-typed
function in the API, so this is the largest single source of churn. See
[DESIGN.md §12.7](DESIGN.md#127-the-two-closest-calls).

### 2.3 Names that read as something else — mandatory

| Before | After | Why |
|---|---|---|
| `Fixed()` | `KeepSize()` | It emits `flex-shrink:0; flex-grow:0`, but `position:fixed` means something entirely different. A reader who knows CSS reads the old name **exactly wrong**. |

Also normalised for grammar, so every option reads as an instruction:
`Scrolls()` → `Scroll()`.

### 2.4 Surfaces — renamed for role, not appearance

| Before | After | Why |
|---|---|---|
| `Selected` | `Highlight` | collided with `widget.Selected`, a `State` |
| `Disabled` | `Inactive` | collided with `widget.Disabled`; also mirrors `--color-disabled` |
| `Muted` | `Subtle` | `Muted` and `Dimmed` both meant "faded"; nothing said which was which |
| `Accent` | `Primary` | it resolves to `--color-primary`; the old name added a translation step |

The collision this ends produced lines like
`When(widget.Selected, "item", style.On(style.Selected))`, where the two
`Selected`s were different types meaning different things.

### 2.5 Other renames

| Before | After |
|---|---|
| `Ratio` in `Split` | `SplitRatio` — `SplitHalf, SplitTwoThirds, SplitThreeQuarters` |
| `Ratio` in `Frame` | `Aspect` — `AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9` |
| `Elevation.Overlay` | `Elevation.Popover` |
| `Text(TextSm)` | `FontSize(TextSm)` |

**`Frame` ratios did not mean what they said:**

| Before | Emitted | Now |
|---|---|---|
| `Frame(RatioHalf)` | `aspect-ratio: 1/1` (a square) | `MediaBox(AspectSquare)` |
| `Frame(RatioTwoThirds)` | `aspect-ratio: 3/2` | `MediaBox(Aspect3x2)` |
| `Frame(RatioThreeQuarters)` | `aspect-ratio: 4/3` | `MediaBox(Aspect4x3)` |

If the intent was a wide media box, `Aspect16x9` is new and probably what was
wanted.

---

## 3. Collapsed calls

| Before | After |
|---|---|
| `On(X)` + `Cue(Hover, p, On(XHover))` + `Cue(Focus, …)` + `Cue(Press, …)` | `Interactive(X)` |
| `On(Panel)` + `Round(RadiusMd)` | `As(Panel)` |
| `Hidden()` + `When(st, p, Shown())` | `RevealedBy(st)` |
| `Center()` + `Width(s)` | `Center(s)` |
| `Center()` alone | `Center(Readable)` |

`Round` and `Raise` still exist as overrides for the case that genuinely differs
from the surface default. `Pad` is **not** folded into surfaces and stays an
explicit choice on every part that needs it — padding follows from what a part
contains, not from what surface it wears.

---

## 4. Deletions

| Removed | Replacement |
|---|---|
| `Hidden()`, `Shown()` | `RevealedBy(state)` |
| `Above()` | derived from `Kind` — see below |
| The 27 `*Hover` / `*Focus` / `*Press` surface constants | private; use `Interactive()` |
| The 18 exported `Color*Hover/Focus/Press` tokens | moved to `tinywasm/css` |
| `Rule`, `Triplet`, and all exported `Sheet` fields | `For()` + options only |

### 4.1 Stacking

`Backdrop(Scope)` no longer emits a hardcoded `z-index: 100`, and `Above()` is
gone. The level comes from the widget's `Kind`.

Consumers that relied on the old integers should check their result: a `Dialog`
backdrop previously sat at `100` — the same level as a dropdown, and *below* a
sticky element at `--z-sticky: 200`. Anything that compensated for a sticky
element covering a modal can drop the workaround.

### 4.2 Direct `Rule` construction

Code of this shape no longer compiles, by design:

```go
sh.PartRules["p"] = style.Rule{HasFlow: true, FlowType: "wobble", HasPad: true, Pad: style.Space2}
```

Rebuild it with `For()` and options.

---

## 5. Behaviour changes without an API change

These compile unchanged and behave differently. Check them.

| Change | Effect |
|---|---|
| Revealing an element restores its flow's `display` | A `Row` or `Grid` that previously became a block when opened now stays a row or grid. **Any CSS written to compensate must be removed.** |
| Focus cue emits `:focus-visible` | The focus ring no longer appears on mouse click. Custom rules that hid it are now unnecessary and may suppress keyboard focus. |
| An invalid sheet panics | A rule pointing at a misspelled part used to emit dead CSS silently; it now fails loudly at emission. Expect to find pre-existing typos. |
| Surfaces carry a radius | Parts that relied on a surface having none now have one. Use `Round(RadiusNone)` where square is intended. Padding is unaffected — it was never folded in. |
| `Split` becomes responsive | It never collapsed before: `container-type` was set on the same selector the `@container` rule targeted, so the rule never applied. It now stacks below ~40rem of its own width. **Layouts tuned around a split that never collapsed will change.** |
| `.fl-*` / `.exc-*` selectors are no longer emitted | Unreachable before — `Class` has no public constructor — so no markup can be affected. Any hand-written CSS targeting them stops matching. |

---

## 6. Full example

Before:

```go
style.Of(m.WidgetName()).
    Root(style.Grid(style.TrackSm, style.Space2), style.On(style.Page), style.Scrolls()).
    Part("master", style.Stack(style.Space1), style.On(style.Panel), style.Round(style.RadiusMd)).
    Part("detail", style.Stack(style.Space2), style.On(style.Panel), style.Round(style.RadiusMd)).
    Part("item",   style.Row(style.Space1),   style.On(style.Muted)).
    When(widget.Selected, "item", style.On(style.Selected)).
    Cue(widget.Hover,     "item", style.On(style.MutedHover))
```

After — 13 options down to 12, and every remaining name says what it does:

```go
style.For(m).
    Root(style.Grid(style.ColumnNarrow, style.Space2), style.As(style.Page), style.Scroll()).
    Part("master", style.Stack(style.Space1), style.As(style.Panel), style.Pad(style.Space3)).
    Part("detail", style.Stack(style.Space2), style.As(style.Panel), style.Pad(style.Space3)).
    Part("item",   style.Row(style.Space1),   style.Interactive(style.Subtle)).
    When(widget.Selected, "item", style.As(style.Highlight))
```

---

## 7. Known consumers

| Module | Impact |
|---|---|
| `tinywasm/form` | `NameField` and its parts are unchanged. Any `RenderCSS()` needs rewriting. |
| `tinywasm/components/fieldset` | Same. This is the global form skin, so it exercises surfaces and states heavily. |
| `tinywasm/layout` | Verify before starting whether it overlaps the flow primitives here. If it does, resolve the overlap in this release rather than after it. |

---

## 8. Upgrading to v0.5.0 — one typed way to write a state

v0.5.0 is breaking in exactly one place: `State.Attr()` stopped returning an
`fmt.KeyValue` and returns a `StateAttr` instead, and `tinywasm/dom` gained the
only sanctioned writers. Everything else is additive. Requirement: upgrade to the
matching `dom` release (`BindState`/`BindStateFunc`/`SetState`) **before** this
release — nothing here compiles against an older `dom`.

### The mapping

Every hand-wired state write collapses to one typed call. The six-line idiom is
gone; `attr.Key`/`attr.Value` no longer exist because there is no pair to pull
apart.

| Before | After |
|---|---|
| `attr := widget.Open.Attr()`<br/>`el.BindAttrFunc(attr.Key, func() string {`<br/>&nbsp;&nbsp;`if on { return attr.Value }`<br/>&nbsp;&nbsp;`return ""`<br/>`})` | `el.BindStateFunc(widget.Open, func() bool { return on })` |
| `el.BindAttrFunc(st.Attr().Key, func() string { … })` | `el.BindStateFunc(st, func() bool { … })` |
| `el.BindAttrBool("data-selected", sig)` | `el.BindState(widget.Selected, sig)` — **this is the bug fix** |
| `el.Attr(st.Attr().Key, st.Attr().Value)` | `el.SetState(st)` (or `BindState`, if reactive) |
| `form`'s package-level `attrInvalid`/`attrLocked` vars (`states.go`, deleted) | `BindStateFunc(widget.Invalid, …)`, `BindStateFunc(widget.Locked, …)` |

### What still compiles is now wrong on purpose

`Key()`/`Value()` exist so `dom` (writing) and `widget/style` (selecting) can meet.
Any other call site that reads them to hand-wire a state is now the wrong path and
it costs more to write than the right one. Reviews should `grep -rn 'BindAttrBool("data-'`
and find nothing.

### The overlap that stays

`widget.Disabled` is `data-disabled="true"`; the HTML `disabled` attribute is
written by `BindAttrBool("disabled", …)` and is a different thing. They are not
confusable anymore: one accepts a `StateAttr`, the other a bare string.

### New primitives (additive)

| What | API |
|---|---|
| Viewport-height flex column | `Cover()` |
| Fixed rail beside fluid content | `Sidebar(side, width, gap)` |
| Edge-anchored overlay panel | `Drawer(side, size)` |
| Device-scoped override | `On(css.Device, "part", …)` |
| Device-only part | `OnlyOn(css.Device, "part", …)` |

### New Sheet methods

`StateAttrs()` returns every `[data-*="true"]` pair the emitted CSS selects on.
Any consumer using `RevealedBy()` should add a test that renders their markup and
asserts every returned pair appears in it.

### Before/after: application shell

Before (hypothetical — the old API had no cover or sidebar):

```go
style.For(app).
    Root(style.Stack(style.SpaceNone), style.As(style.Page)).
    Part("header", style.KeepSize(), style.As(style.Panel)).
    Part("body", style.Fill()).
    Part("rail", style.Width(style.Content), style.As(style.Panel)).
    Part("content", style.Fill())
```

After:

```go
style.For(app).
    Root(style.Cover(), style.As(style.Page)).
    Part("header", style.KeepSize(), style.As(style.Panel)).
    Part("body", style.Fill()).
    Part("rail", style.KeepSize(), style.As(style.Panel)).
    Part("content", style.Fill())
```

Mobile nav rail that becomes a drawer:

```go
style.For(nav).
    Root(style.As(style.Page)).
    Part("rail", style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone), style.As(style.Panel)).
    On(css.Mobile, "rail", style.Drawer(style.SideEnd, style.TwoThirds), style.RevealedBy(widget.Open))
```

---

## Related documents

- [SPECS.md](SPECS.md) — the target API in full.
- [DESIGN.md](DESIGN.md) — why each change was made.
- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure it produces.
