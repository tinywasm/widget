//go:build !wasm

package style_test

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// MasterDetail is an example widget updated to the closed-API release
type MasterDetail struct{}

func (m *MasterDetail) WidgetName() widget.Name { return widget.Name("masterdetail") }
func (m *MasterDetail) WidgetKind() widget.Kind { return widget.Grid }

func (m *MasterDetail) Style() *style.Sheet {
	return style.For(m).
		Root(style.Grid(style.ColumnNarrow, style.Space2), style.As(style.Page), style.Scroll()).
		Part("master", style.Stack(style.Space1), style.As(style.Panel), style.Pad(style.Space3)).
		Part("detail", style.Stack(style.Space2), style.As(style.Panel), style.Pad(style.Space3)).
		Part("item", style.Row(style.Space1), style.Interactive(style.Subtle)).
		When(widget.Selected, "item", style.As(style.Highlight))
}

func (m *MasterDetail) RenderCSS() *css.Stylesheet {
	return m.Style().Stylesheet()
}

func TestConsumerMasterDetail(t *testing.T) {
	wd := &MasterDetail{}
	sheet := wd.Style()
	ss := sheet.Stylesheet()
	cssStr := ss.String()

	t.Logf("Generated CSS:\n%s", cssStr)

	// 1. All class names in stylesheet exist in the expected set
	markupClasses := map[string]bool{
		"masterdetail":         true,
		"masterdetail__master": true,
		"masterdetail__detail": true,
		"masterdetail__item":   true,
	}

	classRegex := regexp.MustCompile(`\.masterdetail[a-zA-Z0-9_\-]*`)
	matches := classRegex.FindAllString(cssStr, -1)

	sheetClasses := make(map[string]bool)
	for _, m := range matches {
		name := strings.TrimPrefix(m, ".")
		sheetClasses[name] = true
	}

	for mc := range markupClasses {
		if !sheetClasses[mc] {
			t.Errorf("Class in markup %q not found in stylesheet", mc)
		}
	}

	for sc := range sheetClasses {
		if !markupClasses[sc] {
			t.Errorf("Class in stylesheet %q not found in markup", sc)
		}
	}

	// 2. Constraints: no !important, no @media (except prefers-reduced-motion if used), no color literals, no vw/vh.
	if strings.Contains(cssStr, "!important") {
		t.Error("Stylesheet contains forbidden '!important'")
	}

	// 3. Stacking layers in the correct order
	idxTokens := strings.Index(cssStr, "tokens")
	idxPrimitives := strings.Index(cssStr, "primitives")
	idxWidgets := strings.Index(cssStr, "widgets")
	idxStates := strings.Index(cssStr, "states")

	if idxTokens == -1 || idxPrimitives == -1 || idxWidgets == -1 || idxStates == -1 {
		t.Error("Missing a required cascade layer (tokens, primitives, widgets, states)")
	} else if !(idxTokens < idxPrimitives && idxPrimitives < idxWidgets && idxWidgets < idxStates) {
		t.Errorf("Incorrect layer order: tokens (%d) < primitives (%d) < widgets (%d) < states (%d)",
			idxTokens, idxPrimitives, idxWidgets, idxStates)
	}

	// 4. Deterministic emission
	cssStr2 := wd.Style().Stylesheet().String()
	if cssStr != cssStr2 {
		t.Error("Stylesheet emission is non-deterministic")
	}

	// 5. GOOS=js build dependency exclusion
	cmd := exec.Command("go", "list", "-deps", "github.com/tinywasm/widget")
	cmd.Env = append(cmd.Environ(), "GOOS=js", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error running 'go list': %v, output: %s", err, string(out))
	}
	if strings.Contains(string(out), "github.com/tinywasm/widget/style") {
		t.Error("WASM consumer depends on 'widget/style' package, violating build tag constraints")
	}

	buildCmd := exec.Command("go", "build", "github.com/tinywasm/widget")
	buildCmd.Env = append(buildCmd.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("GOOS=js GOARCH=wasm go build github.com/tinywasm/widget failed: %v, output: %s", err, string(out))
	}
}

func TestEdgeToEdgeBeatsSurfaceRadius(t *testing.T) {
	// EdgeToEdge must square a part that also carries As(Panel): the surface's
	// default radius used to win because the primitive's border-radius: 0 sits
	// in @layer primitives and the widgets layer outranks it — the crudview
	// root measured 4px despite already carrying EdgeToEdge.
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.As(style.Panel), style.EdgeToEdge()).Stylesheet().String()

	if strings.Contains(s, "border-radius: var(--radius-md") {
		t.Errorf("EdgeToEdge must suppress the surface's default radius, got:\n%s", s)
	}
	if !strings.Contains(s, "border-radius: 0;") {
		t.Errorf("EdgeToEdge must still emit border-radius: 0, got:\n%s", s)
	}

	// An explicit Round() next to EdgeToEdge still wins: it is emitted in the
	// widgets layer, which outranks the primitives-layer 0.
	s2 := style.For(wd).Root(style.As(style.Panel), style.EdgeToEdge(), style.Round(style.RadiusMd)).Stylesheet().String()
	if !strings.Contains(s2, "border-radius: var(--radius-md") {
		t.Errorf("explicit Round must survive next to EdgeToEdge, got:\n%s", s2)
	}
}

func TestRevealedByKeepsFlow(t *testing.T) {
	// a Row with RevealedBy(Open) emits display:flex in the state rule; revert-layer appears nowhere (closes D-1)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part("rowpart", style.Row(style.Space1), style.RevealedBy(widget.Open)).
		Stylesheet().
		String()

	if strings.Contains(s, "revert-layer") {
		t.Error("revert-layer must not be emitted")
	}

	if !strings.Contains(s, ".w__rowpart[data-open=\"true\"] {\n  display: flex;\n}") {
		t.Errorf("expected state rule to emit display:flex for Row with RevealedBy, got:\n%s", s)
	}
}

func TestStackingFromKind(t *testing.T) {
	// a Dialog backdrop emits var(--z-modal…), a Menu emits var(--z-dropdown…); no integer z-index (closes D-2)
	dialog := &testWidget{name: "dlg", kind: widget.Dialog}
	dialogCSS := style.For(dialog).Root(style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(dialogCSS, "z-index: var(--z-modal,300);") {
		t.Errorf("expected Dialog backdrop to emit var(--z-modal,300), got:\n%s", dialogCSS)
	}

	// A Parent backdrop stays out of the layer: it would outrank the panel it
	// is supposed to sit behind.
	parentCSS := style.For(dialog).Part("catcher", style.Backdrop(style.Parent)).Stylesheet().String()
	if strings.Contains(parentCSS, "z-index") {
		t.Errorf("Backdrop(Parent) must not claim a stacking level, got:\n%s", parentCSS)
	}

	menu := &testWidget{name: "mnu", kind: widget.Menu}
	menuCSS := style.For(menu).Root(style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(menuCSS, "z-index: var(--z-dropdown,100);") {
		t.Errorf("expected Menu backdrop to emit var(--z-dropdown,100), got:\n%s", menuCSS)
	}

	// check that no integer z-index is emitted directly
	zindexIntRegex := regexp.MustCompile(`z-index:\s*\d+;`)
	if zindexIntRegex.MatchString(dialogCSS) {
		t.Errorf("Direct integer z-index emitted in dialog backdrop:\n%s", dialogCSS)
	}
}

func TestValidateReportsAll(t *testing.T) {
	// a sheet with an undeclared part, an empty part and a Veil without Backdrop returns three errors (closes D-3)
	wd := &testWidget{name: "w", kind: widget.Listbox} // Grid/Listbox allows Selected state
	sheet := style.For(wd).
		Part("emptypart").                                            // empty part (no options/declarations)
		When(widget.Selected, "undeclared", style.As(style.Primary)). // undeclared part
		Part("veilpart", style.Veil())                                // Veil without Backdrop

	errs := sheet.Validate()
	if len(errs) != 3 {
		t.Errorf("expected exactly 3 validation errors, got %d:\n%v", len(errs), errs)
	}
}

func TestStylesheetPanicsOnInvalid(t *testing.T) {
	// emission panics, and the message names the offending part (closes D-3)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected stylesheet emission to panic on invalid sheet")
			return
		}
		errMsg := r.(error).Error()
		if !strings.Contains(errMsg, "veilpart") {
			t.Errorf("panic message should name the offending part, got:\n%s", errMsg)
		}
	}()

	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).Part("veilpart", style.Veil())
	sheet.Stylesheet()
}

func TestInteractiveDerivesFamily(t *testing.T) {
	// Interactive(Primary) emits the three Primary states and no other family (closes D-4)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.Interactive(style.Primary)).Stylesheet().String()

	for name, want := range map[string]string{
		"hover": css.Hover(css.ColorPrimary),
		"focus": css.Focus(css.ColorPrimary),
		"press": css.Press(css.ColorPrimary),
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected primary %s derivation to be emitted: %s", name, want)
		}
	}

	// no other family leaks in
	if strings.Contains(s, css.Hover(css.ColorSuccess)) {
		t.Error("should not contain another family's derivation")
	}
}

func TestFocusVisible(t *testing.T) {
	// the focus cue emits :focus-visible; bare :focus appears nowhere (closes D-4)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Cue(widget.Focus, "", style.As(style.Primary)).Stylesheet().String()

	if !strings.Contains(s, ":focus-visible") {
		t.Error("expected :focus-visible to be emitted for Focus cue")
	}
	if strings.Contains(s, ":focus ") || strings.HasSuffix(s, ":focus {\n") {
		t.Error("bare :focus must not be emitted")
	}
}

// A simple parser to extract all var(...) calls properly handling nested parenthesis
func extractVarCalls(cssStr string) [][2]string {
	var results [][2]string
	idx := 0
	for {
		start := strings.Index(cssStr[idx:], "var(--")
		if start == -1 {
			break
		}
		startPos := idx + start
		// scan for the matching closing parenthesis
		parenCount := 1
		p := startPos + 4 // after "var("
		for p < len(cssStr) && parenCount > 0 {
			if cssStr[p] == '(' {
				parenCount++
			} else if cssStr[p] == ')' {
				parenCount--
			}
			p++
		}
		if parenCount == 0 {
			fullMatch := cssStr[startPos:p]
			// Extract variable name, which is between "var(" and either "," or ")"
			varNamePart := fullMatch[4 : len(fullMatch)-1]
			varName := varNamePart
			commaIdx := strings.Index(varNamePart, ",")
			if commaIdx != -1 {
				varName = strings.TrimSpace(varNamePart[:commaIdx])
			}
			results = append(results, [2]string{fullMatch, varName})
		}
		idx = startPos + 4
	}
	return results
}

func TestNoInventedValues(t *testing.T) {
	// Extend drift guard to verify every var() matches the catalog + fallbacks (closes D-5)
	wd := &testWidget{name: "w", kind: widget.Dialog}

	// Giant sheet to exercise EVERY option and scale
	sheet := style.For(wd).
		Root(
			style.Stack(style.Space12),
			style.As(style.Page),
			style.Pad(style.Space8),
			style.Round(style.RadiusFull),
			style.Raise(style.Popover),
			style.Width(style.Readable),
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
			style.Animate(style.MotionSlow),
			style.Scroll(),
			style.KeepSize(),
			style.EdgeToEdge(),
			style.HideOverflow(),
			style.Backdrop(style.Parent), // Backdrop(Parent) avoids condition 6 Backdrop(Viewport) under a Split error
			style.Veil(),
		).
		Part("item1", style.Row(style.Space1), style.Interactive(style.Primary)).
		Part("item2", style.Split(style.SplitTwoThirds, style.Space2), style.As(style.Panel)).
		Part("item3", style.Grid(style.ColumnWide, style.Space3), style.As(style.Inset)).
		Part("item4", style.Center(style.Third), style.As(style.Secondary)).
		Part("item5", style.FillCentered(), style.As(style.Highlight)).
		Part("item6", style.ScrollRow(style.Space4), style.As(style.Success)).
		Part("item7", style.MediaBox(style.Aspect16x9), style.As(style.Danger)).
		Part("item8", style.Stack(style.Space6), style.Interactive(style.Subtle)).
		Part("item9", style.Row(style.SpaceNone), style.Interactive(style.Inset)).
		Part("item10", style.Cover()).
		Part("item11", style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).
		Part("item12", style.Drawer(style.SideEnd, style.TwoThirds), style.RevealedBy(widget.Open)).
		Part("item13", style.IconBox(style.IconSm)).
		Part("item14", style.IconBox(style.IconMd)).
		Part("item15", style.IconBox(style.IconLg)).
		On(css.Mobile, "item10", style.Stack(style.Space1))

	cssStr := sheet.Stylesheet().String()

	// Token maps to assert fallbacks exactly
	cssTokens := []css.Token{
		css.ColorPrimary, css.ColorOnPrimary,
		css.ColorSuccess, css.ColorOnSuccess,
		css.ColorDanger, css.ColorOnDanger,
		css.ColorBackground, css.ColorOnBackground,
		css.ColorSurface, css.ColorOnSurface,
		css.ColorSurfaceSunken,
		css.ColorSelection, css.ColorOnSelection,
		css.ColorOutline, css.ColorMuted,
		css.MixHover, css.MixFocus, css.MixPress,
		css.TextXs, css.TextSm, css.TextBase, css.TextLg, css.TextXl, css.Text2xl,
		css.LeadingNormal,
		css.FontWeightRegular, css.FontWeightMedium, css.FontWeightBold,
		css.Space0, css.Space1, css.Space2, css.Space3, css.Space4, css.Space6, css.Space8, css.Space12,
		css.RadiusSm, css.RadiusMd, css.RadiusLg, css.RadiusFull,
		css.ShadowSm, css.ShadowMd, css.ShadowLg,
		css.DurationFast, css.DurationBase, css.DurationSlow,
		css.EaseInOut,
		css.ZBase, css.ZDropdown, css.ZSticky, css.ZModal, css.ZToast, css.ZTooltip,
		css.BpSm, css.BpMd, css.BpLg, css.BpXl,
		css.MaxWReadable,
		css.ColumnNarrow, css.ColumnMedium, css.ColumnWide,
		css.RailNarrow, css.RailWide,
		css.ControlHeight, css.ChipWidth, css.VeilBlur,
		css.ColorAccent, css.ColorOnAccent,
	}

	tokenMap := make(map[string]css.Token)
	for _, tok := range cssTokens {
		tokenMap[tok.Name] = tok
	}

	// Layout variables allowed to not exist in tokenMap
	layoutVariables := map[string]bool{
		"--gap":       true,
		"--ratio":     true,
		"--track":     true,
		"--rail":      true,
		"--max-width": true,
	}
	// Every theme pair's plain light/dark half properties (see
	// css.Token.EnhancedVar, css declareSplit) — a different kind of
	// variable than the ones above, not derived through Token.Var()'s
	// formula, so skipped the same way rather than checked against it.
	for _, tok := range []css.Token{
		css.ColorBackground, css.ColorOnBackground, css.ColorSurface,
		css.ColorOnSurface, css.ColorOutline, css.ColorMuted,
	} {
		layoutVariables[tok.LightVarName()] = true
		layoutVariables[tok.DarkVarName()] = true
	}

	// Check there are no hardcoded hexadecimal colors or rgba except in fallbacks of var()
	// or in a Safari-legacy static fallback (see below).
	varVarRegex := regexp.MustCompile(`var\([^)]+\)`)
	cleanCSS := varVarRegex.ReplaceAllString(cssStr, "")

	// ACCEPTED EXCEPTION — Safari-legacy static fallback (double declaration):
	// every themed color is now emitted twice, static-literal first and
	// light-dark()/color-mix()-enhanced second (see css.Token.LightValue,
	// css.HoverStatic/FocusStatic/PressStatic/FadeStatic, and their call
	// sites in emit.go/emit_decls.go/surface.go). A browser without
	// light-dark()/color-mix() support drops the second declaration
	// (invalid at parse time) and keeps the static first one — permanently
	// the light theme. That first declaration is a bare literal outside
	// var() BY DESIGN, so this drift guard must not flag it. It is still
	// bounded: only literals that exactly match a value this package itself
	// derives from the catalog (never an arbitrary hardcoded color) are
	// allowed.
	knownStaticFallbacks := map[string]bool{}
	// Only literals that are themselves hex colors go in the allowlist — a
	// non-color token's LightValue ("0.75rem", "15%", "0") would otherwise
	// strip that substring out of unrelated hex codes elsewhere in the sheet
	// (e.g. stripping "0" mangles "#ba2c0d" into "#ba2cd").
	addIfHex := func(s string) {
		if strings.HasPrefix(s, "#") {
			knownStaticFallbacks[s] = true
		}
	}
	addIfHex(css.FadeStatic(css.ColorSurface, 0.4)) // Veil() backdrop wash
	for _, tok := range cssTokens {
		addIfHex(tok.LightValue())
		addIfHex(css.HoverStatic(tok))
		addIfHex(css.FocusStatic(tok))
		addIfHex(css.PressStatic(tok))
		// EnhancedVar()/NestedEnhanced() bake BOTH halves of a theme token
		// as literals (e.g. light-dark(#F2F2F7,#161B22)) — see their doc
		// comments in tinywasm/css for why a var() reference can't be used
		// here instead. addIfHex only ever matches a value that IS a bare
		// hex, so this is harmless for tok.Light/tok.Dark that are actually
		// live expressions (e.g. "15%", "0.75rem", or a color-mix() string).
		addIfHex(tok.Light)
		addIfHex(tok.Dark)
	}
	stripKnownStatic := func(s string) string {
		for lit := range knownStaticFallbacks {
			s = strings.ReplaceAll(s, lit, "")
		}
		return s
	}
	cleanCSS = stripKnownStatic(cleanCSS)

	hexColorRegex := regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)
	if hexColorRegex.MatchString(cleanCSS) {
		t.Errorf("Literal color hexadecimal found outside var(): %q", hexColorRegex.FindString(cleanCSS))
	}

	rgbaRegex := regexp.MustCompile(`rgba\([^)]+\)`)
	if rgbaRegex.MatchString(cleanCSS) {
		t.Errorf("Literal rgba wash found outside var(): %q", rgbaRegex.FindString(cleanCSS))
	}

	vwRegex := regexp.MustCompile(`\b[0-9]+vw\b`)
	vhRegex := regexp.MustCompile(`\b[0-9]+vh\b`)
	if vwRegex.MatchString(cleanCSS) {
		t.Errorf("Prohibited unit 'vw' found: %q", vwRegex.FindString(cleanCSS))
	}
	if vhRegex.MatchString(cleanCSS) {
		t.Errorf("Prohibited unit 'vh' found: %q", vhRegex.FindString(cleanCSS))
	}

	// Check all variables are from our catalog
	varMatches := extractVarCalls(cssStr)
	for _, m := range varMatches {
		fullMatch := m[0]
		varName := m[1]

		if layoutVariables[varName] {
			continue
		}

		tok, exists := tokenMap[varName]
		if !exists {
			t.Errorf("CSS variable %q not in css catalog or local private tokens list", varName)
			continue
		}

		// ACCEPTED DRIFT GUARD EXCEPTION:
		// We tolerate a bare "var(--name)" call with no local fallback string only when the variable
		// name is a valid token from the css catalog.
		// Why this is safe: These matches represent nested variable references within larger formulas
		// (such as those returned by ColorSurfaceSunken or ColorSelection from tinywasm/css v0.3.3+).
		// Because they are nested inside an outer var() call, they are already protected by the outer
		// var()'s fallback. Enforcing fallback matching on these nested references is both redundant
		// and structurally impossible here since the formulas themselves are owned and defined by tinywasm/css.
		expectedVarCall := tok.Var()
		if fullMatch != expectedVarCall && fullMatch != "var("+varName+")" {
			t.Errorf("Visual drift detected for %q.\nIn stylesheet: %q\nExpected: %q",
				varName, fullMatch, expectedVarCall)
		}
	}
}

// TestDoubleDeclarationsAreParseTimeSafe is the regression test for the
// iPhone-7-all-blue investigation. Every themed color in this package is
// emitted TWICE per property: a static hex literal first, a
// light-dark()/color-mix() "enhanced" value second — a browser without
// those functions is supposed to drop the second declaration (unrecognized
// function name, invalid AT PARSE TIME) and keep the first.
//
// That only works if the SECOND declaration contains no var() ANYWHERE.
// This is not obvious and was the actual bug: per the CSS Custom Properties
// cascade, a declaration containing var() — even a var() to a property that
// would always itself resolve fine — is a "variable-valued" declaration,
// and validity for the WHOLE thing is deferred to computed-value time
// instead of checked at parse time. It still wins the cascade over the
// earlier static declaration (cascade doesn't know yet that it will turn
// out invalid), and when it then fails to compute, the property falls to
// its INITIAL value (transparent, for a color) — not to the earlier static
// sibling. The static declaration is silently discarded, and the page goes
// solid blue with white text: the demo's own Primary-colored ancestor
// showing through every unpainted surface.
//
// Confirmed empirically, not just by spec reading, before this test
// existed: a minimal repro using an unrecognized function whose arguments
// were LITERALS correctly fell back to the static declaration; the
// identical repro with var() references as those arguments did not — see
// css.Token.NestedEnhanced's doc comment for the full mechanism and
// css.Token.EnhancedVar for the standalone-declaration counterpart.
//
// This walks every rule this package can generate — Interactive() (which
// exercises Hover/Focus/Press), every As(Surface), and Veil() — and, for
// every property declared more than once within a single rule, asserts
// every declaration after the first contains no var() anywhere.
func TestDoubleDeclarationsAreParseTimeSafe(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Dialog}
	sheet := style.For(wd).
		Root(style.As(style.Page), style.Backdrop(style.Parent), style.Veil()).
		Part("item1", style.Row(style.Space1), style.Interactive(style.Primary)).
		Part("item2", style.As(style.Panel)).
		Part("item3", style.As(style.Inset)).
		Part("item4", style.As(style.Secondary)).
		Part("item5", style.As(style.Highlight)).
		Part("item6", style.As(style.Accent)).
		Part("item7", style.As(style.Success)).
		Part("item8", style.As(style.Danger)).
		Part("item9", style.Interactive(style.Subtle)).
		Part("item10", style.Interactive(style.Inset))

	cssStr := sheet.Stylesheet().String()

	// Each {...} match is one rule's declaration body — [^{}]* can't cross
	// into a nested block, so this naturally lands on the innermost
	// (non-@layer-wrapper) blocks, exactly the plain declaration lists this
	// test needs to inspect.
	blockRe := regexp.MustCompile(`\{[^{}]*\}`)
	for _, block := range blockRe.FindAllString(cssStr, -1) {
		body := strings.TrimSuffix(strings.TrimPrefix(block, "{"), "}")
		seen := map[string]int{}
		for _, rawDecl := range strings.Split(body, ";") {
			decl := strings.TrimSpace(rawDecl)
			if decl == "" {
				continue
			}
			parts := strings.SplitN(decl, ":", 2)
			if len(parts) != 2 {
				continue
			}
			prop := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			seen[prop]++
			// Only the combination is broken: a plain var() with nothing
			// risky in it (e.g. "var(--color-primary,#1b5d8c)", where
			// --color-primary is never wrapped in light-dark()/color-mix())
			// is completely safe on any browser — var() itself needs no
			// protecting. light-dark()/color-mix() WITH a nested var() is
			// the failure mode this whole test exists to catch.
			risky := strings.Contains(val, "light-dark(") || strings.Contains(val, "color-mix(")
			if seen[prop] > 1 && risky && strings.Contains(val, "var(") {
				t.Errorf("declaration #%d of %q mixes a light-dark()/color-mix() call with var() — a browser that can't parse the outer function defers the WHOLE declaration to computed-value time and falls to the initial value instead of the earlier static declaration, discarding the Safari-legacy fallback: %q (rule: %q)",
					seen[prop], prop, val, block)
			}
		}
	}
}

func TestSpaceStepsDistinct(t *testing.T) {
	// no two Space steps resolve to the same token (closes D-6)
	steps := []style.Space{
		style.SpaceNone, style.Space1, style.Space2, style.Space3,
		style.Space4, style.Space6, style.Space8, style.Space12,
	}

	wd := &testWidget{name: "w", kind: widget.Region}
	resolved := make(map[string]style.Space)

	for _, s := range steps {
		cssStr := style.For(wd).Root(style.Stack(s)).Stylesheet().String()
		// find `--gap: var(...)`
		re := regexp.MustCompile(`--gap:\s*([^;]+);`)
		match := re.FindStringSubmatch(cssStr)
		if len(match) < 2 {
			t.Fatalf("expected stack to emit --gap declaration for step %d", s)
		}
		val := match[1]
		if prev, exists := resolved[val]; exists {
			t.Errorf("Duplicate resolution: step %d and step %d both resolve to %q", s, prev, val)
		}
		resolved[val] = s
	}
}

func TestSurfaceCarriesShape(t *testing.T) {
	// As(Panel) alone emits radius; Round(RadiusNone) overrides it (closes D-6)
	wd := &testWidget{name: "w", kind: widget.Region}
	s1 := style.For(wd).Root(style.As(style.Panel)).Stylesheet().String()
	if !strings.Contains(s1, "border-radius: var(--radius-md") {
		t.Errorf("expected Panel alone to emit radius-md, got:\n%s", s1)
	}

	s2 := style.For(wd).Root(style.As(style.Panel), style.Round(style.RadiusNone)).Stylesheet().String()
	if strings.Contains(s2, "border-radius: var(--radius-md") {
		t.Errorf("expected Round(RadiusNone) to override Panel's default radius, got:\n%s", s2)
	}
}

func TestNoUnreachableSelectors(t *testing.T) {
	// no selector begins with .fl- or .exc-; no empty @layer block (closes D-7)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.Stack(style.Space1)).Stylesheet().String()

	if strings.Contains(s, ".fl-") {
		t.Error("unreachable .fl- selector emitted")
	}
	if strings.Contains(s, ".exc-") {
		t.Error("unreachable .exc- selector emitted")
	}

	// Should not contain empty layers
	if strings.Contains(s, "@layer states {}") || strings.Contains(s, "@layer widgets {}") {
		t.Error("empty @layer blocks must be omitted")
	}
}

func TestInteractiveRejectsNonInteractive(t *testing.T) {
	// Interactive(Page) and Interactive(Inactive) are reported (closes D-3)
	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).Root(style.Interactive(style.Page))
	if len(sheet.Validate()) == 0 {
		t.Error("Interactive(Page) should be reported as invalid")
	}

	sheet2 := style.For(wd).Root(style.Interactive(style.Inactive))
	if len(sheet2.Validate()) == 0 {
		t.Error("Interactive(Inactive) should be reported as invalid")
	}
}

func TestSplitCollapses(t *testing.T) {
	// the emitted sheet contains no @container and no container-type; side by side / stacked (closes D-9)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.Split(style.SplitTwoThirds, style.Space2)).Stylesheet().String()

	if strings.Contains(s, "@container") {
		t.Error("Split must not emit @container rule")
	}
	if strings.Contains(s, "container-type") {
		t.Error("Split must not emit container-type declaration")
	}
}

func TestSplitRatioIsUnitlessForFlexGrow(t *testing.T) {
	// --ratio feeds flex-grow. An fr unit is invalid at computed-value time:
	// flex-grow silently resets to its initial 0 and the first partition
	// collapses to its content width instead of taking the larger share.
	wd := &testWidget{name: "w", kind: widget.Region}
	for _, tc := range []struct {
		ratio style.SplitRatio
		want  string
	}{
		{style.SplitHalf, "--ratio: 1;"},
		{style.SplitTwoThirds, "--ratio: 2;"},
		{style.SplitThreeQuarters, "--ratio: 3;"},
	} {
		s := style.For(wd).Root(style.Split(tc.ratio, style.Space2)).Stylesheet().String()
		if !strings.Contains(s, tc.want) {
			t.Errorf("expected %q, got:\n%s", tc.want, s)
		}
		if strings.Contains(s, "--ratio: 1fr;") || strings.Contains(s, "--ratio: 2fr;") || strings.Contains(s, "--ratio: 3fr;") {
			t.Errorf("--ratio must not carry an fr unit when it feeds flex-grow, got:\n%s", s)
		}
	}
}

func TestSurfaceCarriesNoPadding(t *testing.T) {
	// As(Panel) emits border-radius but no padding (closes C-2)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.As(style.Panel)).Stylesheet().String()

	if strings.Contains(s, "padding:") {
		t.Error("As(Panel) alone must not emit any padding")
	}
}

func TestSheetParts(t *testing.T) {
	// Parts() returns the declared parts, sorted (closes C-7)
	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).
		Part("beta", style.As(style.Panel)).
		Part("alpha", style.As(style.Panel)).
		Part("gamma", style.As(style.Panel))

	parts := sheet.Parts()
	if len(parts) != 3 || parts[0] != widget.Part("alpha") || parts[1] != widget.Part("beta") || parts[2] != widget.Part("gamma") {
		t.Errorf("expected sorted parts [alpha, beta, gamma], got: %v", parts)
	}
}

type MyZeroWidget struct{}

func (z *MyZeroWidget) WidgetName() widget.Name { return widget.Name("zero") }
func (z *MyZeroWidget) WidgetKind() widget.Kind { return widget.Region }
func (z *MyZeroWidget) RenderCSS() *css.Stylesheet {
	return style.For(z).Root(style.As(style.Panel)).Stylesheet()
}

func TestZeroValueProvider(t *testing.T) {
	// (&T{}).RenderCSS() succeeds without reading fields
	var z MyZeroWidget
	ss := z.RenderCSS()
	if ss == nil || !strings.Contains(ss.String(), ".zero") {
		t.Error("RenderCSS failed on zero value of component")
	}
}
