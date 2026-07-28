---
PLAN: "widget: delete style.Styler and close the motion gap with a typed Motion scale"
EXECUTOR: jules
STATUS: running
SESSION: 2166551128364663081
---

> This plan is dispatched via the CodeJob workflow. See skill: **agents-workflow**.

# Plan — `tinywasm/widget`: one deletion, one closed gap

Two independent changes in one commit. Both are demanded by
[CONSTRUCTION_HARNESS.md](https://github.com/tinywasm/app-releases/blob/main/docs/CONSTRUCTION_HARNESS.md):
one removes surface nobody can reach, the other closes a boundary where a consumer would
otherwise have to reach for an escape hatch.

---

## 1. Change A — delete `style.Styler`

### 1.1 Why

`Styler` was declared for exactly one caller: the `tinywasm/ssr` asset collector, which asserted
`var w twstyle.Styler = inst` inside the `main.go` it generates. Its own doc comment says so:

```go
// Styler es la capacidad "este widget tiene aspecto". La asevera el recolector SSR.
type Styler interface {
	widget.Widget
	Style() *Sheet
}
```

That collector branch has been deleted — `ssr` now discovers a component's sheet through
`RenderCSS() *css.Stylesheet`, the same entry point `tinywasm/css` and `tinywasm/layout` have
always used, and no longer imports anything from `tinywasm/widget`. See
[`ssr/docs/PLAN.md`](https://github.com/tinywasm/ssr/blob/main/docs/PLAN.md).

With that gone, `Styler` is exported surface nobody can reach through a real path. Principle 5 —
*"Export exactly what the author uses… what you cannot see, you cannot misuse"* — makes this a
deletion, not a deprecation. Leaving it exported invites a future component to implement
`Style()` again and rediscover, at runtime and in silence, that nothing collects it.

### 1.2 The deletion

File: `style/sheet.go`, lines ~121-125. Remove the doc comment **and** the interface.

Then check the file's import block: `Styler` embeds `widget.Widget`, so if
`github.com/tinywasm/widget` was imported by `style/sheet.go` **only** for that embed, remove
the import too. The builder methods take `widget.Part`, `widget.State` and `widget.Cue`, so it
almost certainly stays. Let `go build ./...` decide — do not guess.

---

## 2. Change B — close the motion gap

### 2.1 Why this is not optional

`tinywasm/layout`'s `platformd` package has a sliding panel driven by two local tokens,
`tokenSlideDur` (`0.6s`) and `tokenTransitionWait` (`0s`). `widget/style` exposes **no way to
express a transition**, so the only paths available to that consumer are a local `Token`, a
`RawRule(`, or a `Str(` — the three escape hatches the visual-contract work exists to remove.

The harness names this case exactly:

> *"A missing contract at a boundary is a defect in the library, not in the consumer."*
> *"A consumer never re-creates a missing symbol locally. If a library does not expose what you
> need, stop and report it."*
> *"An API gap always surfaces at the leaf (the application), where the agent has no authority
> to publish upstream — so it patches locally. Technical debt is then not an accident: the
> workflow guarantees it."*

So the gap is closed **here**, before `layout` migrates. Otherwise `layout`'s plan blocks on a
report that nobody is scheduled to act on, and its executor patches locally.

### 2.2 The scale already exists — do not invent one

`github.com/tinywasm/css` v0.3.0 **already owns** the motion scale and already declares it in
`RootCSS()` (`css/tokens.go:120-125`, emitted at `css/css.go:91-93`):

```go
DurationFast = Token{"--duration-fast", "150ms"}
DurationBase = Token{"--duration-base", "250ms"}
DurationSlow = Token{"--duration-slow", "400ms"}
EaseIn       = Token{"--ease-in", "cubic-bezier(0.4,0,1,1)"}
EaseOut      = Token{"--ease-out", "cubic-bezier(0,0,0.2,1)"}
EaseInOut    = Token{"--ease-in-out", "cubic-bezier(0.4,0,0.2,1)"}
```

`widget/style` is missing only the typed `Opt` that consumes them — exactly like `Pad` consumes
`css.Space*` and `Raise` consumes `css.Shadow*`. **Do not add a token to `css`.** Do not
hardcode `150ms`/`250ms`/`400ms` anywhere in this repo; reference the tokens through `.Var()`,
the way every other `*Var` helper in `emit.go` does.

### 2.3 The enum — `style/scale.go`

Append at the end of the file, matching the shape of every scale already there:

```go
// Motion es la ÚNICA escala de transición. La duración la posee css (--duration-*);
// aquí solo se elige el peldaño. Valores arbitrarios no existen en esta API.
type Motion uint8

const (
	MotionNone Motion = iota // sin transición
	MotionFast               // realce inmediato: hover, focus
	MotionBase               // cambio de estado
	MotionSlow               // entrada/salida de un panel u overlay
)
```

Prefix every constant with `Motion`. `Fast`/`Slow` unprefixed are too generic and would read
ambiguously next to `Flat`/`Full`/`Content` in autocomplete.

**The scale is closed and has exactly four steps.** There is deliberately no step for `0.6s`:
`platformd`'s current value snaps to `MotionSlow`. That is the scale doing its job — see §2.7.

### 2.4 The `Rule` field — `style/sheet.go`

Add to the `Rule` struct, next to the other `Has*`/value pairs:

```go
	HasMotion bool
	Motion    Motion
```

A field whose name matches its type is the house pattern here — `Surface Surface` already does
it.

### 2.5 The `Opt` — `style/except.go`

Add in the *"Opciones visuales basadas en escalas"* section, next to `Pad`, `Round` and `Raise`:

```go
// Animate aplica una transición según la escala de movimiento. El easing no se
// elige: siempre es --ease-in-out. Una sola forma de hacerlo.
func Animate(m Motion) Opt {
	return func(r *Rule) {
		r.HasMotion = true
		r.Motion = m
	}
}
```

**The easing is not a parameter.** Exposing `EaseIn`/`EaseOut`/`EaseInOut` as a choice would add
a decision with no right answer at the call site — principle 4, *one way to do each thing*. If a
real need for directional easing ever appears, it is a new plan, not a second argument here.

### 2.6 The emission — `style/emit.go`

**Three additions.**

**(a) The value helper.** Next to `radiusVar` / `elevationVar`, following their exact shape:

```go
func motionValue(m Motion) string {
	switch m {
	case MotionNone:
		return "none"
	case MotionFast:
		return "all " + css.DurationFast.Var() + " " + css.EaseInOut.Var()
	case MotionBase:
		return "all " + css.DurationBase.Var() + " " + css.EaseInOut.Var()
	case MotionSlow:
		return "all " + css.DurationSlow.Var() + " " + css.EaseInOut.Var()
	default:
		return "none"
	}
}
```

**(b) The declaration.** In `Rule.Decls()`, in the *"Scales"* group after the `r.HasWeight`
block:

```go
	if r.HasMotion {
		decls = append(decls, "transition: "+motionValue(r.Motion)+";")
	}
```

**(c) The reduced-motion guard.** A transition that ignores
`prefers-reduced-motion` is an accessibility defect, and it is glue every consumer would
otherwise write identically — so the library that owns motion emits it, once.

In `Stylesheet()`, collect the selectors that carry motion while the rules are already being
walked, and after the `@layer states { … }` block closes — immediately before
`return css.NewStylesheet(...)` — emit:

```go
	if len(motionSel) > 0 {
		sb.WriteString("@media (prefers-reduced-motion: reduce) {\n")
		sb.WriteString(formatRule(motionSel, []string{"transition: none;"}))
		sb.WriteString("}\n")
	}
```

**Precision notes — read all four before writing the gathering code:**

- **Do NOT hook this into the existing `collect` closure.** That closure gathers primitive
  selectors and is invoked for `s.RootRule` and `s.PartRules` **only** — it never sees
  `StateRules` or `CueRules`. Appending to `motionSel` from inside it silently misses every
  transition declared in a `When(...)` or `Cue(...)` rule, which is precisely how a sliding
  panel declares one. Test §4.4 exists to catch this.
- `motionSel` must gather the selector of **every** rule with `HasMotion` — `RootRule`,
  `PartRules`, `StateRules` and `CueRules` alike — building each selector the same way that
  rule's own emission already does: `selectorOf(...)` for root and parts, the
  `fmt.Sprintf("%s[%s=\"%s\"]", …)` state-attribute form for states, and
  `selectorOf(...) + cuePseudo(...)` for cues.
- **No explicit sort.** `formatRule` already calls `sort.Strings(selectors)` on what it
  receives, so sorting `motionSel` beforehand is dead code. Determinism is preserved either way.
- Emit **nothing** when no rule declares motion. An empty `@media` block in every sheet would
  break the byte-for-byte determinism the existing tests assert.
- The block goes **outside** `@layer states`, after its closing `}` and immediately before
  `return css.NewStylesheet(...)`. Do not nest it.

### 2.7 What this does to `platformd` — informational

`layout`'s plan previously told its executor to *stop and report* on `tokenSlideDur`. With this
change it maps instead:

| `platformd` today | After |
|---|---|
| `tokenSlideDur` = `0.6s` | `Animate(MotionSlow)` — 400ms |
| `tokenTransitionWait` = `0s` | disappears; no delay is the default |

The `0.6s` → `400ms` change is **intended**, not a rounding error to apologise for. A closed
scale means hand-picked durations stop existing; that is the whole point. `layout`'s plan has
been updated accordingly and no longer reports a gap.

---

## 3. Scope — FORBIDDEN

| Prohibition | Reason |
|---|---|
| Adding a duration or easing token to `tinywasm/css` | The scale already exists there. Adding a `--duration-slower` to preserve `0.6s` reopens exactly the "arbitrary value" hole this closes. |
| Hardcoding `150ms`/`250ms`/`400ms`/`0.6s` or a `cubic-bezier(...)` in this repo | Values live in `css`; this repo references `.Var()`. |
| Making the easing a parameter of `Animate` | Principle 4. One way to do each thing. |
| Adding keyframes, `@keyframes`, or an animation API | Out of scope. `Animate` emits a `transition`, nothing else. |
| Touching `Sheet`, its builder methods (`Root`, `Part`, `When`, `Cue`) or `Stylesheet()`'s existing output | The DSL stays. Change B **adds**; it must not alter a single byte of what an existing sheet already emits. |
| Replacing `Styler` with a `Renderer`/`CSSProvider` interface here | `*css.Stylesheet` is owned by `tinywasm/css`, not by `widget`. Declaring that contract here would invert the dependency. |
| Deprecating `Styler` instead of deleting it | There is no published consumer left. |
| Using `go test` | This repo uses `gotest`. |

---

## 4. Tests

`style/consumer_test.go` and `style/overlay_test.go` exercise the `Sheet` builder directly and
never mention `Styler`; they must keep passing **unmodified**. If either fails to compile,
change A went too far.

Add `style/motion_test.go` covering:

1. `Animate(MotionSlow)` on `Root` emits `transition: all var(--duration-slow) var(--ease-in-out);`
   and no literal `400ms`.
2. `Animate(MotionNone)` emits `transition: none;`.
3. A sheet with **no** `Animate` call contains no `prefers-reduced-motion` block and no
   `transition:` declaration at all.
4. A sheet whose motion is declared **only** inside `When(widget.Open, "", Animate(MotionSlow))`
   still emits the `prefers-reduced-motion` block, and that block's selector is the state
   selector.
5. Two consecutive `Stylesheet()` calls on equivalent sheets produce **byte-identical** output
   (extends the determinism guarantee to the new block).

---

## 5. Acceptance criteria — grep-verifiable

1. `gotest` green.
2. `grep -rn "Styler" .` → **empty**.
3. `grep -rnE '[0-9]+ms|0\.6s|cubic-bezier' --include='*.go' .` → **empty outside `*_test.go`**
   (durations come from `css` tokens, never from a literal here).
4. `grep -c "css.Duration" style/emit.go` → **3**, and
   `grep -c "css.EaseInOut" style/emit.go` → **3**. (Each `motionValue` case pairs one duration
   with the easing on a single line, so a combined `grep -n` returns three lines, not six —
   count them separately.)
5. `grep -n "func Animate(m Motion) Opt" style/except.go` → **present**.
6. `grep -n "prefers-reduced-motion" style/emit.go` → **present**.
7. `grep -rn "func (s \*Sheet) Stylesheet()" style/emit.go` → **still present**; the DSL's exit
   point is untouched.
8. `go build ./...` and `GOOS=js GOARCH=wasm go build ./...` both succeed.
9. `git diff --stat` touches only `style/sheet.go`, `style/scale.go`, `style/except.go`,
   `style/emit.go` and the new `style/motion_test.go`. Anything else is out of scope.

---

## 6. Go quality checklist (mandatory)

- Errors via `github.com/tinywasm/fmt`, never stdlib `errors`/`fmt`.
  **Anti-footgun:** `style/emit.go` imports stdlib `sort` for deterministic ordering. That file
  is `!wasm`-only emission code and the import is legitimate — do **not** "fix" it.
- Every file in `style/` keeps its `//go:build !wasm` tag.
- No repeated string literals: `"all "`, the token `.Var()` calls and `"transition: "` appear in
  exactly one place each (`motionValue` / the single `Decls()` line).
- Zero `any`, zero `map` in new API.
- Delete, do not comment out. No `// removed: …` marker.

---

## 7. Stages table

| # | Stage | Files | Gate |
|---|---|---|---|
| 0 | *(gate)* `ssr` published without the `Style()` branch | — | `grep -rn "Styler" "$(go env GOMODCACHE)"/github.com/tinywasm/ssr@*/` prints nothing for the newest version |
| 1 | Change A — delete `Styler` | `style/sheet.go` | `go build ./...` |
| 2 | Change B — `Motion` enum + `Rule` field | `style/scale.go`, `style/sheet.go` | `go build ./...` |
| 3 | Change B — `Animate` + emission + reduced-motion | `style/except.go`, `style/emit.go` | `go build ./...` |
| 4 | Tests | `style/motion_test.go` | `gotest` green |

Stage 0 is a **gate**: publishing stage 1 before `ssr` drops its branch breaks asset extraction
for every app, because the generated `main.go` still references `twstyle.Styler`. If the gate
check fails, **stop and report it**.

Stages 2-4 (change B) do **not** depend on that gate — they add API and break nothing. If the
gate blocks, ship change B alone and leave `Styler` for a follow-up; say so explicitly in the PR.

---

## 8. Downstream — informational, not this agent's work

[`tinywasm/layout`](https://github.com/tinywasm/layout/blob/main/docs/PLAN.md) consumes
`Animate(MotionSlow)` in `platformd` and gates on a `widget` version published from this plan.
Do not attempt that migration from this repo.
