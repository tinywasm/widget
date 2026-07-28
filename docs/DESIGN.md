# Design decisions — `tinywasm/widget`

Justifies the technical decisions behind [ARCHITECTURE.md](ARCHITECTURE.md) and
records the alternatives that were rejected. Does not restate the architecture,
and does not specify behaviour — that is [SPECS.md](SPECS.md).

Read this when a decision looks arbitrary and you need to know whether it can be
changed.

---

## 1. Why `tinywasm/css` stays

**Decision.** Keep the dependency, and narrow this module's use of it to one rule:
*reference tokens, never invent values.*

The question is fair, because the dependency looks thinner than it is. Only three
things are used from `css`, and one of them is nearly free:

1. The **token catalog** — `ColorPrimary`, `Space4`, `RadiusMd`, `DurationFast` —
   resolved through `.Var()`.
2. **`css.Token`**, the type.
3. **`css.Stylesheet`** as an opaque container. The typed DSL inside `css`
   (`rule`, `decl`, `media`, `keyframes`) is unexported, so the engine builds its
   whole sheet as a string and wraps it in `css.Raw`.

Point 3 contributes a return type and nothing else. It would be tempting to
conclude the dependency is ceremonial. It is not, for two reasons:

**Someone must declare the variables.** This module emits *references*:
`var(--color-primary, #1E6B9E)`. A reference with no declaration falls back to a
literal, which means no theming and no dark mode. `css.RootCSS()` and
`css.Theme(css.Set(…))` are what make the fallback a fallback rather than the
value.

**The contrast guarantee cannot live here.** `css` carries a contrast test across
the palette. That test is precisely what makes `On(Accent)` safe to hand to
someone who cannot evaluate contrast themselves — the entire premise of this
library. A guarantee about values belongs with the values; duplicating the palette
here to avoid a dependency would move the guarantee somewhere it cannot be
maintained.

**Rejected: vendoring the tokens.** Removes the dependency, duplicates the
palette, and splits the contrast guarantee across two repositories that will
drift. The drift is not hypothetical — see §2.

**Rejected: exporting the `css` DSL and building sheets with it.** Larger API in
`css`, no benefit here: the engine's output is fully determined by the sheet, and
a string builder emits it deterministically in one pass. Revisit only if `css`
needs the DSL for its own reasons.

**Consequence.** `css.Stylesheet` stays as the return type even though the DSL is
unused, because it is the integration contract `ssr` matches on.

---

## 2. Why the boundary needs enforcing, not just stating

The rule "never invent a value" was already implied by the architecture and is
already violated in four places. That is the evidence that a stated rule is not
enough:

- Eighteen tokens with hardcoded hex live in the style package instead of `css`,
  outside the reach of the contrast test.
- Two surfaces return literal `rgba(0,0,0,…)` washes, which are invisible on a
  dark surface — broken in dark mode by construction, and invisible to a guard
  that only inspects `var()` calls.
- One overlay declaration emits a `var()` with no fallback, drifting from the
  catalog.
- Stacking is hardcoded to integers while `css` publishes a `--z-*` scale, so a
  sticky element can render above a modal.

**Decision.** The drift guard becomes the enforcement mechanism: every `var()` is
compared against the catalog *including its fallback*, the sheet is asserted to
contain no colour literals, and the fixture must exercise **every** option — the
current fixture does not, which is how the overlay drift got in.

**Rejected: a lint rule.** Slower to write, easier to disable, and it cannot see
the emitted output, which is where the invariant actually matters.

---

## 3. Why visibility is one declaration, not a pair

**Decision.** A single option binds hiding to a state; the engine restores the
display mode the element's flow requires.

The pair form has two independent faults. The ordering rule ("the reveal goes in a
state rule, never in the base") lives only in a comment and nothing checks it. And
the reveal emitted a fixed `display`, which overrides the flow primitive's own
`display` from an earlier cascade layer — so revealing a row turned it into a
block, stacked its children, and killed its gap. The Go read correctly and the
symptom existed only in the revealed state.

**Rejected: `display: revert-layer`.** This is the obvious fix and it does not
work here. The base `display: none` sits in the `widgets` layer, so reverting from
the `states` layer rolls back exactly to that `none` and the element stays hidden.
Making it work would mean moving the base declaration out of `widgets`, which
complicates emission for no gain over resolving the value directly.

**Rejected: `[hidden]` in markup.** Moves a styling concern into the DOM and gives
up the state-attribute model the rest of the library uses.

---

## 4. Why an invalid sheet panics

**Decision.** Validation returns all problems; emission panics if any exist.

The failure this prevents is a rule attached to a part that does not exist —
typically a typo. It emits CSS that matches nothing, so the page is simply
unstyled in one spot, with no error anywhere. An author who does not already know
the library has no way to find it.

A returned `error` does not fix this. The code path that builds a sheet is
initialisation-shaped, the error would be handled by ignoring it, and the failure
would stay silent. A panic naming the offending part is actionable in seconds.

This mirrors `regexp.MustCompile`: a malformed pattern known at authoring time is
a programming error, not a runtime condition.

**Rejected: returning `(*css.Stylesheet, error)`.** Changes every call site into
ceremony to guard against a condition that is never recoverable.

**Rejected: logging a warning.** Warnings in a build that stays green are how the
original problem shipped.

---

## 5. Why a surface carries shape

**Decision.** A surface resolves background, text, border, radius and padding
together, not colour alone.

Splitting them means an author who has already said "this is a Panel" is then
asked to choose a radius and a padding — two decisions they have no basis for, on
a thing whose identity already implies both. In practice every panel in a codebase
gets the same radius, chosen by copy-paste, until one does not.

Folding shape into the surface removes the two most frequently repeated options
from the common path. They survive as explicit overrides for the cases that
genuinely differ.

**Rejected: presets like `Card()` or `Toolbar()`.** More names, less
composability, and the set is never complete — every project needs the ninth one.

---

## 6. Why interaction states are derived, not chosen

**Decision.** One option applies a surface and derives its hover, focus and press
treatments.

The alternative — a constant per state per family — produced a large flat
namespace where nothing prevented pairing one family's base with another family's
hover. That combination compiles, emits, and looks like a bug rather than a
mistake. It is also the exact decision the library claims to have removed from the
author.

**Consequence.** The per-state constants become private. They are still needed
internally, but they are no longer a menu the author can pick wrongly from.

**Consequence.** The cue mechanism survives as an escape hatch for the rare
browser-owned state that is not part of an interaction family, not as a normal
tool.

---

## 7. Why stacking derives from `Kind`

**Decision.** Overlay stacking is looked up from the widget's pattern, mapped onto
the token scale. The author never names a level.

Z-index is the property non-specialists get wrong most reliably, because
correctness is global: the right value depends on every other overlay in the
application. No local decision can be correct in isolation, so asking for one
guarantees eventual breakage — which is what the hardcoded integers produced.

A pattern, by contrast, determines a level unambiguously: a dialog is above a
dropdown, a toast is above both. The author already declares the pattern.

**Consequence.** This is also the answer to "`Kind` is required but nothing reads
it." Either it earns its place this way, or it and the capability interfaces
should be deleted. An API that demands data it ignores teaches people not to trust
its other demands.

**Rejected: an explicit level scale.** Smaller change, keeps the decision with the
author, and keeps the failure mode.

---

## 8. Why the scales were resized

**Decision.** Scales are sized to what the token catalog can actually distinguish,
and each scale means exactly one thing.

The spacing scale advertised thirteen steps and resolved to six distinct values;
adjacent steps produced identical output. An author who changes a step and sees no
change concludes the library is broken, and they are not wrong.

One scale served two different meanings depending on the primitive consuming it —
a column share in one, an aspect ratio in another — so the same constant produced
a square where its name promised a half. Two meanings need two types.

One option changed meaning depending on which other option accompanied it. That is
undiscoverable by construction: the signature cannot express it and the name does
not hint at it. The primitive now takes the value it needs directly.

**Rejected: keeping the numeric spacing scale and documenting the duplicates.**
Documentation cannot fix a scale whose steps are indistinguishable at the point of
use.

---

## 9. Why the utility selectors are dropped

**Decision.** Stop emitting `.fl-*` / `.exc-*` selectors alongside widget
selectors.

`Class` has no public constructor, so no consumer can put those class names in
markup. The selectors are unreachable. They were also emitted once per sheet, and
`ssr` concatenates sheets, so the dead weight multiplied by the number of
components.

---

## 10. Why deduplication belongs to `ssr`, not here

**Decision.** This module minimises what a single sheet repeats; cross-sheet
merging is `ssr`'s responsibility.

A sheet is built from one widget's declarations and cannot know how many other
widgets exist, so it cannot decide what is redundant. `ssr` merges all sheets and
is the only layer that can see the duplication. Attempting it here would require
global state across sheet construction, which would break both determinism and the
zero-value provider contract.

---

## 11. Why one breaking release

**Decision.** Ship every change together rather than phasing it.

The module has no published tag, so the cost of breaking is at its lifetime
minimum. Phasing would mean designing several signatures twice — once
compatibly, once correctly — and would leave consumers migrating in two hops for
no benefit. The known consumers are inside the same suite.

---

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure these decisions produce.
- [SPECS.md](SPECS.md) — the exact behaviour they specify.
- [MIGRATION.md](MIGRATION.md) — what changes for a consumer.
