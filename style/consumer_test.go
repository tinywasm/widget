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

// MasterDetail es un widget de ejemplo para el test de consumidor
type MasterDetail struct{}

func (m *MasterDetail) WidgetName() widget.Name { return widget.Name("masterdetail") }
func (m *MasterDetail) WidgetKind() widget.Kind { return widget.Grid }

func (m *MasterDetail) Style() *style.Sheet {
	return style.Of(m.WidgetName()).
		Root(style.Grid(style.TrackSm, style.Space2), style.On(style.Page), style.Scrolls()).
		Part("master", style.Stack(style.Space1), style.On(style.Panel), style.Round(style.RadiusMd)).
		Part("detail", style.Stack(style.Space2), style.On(style.Panel), style.Round(style.RadiusMd)).
		Part("item", style.Row(style.Space1), style.On(style.Muted)).
		When(widget.Selected, "item", style.On(style.Selected)).
		Cue(widget.Hover, "item", style.On(style.MutedHover))
}

func TestConsumerMasterDetail(t *testing.T) {
	wd := &MasterDetail{}
	sheet := wd.Style()
	ss := sheet.Stylesheet()
	cssStr := ss.String()

	t.Logf("Generated CSS:\n%s", cssStr)

	// 1. Toda clase presente en el markup existe en la hoja, y viceversa.
	markupClasses := map[string]bool{
		"masterdetail":         true,
		"masterdetail__master": true,
		"masterdetail__detail": true,
		"masterdetail__item":   true,
	}

	// Extraer todas las clases que empiezan con .masterdetail del CSS
	classRegex := regexp.MustCompile(`\.masterdetail[a-zA-Z0-9_\-]*`)
	matches := classRegex.FindAllString(cssStr, -1)

	sheetClasses := make(map[string]bool)
	for _, m := range matches {
		name := strings.TrimPrefix(m, ".")
		sheetClasses[name] = true
	}

	for mc := range markupClasses {
		if !sheetClasses[mc] {
			t.Errorf("Clase en el markup %q no encontrada en la hoja de estilos", mc)
		}
	}

	for sc := range sheetClasses {
		if !markupClasses[sc] {
			t.Errorf("Clase en la hoja de estilos %q no encontrada en el markup", sc)
		}
	}

	// 2. La hoja no contiene: !important, @media, ningún literal de color, ninguna unidad vw/vh.
	if strings.Contains(cssStr, "!important") {
		t.Error("La hoja de estilos contiene '!important', lo cual está prohibido")
	}

	if strings.Contains(cssStr, "@media") {
		t.Error("La hoja de estilos contiene '@media', lo cual está prohibido (se debe usar @container)")
	}

	// Para verificar literales de color y unidades prohibidas sin falsos positivos en las llamadas a var() con fallbacks:
	// Eliminamos las referencias var(...) antes de realizar la validación.
	varVarRegex := regexp.MustCompile(`var\([^)]+\)`)
	cleanCSS := varVarRegex.ReplaceAllString(cssStr, "")

	// Literal de color hexadecimal
	hexColorRegex := regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)
	if hexColorRegex.MatchString(cleanCSS) {
		t.Errorf("La hoja de estilos contiene un literal de color hexadecimal fuera de var(): %q", hexColorRegex.FindString(cleanCSS))
	}

	// Unidades vw/vh
	vwRegex := regexp.MustCompile(`\b[0-9]+vw\b`)
	vhRegex := regexp.MustCompile(`\b[0-9]+vh\b`)
	if vwRegex.MatchString(cleanCSS) {
		t.Errorf("La hoja de estilos contiene unidad prohibida 'vw': %q", vwRegex.FindString(cleanCSS))
	}
	if vhRegex.MatchString(cleanCSS) {
		t.Errorf("La hoja de estilos contiene unidad prohibida 'vh': %q", vhRegex.FindString(cleanCSS))
	}

	// 3. Las capas aparecen en el orden declarado.
	idxTokens := strings.Index(cssStr, "tokens")
	idxPrimitives := strings.Index(cssStr, "primitives")
	idxWidgets := strings.Index(cssStr, "widgets")
	idxStates := strings.Index(cssStr, "states")

	if idxTokens == -1 || idxPrimitives == -1 || idxWidgets == -1 || idxStates == -1 {
		t.Error("Falta alguna de las capas (tokens, primitives, widgets, states) en la hoja")
	} else if !(idxTokens < idxPrimitives && idxPrimitives < idxWidgets && idxWidgets < idxStates) {
		t.Errorf("El orden de las capas es incorrecto: tokens (%d) < primitives (%d) < widgets (%d) < states (%d)",
			idxTokens, idxPrimitives, idxWidgets, idxStates)
	}

	// 4. Cada Surface usada resuelve a un token que existe en el catálogo de css (y coincide exactamente en su valor/fallback).
	cssTokens := []css.Token{
		css.ColorPrimary, css.ColorOnPrimary, css.ColorSecondary, css.ColorOnSecondary,
		css.ColorSuccess, css.ColorOnSuccess, css.ColorError, css.ColorOnError,
		css.ColorBackground, css.ColorSurface, css.ColorSurfaceSunken, css.ColorOnSurface,
		css.ColorOutline, css.ColorMuted, css.ColorHover, css.ColorSelection, css.ColorOnSelection,
		css.ColorDisabled, css.ColorOnDisabled,
		css.ColorBackgroundLight, css.ColorBackgroundDark, css.ColorSurfaceLight, css.ColorSurfaceDark,
		css.ColorSurfaceSunkenLight, css.ColorSurfaceSunkenDark, css.ColorOnSurfaceLight, css.ColorOnSurfaceDark,
		css.ColorOutlineLight, css.ColorOutlineDark, css.ColorMutedLight, css.ColorMutedDark,
		css.ColorHoverLight, css.ColorHoverDark, css.ColorSelectionLight, css.ColorSelectionDark,
		css.ColorOnSelectionLight, css.ColorOnSelectionDark, css.ColorDisabledLight, css.ColorDisabledDark,
		css.ColorOnDisabledLight, css.ColorOnDisabledDark,
		css.TextXs, css.TextSm, css.TextBase, css.TextLg, css.TextXl, css.Text2xl,
		css.LeadingTight, css.LeadingNormal, css.LeadingRelaxed,
		css.FontWeightRegular, css.FontWeightMedium, css.FontWeightBold,
		css.TrackingTight, css.TrackingNormal, css.TrackingWide,
		css.Space1, css.Space2, css.Space3, css.Space4, css.Space6, css.Space8, css.Space12,
		css.RadiusSm, css.RadiusMd, css.RadiusLg, css.RadiusFull,
		css.ShadowSm, css.ShadowMd, css.ShadowLg, css.ShadowXl,
		css.DurationFast, css.DurationBase, css.DurationSlow,
		css.EaseIn, css.EaseOut, css.EaseInOut,
		css.ZBase, css.ZDropdown, css.ZSticky, css.ZModal, css.ZToast, css.ZTooltip,
		css.BpSm, css.BpMd, css.BpLg, css.BpXl,
		css.MaxWProse, css.MaxWContent, css.MaxWScreen,
	}

	localTokens := []css.Token{
		style.ColorBackgroundHover, style.ColorSurfaceHover, style.ColorPrimaryHover, style.ColorSecondaryHover,
		style.ColorSelectionHover, style.ColorSuccessHover, style.ColorErrorHover, style.ColorMutedHover,
		style.ColorBackgroundFocus, style.ColorSurfaceFocus, style.ColorPrimaryFocus, style.ColorSecondaryFocus,
		style.ColorSelectionFocus,
		style.ColorBackgroundPress, style.ColorSurfacePress, style.ColorPrimaryPress, style.ColorSecondaryPress,
		style.ColorSelectionPress,
	}

	tokenMap := make(map[string]css.Token)
	for _, tok := range cssTokens {
		tokenMap[tok.Name] = tok
	}
	for _, tok := range localTokens {
		tokenMap[tok.Name] = tok
	}

	layoutVariables := map[string]bool{
		"--gap":       true,
		"--ratio":     true,
		"--track":     true,
		"--max-width": true,
	}

	// Regex para encontrar cada llamada a var(...) incluyendo su posible valor de fallback
	varCallRegex := regexp.MustCompile(`var\((--[a-zA-Z0-9_\-]+)(?:,\s*([^)]+))?\)`)
	varMatches := varCallRegex.FindAllStringSubmatch(cssStr, -1)
	for _, m := range varMatches {
		fullMatch := m[0]
		varName := m[1]

		if layoutVariables[varName] {
			continue // las variables de disposición dinámica no requieren comprobación de catálogo
		}

		tok, exists := tokenMap[varName]
		if !exists {
			t.Errorf("Variable CSS %q no existe en el catálogo de css ni en los tokens locales", varName)
			continue
		}

		// Aseveración estricta de coherencia total (incluyendo fallback):
		// El string literal de var(...) en la hoja de estilos generada debe coincidir al 100%
		// con lo que el token de referencia produce bajo tok.Var(). ¡Esto elimina cualquier deriva!
		expectedVarCall := tok.Var()
		if fullMatch != expectedVarCall {
			t.Errorf("Deriva visual detectada en el valor/fallback para %q.\nEn la hoja: %q\nEn el catálogo: %q",
				varName, fullMatch, expectedVarCall)
		}
	}

	// 5. Emisión determinista: dos ejecuciones, salida idéntica.
	cssStr2 := wd.Style().Stylesheet().String()
	if cssStr != cssStr2 {
		t.Error("La salida de la hoja de estilos no es determinista (difiere entre ejecuciones)")
	}

	// 6. GOOS=js GOARCH=wasm go list -deps sobre un consumidor de ejemplo no contiene widget/style.
	cmd := exec.Command("go", "list", "-deps", "github.com/tinywasm/widget/example")
	cmd.Env = append(cmd.Environ(), "GOOS=js", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error ejecutando 'go list': %v, salida: %s", err, string(out))
	}

	outStr := string(out)
	if strings.Contains(outStr, "github.com/tinywasm/widget/style") {
		t.Error("El binario WASM del consumidor de ejemplo depende de 'widget/style', lo cual viola las restricciones de build tag")
	}
}
