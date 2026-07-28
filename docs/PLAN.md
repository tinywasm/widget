# Plan: `tinywasm/widget` intuitiva para un junior — una sola etapa, con ruptura

Objetivo: que alguien **sin conocimientos de diseño** construya un widget
correcto y accesible sin leer el código fuente, y que la librería le impida
equivocarse en vez de dejarle equivocarse en silencio.

Se hace **de una vez y rompiendo la API**. No hay periodo de convivencia ni
alias de compatibilidad: el módulo no tiene ningún tag publicado, así que el
coste de romper ahora es el más bajo que va a ser nunca, y partir el trabajo en
fases obligaría a diseñar dos veces las mismas firmas.

Diagnóstico y API objetivo verificados contra el CSS que la librería emite hoy
(commit `b291d31`), no supuestos.

---

## 1. Qué está mal hoy

Cinco problemas, todos comprobados ejecutando la librería.

### 1.1 `Shown()` destruye el layout del elemento que muestra — es un fallo, no un gusto

```go
Part("actions", Row(Space2), Hidden()).
When(widget.Open, "actions", Shown())
```

```css
@layer primitives { .bar__actions, .fl-row { display: flex; gap: var(--gap); … } }
@layer widgets    { .bar__actions { display: none; } }
@layer states     { .bar__actions[data-open="true"] { display: block; } }  /* ← gana */
```

Al abrirse, el `Row` deja de ser row: los hijos se apilan en vertical y el `gap`
deja de aplicar. Afecta a **todos** los primitivos de flujo (`Stack`, `Row`,
`Split`, `Grid`, `Cover`, `Reel`, `Frame`). El Go se lee perfecto y el síntoma
solo existe en el estado abierto: un junior no tiene por dónde diagnosticarlo.

Además, el par `Hidden()`/`Shown()` obliga a recordar una regla de orden que solo
está en un comentario («`Shown` va en una regla de estado, nunca en la base»), y
nada la comprueba.

### 1.2 Los errores del usuario no producen ningún error

Una errata en el nombre de una parte se compila, se ejecuta y emite CSS muerto:

```go
Part("item", On(Muted)).
When(widget.Selected, "itm", On(Selected))   // errata
```
```css
@layer states { .list__itm[data-selected="true"] { … } }   /* no casa con nada */
```

No hay validación de ningún tipo. Tampoco avisa `Scrim()` sin `Backdrop()`, ni
una parte declarada que no emite nada.

### 1.3 La API sigue exigiendo criterio de diseño donde prometía no exigirlo

`Surface` expone 37 constantes en familias de cuatro. Un botón obliga a
emparejarlas a mano, y nada impide emparejarlas mal:

```go
Part("cta", On(Accent)).
Cue(widget.Hover, "cta", On(DangerHover))   // el botón se vuelve rojo al pasar el ratón
```

Eso compila y emite exactamente eso. Es justo la decisión que la librería decía
haberle quitado al junior.

Al lado, dos huecos de accesibilidad: el cue de foco usa `:focus` en vez de
`:focus-visible` (dibuja el anillo también al hacer clic con ratón, que es lo que
todo el mundo acaba quitando mal), y `Kind` es obligatorio en la interfaz
`Widget` pero **nadie lo lee**: no emite `role`, no valida estados, no hace nada.
Igual que `Selectable`/`Dismissible`/`Expandable`, declaradas y nunca consumidas.

### 1.4 Escalas que engañan

`Space` promete 13 pasos y entrega 6 valores distintos:

```
Space4 → --space-4        Space7, Space8              → --space-8
Space5, Space6 → --space-6    Space9…Space12          → --space-12
```

Cambiar `Space5` por `Space6` no hace nada, y el junior concluye que no entiende
la librería. `Ratio` es peor: significa dos cosas según dónde se use. En `Split`
reparte columnas y los nombres cuadran; en `Frame` es proporción de aspecto y no
cuadran:

```go
Frame(RatioHalf) → aspect-ratio: 1/1    // un cuadrado, no una mitad
```

Y `Width()` cambia de significado según qué otro `Opt` lo acompañe:

```go
Root(Width(Half))            → width: 50%
Root(Center(), Width(Half))  → --max-width: 50%
```

Además `Space` es numérica mientras `Radius`, `TextSize` y `Elevation` son
semánticas: cuatro escalas, dos convenciones.

### 1.5 Colisiones de nombres y fugas de la API

En el propio test de consumidor de los autores:

```go
When(widget.Selected, "item", style.On(style.Selected))
```

`widget.Selected` es un `State`, `style.Selected` es una `Surface`. Igual
`widget.Disabled` vs `style.Disabled`. Y `style.Overlay` es un nivel de
`Elevation` sin relación con `overlay.go` — el comentario del código ya explica
que `Backdrop` se llama así por no poder llamarse `Overlay`.

Y la promesa de «cero escapes» tiene un agujero: `Sheet` y `Rule` tienen todos
los campos públicos, incluido `FlowType string`. Esto compila:

```go
sh.PartRules["p"] = Rule{HasFlow: true, FlowType: "wobble", HasPad: true, Pad: Space2}
```

El autocompletado se lo ofrece al junior como si fuera una forma legítima de usar
la librería.

### 1.6 No hay por dónde empezar

README de 8 líneas sin un solo bloque de código. `example/main.go` imprime
`"button"` por consola: no construye una hoja, no genera CSS, no muestra markup.
Sin `doc.go` en ninguno de los dos paquetes, sin `Example` functions, y en
ninguna parte se explica cómo el `*css.Stylesheet` llega al navegador ni cómo
escribir el markup con las clases correctas. Los comentarios están en español en
`widget.go`, `kind.go`, `state.go` y todo `style/`, y en inglés en `field.go` y
`ARCHITECTURE.md`.

---

## 2. API objetivo

Todo lo que sigue entra en una única versión.

### 2.1 `widget` — se añade el uso de `Kind`

```go
func (k Kind) Role() fmt.KeyValue      // NUEVO: el role ARIA del patrón
func (k Kind) Allows(s State) bool     // NUEVO: qué estados tienen sentido
```

Misma forma que `State.Attr()`, que ya existe y funciona. `Allows` es lo que
permite que la validación avise de un `Open` en un `Listbox`. Si se decide no
implementarlas, hay que **borrar `Kind` y las tres capacidades**: una API que
exige datos que ignora enseña a no fiarse de ella.

### 2.2 Escalas — todas semánticas, sin pasos repetidos

```go
type Space uint8
const (
    SpaceNone Space = iota // --space-0   0
    SpaceXs                // --space-1   0.25rem
    SpaceSm                // --space-2   0.5rem
    SpaceMd                // --space-3   0.75rem
    SpaceLg                // --space-4   1rem
    SpaceXl                // --space-6   1.5rem
    Space2xl               // --space-8   2rem
    Space3xl               // --space-12  3rem
)
```

Ocho pasos, ocho valores distintos, y la misma convención de nombres que
`Radius`, `TextSize` y `Elevation`. No es un número elegido a ojo: son
exactamente los ocho tokens `--space-*` que publica `tinywasm/css@v0.3.0`, de
modo que la escala deja de inventar pasos que el sistema de tokens no tiene. `Ratio` se parte en dos porque hoy son dos
cosas:

```go
type SplitRatio uint8   // reparto de columnas en Split
const (SplitHalf, SplitTwoThirds, SplitThreeQuarters)

type Aspect uint8       // proporción de aspecto en Frame
const (AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9)
```

`Elevation.Overlay` → `Elevation.Popover`, que libera el nombre y deja de sugerir
parentesco con `Backdrop`. `Flat`, `Raised` y `Floating` se quedan: son
semánticos y le dicen más a un junior que un `ElevationMd`.

### 2.3 `Surface` — de 37 constantes a 10

```go
type Surface uint8
const (
    Page      // fondo de la aplicación
    Panel     // tarjeta, panel
    Sunken    // hueco: campo de entrada
    Accent    // acción principal
    Secondary // acción secundaria
    Highlight // ítem seleccionado      (era Selected)
    Success
    Danger
    Muted     // texto secundario, sin fondo
    Dimmed    // deshabilitado          (era Disabled)
)
```

Las 27 variantes `*Hover`/`*Focus`/`*Press` pasan a **privadas**: dejan de ser
API. `Highlight` y `Dimmed` eliminan la colisión con `widget.Selected` y
`widget.Disabled`.

La forma de estilar algo pulsable pasa a ser una sola llamada:

```go
// On aplica una superficie estática.
func On(s Surface) Opt

// Interactive aplica s y deriva sus estados hover, focus y press.
// Es la forma normal de estilar cualquier cosa pulsable.
func Interactive(s Surface) Opt
```

```go
Part("cta", Interactive(Accent))   // sustituye a cuatro líneas y no se puede emparejar mal
```

### 2.4 Flujo — `Center` y `Frame` dejan de tener significado oculto

```go
func Stack(gap Space) Opt
func Row(gap Space) Opt
func Split(r SplitRatio, gap Space) Opt
func Grid(min Track, gap Space) Opt
func Center(max Size) Opt      // recibe el tope; ya no lo lee de Width()
func Cover() Opt
func Reel(gap Space) Opt
func Frame(a Aspect) Opt       // recibe una proporción, no un reparto
```

Con `Center(Size)`, `Width(Size)` recupera un único significado.

### 2.5 Visibilidad — `Hidden`/`Shown` desaparecen

```go
// RevealedBy oculta el elemento y lo muestra cuando el widget posee st,
// restituyendo el display que le corresponde a su flujo.
func RevealedBy(st widget.State) Opt
```

```go
Part("actions", Row(SpaceSm), RevealedBy(widget.Open))
```

Una llamada en vez de un par repartido entre dos reglas. Desaparece la regla de
orden que había que recordar, y desaparece el fallo de 1.1: la hoja conoce el
`FlowType` de la parte y emite el `display` correcto.

| flujo de la parte | `display` que emite la regla de estado |
|---|---|
| `Stack`, `Row`, `Reel`, `Frame` | `flex` |
| `Split`, `Grid`, `Cover` | `grid` |
| `Center` o sin flujo | `block` |

> Aviso para quien lo implemente: la solución que parece obvia —emitir
> `display: revert-layer`— **no funciona**. El `display: none` base vive en la
> capa `widgets`, así que `revert-layer` desde `states` retrocede justo hasta ese
> `none`. Hay que resolver el valor desde el `FlowType`.

### 2.6 Resto de opciones

```go
func Pad(Space) Opt
func Round(Radius) Opt
func Raise(Elevation) Opt
func Width(Size) Opt
func FontSize(TextSize) Opt    // era Text(): par consistente con FontWeight
func FontWeight(Weight) Opt
func Animate(Motion) Opt
func Fill() Opt
func Scrolls() Opt
func Fixed() Opt
func Flush() Opt
func Clip() Opt
func Backdrop(Scope) Opt
func Above() Opt
func Scrim() Opt
```

### 2.7 `Sheet` — se cierra y se valida

```go
func Of(n widget.Name) *Sheet
func (s *Sheet) Root(opts ...Opt) *Sheet
func (s *Sheet) Part(p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Opt) *Sheet
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Opt) *Sheet

// Validate devuelve TODOS los problemas de la hoja, no solo el primero.
func (s *Sheet) Validate() []error

// Stylesheet emite el CSS. Entra en pánico si Validate() encuentra algo:
// una hoja mal construida es un error de programación, no una condición
// de ejecución — mismo criterio que regexp.MustCompile.
func (s *Sheet) Stylesheet() *css.Stylesheet
```

`Rule`, `stateKey`, `cueKey` y todos los campos de `Sheet` pasan a privados, y
`FlowType` deja de ser `string` para ser un enum. Con eso, la única forma de
construir una hoja es `Of` + los `Opt`, que es lo que la arquitectura ya decía.

`Validate()` detecta como mínimo:

- `When`/`Cue` sobre una parte nunca declarada con `Part()` — la errata de 1.2
- una parte declarada que no produce ninguna declaración
- `Scrim()` sin `Backdrop()`
- `Above()` sin ningún `Backdrop()` en la hoja
- un `State` que `Kind.Allows` rechaza para ese widget

El pánico es deliberado: un junior que solo recibe un `error` devuelto lo ignora;
uno que recibe un pánico con el nombre de la parte mal escrita lo arregla en diez
segundos.

### 2.8 Foco

`Cue(widget.Focus, …)` emite `:focus-visible` en vez de `:focus`. Es una línea y
es la decisión correcta en el 100% de los casos: no hay motivo para dejársela al
consumidor.

---

## 3. Documentación

Es la primera barrera en la práctica, y entra en la misma etapa.

**`docs/GUIDE.md`**, orientado a tarea y no a concepto, con un widget real de
principio a fin: declarar `Name`/`Part` → implementar `Widget` → escribir
`Style()` → generar y servir el CSS → escribir el markup con
`Name.Class(Part).AsAttr()`, `State.Attr()` y `Kind.Role()`.

Y una **tabla de decisión**, que es lo que de verdad sustituye al criterio de
diseño: el junior no elige, busca.

| Quiero… | Uso |
|---|---|
| una columna de cosas | `Stack(SpaceSm)` |
| una fila de botones | `Row(SpaceXs)` |
| una rejilla que se adapta sola | `Grid(TrackSm, SpaceSm)` |
| lista + detalle | `Split(SplitTwoThirds, SpaceMd)` |
| centrar una columna de texto | `Center(Prose)` |
| una tira horizontal con scroll | `Reel(SpaceSm)` |
| una imagen con proporción fija | `Frame(Aspect16x9)` |
| el fondo de la página | `On(Page)` |
| una tarjeta o panel | `On(Panel)` + `Round(RadiusMd)` |
| algo pulsable | `Interactive(Accent)` |
| algo pulsable secundario | `Interactive(Secondary)` |
| el ítem seleccionado de una lista | `When(widget.Selected, "item", On(Highlight))` |
| texto secundario | `On(Muted)` |
| un error | `On(Danger)` |
| que ocupe el alto restante | `Fill()` |
| que haga scroll interno | `Scrolls()` |
| algo que se despliega | `RevealedBy(widget.Open)` |
| un diálogo modal | `Backdrop(Viewport)` + `Scrim()` + panel con `Above()` |

Además: `doc.go` en ambos paquetes; `Example` functions para `Of`, `Stack`,
`Split`, `Interactive`, `RevealedBy` y `Backdrop` (las verifica el compilador y
salen en pkg.go.dev); `example/main.go` reescrito para construir una hoja de
verdad e imprimir su CSS; y un README con un bloque de código que se entienda en
diez segundos.

**Idioma**: una sola regla, no la mezcla actual. Propuesta: godoc y comentarios
de código en inglés (es una librería pública, se lee en pkg.go.dev), `docs/` en
español (lo lee el equipo). Conviene confirmarlo antes de tocar 1.700 líneas de
comentarios.

---

## 4. Tabla de migración

| Antes | Ahora |
|---|---|
| `Space0…Space12` | `SpaceNone, SpaceXs, SpaceSm, SpaceMd, SpaceLg, SpaceXl, Space2xl, Space3xl` |
| `Split(RatioTwoThirds, g)` | `Split(SplitTwoThirds, g)` |
| `Frame(RatioHalf)` | `Frame(AspectSquare)` |
| `Center()` + `Width(s)` | `Center(s)` |
| `Center()` a secas | `Center(Prose)` |
| `Raise(Overlay)` | `Raise(Popover)` |
| `On(Selected)` | `On(Highlight)` |
| `On(Disabled)` | `On(Dimmed)` |
| `On(X)` + `Cue(Hover, On(XHover))` + `Focus` + `Press` | `Interactive(X)` |
| `Text(TextSm)` | `FontSize(TextSm)` |
| `Hidden()` + `When(st, p, Shown())` | `RevealedBy(st)` |
| `Rule{…}` literal | no existe: solo `Of` + `Opt` |

---

## 5. Orden de implementación

Un solo entregable; el orden es de dependencia, no de riesgo.

1. **Escalas y renombrados**: `Space`, `SplitRatio`/`Aspect`, `Popover`,
   `Highlight`/`Dimmed`, `FontSize`. Toca `scale.go`, `surface.go`, `except.go`,
   `flow.go` y `emit.go`. Es mecánico y desbloquea todo lo demás.
2. **`Surface` a 10 constantes** + `Interactive()`, con las variantes de
   interacción privadas y derivadas en la emisión.
3. **`RevealedBy()`**, borrando `Hidden`/`Shown` y resolviendo el `display` desde
   el `FlowType`. Cierra el fallo de 1.1.
4. **`Center(Size)`** y `Width` con un único significado.
5. **`:focus-visible`** en `cuePseudo`.
6. **`Kind.Role()` y `Kind.Allows()`** en el paquete raíz.
7. **Cerrar la API**: `Rule` y los campos de `Sheet` a privados, `FlowType` a
   enum. Debe ir después de 1–6 para no reescribir dos veces la emisión.
8. **`Validate()`** y el pánico en `Stylesheet()`, apoyándose en 6 y 7.
9. **Tests**: regresión del `Row` con `RevealedBy` abierto (debe seguir siendo
   `flex`), un test por cada chequeo de `Validate()`, y actualizar
   `consumer_test.go` a la API nueva.
10. **Documentación**: `GUIDE.md`, `doc.go` ×2, `Example` functions,
    `example/main.go`, README, y unificación de idioma.
11. **Actualizar `ARCHITECTURE.md`**, que hoy describe escalas que dejan de
    existir.

Consumidores conocidos a avisar: `tinywasm/form` y
`tinywasm/components/fieldset`, que dependen de `NameField` y sus `Part` — esos
no cambian, pero sí cualquier `Style()` que tengan escrito.

---

## 6. Criterio de aceptación

La pregunta «¿es intuitiva para un junior?» tiene respuesta objetiva. Se
comprueba dando la `GUIDE.md` a alguien que no sepa de diseño y pidiéndole un
panel desplegable con una lista seleccionable, sin ayuda y sin abrir el código de
la librería.

Hoy no lo consigue: se le rompe el layout al abrir (1.1), no se entera si comete
una errata (1.2), y no sabe qué superficies emparejar (1.3).

Con esta etapa entregada, debería conseguirlo — y si comete la errata, la
librería se lo dice por su nombre.
