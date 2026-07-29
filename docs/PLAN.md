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
| whether a limitation you hit is a bug or a known cost | [TRADEOFFS.md](TRADEOFFS.md) |
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
| Options to style the reference widget | 13 | 12 |
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

**D-9 — `Split` has no responsive behaviour at all.** It sets
`container-type: inline-size` on the same selector its `@container` rule targets,
and an element is never its own query container, so the rule never applies.
Measured in Chromium at a 320px viewport, below the 40rem breakpoint: the
published form stays at `213.328px 106.656px` — two columns — while the
ancestor-query form collapses to `320px`. SPECS §4.1 specifies a replacement that
needs no query and no wrapper element.

---

## 3. Prerequisite: `tinywasm/css` v0.3.2 migration

Executed first in [PLAN-css-v032.md](PLAN-css-v032.md). This plan assumes that
plan has completed and the code compiles against css v0.3.2.

When the css catalog was finalised, interaction tokens (Hover/Focus/Press) were
removed and replaced with inline `color-mix()` expressions. The migration plan
rewrites `familyTokens()` and the surface resolver accordingly, and syncs the
consumer test token list.

---

## 3b. Trade-off verdicts

The architecture review raised eight costs. Six are in this release, two are not.
The test applied to each: **does deferring it cost more than doing it now?**

**In this release**

| Improvement | Step | Why it cannot wait |
|---|---|---|
| Padding out of surfaces | 2 | Changes what `As(Panel)` *emits* without changing its signature. Ship it wrong and reverting later silently changes every consumer's spacing with no compile error. Behaviour breaks with no compile break are the worst thing to defer. |
| `Split` without a container query | 4 | Its premise turned out to be a live defect, D-9: the responsive collapse never fires. Replacing the mechanism fixes that and removes the hazard, rather than validating around it. |
| Diagnosable validation messages | 8 | Already specified in SPECS §6.1. Free. |
| `Sheet.Parts()` | 9 | Ten lines, and the last silent-CSS path the emitter structurally cannot see. Shipping "we closed silent CSS" while leaving the most common instance open undercuts the release's own claim. |
| Custom-property route for data-dependent appearance | 11 | Documentation only. The zero-value contract is being written down for the first time here; leaving the sanctioned alternative unstated guarantees the first person who hits it invents something worse. |
| Component-versus-shell responsiveness | 11 | Documentation only. One paragraph, and the rule is now load-bearing since no primitive emits a query at all. |

Their reasoning now lives in [DESIGN.md](DESIGN.md) and their behaviour in
[SPECS.md](SPECS.md); they are no longer listed as costs.

**Deliberately out**, with triggers, in [TRADEOFFS.md](TRADEOFFS.md):

- **A sanctioned escape (`Custom`).** Adding a function is not breaking, so the
  one-window argument does not apply. Shipping an escape before knowing which
  values are genuinely missing is how it becomes the default path, and it is the
  only proposal that weakens the core guarantee.
- **Relative stacking for nested overlays.** Real, but unevidenced in this suite.
- **Generating the surface table from the catalog.** Internal tooling, no API,
  lands any time; the drift it prevents is already caught in step 10.

One requirement crossed the boundary: recovering per producer so a panicking
sheet names its package belongs to `tinywasm/ssr`, and is E-7 in that
repository's plan.

---

## 4. Implementation order

Dependency order, not risk order. Each step lists the file it lands in and the
SPECS section that defines it.

**Step 1 — scales and the naming pass.** `scale.go`, `flow.go`, `except.go`,
`overlay.go`. Mechanical, and unblocks everything. SPECS §2.

Split `Ratio` into `SplitRatio` and `Aspect`; resize `Space` to eight steps
mirroring `--space-N`; `Overlay`→`Popover`. Apply the full rename table in
[MIGRATION.md §2](MIGRATION.md#2-renames) in the same commit — it is one
mechanical sweep, and splitting it means renaming the same identifiers twice.

Two entries in that table are not cosmetic and must not be dropped if the pass is
trimmed: `Fixed()`→`KeepSize()`, because the old name reads as `position:fixed`
and means the opposite; and `Muted`/`Dimmed`→`Subtle`/`Inactive`, because the old
pair was two words for "faded" with nothing to tell them apart. Reasoning in
[DESIGN.md §12](DESIGN.md#12-naming).

**Step 2 — surfaces.** `surface.go`. SPECS §3. Ten constants; interaction variants
unexported; a surface resolves background, text, border and **radius** — not
padding — see [DESIGN.md §5](DESIGN.md#5-why-a-surface-carries-shape); reject `Interactive` on `Page` and `Inactive`;
delete the eighteen local tokens in favour of the `css` ones from the
prerequisite.

```go
// Interactive applies s and derives its hover, focus and press treatments.
func Interactive(s Surface) Option {
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
    case flowStack, flowRow, flowScrollRow, flowMediaBox:
        return "flex"
    case flowSplit, flowGrid, flowFillCentered:
        return "grid"
    default:
        return "block"
    }
}
```

**Step 4 — `Center(Size)` and the `Split` rewrite.** `flow.go` + `emit.go`.
`Width` recovers one meaning. SPECS §4, §4.1. Closes **D-9**.

`Split` drops `container-type` and `@container` entirely for intrinsic sizing.
After this step the emitted sheet contains neither, anywhere — assert it.

```go
// Split: flex-wrap plus a flex-basis that is either huge or negative,
// so the row wraps below ~40rem of its own width. No query, no wrapper.
"display:flex;", "flex-wrap:wrap;", "gap:var(--gap);"
// > *            flex-grow:1; flex-basis:calc((40rem - 100%) * 999)
// > :first-child flex-grow:var(--ratio)
```

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
**D-3** and **D-9**.

Six conditions, not four. Two were found while specifying and are easy to miss:
`Interactive()` on `Page` or `Inactive` (SPECS §3.2), and `Backdrop(Viewport)`
under a `Split` in the same sheet (SPECS §4.1). Every message names the sheet and
the part, because the panic surfaces from inside `ssr`'s generated program.

```go
func (s *Sheet) Stylesheet() *css.Stylesheet {
    if errs := s.Validate(); len(errs) > 0 {
        panic(fmt.Err("widget/style:", errs))   // see DESIGN.md §4
    }
    …
}
```

**Step 9 — `Sheet.Parts()`.** `sheet.go`. SPECS §6. Returns the
declared parts, sorted, so a component test can compare them against the parts its
render actually emits. Ten lines, and it closes the last silent-CSS path the
emitter structurally cannot see.

**Step 10 — tests.** Everything in §5.

**Step 11 — documentation.** `GUIDE.md` (task-oriented, with the decision table
below), `doc.go` in both packages, `Example` functions for `For`, `Stack`, `Split`,
`Interactive`, `RevealedBy` and `Backdrop`, a rewritten `example/main.go` that
builds a real sheet, and a README code block. Closes **D-8**.

Comment language: English throughout, decided in this release. `widget.go`,
`kind.go`, `state.go` and all of `style/` are Spanish today.

**Step 12 — remove the STATUS markers** from `ARCHITECTURE.md`, `SPECS.md` and
`MIGRATION.md`. They exist because those documents were written ahead of the
implementation; removing them is the last act of this plan.

---

### Decision table for `GUIDE.md`

The substitute for design judgement: the author does not choose, they look up.

| I want… | Use |
|---|---|
| a column of things | `Stack(Space2)` |
| a row of buttons | `Row(Space1)` |
| a grid that adapts by itself | `Grid(ColumnNarrow, Space2)` |
| list plus detail | `Split(SplitTwoThirds, Space3)` |
| a centred column of text | `Center(Readable)` |
| a horizontal scrolling strip | `ScrollRow(Space2)` |
| an image with a fixed proportion | `MediaBox(Aspect16x9)` |
| the page background | `As(Page)` |
| a card or panel | `As(Panel)` |
| something clickable | `Interactive(Primary)` |
| something clickable, secondary | `Interactive(Secondary)` |
| the selected item of a list | `When(widget.Selected, "item", As(Highlight))` |
| secondary text | `As(Subtle)` |
| an error | `As(Danger)` |
| to fill the remaining height | `Fill()` |
| to scroll internally | `Scroll()` |
| something that expands | `RevealedBy(widget.Open)` |
| a modal dialog | `Backdrop(Viewport)` + `Veil()` |

---

## 5. Test strategy

Every test names the defect it closes, so a regression is a named failure.

| Test | Asserts | Closes |
|---|---|---|
| `TestRevealedByKeepsFlow` | a `Row` with `RevealedBy(Open)` emits `display:flex` in the state rule; `revert-layer` appears nowhere | D-1 |
| `TestStackingFromKind` | a `Dialog` backdrop emits `var(--z-modal…)`, a `Menu` emits `var(--z-dropdown…)`; no integer `z-index` is emitted | D-2 |
| `TestValidateReportsAll` | a sheet with an undeclared part, an empty part and a `Veil` without `Backdrop` returns three errors, not one | D-3 |
| `TestStylesheetPanicsOnInvalid` | emission panics, and the message names the offending part | D-3 |
| `TestInteractiveDerivesFamily` | `Interactive(Primary)` emits the three `Primary` states and no other family | D-4 |
| `TestFocusVisible` | the focus cue emits `:focus-visible`; bare `:focus` appears nowhere | D-4 |
| `TestNoInventedValues` | **extend the existing drift guard**: every `var()` matches the catalog including its fallback, and the output has no `#`, no `rgba(`, no `vw`/`vh`. The fixture must exercise **every** option — the current one does not, which is how the overlay drift got in | D-5 |
| `TestSpaceStepsDistinct` | no two `Space` steps resolve to the same token | D-6 |
| `TestSurfaceCarriesShape` | `As(Panel)` alone emits radius and padding; `Round(RadiusNone)` overrides it | D-6 |
| `TestNoUnreachableSelectors` | no selector begins with `.fl-` or `.exc-`; no empty `@layer` block | D-7 |
| `TestInteractiveRejectsNonInteractive` | `Interactive(Page)` and `Interactive(Inactive)` are reported | D-3 |
| `TestSplitCollapses` | the emitted sheet contains no `@container` and no `container-type`; a browser check at 320px stacks and at 800px gives 2:1 | D-9 |
| `TestSurfaceCarriesNoPadding` | `As(Panel)` emits `border-radius` but no `padding` | C-2 |
| `TestSheetParts` | `Parts()` returns the declared parts, sorted | C-7 |
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
