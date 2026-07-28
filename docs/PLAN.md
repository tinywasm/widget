---
PLAN: "widget/style: vocabulario de overlay (backdrop, scrim, apilado y visibilidad por estado)"
TAG: v0.3.0
EXECUTOR: jules
STATUS: running
SESSION: 107778640786533136
---

> Este plan se despacha con el flujo CodeJob. Ver skill: **agents-workflow**.

# Plan — `tinywasm/widget` v0.3.0: vocabulario de overlay

## ⚠️ 0. Alcance — LEE ESTO ANTES DE TOCAR NADA

Cambio **puramente aditivo** sobre el subpaquete `style/`. Cinco constructores nuevos y dos
enums cerrados. No se modifica ni un símbolo existente.

**PROHIBIDO:**

| Prohibición | Motivo |
|---|---|
| Tocar el paquete raíz (`widget.go`, `kind.go`, `state.go`, `capability.go`, `field.go`) | Este plan es solo de emisión visual. El raíz solo depende de `tinywasm/fmt` y así se queda. |
| Importar algo que no sea `github.com/tinywasm/css`, `github.com/tinywasm/widget` y stdlib en `style/` | `style/` ya depende de esos dos; no añadas más. |
| Quitar el `//go:build !wasm` de cualquier archivo de `style/` | `style/` no puede entrar al binario WASM. |
| Cambiar el valor de cualquier constante existente | v0.3.0 es aditiva. |
| Exponer un constructor público de `Class` | Sigue prohibido: la única forma de obtener una `Class` es derivarla de un `Name`. |
| Añadir un escape hatch (una `Opt` que reciba CSS crudo) | Es exactamente lo que este plan existe para eliminar. |

---

## 1. El hueco, con evidencia

`tinywasm/components` tuvo que **reabrir el escape hatch** porque este vocabulario no existe.
Está hoy en `main`, en dos componentes, con un comentario que lo admite:

```go
// components/targetlist/css.go
func (t *TargetList) RenderCSS() *css.Stylesheet {
	return css.NewStylesheet(css.Raw(
		backdrop + "{display:none;position:fixed;top:0;left:0;right:0;bottom:0;z-index:4;}" +
			root + ":has(" + menu + "[open]) " + backdrop + "{display:block;}",
	))
}

// components/modaldialog/css.go
	backdrop + "{position:absolute;top:0;left:0;width:100%;height:100%;" +
		"background-color:color-mix(in srgb, var(--color-surface) 60%, transparent);z-index:1;}"
```

Peor: al intentar expresarlo con el vocabulario existente se usó `style.Fixed()`, que **no es
posicionamiento** — es la excepción *no-reflota* (`flex-shrink: 0; flex-grow: 0`). El resultado
fue un backdrop siempre visible en el flujo y el menú ⋮ que dejó de cerrarse al hacer clic
fuera. Un fallo silencioso, que es justo lo que el arnés prohíbe.

Las dos necesidades reales son:

| Componente | Qué necesita |
|---|---|
| `targetlist` | Cazaclics a **viewport** completo, invisible, oculto salvo mientras un menú ⋮ está abierto. Debe quedar **por encima** de las filas y **por debajo** del desplegable. |
| `modaldialog` | Velo atenuado que cubre **el ancestro posicionado**, translúcido, por debajo del panel del diálogo. |

---

## 2. La decisión de diseño (ya tomada, no la reabras)

### 2.1 Nada de `:has()` — la visibilidad la escribe Go como estado

El CSS actual usa `:has(.targetlist__menu[open])` para mostrar el backdrop. **No se añade
vocabulario para `:has()`.** Es innecesario y contradice el principio de la casa, escrito en
`state.go`:

> *"State es un estado que POSEE el widget: lo escribe Go, lo lee la hoja de estilos."*

`targetlist` ya sabe en Go cuándo hay un menú abierto (tiene `closeAllMenus()`). Poner
`widget.Open` sobre el backdrop y dejar que la hoja seleccione por `[data-open="true"]` es más
simple, no necesita `:has()` (que además no está soportado en todos los navegadores antiguos),
y **es el patrón que el resto del sistema ya usa**.

Por eso este plan aporta `Hidden()` y `Shown()`: la regla base oculta, y una regla de estado
(`When(widget.Open, …)`) muestra.

### 2.2 Dos niveles de apilado, nombrados por rol

No se expone un `z-index` numérico — sería un valor, no una intención. Se exponen exactamente
los dos niveles que el problema real necesita:

- `Overlay(scope)` — la capa que **respalda** algo (backdrop, scrim). Por encima del contenido.
- `Above()` — lo que se apoya **sobre** un overlay (menú desplegable, panel del diálogo).

Orden garantizado: contenido normal < `Overlay` < `Above`.

---

## Etapa 1 — `style/overlay.go` (archivo NUEVO)

Archivo: `style/overlay.go`. **Con `//go:build !wasm`**, igual que el resto de `style/`.

```go
//go:build !wasm

package style

// Scope dice contra qué se dimensiona un overlay.
type Scope uint8

const (
	// Parent cubre el ancestro posicionado más cercano (position: absolute).
	Parent Scope = iota
	// Viewport cubre toda la ventana (position: fixed).
	Viewport
)

// Overlay saca el elemento del flujo y lo estira sobre todo su Scope. Es la capa
// que respalda a otra: un cazaclics o un velo. Queda por encima del contenido
// normal y por debajo de cualquier Above().
func Overlay(s Scope) Opt

// Above coloca el elemento por encima de cualquier Overlay: el desplegable de un
// menú, el panel de un diálogo. No lo saca del flujo por sí solo.
func Above() Opt

// Scrim rellena el elemento con un velo translúcido sobre la superficie de la
// página. Solo tiene sentido junto a Overlay.
func Scrim() Opt

// Hidden oculta el elemento. Es la regla base de un overlay que solo debe verse
// en cierto estado; se revierte con Shown() en una regla When().
func Hidden() Opt

// Shown revierte Hidden(). Va en una regla de estado, nunca en la base.
func Shown() Opt
```

### 1.1 CSS que debe emitir cada uno

Estos valores son el contrato. Emítelos exactamente:

| Constructor | Declaraciones |
|---|---|
| `Overlay(Parent)` | `position: absolute; inset: 0; z-index: 100;` |
| `Overlay(Viewport)` | `position: fixed; inset: 0; z-index: 100;` |
| `Above()` | `z-index: 101;` |
| `Scrim()` | `background-color: color-mix(in srgb, var(--color-surface) 60%, transparent);` |
| `Hidden()` | `display: none;` |
| `Shown()` | `display: block;` |

`inset: 0` sustituye a `top/left/right/bottom: 0` — es la forma moderna y equivalente.

### 1.2 Integración con `Rule` y `emit.go`

Sigue el patrón exacto que ya usan `Fill()`, `Scrolls()`, `Fixed()`, `Flush()` y `Clip()` en
`style/except.go` y `style/emit.go`:

1. Añade los campos al struct `Rule` en `style/sheet.go`:
   ```go
   HasOverlay   bool
   OverlayScope Scope
   Above        bool
   Scrim        bool
   Hidden       bool
   Shown        bool
   ```
2. Cada constructor es una `Opt` que activa su campo.
3. En `emit.go`, emítelos en `Rule.Decls()` junto a las demás declaraciones.

**Importante:** a diferencia de `Fill()`/`Scrolls()`, que se emiten como clases compartidas en
`@layer primitives` (`.exc-fill`, `.exc-scrolls`), estos van en `Rule.Decls()` — es decir, en la
regla del propio widget dentro de `@layer widgets` y `@layer states`. Motivo: `Shown()` tiene que
poder ganarle a `Hidden()` desde una regla de estado, y eso lo resuelve el orden de capas
(`states` después de `widgets`), no la especificidad.

---

## Etapa 2 — Tests

Archivo: `style/overlay_test.go` (paquete `style`, junto al código; este repo no tiene `tests/`).

Assertions de stdlib solamente (`if got != want { t.Errorf(...) }`). Ejecuta con `gotest`.

1. `Of("w").Root(Overlay(Viewport))` emite `position: fixed` **y** `inset: 0` **y**
   `z-index: 100`.
2. `Of("w").Root(Overlay(Parent))` emite `position: absolute` (no `fixed`).
3. `Of("w").Part("p", Above())` emite `z-index: 101`.
4. `Of("w").Root(Scrim())` emite un `background-color` con `color-mix(` y `var(--color-surface)`,
   y **no** contiene ningún literal hexadecimal.
5. **El caso que motiva el plan:** una hoja con
   `.Part(p, Overlay(Viewport), Hidden()).When(widget.Open, p, Shown())` emite `display: none`
   dentro de `@layer widgets` y `display: block` dentro de `@layer states`, **en ese orden**, y
   el selector de estado es `.w__p[data-open="true"]`.
6. La hoja emitida no contiene `:has(`.

---

## 3. Lo que este plan NO hace

- **No** añade vocabulario para `:has()` ni para ningún selector relacional (§2.1).
- **No** añade transiciones ni animaciones. `layout/platformd` las necesitará
  (`--pd-slide-duration`); es un hueco distinto y va en su propio plan.
- **No** toca el paquete raíz `widget`.
- **No** modifica `Fill`, `Scrolls`, `Fixed`, `Flush` ni `Clip`. En particular, **`Fixed()`
  sigue significando *no-reflota***; si su nombre resulta confuso frente a `position: fixed`, eso
  es un renombrado que se decide aparte, no aquí.

---

## 4. Consumidor (OTRO repo — no lo toques desde aquí)

`tinywasm/components` borrará sus dos `RenderCSS()` con `css.Raw` en cuanto esta versión se
publique, y cerrará la puerta en su test de conformidad.
Plan: <https://github.com/tinywasm/components/blob/main/docs/PLAN.md>

**No incluyas esos cambios en este PR.**

---

## 5. Criterios de aceptación — verificables con grep

1. `gotest` en verde, incluidos los seis tests nuevos.
2. `ls style/overlay.go style/overlay_test.go` → ambos existen.
3. `head -1 style/overlay.go` → `//go:build !wasm`.
4. `go list -deps . | grep tinywasm` → exactamente dos líneas:
   ```
   github.com/tinywasm/fmt
   github.com/tinywasm/widget
   ```
   (el paquete raíz sigue sin ganar dependencias).
5. `GOOS=js GOARCH=wasm go build ./...` compila, y
   `GOOS=js GOARCH=wasm go list -deps ./... | grep widget/style` → **vacío**.
6. `git diff --stat v0.2.0 -- widget.go kind.go state.go capability.go field.go` → **vacío**.
7. `grep -rn ":has(" style/` → **vacío**.
8. `grep -rnE '#[0-9a-fA-F]{3,6}' style/overlay.go` → **vacío**.

---

## 6. Checklist de calidad Go (obligatorio)

- **Sin strings repetidos**: cada declaración CSS se escribe una vez.
- **Sin stdlib prohibida**: en `style/` se permite `fmt`/`sort`/`strings` de stdlib porque es
  código `!wasm` (ya lo hace `emit.go`); en el paquete raíz, solo `tinywasm/fmt`.
- **Cero `any`, cero `map`** en API nueva.
- **Enums cerrados**: `Scope` es un `uint8` con constantes; nada de strings.
- Comentarios de código en inglés, según la convención del repo.

---

## 7. Tabla de etapas

| # | Etapa | Archivos | Gate |
|---|---|---|---|
| 1 | Vocabulario | `style/overlay.go` (nuevo), `style/sheet.go`, `style/emit.go` | `go build ./...` |
| 2 | Tests | `style/overlay_test.go` (nuevo) | `gotest` verde |

Secuenciales. La 2 es el gate.
