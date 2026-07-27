# Architecture of `tinywasm/widget`

This document defines the **What** and **Why** of visual contracts, anatomy, states, layout, and the style system for components in the `tinywasm/widget` architecture.

---

## What is `tinywasm/widget`

`tinywasm/widget` is the module governing the structure of visual components. It names and unifies the pieces a visual component consists of, the states it can occupy, and how those pieces are laid out—completely independent of data, transport, DOM, or free CSS text.

The system is strictly divided into two packages with opposing goals and constraints:

1.  **`github.com/tinywasm/widget` (Root, WASM-compatible):**
    *   **Purpose:** Defines identity, component anatomy (Open UI), states, browser cues, and ARIA kinds.
    *   **Constraint:** Must be extremely lightweight (zero style logic or emission) because it **travels inside the WASM binary**.

2.  **`github.com/tinywasm/widget/style` (Stylesheet engine, WASM-exempt):**
    *   **Purpose:** Defines closed scales for sizing, typography, elevation, color surfaces, flow primitives, and layout exceptions. It produces deterministic CSS stylesheets inside fixed cascade layers.
    *   **Constraint:** Excluded from compiling on WebAssembly using the `//go:build !wasm` build constraint to prevent it from ever traveling to the client binary.

---

## Core Structure and Components

### 1. Anatomy and States Contract (Root Package)

*   **`Name`:** Identifies a widget. It is used as a prefix for all its emitted classes and selectors, preventing name collisions.
*   **`Part`:** A named local slot of a widget anatomy (e.g., `"row"`, `"menu"`, `"header"`).
*   **`Class`:** A CSS class name derived deterministically from a `Name` and a `Part` (e.g., `.targetlist__row`). It is an opaque type without a public constructor outside the package, ensuring naming consistency by construction.
*   **`State` vs `Cue`:**
    *   `State` represents logical states owned by the widget in Go (such as `Selected`, `Disabled`, `Open`, `Invalid`). It maps directly to DOM data attributes (`data-selected="true"`) by construction.
    *   `Cue` represents interactive browser-only states (such as `Hover`, `Focus`, `Press`, `Target`) that are styled in the stylesheet via CSS pseudo-classes (`:hover`, `:focus`).

---

## Stylesheet and Layout System (`widget/style`)

The stylesheet engine is designed under the **zero escape** principle: it does not accept free strings, viewport units (`vw`/`vh`), or arbitrary custom styling utilities.

### 1. Closed Scales and Enums
All visual properties are specified through strict, typed scales:
*   `Space`: System spacing scale (0 to 12).
*   `Radius`: Border radius options (`RadiusSm`, `RadiusMd`, `RadiusLg`, `RadiusFull`).
*   `TextSize` & `Weight`: A closed typography system.
*   `Elevation`: Shadows representing height (`Flat`, `Raised`, `Floating`, `Overlay`).
*   `Size`: Relative container-based width options (`Content`, `Prose`, `Third`, `Half`, `TwoThirds`, `Full`). Height cannot be directly declared, promoting automatic content flow.

### 2. Colors as Complete Decisions: `Surface`
Arbitrary color selections for backgrounds or text are not allowed. A `Surface` unifies background, text color, and borders into a single cohesive visual decision (e.g., `Page`, `Panel`, `Sunken`, `Accent`), ensuring proper contrast ratio by design.

### 3. Responsive Flow Primitives (`Flow`)
Inspired by *Every Layout*, it defines primitives like `Stack`, `Row`, `Split`, and `Grid`. These are inherently responsive and use container queries (`@container`) instead of media queries (`@media`), reacting directly to their own container's width.

---

## CSS Emission Guarantees

The stylesheet generator (`emit.go`) outputs CSS conforming to the following guarantees:

1.  **Fixed Layer Order (Cascade Layers):**
    ```css
    @layer tokens, primitives, widgets, states;
    ```
    This completely eliminates specificity conflicts and avoids any use of `!important`.
2.  **Stable Output:** Generating the stylesheet twice yields byte-for-byte identical output, keeping version control diffs clean.
3.  **Flat Specificity:** All generated widget rules have a flat specificity (e.g., `.class` or `.class[data-state]`), avoiding coupling to the DOM structure.
