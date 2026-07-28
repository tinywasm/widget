//go:build !wasm

package style

import (
	"strings"
	"testing"

	"github.com/tinywasm/widget"
)

func TestMotion_AnimateSlow(t *testing.T) {
	// 1. Animate(MotionSlow) on Root emits transition: all var(--duration-slow) var(--ease-in-out); and no literal 400ms.
	sheet := Of("panel").Root(Animate(MotionSlow))
	cssOut := sheet.Stylesheet().String()

	// Verify it contains the CSS variables rather than hardcoded transitions
	if !strings.Contains(cssOut, "var(--duration-slow") || !strings.Contains(cssOut, "var(--ease-in-out") {
		t.Errorf("Expected slow transition declaration using CSS variables, got:\n%s", cssOut)
	}

	// Ensure we don't have "all 400ms" or literal "400ms" as a standalone duration (outside variable fallbacks)
	// If it contains "transition: all 400ms", that's wrong.
	if strings.Contains(cssOut, "transition: all 400ms") || strings.Contains(cssOut, "transition: 400ms") {
		t.Errorf("Transition should not contain literal 400ms duration, got:\n%s", cssOut)
	}
}

func TestMotion_AnimateNone(t *testing.T) {
	// 2. Animate(MotionNone) emits transition: none;.
	sheet := Of("panel").Root(Animate(MotionNone))
	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, "transition: none;") {
		t.Errorf("Expected 'transition: none;', got:\n%s", cssOut)
	}
}

func TestMotion_NoAnimate(t *testing.T) {
	// 3. A sheet with no Animate call contains no prefers-reduced-motion block and no transition: declaration at all.
	sheet := Of("panel")
	cssOut := sheet.Stylesheet().String()

	if strings.Contains(cssOut, "prefers-reduced-motion") {
		t.Errorf("Expected no prefers-reduced-motion block when no motion is defined, got:\n%s", cssOut)
	}
	if strings.Contains(cssOut, "transition:") {
		t.Errorf("Expected no transition declarations, got:\n%s", cssOut)
	}
}

func TestMotion_OnlyInWhen(t *testing.T) {
	// 4. A sheet whose motion is declared only inside When(widget.Open, "", Animate(MotionSlow))
	// still emits the prefers-reduced-motion block, and that block's selector is the state selector.
	sheet := Of("panel").When(widget.Open, "", Animate(MotionSlow))
	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, "@media (prefers-reduced-motion: reduce) {") {
		t.Fatalf("Expected prefers-reduced-motion block, got:\n%s", cssOut)
	}

	// Verify selector is state selector in reduced-motion
	expectedBlock := ".panel[data-open=\"true\"] {\n  transition: none;\n}"
	if !strings.Contains(cssOut, expectedBlock) {
		t.Errorf("Expected reduced motion selector block:\n%s\n\nGot:\n%s", expectedBlock, cssOut)
	}
}

func TestMotion_Determinism(t *testing.T) {
	// 5. Two consecutive Stylesheet() calls on equivalent sheets produce byte-identical output (extends the determinism guarantee to the new block).
	s1 := Of("panel").Root(Animate(MotionSlow)).When(widget.Open, "", Animate(MotionBase)).Cue(widget.Hover, "item", Animate(MotionFast))
	s2 := Of("panel").Root(Animate(MotionSlow)).When(widget.Open, "", Animate(MotionBase)).Cue(widget.Hover, "item", Animate(MotionFast))

	out1 := s1.Stylesheet().String()
	out2 := s2.Stylesheet().String()

	if out1 != out2 {
		t.Errorf("Stylesheet generation is non-deterministic.\n\nCSS 1:\n%s\n\nCSS 2:\n%s", out1, out2)
	}
}
