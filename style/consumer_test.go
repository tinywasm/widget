//go:build !wasm

package style_test

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"

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

	// 4. Cada Surface usada resuelve a un token que existe en el catálogo de css.
	allowedVariables := map[string]bool{
		// CSS package tokens
		"--color-primary":      true,
		"--color-on-primary":   true,
		"--color-secondary":    true,
		"--color-on-secondary": true,
		"--color-success":      true,
		"--color-error":        true,
		"--color-background":   true,
		"--color-surface":      true,
		"--color-on-surface":   true,
		"--color-muted":        true,
		"--color-hover":        true,

		"--color-background-light": true,
		"--color-background-dark":  true,
		"--color-surface-light":    true,
		"--color-surface-dark":     true,
		"--color-on-surface-light": true,
		"--color-on-surface-dark":  true,
		"--color-muted-light":      true,
		"--color-muted-dark":       true,
		"--color-hover-light":      true,
		"--color-hover-dark":       true,

		"--text-xs":   true,
		"--text-sm":   true,
		"--text-base": true,
		"--text-lg":   true,
		"--text-xl":   true,
		"--text-2xl":  true,

		"--leading-tight":   true,
		"--leading-normal":  true,
		"--leading-relaxed": true,

		"--font-weight-regular": true,
		"--font-weight-medium":  true,
		"--font-weight-bold":    true,

		"--tracking-tight":  true,
		"--tracking-normal": true,
		"--tracking-wide":   true,

		"--space-1":  true,
		"--space-2":  true,
		"--space-3":  true,
		"--space-4":  true,
		"--space-6":  true,
		"--space-8":  true,
		"--space-12": true,

		"--radius-sm":   true,
		"--radius-md":   true,
		"--radius-lg":   true,
		"--radius-full": true,

		"--shadow-sm": true,
		"--shadow-md": true,
		"--shadow-lg": true,
		"--shadow-xl": true,

		"--duration-fast": true,
		"--duration-base": true,
		"--duration-slow": true,

		"--ease-in":     true,
		"--ease-out":    true,
		"--ease-in-out": true,

		"--z-base":     true,
		"--z-dropdown": true,
		"--z-sticky":   true,
		"--z-modal":    true,
		"--z-toast":    true,
		"--z-tooltip":  true,

		"--bp-sm": true,
		"--bp-md": true,
		"--bp-lg": true,
		"--bp-xl": true,

		"--max-w-prose":   true,
		"--max-w-content": true,
		"--max-w-screen":  true,

		// Local style package tokens (surface.go)
		"--color-surface-sunken": true,
		"--color-outline":        true,
		"--color-selection":      true,
		"--color-on-selection":   true,
		"--color-on-success":     true,
		"--color-on-error":       true,
		"--color-disabled":       true,
		"--color-on-disabled":    true,

		"--color-background-hover": true,
		"--color-surface-hover":    true,
		"--color-primary-hover":    true,
		"--color-secondary-hover":  true,
		"--color-selection-hover":  true,
		"--color-success-hover":    true,
		"--color-error-hover":      true,
		"--color-muted-hover":      true,

		"--color-background-focus": true,
		"--color-surface-focus":    true,
		"--color-primary-focus":    true,
		"--color-secondary-focus":  true,
		"--color-selection-focus":  true,

		"--color-background-press": true,
		"--color-surface-press":    true,
		"--color-primary-press":    true,
		"--color-secondary-press":  true,
		"--color-selection-press":  true,

		// Primitives layout custom properties
		"--gap":       true,
		"--ratio":     true,
		"--track":     true,
		"--max-width": true,
	}

	varVarSubRegex := regexp.MustCompile(`var\((--[a-zA-Z0-9_\-]+)`)
	varMatches := varVarSubRegex.FindAllStringSubmatch(cssStr, -1)
	for _, m := range varMatches {
		varName := m[1]
		if !allowedVariables[varName] {
			t.Errorf("Variable CSS inválida/no registrada encontrada: %q", varName)
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
