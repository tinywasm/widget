# PLAN — closed-API release

Execution document. Steps, reference code, test strategy. **Ephemeral**: not
indexed by `README.md`, and no permanent document links here.

Everything this plan needs is specified elsewhere. Consult those documents only
when a step is ambiguous — do not re-derive their content here:

| If you need… | Read |
|---|---|
| the exact target API, values, and emitted output | [SPECS.md](SPECS.md) |
| why a decision was made, or what was already rejected | [DESIGN.md](DESIGN.md) |
| the structure and invariants being preserved | [ARCHITECTURE.md](ARCHITECTURE.md) |
| what changes for a consumer | [MIGRATION.md](MIGRATION.md) |
| how the modules divide the problem | [diagrams/BOUNDARIES.md](diagrams/BOUNDARIES.md) |

---

## Development Rules

- **Documentation first.** The docs above are already written to the target. Code
  follows them; if code and SPECS disagree, SPECS is wrong or the code is — decide
  and update SPECS in the same commit.
- **WASM boundary.** `widget` stays identity-only and imports only
  `tinywasm/fmt` — never `tinywasm/css`. `widget/style` keeps `//go:build !wasm`.
  The existing `go list -deps` assertion stays.
- **Zero escape.** No free strings, no `vw`/`vh`, no arbitrary values, no
  `!important`.
- **Never invent a value.** Every emitted value is a `tinywasm/css` token
  reference. Geometry exceptions are enumerated in SPECS §2.2 and nowhere else.
- **Deterministic output.** Two emissions are byte-identical.
- **Zero-value providers.** `RenderCSS()` runs on `&T{}` and must not read fields.
- **Single release.** No aliases, no deprecation shims, no compatibility period.

---

## 1. Goal

Someone who does not know design builds a correct, accessible widget without
reading this library's source.

| Metric | Before | Target |
|---|---|---|
| Public identifiers in `widget/style` | ~150 | ~90 |
| Options to style the reference widget | 13 | 9 |
| Options to style one interactive part | 4 | 1 |

---

## 2. Defects this release closes

Each was reproduced by executing the library at `b291d31`. Keep these reproducers
— they become the regression tests in §5.

**D-1 — revealing an element destroys its layout.** A part with a flow primitive
plus the hide/show pair emits `display:block` in the `states` layer, overriding
the primitive's `display:flex` from `primitives`. A revealed `Row` stacks
vertically and loses its gap. Affects every flow primitive.

**D-2 — stacking is hardcoded.** `z-index: 100` / `101` are emitted while `css`
publishes `--z-base` through `--z-tooltip`. A sticky element at `--z-sticky: 200`
renders above an open modal.

**D-3 — mistakes are silent.** A rule naming a misspelled part emits CSS that
matches nothing. No validation exists.

**D-4 — surfaces can be mispaired.** A base surface from one family and a hover
from another compiles and emits.

**D-5 — the boundary leaks.** Eighteen tokens with hardcoded hex live here; two
surfaces emit literal `rgba()` washes broken in dark mode; one overlay declaration
emits a `var()` with no fallback and the drift guard misses it because the fixture
never exercises that option.

**D-6 — scales lie.** Spacing offers 13 steps resolving to 6 values; one ratio
type means two different things; one option changes meaning depending on its
neighbours.

**D-7 — dead CSS ships per widget.** Unreachable `.fl-*` / `.exc-*` selectors are
emitted in every sheet, and `ssr` concatenates sheets.

**D-8 — no entry point.** No package docs, no runnable examples, an example
program that prints a class name, and mixed comment languages.

---

## 3. Prerequisite: `tinywasm/css` release

Nothing below compiles until this lands. Details in
[MIGRATION.md §1](MIGRATION.md#1-prerequisite-tinywasmcss).

- Interaction tokens for every surface family, contrast-tested alongside the rest
  of the palette.
- A real token behind the `Muted` interaction states, replacing the `rgba()`
  washes.
- `--track-sm`, `--track-md`, `--track-lg`.

---

## 4. Implementation order

Dependency order, not risk order. Each step lists the file it lands in and the
SPECS section that defines it.

**Step 1 — scales and renames.** `scale.go`, mechanical, unblocks everything.
SPECS §2. Split `Ratio` into `SplitRatio` and `Aspect`; resize `Space` to eight
steps; `Overlay`→`Popover`.

**Step 2 — surfaces.** `surface.go`. SPECS §3. Ten constants; interaction variants
unexported; a surface resolves radius and padding alongside colour; delete the
eighteen local tokens in favour of the `css` ones from step 3 of the prerequisite.

```go
// Interactive applies s and derives its hover, focus and press treatments.
func Interactive(s Surface) Opt {
    return func(r *rule) { r.hasSurface, r.surface, r.interactive = true, s, true }
}
```

**Step 3 — `RevealedBy`.** `overlay.go` + `emit.go`. SPECS §5.1. Delete the
hide/show pair. Closes **D-1**.

The emitter resolves `display` from the part's own flow — it must **not** emit
`revert-layer`, which resolves back to the base `display:none`. See
[DESIGN.md §3](DESIGN.md#3-why-visibility-is-one-declaration-not-a-pair).

```go
func displayFor(f flowType) string {
    switch f {
    case flowStack, flowRow, flowReel, flowFrame:
        return "flex"
    case flowSplit, flowGrid, flowCover:
        return "grid"
    default:
        return "block"
    }
}
```

**Step 4 — `Center(Size)`.** `flow.go`. `Width` recovers one meaning. SPECS §4.

**Step 5 — focus and dead CSS.** `emit.go`. `cuePseudo` emits `:focus-visible`;
`emitPrimitive` stops prepending `.fl-*` / `.exc-*`; empty layer blocks are
omitted. Closes **D-7**. SPECS §7.

**Step 6 — `Kind` earns its place.** `kind.go` + `emit.go`. SPECS §1.3.
`Role()`, `Layer()`, `Allows()`. `Layer` is an enum, not a `css.Token` — the root
package must not import `css`. Wire `Backdrop` to it and delete `Above()`. Closes
**D-2**.

If the team decides against this, delete `Kind` and the three capability
interfaces instead; leaving them inert is not an option. See
[DESIGN.md §7](DESIGN.md#7-why-stacking-derives-from-kind).

**Step 7 — close the API.** `sheet.go`. `rule`, `Triplet`, `stateKey`, `cueKey`
and every `Sheet` field unexported; `FlowType` becomes an enum. After steps 1–6 so
the emitter is not rewritten twice. SPECS §6.

**Step 8 — validation.** `sheet.go`. SPECS §6.1. Rests on steps 6 and 7. Closes
**D-3**.

```go
func (s *Sheet) Stylesheet() *css.Stylesheet {
    if errs := s.Validate(); len(errs) > 0 {
        panic(fmt.Err("widget/style:", errs))   // see DESIGN.md §4
    }
    …
}
```

**Step 9 — tests.** Everything in §5.

**Step 10 — documentation.** `GUIDE.md` (task-oriented, with the decision table
below), `doc.go` in both packages, `Example` functions for `Of`, `Stack`, `Split`,
`Interactive`, `RevealedBy` and `Backdrop`, a rewritten `example/main.go` that
builds a real sheet, and a README code block. Closes **D-8**.

Comment language: English throughout, decided in this release. `widget.go`,
`kind.go`, `state.go` and all of `style/` are Spanish today.

**Step 11 — remove the STATUS markers** from `ARCHITECTURE.md`, `SPECS.md` and
`MIGRATION.md`. They exist because those documents were written ahead of the
implementation; removing them is the last act of this plan.

---

### Decision table for `GUIDE.md`

The substitute for design judgement: the author does not choose, they look up.

| I want… | Use |
|---|---|
| a column of things | `Stack(SpaceSm)` |
| a row of buttons | `Row(SpaceXs)` |
| a grid that adapts by itself | `Grid(TrackSm, SpaceSm)` |
| list plus detail | `Split(SplitTwoThirds, SpaceMd)` |
| a centred column of text | `Center(Prose)` |
| a horizontal scrolling strip | `Reel(SpaceSm)` |
| an image with a fixed proportion | `Frame(Aspect16x9)` |
| the page background | `On(Page)` |
| a card or panel | `On(Panel)` |
| something clickable | `Interactive(Accent)` |
| something clickable, secondary | `Interactive(Secondary)` |
| the selected item of a list | `When(widget.Selected, "item", On(Highlight))` |
| secondary text | `On(Muted)` |
| an error | `On(Danger)` |
| to fill the remaining height | `Fill()` |
| to scroll internally | `Scrolls()` |
| something that expands | `RevealedBy(widget.Open)` |
| a modal dialog | `Backdrop(Viewport)` + `Scrim()` |

---

## 5. Test strategy

Every test names the defect it closes, so a regression is a named failure.

| Test | Asserts | Closes |
|---|---|---|
| `TestRevealedByKeepsFlow` | a `Row` with `RevealedBy(Open)` emits `display:flex` in the state rule; `revert-layer` appears nowhere | D-1 |
| `TestStackingFromKind` | a `Dialog` backdrop emits `var(--z-modal…)`, a `Menu` emits `var(--z-dropdown…)`; no integer `z-index` is emitted | D-2 |
| `TestValidateReportsAll` | a sheet with an undeclared part, an empty part and a `Scrim` without `Backdrop` returns three errors, not one | D-3 |
| `TestStylesheetPanicsOnInvalid` | emission panics, and the message names the offending part | D-3 |
| `TestInteractiveDerivesFamily` | `Interactive(Accent)` emits the three `Accent` states and no other family | D-4 |
| `TestFocusVisible` | the focus cue emits `:focus-visible`; bare `:focus` appears nowhere | D-4 |
| `TestNoInventedValues` | **extend the existing drift guard**: every `var()` matches the catalog including its fallback, and the output has no `#`, no `rgba(`, no `vw`/`vh`. The fixture must exercise **every** option — the current one does not, which is how the overlay drift got in | D-5 |
| `TestSpaceStepsDistinct` | no two `Space` steps resolve to the same token | D-6 |
| `TestSurfaceCarriesShape` | `On(Panel)` alone emits radius and padding; `Round(RadiusNone)` overrides it | D-6 |
| `TestNoUnreachableSelectors` | no selector begins with `.fl-` or `.exc-`; no empty `@layer` block | D-7 |
| `TestZeroValueProvider` | `(&T{}).RenderCSS()` succeeds without reading a field | — |
| existing WASM guard | `GOOS=js go list -deps` still excludes `widget/style`; `widget` does not import `css` | — |
| existing determinism check | two emissions byte-identical | — |

`TestNoInventedValues` is the highest-value item: it is the only mechanical
enforcement of the `css`↔`widget` boundary, and the four leaks in D-5 all got in
past a version of it that had gaps.

Update `consumer_test.go` to the new API; keep its layer-order and determinism
assertions unchanged.

---

## 6. Acceptance criterion

Hand `GUIDE.md` to someone who does not know design and ask for a collapsible
panel containing a selectable list — no help, no reading the source.

Today they fail three times: the layout breaks on open (D-1), a typo tells them
nothing (D-3), and they have no basis for pairing surfaces (D-4).

After this release they should succeed, and a typo should name the part.
