# Migration — closed-API release

Mapping from the published API to the one specified in [SPECS.md](SPECS.md). One
breaking release; there is no compatibility period and no aliases. Reasoning for
that choice is in [DESIGN.md §11](DESIGN.md#11-why-one-breaking-release).

> **STATUS (remove this note when the closed-API release lands):** the release
> described here has not shipped. Until it does, the "Before" column is current.

---

## 1. Prerequisite: `tinywasm/css`

This release depends on a `css` release first. Nothing here compiles until it
lands.

| Addition to `css` | Why |
|---|---|
| `--color-<family>-hover|focus|press` for every surface family | These tokens currently live in `widget/style` with hardcoded hex, outside the reach of the contrast test |
| A real token behind the `Muted` interaction states | They are literal `rgba(0,0,0,0.05)` washes today, invisible on a dark surface |
| `--track-sm`, `--track-md`, `--track-lg` | `Grid` column minimums are hardcoded `rem` values with no token |

---

## 2. Renames

| Before | After |
|---|---|
| `Space0 … Space12` | `SpaceNone, SpaceXs, SpaceSm, SpaceMd, SpaceLg, SpaceXl, Space2xl, Space3xl` |
| `Ratio` in `Split` | `SplitRatio` — `SplitHalf, SplitTwoThirds, SplitThreeQuarters` |
| `Ratio` in `Frame` | `Aspect` — `AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9` |
| `Elevation.Overlay` | `Elevation.Popover` |
| `Surface.Selected` | `Surface.Highlight` |
| `Surface.Disabled` | `Surface.Dimmed` |
| `Text(TextSm)` | `FontSize(TextSm)` |

`Highlight` and `Dimmed` exist to end the collision with `widget.Selected` and
`widget.Disabled`, which previously produced lines like
`When(widget.Selected, "item", style.On(style.Selected))`.

### 2.1 Spacing is not a 1:1 remap

The old scale advertised 13 steps and resolved to 6 values. Adjacent steps were
identical, so a mechanical rename is not possible — pick by intent:

| Before | Resolved to | Now |
|---|---|---|
| `Space0` | `0` | `SpaceNone` |
| `Space1` | `--space-1` | `SpaceXs` |
| `Space2` | `--space-2` | `SpaceSm` |
| `Space3` | `--space-3` | `SpaceMd` |
| `Space4` | `--space-4` | `SpaceLg` |
| `Space5`, `Space6` | `--space-6` | `SpaceXl` |
| `Space7`, `Space8` | `--space-8` | `Space2xl` |
| `Space9` … `Space12` | `--space-12` | `Space3xl` |

### 2.2 `Frame` ratios did not mean what they said

| Before | Emitted | Now |
|---|---|---|
| `Frame(RatioHalf)` | `aspect-ratio: 1/1` (a square) | `Frame(AspectSquare)` |
| `Frame(RatioTwoThirds)` | `aspect-ratio: 3/2` | `Frame(Aspect3x2)` |
| `Frame(RatioThreeQuarters)` | `aspect-ratio: 4/3` | `Frame(Aspect4x3)` |

If the intent was a wide media box, `Aspect16x9` is new and probably what was
wanted.

---

## 3. Collapsed calls

| Before | After |
|---|---|
| `On(X)` + `Cue(Hover, p, On(XHover))` + `Cue(Focus, …)` + `Cue(Press, …)` | `Interactive(X)` |
| `On(Panel)` + `Round(RadiusMd)` + `Pad(s)` | `On(Panel)` |
| `Hidden()` + `When(st, p, Shown())` | `RevealedBy(st)` |
| `Center()` + `Width(s)` | `Center(s)` |
| `Center()` alone | `Center(Prose)` |

`Round`, `Pad` and `Raise` still exist. They are now overrides for the case that
genuinely differs from the surface default, not part of the common path.

---

## 4. Deletions

| Removed | Replacement |
|---|---|
| `Hidden()`, `Shown()` | `RevealedBy(state)` |
| `Above()` | derived from `Kind` — see below |
| The 27 `*Hover` / `*Focus` / `*Press` surface constants | private; use `Interactive()` |
| The 18 exported `Color*Hover/Focus/Press` tokens | moved to `tinywasm/css` |
| `Rule`, `Triplet`, and all exported `Sheet` fields | `Of()` + options only |

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

Rebuild it with `Of()` and options.

---

## 5. Behaviour changes without an API change

These compile unchanged and behave differently. Check them.

| Change | Effect |
|---|---|
| Revealing an element restores its flow's `display` | A `Row` or `Grid` that previously became a block when opened now stays a row or grid. **Any CSS written to compensate must be removed.** |
| Focus cue emits `:focus-visible` | The focus ring no longer appears on mouse click. Custom rules that hid it are now unnecessary and may suppress keyboard focus. |
| An invalid sheet panics | A rule pointing at a misspelled part used to emit dead CSS silently; it now fails loudly at emission. Expect to find pre-existing typos. |
| Surfaces carry radius and padding | Parts that relied on a surface having neither will now have both. Use `Round(RadiusNone)` / `Pad(SpaceNone)` where flat is intended. |
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

After — 13 options down to 9:

```go
style.Of(m.WidgetName()).
    Root(style.Grid(style.TrackSm, style.SpaceSm), style.On(style.Page), style.Scrolls()).
    Part("master", style.Stack(style.SpaceXs), style.On(style.Panel)).
    Part("detail", style.Stack(style.SpaceSm), style.On(style.Panel)).
    Part("item",   style.Row(style.SpaceXs),   style.Interactive(style.Muted)).
    When(widget.Selected, "item", style.On(style.Highlight))
```

---

## 7. Known consumers

| Module | Impact |
|---|---|
| `tinywasm/form` | `NameField` and its parts are unchanged. Any `RenderCSS()` needs rewriting. |
| `tinywasm/components/fieldset` | Same. This is the global form skin, so it exercises surfaces and states heavily. |
| `tinywasm/layout` | Verify before starting whether it overlaps the flow primitives here. If it does, resolve the overlap in this release rather than after it. |

---

## Related documents

- [SPECS.md](SPECS.md) — the target API in full.
- [DESIGN.md](DESIGN.md) — why each change was made.
- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure it produces.
