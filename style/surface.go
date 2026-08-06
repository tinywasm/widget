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
	Accent
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
	case Accent:
		return "Accent"
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
//
// The Static fields are the browser-safe counterparts of bg/text/border —
// see css.Token.LightValue(). emit_decls.go emits each one as the first of a
// double declaration, immediately before its light-dark()-enhanced sibling.
type triplet struct {
	bg     string
	text   string
	border string

	bgStatic     string
	textStatic   string
	borderStatic string
}

// resolve translates a Surface to its corresponding triplet of styles.
func (s Surface) resolve() triplet {
	switch s {
	case Page:
		return triplet{
			bg: css.ColorBackground.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(),
			bgStatic: css.ColorBackground.LightValue(), textStatic: css.ColorOnSurface.LightValue(),
		}
	case Panel:
		return triplet{
			bg: css.ColorSurface.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(), border: "1px solid " + css.ColorOutline.EnhancedVar(),
			bgStatic: css.ColorSurface.LightValue(), textStatic: css.ColorOnSurface.LightValue(), borderStatic: "1px solid " + css.ColorOutline.LightValue(),
		}
	case Inset:
		return triplet{
			bg: css.ColorSurfaceSunken.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(), border: "1px solid " + css.ColorOutline.EnhancedVar(),
			bgStatic: css.ColorSurfaceSunken.LightValue(), textStatic: css.ColorOnSurface.LightValue(), borderStatic: "1px solid " + css.ColorOutline.LightValue(),
		}
	case Primary:
		return triplet{
			bg: css.ColorPrimary.EnhancedVar(), text: css.ColorOnPrimary.EnhancedVar(),
			bgStatic: css.ColorPrimary.LightValue(), textStatic: css.ColorOnPrimary.LightValue(),
		}
	case Secondary:
		return triplet{
			bg: css.ColorSurface.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(),
			bgStatic: css.ColorSurface.LightValue(), textStatic: css.ColorOnSurface.LightValue(),
		}
	case Highlight:
		return triplet{
			bg: css.ColorSelection.EnhancedVar(), text: css.ColorOnSelection.EnhancedVar(),
			bgStatic: css.ColorSelection.LightValue(), textStatic: css.ColorOnSelection.LightValue(),
		}
	case Accent:
		return triplet{
			bg: css.ColorAccent.EnhancedVar(), text: css.ColorOnAccent.EnhancedVar(),
			bgStatic: css.ColorAccent.LightValue(), textStatic: css.ColorOnAccent.LightValue(),
		}
	case Success:
		return triplet{
			bg: css.ColorSuccess.EnhancedVar(), text: css.ColorOnSuccess.EnhancedVar(),
			bgStatic: css.ColorSuccess.LightValue(), textStatic: css.ColorOnSuccess.LightValue(),
		}
	case Danger:
		return triplet{
			bg: css.ColorDanger.EnhancedVar(), text: css.ColorOnDanger.EnhancedVar(),
			bgStatic: css.ColorDanger.LightValue(), textStatic: css.ColorOnDanger.LightValue(),
		}
	case Subtle:
		return triplet{
			bg: "transparent", text: css.ColorMuted.EnhancedVar(),
			bgStatic: "transparent", textStatic: css.ColorMuted.LightValue(),
		}
	case Inactive:
		return triplet{
			bg: css.ColorSurface.EnhancedVar(), text: css.ColorMuted.EnhancedVar(),
			bgStatic: css.ColorSurface.LightValue(), textStatic: css.ColorMuted.LightValue(),
		}
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
