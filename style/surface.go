//go:build !wasm

package style

import "github.com/tinywasm/css"

// Surface es una decisión visual completa: fondo, texto y borde resueltos como
// UNA tripleta. No hay forma de elegir un fondo sin su texto.
type Surface uint8

const (
	Page Surface = iota
	PageHover
	PageFocus
	PagePress

	Panel
	PanelHover
	PanelFocus
	PanelPress

	Sunken
	SunkenHover
	SunkenFocus
	SunkenPress

	Accent
	AccentHover
	AccentFocus
	AccentPress

	Secondary
	SecondaryHover
	SecondaryFocus
	SecondaryPress

	Selected
	SelectedHover
	SelectedFocus
	SelectedPress

	Success
	SuccessHover
	SuccessFocus
	SuccessPress

	Danger
	DangerHover
	DangerFocus
	DangerPress

	Muted
	MutedHover
	MutedFocus
	MutedPress

	Disabled
)

// Tokens de estados interactivos locales (Hover, Focus, Press)
var (
	ColorBackgroundHover = css.Token{Name: "--color-background-hover", Fallback: "#f2f2f7"}
	ColorSurfaceHover    = css.Token{Name: "--color-surface-hover", Fallback: "#e5e5ea"}
	ColorPrimaryHover    = css.Token{Name: "--color-primary-hover", Fallback: "#0096bc"}
	ColorSecondaryHover  = css.Token{Name: "--color-secondary-hover", Fallback: "#503cd1"}
	ColorSelectionHover  = css.Token{Name: "--color-selection-hover", Fallback: "#0062cc"}
	ColorSuccessHover    = css.Token{Name: "--color-success-hover", Fallback: "#34a243"}
	ColorErrorHover      = css.Token{Name: "--color-error-hover", Fallback: "#cf3c15"}
	ColorMutedHover      = css.Token{Name: "--color-muted-hover", Fallback: "#55555a"}

	ColorBackgroundFocus = css.Token{Name: "--color-background-focus", Fallback: "#ffffff"}
	ColorSurfaceFocus    = css.Token{Name: "--color-surface-focus", Fallback: "#f2f2f7"}
	ColorPrimaryFocus    = css.Token{Name: "--color-primary-focus", Fallback: "#0096bc"}
	ColorSecondaryFocus  = css.Token{Name: "--color-secondary-focus", Fallback: "#503cd1"}
	ColorSelectionFocus  = css.Token{Name: "--color-selection-focus", Fallback: "#0062cc"}

	ColorBackgroundPress = css.Token{Name: "--color-background-press", Fallback: "#e5e5ea"}
	ColorSurfacePress    = css.Token{Name: "--color-surface-press", Fallback: "#d1d1d6"}
	ColorPrimaryPress    = css.Token{Name: "--color-primary-press", Fallback: "#0080a0"}
	ColorSecondaryPress  = css.Token{Name: "--color-secondary-press", Fallback: "#4532be"}
	ColorSelectionPress  = css.Token{Name: "--color-selection-press", Fallback: "#0052b3"}
)

// Triplet representa la combinación de fondo, texto y borde.
type Triplet struct {
	Bg     string
	Text   string
	Border string
}

// Resolve traduce una Surface a su correspondiente Triplet de estilos.
func (s Surface) Resolve() Triplet {
	switch s {
	case Page:
		return Triplet{Bg: css.ColorBackground.Var(), Text: css.ColorOnSurface.Var()}
	case PageHover:
		return Triplet{Bg: ColorBackgroundHover.Var(), Text: css.ColorOnSurface.Var()}
	case PageFocus:
		return Triplet{Bg: ColorBackgroundFocus.Var(), Text: css.ColorOnSurface.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case PagePress:
		return Triplet{Bg: ColorBackgroundPress.Var(), Text: css.ColorOnSurface.Var()}

	case Panel:
		return Triplet{Bg: css.ColorSurface.Var(), Text: css.ColorOnSurface.Var(), Border: "1px solid " + css.ColorOutline.Var()}
	case PanelHover:
		return Triplet{Bg: ColorSurfaceHover.Var(), Text: css.ColorOnSurface.Var(), Border: "1px solid " + css.ColorOutline.Var()}
	case PanelFocus:
		return Triplet{Bg: ColorSurfaceFocus.Var(), Text: css.ColorOnSurface.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case PanelPress:
		return Triplet{Bg: ColorSurfacePress.Var(), Text: css.ColorOnSurface.Var(), Border: "1px solid " + css.ColorOutline.Var()}

	case Sunken:
		return Triplet{Bg: css.ColorSurfaceSunken.Var(), Text: css.ColorOnSurface.Var()}
	case SunkenHover:
		return Triplet{Bg: ColorSurfaceHover.Var(), Text: css.ColorOnSurface.Var()}
	case SunkenFocus:
		return Triplet{Bg: ColorSurfaceFocus.Var(), Text: css.ColorOnSurface.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case SunkenPress:
		return Triplet{Bg: ColorSurfacePress.Var(), Text: css.ColorOnSurface.Var()}

	case Accent:
		return Triplet{Bg: css.ColorPrimary.Var(), Text: css.ColorOnPrimary.Var()}
	case AccentHover:
		return Triplet{Bg: ColorPrimaryHover.Var(), Text: css.ColorOnPrimary.Var()}
	case AccentFocus:
		return Triplet{Bg: ColorPrimaryFocus.Var(), Text: css.ColorOnPrimary.Var(), Border: "1px solid " + css.ColorOnPrimary.Var()}
	case AccentPress:
		return Triplet{Bg: ColorPrimaryPress.Var(), Text: css.ColorOnPrimary.Var()}

	case Secondary:
		return Triplet{Bg: css.ColorSecondary.Var(), Text: css.ColorOnSecondary.Var()}
	case SecondaryHover:
		return Triplet{Bg: ColorSecondaryHover.Var(), Text: css.ColorOnSecondary.Var()}
	case SecondaryFocus:
		return Triplet{Bg: ColorSecondaryFocus.Var(), Text: css.ColorOnSecondary.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case SecondaryPress:
		return Triplet{Bg: ColorSecondaryPress.Var(), Text: css.ColorOnSecondary.Var()}

	case Selected:
		return Triplet{Bg: css.ColorSelection.Var(), Text: css.ColorOnSelection.Var()}
	case SelectedHover:
		return Triplet{Bg: ColorSelectionHover.Var(), Text: css.ColorOnSelection.Var()}
	case SelectedFocus:
		return Triplet{Bg: ColorSelectionFocus.Var(), Text: css.ColorOnSelection.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case SelectedPress:
		return Triplet{Bg: ColorSelectionPress.Var(), Text: css.ColorOnSelection.Var()}

	case Success:
		return Triplet{Bg: css.ColorSuccess.Var(), Text: css.ColorOnSuccess.Var()}
	case SuccessHover:
		return Triplet{Bg: ColorSuccessHover.Var(), Text: css.ColorOnSuccess.Var()}
	case SuccessFocus:
		return Triplet{Bg: css.ColorSuccess.Var(), Text: css.ColorOnSuccess.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case SuccessPress:
		return Triplet{Bg: css.ColorSuccess.Var(), Text: css.ColorOnSuccess.Var()}

	case Danger:
		return Triplet{Bg: css.ColorError.Var(), Text: css.ColorOnError.Var()}
	case DangerHover:
		return Triplet{Bg: ColorErrorHover.Var(), Text: css.ColorOnError.Var()}
	case DangerFocus:
		return Triplet{Bg: css.ColorError.Var(), Text: css.ColorOnError.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case DangerPress:
		return Triplet{Bg: css.ColorError.Var(), Text: css.ColorOnError.Var()}

	case Muted:
		return Triplet{Bg: "transparent", Text: css.ColorMuted.Var()}
	case MutedHover:
		return Triplet{Bg: "rgba(0,0,0,0.05)", Text: css.ColorOnSurface.Var()}
	case MutedFocus:
		return Triplet{Bg: "transparent", Text: css.ColorMuted.Var(), Border: "1px solid " + css.ColorPrimary.Var()}
	case MutedPress:
		return Triplet{Bg: "rgba(0,0,0,0.1)", Text: css.ColorOnSurface.Var()}

	case Disabled:
		return Triplet{Bg: css.ColorDisabled.Var(), Text: css.ColorOnDisabled.Var()}

	default:
		return Triplet{}
	}
}
