# PLAN — Migrar a tinywasm/css v0.3.2

Ejecutar antes del plan principal `docs/PLAN.md`.

## Prerrequisito

```bash
go get github.com/tinywasm/css@v0.3.2
go mod tidy
```

## 1. Mapeo de tokens eliminados

| Token removido (v0.2) | Reemplazo en v0.3.2 |
|---|---|
| `ColorError` / `ColorOnError` | `ColorDanger` / `ColorOnDanger` |
| `ColorSecondary` / `ColorOnSecondary` | usar `ColorPrimary` o `ColorSurface` según contexto |
| `ColorSurfaceSunken` | `color-mix(in oklab, var(--color-surface), var(--color-background))` |
| `ColorSelection` / `ColorOnSelection` | `color-mix(in oklab, var(--color-primary), transparent 85%)` / `ColorOnSurface` |
| `ColorDisabled` / `ColorOnDisabled` | `ColorSurface` / `ColorMuted` |
| `ColorFocusRing` | `ColorPrimary` |
| `ColorHover` | `color-mix(in oklab, var(--color-surface), black 8%)` |
| `Color*Hover/Focus/Press` (27 tokens) | reemplazar por `color-mix()` inline (ver §2) |
| `ColorBackgroundLight/Dark` y demás Light/Dark twins | eliminados — usar `SetTheme()` para overrides |
| `Source` type | eliminado — usar `Token` |

## 2. `style/emit.go` — `familyTokens()` (líneas 346–363)

Los tokens Hover/Focus/Press ya no existen. `familyTokens()` debe devolver strings con `color-mix()` en lugar de `css.Token`.

### Nuevo diseño

```go
func interactionColor(s Surface, cue widget.Cue) string {
    switch s {
    case Panel, Inset, Highlight:
        base := css.ColorSurface.Var()
        switch cue {
        case widget.Hover:  return "color-mix(in oklab, " + base + ", black 8%)"
        case widget.Focus:  return "color-mix(in oklab, " + base + ", black 12%)"
        case widget.Press:  return "color-mix(in oklab, " + base + ", black 16%)"
        }
    case Primary:
        base := css.ColorPrimary.Var()
        switch cue {
        case widget.Hover:  return "color-mix(in oklab, " + base + ", black 12%)"
        case widget.Focus:  return "color-mix(in oklab, " + base + ", black 24%)"
        case widget.Press:  return "color-mix(in oklab, " + base + ", black 36%)"
        }
    case Secondary:
        base := css.ColorSurface.Var()
        switch cue {
        case widget.Hover:  return "color-mix(in oklab, " + base + ", black 6%)"
        case widget.Focus:  return "color-mix(in oklab, " + base + ", black 10%)"
        case widget.Press:  return "color-mix(in oklab, " + base + ", black 14%)"
        }
    case Success:
        base := css.ColorSuccess.Var()
        switch cue {
        case widget.Hover:  return "color-mix(in oklab, " + base + ", black 12%)"
        case widget.Focus:  return "color-mix(in oklab, " + base + ", black 24%)"
        case widget.Press:  return "color-mix(in oklab, " + base + ", black 36%)"
        }
    case Danger:
        base := css.ColorDanger.Var()
        switch cue {
        case widget.Hover:  return "color-mix(in oklab, " + base + ", black 12%)"
        case widget.Focus:  return "color-mix(in oklab, " + base + ", black 24%)"
        case widget.Press:  return "color-mix(in oklab, " + base + ", black 36%)"
        }
    case Subtle:
        base := css.ColorMuted.Var()
        switch cue {
        case widget.Hover:  return "color-mix(in oklab, " + base + ", transparent 60%)"
        case widget.Focus:  return "color-mix(in oklab, " + base + ", transparent 40%)"
        case widget.Press:  return "color-mix(in oklab, " + base + ", transparent 20%)"
        }
    }
    return ""
}
```

### Cambio en la emisión (líneas 640–656)

El bloque que usa `familyTokens` ahora debe usar `interactionColor` directamente:

```go
addInteractive := func(p widget.Part, r rule) {
    if r.interactive {
        if s := interactionColor(r.surface, widget.Hover); s != "" {
            k := cueKey{cue: widget.Hover, part: p}
            cueDecls[k] = append(cueDecls[k], "background-color: "+s+";")
        }
        if s := interactionColor(r.surface, widget.Focus); s != "" {
            k := cueKey{cue: widget.Focus, part: p}
            cueDecls[k] = append(cueDecls[k], "background-color: "+s+";")
        }
        if s := interactionColor(r.surface, widget.Press); s != "" {
            k := cueKey{cue: widget.Press, part: p}
            cueDecls[k] = append(cueDecls[k], "background-color: "+s+";")
        }
    }
}
```

## 3. `style/surface.go` — `resolve()` (líneas 58–83)

Reemplazar tokens obsoletos:

| Línea | Antes | Después |
|---|---|---|
| 65 | `css.ColorSurfaceSunken.Var()` | `"color-mix(in oklab, var(--color-surface,"+css.ColorSurface.Dark+"), var(--color-background,"+css.ColorBackground.Dark+"))"` |
| 68 | `css.ColorSecondary.Var()` | `css.ColorSurface.Var()` |
| 69 | `css.ColorOnSecondary.Var()` | `css.ColorOnSurface.Var()` |
| 71 | `css.ColorSelection.Var()` | `"color-mix(in oklab, var(--color-primary,"+css.ColorPrimary.Dark+"), transparent 85%)"` |
| 71 | `css.ColorOnSelection.Var()` | `css.ColorOnSurface.Var()` |
| 75 | `css.ColorError.Var()` | `css.ColorDanger.Var()` |
| 75 | `css.ColorOnError.Var()` | `css.ColorOnDanger.Var()` |
| 79 | `css.ColorDisabled.Var()` | `css.ColorSurface.Var()` |
| 79 | `css.ColorOnDisabled.Var()` | `css.ColorMuted.Var()` |

## 4. `style/consumer_test.go` — lista de tokens (líneas 292–320)

Reemplazar la lista completa de `cssTokens` por el catálogo actual de v0.3.2:

```go
cssTokens := []css.Token{
    css.ColorPrimary, css.ColorOnPrimary,
    css.ColorSuccess, css.ColorOnSuccess,
    css.ColorDanger, css.ColorOnDanger,
    css.ColorBackground, css.ColorOnBackground,
    css.ColorSurface, css.ColorOnSurface,
    css.ColorOutline, css.ColorMuted,
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
}
```

## 5. Verificación

```bash
gotest
```

`vendor ✅, race ✅, tests ✅, wasm ✅` sin errores.
