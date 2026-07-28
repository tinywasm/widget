# Trade-offs — `tinywasm/widget`

What this architecture buys, what it costs, and what remains unsolved.

Distinct from [DESIGN.md](DESIGN.md): that document justifies decisions *taken*.
This one records the price of those decisions and the weaknesses that survive
them. Each weakness carries a proposed improvement with its justification, and an
honest note where the proposal is a judgement call rather than a fix.

Read this before extending the library, and before concluding that a limitation
you hit is a bug.

---

## Part 1 — What the architecture buys

### P-1. The largest class of CSS bugs becomes untypeable

Closed scales, no free strings, no arbitrary values. A misspelt property, a unit
mistake, a colour that fails contrast, a magic number — none can be expressed.
This is the whole point: CSS is error-prone largely because its surface is
enormous and untyped, while what is actually needed to build is small.

### P-2. Markup and stylesheet agree by construction

`Class` has no public constructor. A class name can only be derived from a `Name`
and a `Part`, so the selector in the sheet and the attribute in the markup come
from the same expression. Divergence is not a discipline problem; it is
impossible.

### P-3. Design knowledge is encoded rather than required

A surface is a complete decision, contrast is guaranteed by the token catalog's
own test, interaction states are derived, and role and stacking follow from
`Kind`. The author supplies product judgement — *this is a dialog, this row is
selected* — and never visual judgement.

### P-4. The cascade is deterministic

Fixed layer order, flat specificity, no `!important`, no descendant selectors. A
rule's effect can be predicted from the rule alone, which is what makes the
output safe to concatenate across dozens of components.

### P-5. Zero client cost

The style engine is excluded from WebAssembly by build tag, enforced by test. The
binary shipped to the browser carries identity only.

### P-6. Output is diffable and cacheable

Byte-stable emission means a stylesheet diff shows intent, and `ssr`'s
content-hash cache is sound.

### P-7. The boundary is mechanically enforced

"Never invent a value" is checked by a drift guard over the emitted CSS, not
asserted in prose. That is the difference between an architecture and an
intention — and the four leaks it has already let through are the argument for
tightening it rather than trusting it.

---

## Part 2 — What it costs, and what to do about it

### C-1. Zero escape has no sanctioned exit

**The cost.** Every real project eventually needs a value the scale does not
have. Zero escape means the author's only recourse is hand-written CSS outside
the system — which is strictly worse than a controlled escape: it is invisible to
the drift guard, sits outside the layer model, and is not themeable.

The constraint is stated as "no new values". That is not actually what the
architecture needs.

**Proposed improvement.** Restate the rule as **no *undeclared* values**, and add
one option that accepts a `css.Token` and nothing else:

```go
func Custom(prop string, t css.Token) Option
```

A one-off must therefore still be declared as a token in `tinywasm/css`, where it
is themeable, dark-mode-aware and contrast-tested. The escape stays inside the
model.

**Justification.** The property that matters is not the size of the scale — it is
that every value has a declaration someone can theme and test. A `css.Token`
parameter preserves that while removing the incentive to leave the system
entirely. A `string` parameter would not, which is why the signature takes a
token.

**Risk, stated plainly.** `prop` is a free string, so this is the one place the
zero-escape guarantee is relaxed. Constrain it to an allow-list of properties the
engine does not otherwise emit, and have the drift guard assert that `Custom` is
the only source of such declarations.

### C-2. Padding does not belong to a surface

**The cost.** [DESIGN.md §5](DESIGN.md#5-why-a-surface-carries-shape) folds
radius *and padding* into the surface. Radius is genuinely part of a panel's
identity. Padding is not: the same `Panel` is a padded card in one place and a
flush container for a table in another, so the "one option" win evaporates
exactly in the cases that differ, replaced by an override the author must
remember.

This is an over-reach in the current proposal, not an inherited defect.

**Proposed improvement.** A surface resolves **background, text, border and
radius**. Padding returns to being an explicit `Pad()`.

**Justification.** The test for folding a property into a surface is whether its
value is implied by the surface's *identity*. Two panels always want the same
radius — it is what makes them look like the same system. Two panels frequently
want different padding, because padding is a function of what the part contains,
which is a layout decision the surface cannot see. Folding in a property that
varies converts a saved call into a remembered exception, which is a net loss.

**Cost of the change.** The reference example goes from 10 options back to 12.
That is the correct number: the two `Pad()` calls are real decisions, and hiding
them was the mistake.

### C-3. Stacking derived from `Kind` cannot see composition

**The cost.** `Kind` yields one stacking level per pattern. Correct stacking is a
property of the *composition*, not the pattern: a dialog opened from within a
dialog needs to sit above it, and both resolve to `--z-modal`. Nesting the same
pattern is unrepresentable.

Worse, the containment interaction in [SPECS.md §4.1](SPECS.md#41-split-establishes-containment)
means a `Backdrop(Viewport)` inside another widget's `Split` is contained to that
split. A sheet can detect this within itself, but **cannot see across widgets** —
and cross-widget nesting is the normal case.

**Proposed improvement.** Two parts.

1. Keep `Kind` as the default, and let a widget that genuinely nests declare a
   relative bump: `Backdrop(Viewport, Above(parentKind))`, resolving to the
   parent's level plus one step. Bounded and typed, unlike a free integer.
2. For the containment problem, stop using `container-type` on `Split` when the
   sheet also declares a viewport backdrop, and document that a `Dialog` must be
   rendered from a portal-style root rather than inline. This is a *markup*
   contract, so it belongs in `ARCHITECTURE.md` and cannot be enforced by the
   emitter.

**Justification.** Role and valid states genuinely follow from the pattern and
should stay derived. Stacking does not — it depends on what contains what, which
`Kind` cannot know by construction. Deriving it removes the common error while
leaving the rare one unrepresentable, so the escape has to exist; making it
relative rather than absolute keeps the author from choosing a number.

**Judgement call.** If nested overlays are out of scope for this suite, decline
both and record the limitation instead. The cost of being wrong is a component
that cannot be built without leaving the library.

### C-4. Panicking couples a cosmetic mistake to a failed build

**The cost.** `Stylesheet()` panics on an invalid sheet, and `ssr` calls it from
a generated program at build time. A misspelt part name therefore fails a deploy,
and the panic surfaces from a program the author never wrote, with a stack that
points at generated code.

The decision is right — see [DESIGN.md §4](DESIGN.md#4-why-an-invalid-sheet-panics)
— but its blast radius is larger than that section acknowledges.

**Proposed improvement.** Keep the panic; make it diagnosable, on both sides of
the boundary.

- Every message names the sheet, the part and the option — specified in
  [SPECS.md §6.1](SPECS.md#61-validation-conditions).
- `ssr` recovers per producer and reports *which package and type* panicked,
  rather than failing the whole extraction opaquely. This is a concrete
  requirement on the other repository, not a suggestion.

**Justification.** The objection to panicking is really an objection to
undiagnosable panics. A build that fails with `sheet "targetlist": rule for
undeclared part "itm"` is a thirty-second fix; the alternative — a warning — is
how the original silent-CSS problem shipped.

### C-5. One visual change is a three-repository change

**The cost.** Adding a surface family means a token plus a contrast test in
`css`, a constant plus a resolution entry in `widget`, and a release of each in
order. What a designer thinks of as "add a colour" is a coordinated multi-repo
change, and the boundary that makes the architecture sound is what makes this
slow.

**Proposed improvement.** Generate the surface resolution table in `widget` from
the `css` catalog, with a `go:generate` step and a test asserting the generated
table is current. The `widget` side of a new family becomes mechanical and
verified rather than hand-written and forgettable.

**Justification.** The boundary is worth its cost — see
[DESIGN.md §1](DESIGN.md#1-why-tinywasmcss-stays) — so the answer is to reduce
the friction rather than remove the boundary. Generation also closes the drift
class directly: a hand-written table can disagree with the catalog, and four such
disagreements are already in the published code.

**Not proposed: merging `css` into `widget`.** It would collapse the contrast
guarantee into the component library and make `css` unusable on its own, which is
a much larger loss than the coordination cost.

### C-6. Appearance cannot depend on component data

**The cost.** Producers run on a zero value, so a component whose appearance
depends on configuration — a table with a variable column count, a chart with N
series — cannot express it in its sheet.

**Proposed improvement.** Accept the constraint, and document the sanctioned
alternative: the component writes a custom property on the element
(`style="--columns: 7"`) and the sheet consumes it via `var(--columns)`. The
sheet stays static and cacheable; the variation stays in the markup, where the
data is.

**Justification.** Static sheets are what make extraction, caching and
deduplication possible at all. The alternative — sheets parameterised at
build time — multiplies the emitted CSS by the number of configurations and
defeats `ssr`'s content-hash cache. The custom-property route costs one
declaration and keeps every guarantee.

**Limitation of the proposal.** That inline `style` attribute is a genuine hole
in "no free strings": it is written by the component, not by the engine, and
nothing checks it. Confining it to numeric custom properties keeps the hole
small, but it is a hole.

### C-7. A part can exist in markup and not in the sheet, or the reverse

**The cost.** `Validate()` catches a `When` naming an undeclared part, because
both are in the sheet. It cannot catch the more common pair: markup that
references a part the sheet never declares, or a declared part no markup ever
carries. The sheet cannot see the markup.

**Proposed improvement.** Expose the declared set:

```go
func (s *Sheet) Parts() []widget.Part
```

and provide a test helper that compares it against the parts a component's render
actually emits. Component authors already declare parts as constants — the
pattern in `field.go` — so the comparison is cheap.

**Justification.** This is the one remaining silent-CSS path after this release.
It cannot be closed by the emitter, only by a test the component owns, so the
library's job is to make that test one line rather than to attempt detection it
structurally cannot perform.

### C-8. Container queries are the only responsive mechanism

**The cost.** Reacting to the container rather than the viewport is right for
components, and it is genuinely better than media queries here. But some
decisions are legitimately viewport-scoped — a mobile navigation pattern, a print
layout — and the API has no way to express them, having removed `vw`/`vh` and
media queries entirely.

**Proposed improvement.** Leave the component API as it is, and place
viewport-level decisions in the application shell, where `css` already publishes
`--bp-sm` through `--bp-xl`. Document the split: **components respond to their
container; the shell responds to the viewport.**

**Justification.** A component that reacts to the viewport is unusable inside a
sidebar, which is exactly the bug container queries exist to prevent. Allowing
the escape at component level would reintroduce it. The breakpoint tokens already
exist and go unused, so the shell-level path needs documentation rather than new
API.

---

## Part 3 — Accepted limitations

Recorded so they are not rediscovered as bugs.

| Limitation | Why it is accepted |
|---|---|
| One token catalog for all widgets | Per-widget theming would defeat the single visual system the library exists to enforce. Scope-level overrides of `--color-*` on a subtree remain available. |
| `Kind` is a closed enum | A component fitting no ARIA pattern is almost always two components. Extending the enum is a deliberate act, which is the intent. |
| No animation beyond a transition scale | Keyframes are an open-ended language; admitting them would reopen the whole surface the closed scales exist to shut. Components needing them fall back to application CSS. |
| Height cannot be declared | Content-driven height is what makes the primitives composable. `Fill()` and `Scroll()` cover the cases that need it. |

---

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure being assessed.
- [DESIGN.md](DESIGN.md) — why each decision was taken.
- [SPECS.md](SPECS.md) — exact behaviour, including the validations referenced here.
