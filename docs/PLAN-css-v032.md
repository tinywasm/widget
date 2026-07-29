# PLAN — Migrar a tinywasm/css (API de interacción)

Ejecutar antes del plan principal `docs/PLAN.md`. Las restricciones que ningún
paso puede violar están en [../AGENTS.md](../AGENTS.md) — en particular §1
(frontera WASM) y §2 (never invent a value), que son las que este plan roza.

## Prerrequisito: css publica API de interacción

Este plan depende de que `tinywasm/css` publique primero **v0.3.3** (el plan de
css es efímero y se borra al publicar; la referencia es el tag, no el archivo).
Lo que esa versión añade:

- `Hover(t Token) string`, `Focus(t Token) string`, `Press(t Token) string`
- Tokens `ColorSurfaceSunken`, `ColorSelection`, `ColorOnSelection`
- Tokens de intensidad `MixHover`, `MixFocus`, `MixPress` — las funciones los
  referencian por `var()`, así que aparecen en el CSS emitido por widget aunque
  widget no los nombre nunca (impacta §5)

Sin eso, este plan no compila. No definir fórmulas `color-mix()` en widget —
la derivación de estados es responsabilidad de css. Si un paso de este plan
parece necesitar una fórmula nueva, el plan de css está incompleto: corregir allá,
no aquí.

```bash
go get github.com/tinywasm/css@latest
go mod tidy
```

Confirmar antes de seguir que `go list -m github.com/tinywasm/css` reporta
`v0.3.3` o posterior.

## 1. Mapeo de tokens eliminados

| Token removido | Reemplazo |
|---|---|
| `ColorError` / `ColorOnError` | `ColorDanger` / `ColorOnDanger` |
| `ColorSecondary` / `ColorOnSecondary` | `ColorSurface` / `ColorOnSurface` |
| `ColorSurfaceSunken` | `css.ColorSurfaceSunken` (restaurado en css) |
| `ColorSelection` / `ColorOnSelection` | `css.ColorSelection` / `css.ColorOnSelection` (restaurados) |
| `ColorDisabled` / `ColorOnDisabled` | `css.ColorSurface` / `css.ColorMuted` |
| `ColorFocusRing` | `css.ColorPrimary` |
| `ColorHover` | `css.Hover(css.ColorSurface)` |
| `Color*Hover/Focus/Press` (27 tokens) | `css.Hover/Focus/Press(base)` (ver §2) |
| `Source` type, Light/Dark twins | eliminados — usar `Token`, `SetTheme()` |

## 2. `style/emit.go` — `familyTokens()` (líneas 346–363)

Widget solo decide qué token base usa cada superficie. La derivación del estado
la hace css:

```go
// familyBase maps each interactive Surface to its base token.
func familyBase(s Surface) css.Token {
    switch s {
    case Panel, Inset, Highlight:
        return css.ColorSurface
    case Primary:
        return css.ColorPrimary
    case Secondary:
        return css.ColorSurface
    case Success:
        return css.ColorSuccess
    case Danger:
        return css.ColorDanger
    case Subtle:
        return css.ColorMuted
    default:
        return css.Token{}
    }
}
```

### Emisión (líneas 640–656)

```go
addInteractive := func(p widget.Part, r rule) {
    if r.interactive {
        base := familyBase(r.surface)
        if base.Name == "" {
            return
        }
        cueDecls[cueKey{cue: widget.Hover, part: p}] =
            append(cueDecls[cueKey{cue: widget.Hover, part: p}], "background-color: "+css.Hover(base)+";")
        cueDecls[cueKey{cue: widget.Focus, part: p}] =
            append(cueDecls[cueKey{cue: widget.Focus, part: p}], "background-color: "+css.Focus(base)+";")
        cueDecls[cueKey{cue: widget.Press, part: p}] =
            append(cueDecls[cueKey{cue: widget.Press, part: p}], "background-color: "+css.Press(base)+";")
    }
}
```

El valor emitido deja de ser un `var()` simple y pasa a ser una expresión
`color-mix()` con `var()` anidados, incluida la intensidad:

```css
background-color: color-mix(in oklab, var(--color-primary,#1b5d8c), light-dark(black, white) var(--mix-hover,15%));
```

Es correcto y el drift guard lo tolera, siempre que la lista de §5 incluya
`css.MixHover/MixFocus/MixPress` — el guard valida **cada** `var()` emitido, y
`--mix-hover` aparece aunque widget no lo nombre en su código.

## 3. `style/surface.go` — `resolve()` (líneas 58–83)

| Línea | Antes | Después |
|---|---|---|
| 65 | `css.ColorSurfaceSunken.Var()` | `css.ColorSurfaceSunken.Var()` (sin cambio de API — el token vuelve a existir) |
| 68–69 | `css.ColorSecondary` / `ColorOnSecondary` | `css.ColorSurface` / `css.ColorOnSurface` |
| 71 | `css.ColorSelection` / `ColorOnSelection` | `css.ColorSelection` / `css.ColorOnSelection` (restaurados) |
| 75 | `css.ColorError` / `ColorOnError` | `css.ColorDanger` / `css.ColorOnDanger` |
| 79 | `css.ColorDisabled` / `ColorOnDisabled` | `css.ColorSurface` / `css.ColorMuted` |

**Cambio visual esperado, no un defecto:** `Highlight` era naranja sólido
(`#f5a623` con texto oscuro); ahora es un tinte del 15% de `ColorPrimary` sobre
la superficie. `Secondary` deja de tener color propio y se ve como `Panel` — es lo
que dice el mapeo, y la distinción la aporta `Interactive()`. Verificar ambos
visualmente antes de cerrar el plan.

## 4. `style/emit.go` — `Veil()` (línea 310)

⚠️ **Sin este paso `TestNoInventedValues` falla.** La declaración escribe el
`var()` a mano con el fallback de la v0.3.1:

```go
// antes — el fallback literal ya no coincide con css.ColorSurface.Var()
decls = append(decls, "background-color: color-mix(in srgb, var(--color-surface,#F2F2F7) 60%, transparent);")
```

En v0.3.2+ `css.ColorSurface.Var()` es
`var(--color-surface,light-dark(#F2F2F7, #161B22))`, así que el guard reporta
`Visual drift detected for "--color-surface"`. Y el fallback congelado en `#F2F2F7`
es exactamente el hardcodeo que el defecto D-5 cierra. Reemplazar por:

```go
decls = append(decls, "background-color: color-mix(in srgb, "+css.ColorSurface.Var()+" 60%, transparent);")
```

`docs/SPECS.md` (tabla de `Veil()`, línea ~286) ya describe el fallback como
`<fallback>`, así que no requiere cambio.

## 5. `style/consumer_test.go` — guard de drift

### Lista de tokens (líneas 292–320)

Reemplazar por el catálogo actual + los tokens restaurados:

```go
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
}
```

### Test de familia (líneas 184–205)

⚠️ **Segundo test que falla sin este paso.** `TestInteractiveDerivesFamily`
afirma que `Interactive(Primary)` emite `--color-primary-hover/-focus/-press` y
ningún `--color-secondary-hover`. Esos tokens dejan de existir. El defecto que
cierra (D-4: mezclar familias) sigue siendo válido, así que se reescribe la
aserción contra la derivación, no contra los nombres:

```go
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
```

La comprobación negativa cambia de familia (`Success` en vez de `Secondary`)
porque `Secondary` ahora deriva de `ColorSurface` y su expresión de hover es
idéntica a la de `Panel` — un `!Contains` sobre ella no probaría nada.

### Excepción muerta (línea ~375)

`--color-hover` ya no existe; borrar la tolerancia y su comentario para que el
guard vuelva a ser estricto:

```go
// antes
if fullMatch != expectedVarCall && varName != "--color-hover" {
// después
if fullMatch != expectedVarCall {
```

### No modificar el mecanismo del guard

Verificado sobre el código actual: `extractVarCalls` (línea 220) balancea
paréntesis, así que un `var(--color-surface-sunken, color-mix(… var(…) …))`
anidado se extrae completo e iguala `tok.Var()` exacto; y el regex anti-hex borra
cada `var()` antes de buscar `#`, por lo que las expresiones `color-mix()` de §2
pasan sin tocar nada. Si el guard falla, la causa es §4, la lista de tokens o la
excepción muerta — no el extractor.

Nota sobre `--mix-*`: el extractor sigue escaneando después de cada `var()`
cerrado, así que detecta también la intensidad al final de la expresión. Si
reporta `CSS variable "--mix-hover" not in css catalog`, faltan los tres tokens
en la lista de arriba.

## 6. Documentación — mismo commit

`docs/PLAN.md` fija la regla: si el código y SPECS discrepan, uno de los dos está
mal y se corrigen juntos. Esta migración cambia lo que se emite, así que SPECS
cambia:

- **§3, tabla de superficies** (líneas ~152-163): `Secondary` → `--color-surface`
  / `--color-on-surface`; `Danger` → `--color-danger` / `--color-on-danger`;
  `Inactive` → `--color-surface` / `--color-muted`. `Inset` y `Highlight` no
  cambian de nombre de token.
- **§3.1 Interaction families** (líneas ~167-181): el patrón
  `--color-<family>-hover|focus|press` **desaparece**. Los estados pasan a ser
  `color-mix()` derivados de un token base — describir la tabla en esos términos
  y nombrar `css.Hover/Focus/Press`.
- **§3.1, nota de `Subtle`**: "resolves to `--color-hover` so it remains visible
  in dark mode" queda obsoleta y su premisa ya no aplica. `Subtle` deriva de
  `ColorMuted` y el mezclador es `light-dark(black, white)`, así que ninguna
  familia produce un black wash en tema oscuro — la excepción deja de existir.

También:

- `docs/MIGRATION.md` línea ~16: la entrada dice que los tokens de familia "viven
  en widget/style con hex hardcodeado, fuera del alcance del contrast test".
  Actualizar: ya no viven aquí, y el hueco de contraste que queda está
  documentado en el SPECS de css, no en el de widget.
- `../AGENTS.md` §2: el bullet "do not define a `color-mix()` formula here" queda
  confirmado por esta migración — verificar que sigue describiendo el código, ya
  que §4 de este plan elimina el último `var()` escrito a mano.
- `docs/DESIGN.md` línea ~336 y `docs/MIGRATION.md` línea ~94: mencionan
  `--color-disabled` como justificación del nombre `Inactive`. El nombre se
  queda; corregir la justificación, que ya no puede apoyarse en un token
  inexistente.

## 7. Verificación

```bash
gotest
```

`vet ✅, race ✅, tests ✅, wasm ✅` sin errores.

- `TestNoInventedValues` pasa: los `color-mix()` emitidos referencian solo tokens
  css, sin hex ni `rgba()` literales fuera de `var()`.
- `TestInteractiveDerivesFamily` pasa con las aserciones reescritas de §5, y
  sigue cerrando D-4: una familia no filtra la derivación de otra.
- La aserción de frontera WASM (consumer_test.go:97) sigue valiendo: con
  `GOOS=js` el grafo de `github.com/tinywasm/widget` no contiene `widget/style`,
  y por tanto tampoco `css`. Toda la derivación nueva vive en `widget/style`, que
  ya es `//go:build !wasm` — este plan no toca ningún archivo del paquete raíz.
- Revisión visual en ambos temas de: `Inset`, `Highlight`, `Secondary` y los tres
  estados de `Interactive()` — `Press` (45%) es el caso extremo de contraste y
  queda fuera del audit hex-based de css.

Al terminar, continuar con `docs/PLAN.md`.
