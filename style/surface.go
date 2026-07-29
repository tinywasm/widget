//go:build !wasm

package style

import "github.com/tinywasm/css"

// Surface is a complete visual decision: background, text, and border resolved together.
type Surface uint8

const (
	Page Surface = iota
	Panel
	Inset
	Primary
	Secondary
	Highlight
	Success
	Danger
	Subtle
	Inactive
)

func (s Surface) String() string {
	switch s {
	case Page:
		return "Page"
	case Panel:
		return "Panel"
	case Inset:
		return "Inset"
	case Primary:
		return "Primary"
	case Secondary:
		return "Secondary"
	case Highlight:
		return "Highlight"
	case Success:
		return "Success"
	case Danger:
		return "Danger"
	case Subtle:
		return "Subtle"
	case Inactive:
		return "Inactive"
	default:
		return "Unknown"
	}
}

// triplet represents the combination of background, text, and border.
type triplet struct {
	bg     string
	text   string
	border string
}

// resolve translates a Surface to its corresponding triplet of styles.
func (s Surface) resolve() triplet {
	switch s {
	case Page:
		return triplet{bg: css.ColorBackground.Var(), text: css.ColorOnSurface.Var()}
	case Panel:
		return triplet{bg: css.ColorSurface.Var(), text: css.ColorOnSurface.Var(), border: "1px solid " + css.ColorOutline.Var()}
	case Inset:
		return triplet{bg: css.ColorSurfaceSunken.Var(), text: css.ColorOnSurface.Var(), border: "1px solid " + css.ColorOutline.Var()}
	case Primary:
		return triplet{bg: css.ColorPrimary.Var(), text: css.ColorOnPrimary.Var()}
	case Secondary:
		return triplet{bg: css.ColorSurface.Var(), text: css.ColorOnSurface.Var()}
	case Highlight:
		return triplet{bg: css.ColorSelection.Var(), text: css.ColorOnSelection.Var()}
	case Success:
		return triplet{bg: css.ColorSuccess.Var(), text: css.ColorOnSuccess.Var()}
	case Danger:
		return triplet{bg: css.ColorDanger.Var(), text: css.ColorOnDanger.Var()}
	case Subtle:
		return triplet{bg: "transparent", text: css.ColorMuted.Var()}
	case Inactive:
		return triplet{bg: css.ColorSurface.Var(), text: css.ColorMuted.Var()}
	default:
		return triplet{}
	}
}

// defaultRadius returns the default radius associated with a Surface.
func (s Surface) defaultRadius() Radius {
	switch s {
	case Page, Subtle:
		return RadiusNone
	case Panel:
		return RadiusMd
	default:
		return RadiusSm
	}
}

// Interactive applies s and derives its hover, focus, and press treatments.
func Interactive(s Surface) Option {
	return func(r *rule) {
		r.hasSurface, r.surface, r.interactive = true, s, true
	}
}
