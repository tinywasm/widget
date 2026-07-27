---
PLAN: "widget: añadir la anatomía compartida del campo de formulario (NameField + partes)"
TAG: v0.2.0
EXECUTOR: jules
STATUS: running
SESSION: 2580046364401006697
---

> Este plan se despacha con el flujo CodeJob. Ver skill: **agents-workflow**.
> El plan que creó la v0.1.0 está en [`PLAN_EXECUTED.md`](PLAN_EXECUTED.md) — ya ejecutado, no
> lo reabras.

# Plan — `tinywasm/widget` v0.2.0: la anatomía del campo de formulario

## ⚠️ 0. Alcance — LEE ESTO ANTES DE TOCAR NADA

Este cambio es **puramente aditivo**: un archivo nuevo con cinco constantes. No modifica ni un
símbolo existente. Si te ves cambiando algo que ya estaba, te saliste del plan.

**PROHIBIDO:**

| Prohibición | Motivo |
|---|---|
| Importar cualquier cosa que no sea `github.com/tinywasm/fmt` en el paquete raíz | Es **la** propiedad que justifica que `widget` sea un repo aparte: `form` y `view` (librerías de datos) pueden depender de él sin arrastrar una librería de estilo. Romperla invalida el diseño entero. |
| Tocar el subpaquete `style/` | Va con `//go:build !wasm` y no interviene aquí. |
| Añadir un constructor público de `Class` | `Class` no tiene constructor a propósito: la única forma de obtener una es derivarla de un `Name`. `Class("lo-que-sea")` **no debe compilar** fuera del paquete. Es el arnés. |
| Cambiar el valor de cualquier constante existente | v0.2.0 es aditiva. Cambiar un valor rompe a todos los consumidores en silencio. |
| Añadir un `Kind` nuevo | `Form` ya existe en `kind.go`. No hace falta ninguno. |
| Añadir `WidgetName()`/`WidgetKind()` a nada | Este plan solo aporta vocabulario (constantes). No declara tipos ni implementa interfaces. |
| Usar la librería estándar de Go | `widget` compila a WASM. Solo `github.com/tinywasm/fmt`. |

---

## 1. El hueco que se está tapando

`tinywasm/form` emite el markup de cada campo de formulario:

```html
<div class="tw-field">
  <label for="...">Nombre</label>
  <input id="..." />
  <span class="tw-field-error">…</span>
</div>
```

`tinywasm/components/fieldset` existe **solo** para pintarlo. Su propio doc de paquete lo dice:

> *"It renders nothing itself… It **deliberately** styles form elements (`.tw-field`)… it is the
> one place the ecosystem centralizes form appearance."*

Hoy los dos lados se acoplan por un **string repetido en dos repos** (`".tw-field"`), sin ningún
tipo que lo nombre. Es exactamente el defecto que la regla lego describe y que el plan maestro
cita en su §5.1:

> *"A missing contract at a boundary is a defect in the library, not in the consumer. If two
> libraries meet and there is no type to name the thing that crosses between them, **the type is
> missing upstream**. Do not declare a local intersection to paper over it."*

La pieza que cruza entre `form` (que emite) y `fieldset` (que pinta) es **la anatomía del
campo**. El sitio correcto para nombrarla es `widget`, que es la única pieza por debajo de las
dos. La alternativa —que `components` dependa de `form`— añade una arista que el diagrama de
dependencias del maestro (§5.3) no contempla, y el propio maestro prescribe la salida en §10:

> *"si ahí falta algo, **se amplía el enum aguas arriba en `widget`/`css`, se publica una versión
> nueva y se sigue**"*

Esto es esa ampliación.

---

## 2. Etapa 1 — `field.go` (archivo NUEVO, único cambio de código)

Archivo: `field.go`, en la raíz del módulo. **Sin build tag**: el markup lo escriben tanto el
lado WASM como el SSR.

```go
package widget

// Anatomía compartida del campo de formulario.
//
// La emite github.com/tinywasm/form (que construye el markup) y la estiliza
// github.com/tinywasm/components/fieldset (que es la piel global de los
// formularios). Vive aquí porque es lo que cruza entre esas dos librerías, y
// ninguna de las dos puede poseerlo sin que la otra dependa de ella.
const NameField = Name("tw-field")

// Partes del campo. Los nombres son genéricos a propósito: Part es local a su
// widget y sólo se convierte en clase a través de un Name, así que "label" aquí
// produce "tw-field__label" y jamás colisiona con el "label" de otro widget.
const (
	PartLabel      = Part("label")
	PartInput      = Part("input")
	PartError      = Part("error")
	PartRadioGroup = Part("radio-group")
)
```

### 2.1 Los valores son un contrato — no los cambies

| Expresión | Valor exacto que debe producir |
|---|---|
| `NameField.Root()` | `tw-field` |
| `NameField.Class(PartLabel)` | `tw-field__label` |
| `NameField.Class(PartInput)` | `tw-field__input` |
| `NameField.Class(PartError)` | `tw-field__error` |
| `NameField.Class(PartRadioGroup)` | `tw-field__radio-group` |

`NameField` vale `"tw-field"` y **no** otra cosa: es la clase raíz que `form` ya emite hoy y
sobre la que hay CSS y tests existentes en el ecosistema. Cambiarla rompería a los consumidores
sin que nada falle en build.

---

## 3. Etapa 2 — Tests

Archivo: `field_test.go`, en la raíz (paquete `widget`, igual que el código — este repo no tiene
directorio `tests/`).

Aserciones, todas sobre valores literales:

1. `NameField.Root().String()` es exactamente `"tw-field"`.
2. `NameField.Class(PartLabel).String()` es exactamente `"tw-field__label"`.
3. Lo mismo para `PartInput`, `PartError` y `PartRadioGroup`, con los valores de la tabla §2.1.
4. `NameField.Class(PartLabel).AsAttr()` devuelve `fmt.KeyValue{Key: "class", Value: "tw-field__label"}`.

Assertions de stdlib solamente (`if got != want { t.Errorf(...) }`). Sin librerías de aserción.

Ejecutar con `gotest`, nunca `go test`.

---

## 4. Lo que este plan NO hace

- **No** añade una implementación de `Widget` para el campo. `form` solo necesita las
  constantes para derivar clases; no necesita satisfacer ninguna interfaz.
- **No** añade estilos. `widget/style` no se toca; la hoja la escribe `fieldset` en
  `components`.
- **No** toca `state.go`. `Invalid` y `Locked` ya existen y son lo que `form` va a emitir.

---

## 5. Consumidores (OTROS repos — no los toques desde aquí)

Este cambio se publica como **v0.2.0** y lo consumen, en este orden:

1. **`tinywasm/form`** — emite `widget.NameField.Root()` y las partes, más los estados
   `data-invalid`/`data-locked`. Plan: <https://github.com/tinywasm/form/blob/main/docs/PLAN.md>
2. **`tinywasm/components`** — `fieldset` pasa a `style.Of(widget.NameField)` y usa
   `widget.PartLabel`/`PartInput`/`PartError` en sus `Part(...)`.
   Contexto: <https://github.com/tinywasm/components/pull/14>

**No incluyas esos cambios en este PR.** Este plan se cierra dentro de `tinywasm/widget`.

---

## 6. Criterios de aceptación — verificables con grep

1. `gotest` en verde.
2. `ls field.go` → existe; `ls field_test.go` → existe.
3. `go list -deps . | grep tinywasm` → debe imprimir **exactamente estas dos líneas**, ni una
   más (la segunda es el propio paquete):
   ```
   github.com/tinywasm/fmt
   github.com/tinywasm/widget
   ```
   Una tercera línea significa que el paquete raíz ganó una dependencia y el cambio está mal.
4. `GOOS=js GOARCH=wasm go build ./...` compila.
5. `git diff --stat v0.1.0 -- widget.go kind.go state.go capability.go style/` → **vacío**:
   ningún símbolo preexistente cambió.
6. `grep -rn "func Class(" .` → **vacío**: no se añadió constructor público de `Class`.

---

## 7. Checklist de calidad Go (obligatorio)

- **Sin strings repetidos**: los cinco valores se declaran una vez, en `field.go`, y todo lo
  demás los deriva. Ningún literal `"tw-field"` en otro sitio.
- **Sin stdlib**: solo `github.com/tinywasm/fmt`.
- **Cero `any`, cero `map`**.
- Constantes **exportadas** a propósito: son contrato entre repos.
- Comentarios en el código en inglés, según la convención del repo.

---

## 8. Tabla de etapas

| # | Etapa | Archivos | Gate |
|---|---|---|---|
| 1 | Anatomía | `field.go` (nuevo) | `go build ./...` |
| 2 | Tests | `field_test.go` (nuevo) | `gotest` verde |

Dos etapas, secuenciales. La 2 es el gate.

---

## 9. Anexo — vocabulario existente, para referencia

De `widget.go` (no se toca):

```go
type Name string
type Part string
type Class string

func (c Class) String() string   { return string(c) }
func (c Class) AsAttr() fmt.KeyValue

func (n Name) Root() Class        { return Class(n) }
func (n Name) Class(p Part) Class { return Class(string(n) + "__" + string(p)) }
```

De `state.go` (no se toca) — lo que `form` va a emitir:

```go
Locked   // -> data-locked="true"
Invalid  // -> data-invalid="true"

func (s State) Attr() fmt.KeyValue
```
