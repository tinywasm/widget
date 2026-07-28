# Plan de usabilidad: hacer `tinywasm/widget` intuitiva para un junior

Objetivo: que alguien **sin conocimientos de diseño** pueda construir un widget
correcto y bonito sin leer el código fuente de la librería, y que la librería le
impida equivocarse en vez de dejarle equivocarse en silencio.

Este documento es un diagnóstico + plan de acción. Cada hallazgo está verificado
contra el CSS que la librería emite hoy (commit `b291d31`), no supuesto.

---

## Resumen del diagnóstico

La arquitectura es buena y la intención es correcta: escalas cerradas, `Surface`
como decisión completa, capas de cascada fijas, salida determinista. Un junior no
puede escribir `#ff0000` ni `margin-top: 13px`, y eso ya elimina la mayor fuente
de fealdad accidental.

Pero **hoy la librería no es usable por un junior sin ayuda**, por tres razones
distintas que conviene no mezclar:

| | Problema | Gravedad |
|---|---|---|
| A | Hay un fallo real: `Shown()` rompe el layout del elemento que muestra | Bug |
| B | Los errores del usuario no producen error: se emite CSS muerto en silencio | Alto |
| C | La API sigue exigiendo criterio de diseño donde prometía no exigirlo | Alto |
| D | No hay por dónde empezar: README de 4 líneas, ejemplo que no enseña nada | Alto |

El resto del documento desarrolla cada uno con el caso concreto y la corrección
propuesta.

---

## A. Fallo real: `Shown()` destruye el layout del elemento

Es lo primero que hay que arreglar porque no es una cuestión de gusto: produce
una página rota.

Caso: una barra de acciones que se despliega. Es lo más natural del mundo
escribirlo así:

```go
style.Of("bar").
    Part("actions", style.Row(style.Space2), style.Hidden()).
    When(widget.Open, "actions", style.Shown())
```

CSS emitido hoy:

```css
@layer primitives {
  .bar__actions, .fl-row { display: flex; flex-wrap: wrap; gap: var(--gap); align-items: center; }
}
@layer widgets {
  .bar__actions { --gap: var(--space-2,0.5rem); display: none; }
}
@layer states {
  .bar__actions[data-open="true"] { display: block; }   /* ← gana */
}
```

Cuando la barra se abre, `display: block` gana a `display: flex` (capa `states`
va después de `primitives`). El `Row` deja de ser un row: los hijos se apilan en
vertical y el `gap` deja de aplicarse. Lo mismo con `Stack`, `Grid`, `Split`,
`Cover`, `Reel` y `Frame` — es decir, con **todos** los primitivos de flujo.

Un junior no tiene forma de diagnosticar esto: el código Go se lee perfecto, y el
síntoma aparece solo en el estado abierto.

### Corrección

`Shown()` no debe imponer un `display` fijo: debe restituir **el `display` que le
corresponde a ese elemento**. La hoja ya tiene esa información — conoce el
`FlowType` de la regla base de la parte — así que la emisión puede resolverlo:

| flujo de la parte | `display` que debe emitir `Shown()` |
|---|---|
| `Stack`, `Row`, `Reel`, `Frame` | `flex` |
| `Split`, `Grid`, `Cover` | `grid` |
| `Center` o sin flujo | `block` |

En `Stylesheet()`, al emitir una regla de estado con `Shown`, basta con mirar
`s.PartRules[key.Part].FlowType` y emitir el `display` correspondiente en vez del
`block` incondicional de hoy.

Un aviso para quien lo implemente: la solución que parece obvia —emitir
`display: revert-layer`— **no funciona aquí**. `Hidden()` emite su `display: none`
en la capa `widgets`, así que `revert-layer` desde la capa `states` retrocede
justo hasta ese `none` y el elemento sigue invisible. Sería viable solo moviendo
también `Hidden` fuera de `widgets`, lo que complica la emisión sin ganar nada
frente a la tabla de arriba.

**Acción**: resolver el `display` de `Shown()` desde el `FlowType` de la parte, y
añadir un test que verifique que un `Row` con `Hidden`/`Shown` sigue siendo
`flex` cuando está abierto.

---

## B. Los errores no dan error

Un junior aprende por realimentación. Hoy la librería no le da ninguna: los tres
errores más probables se compilan, se ejecutan y emiten CSS que no hace nada.

### B.1 Una errata en el nombre de una parte no falla

```go
style.Of("list").
    Part("item", style.On(style.Muted)).
    When(widget.Selected, "itm", style.On(style.Selected))   // ← errata
```

emite:

```css
@layer widgets { .list__item { ... } }
@layer states  { .list__itm[data-selected="true"] { ... } }   /* nunca casa con nada */
```

Sin aviso. El junior ve que la selección "no funciona" y no tiene dónde mirar.
Esto pasa porque `Part` es un `type Part string` desnudo: cualquier literal vale.

**Acción**: `Sheet` ya conoce todas sus partes declaradas. Añadir validación:

```go
// Validate comprueba que la hoja es coherente consigo misma.
// Devuelve todos los problemas encontrados, no solo el primero.
func (s *Sheet) Validate() []error
```

que detecte al menos:
- `When`/`Cue` sobre una parte que nunca se declaró con `Part()`
- una parte declarada que no produce ninguna declaración (regla vacía)
- `Scrim()` sin `Backdrop()` (hoy no hace nada y no avisa)
- `Shown()` en la regla base en vez de en un `When()`

y llamarla desde `Stylesheet()` en modo pánico o log según convenga. Como mínimo,
que exista y que el test de consumidor la use.

### B.2 La promesa de "cero escapes" tiene un agujero

`Sheet` y `Rule` tienen todos los campos públicos, incluido `FlowType string`.
Esto compila y emite un `padding` huérfano sin flujo ninguno:

```go
sh := style.Of("x")
sh.PartRules["p"] = style.Rule{HasFlow: true, FlowType: "wobble", HasPad: true, Pad: style.Space2}
```

No es que un junior vaya a hacerlo a propósito; es que el autocompletado del
editor se lo ofrece y le sugiere que ésa es una forma legítima de usar la
librería. La API pública debería ser solo `Of` + los `Opt`.

**Acción**: pasar `Rule`, `stateKey`, `cueKey` y los campos de `Sheet` a privado.
Si algo externo necesita leer la hoja, exponer accesores concretos. `FlowType`
debe ser un enum, no un `string`.

### B.3 La escala `Space` promete 13 pasos y entrega 6

```
Space4  → var(--space-4)
Space5  → var(--space-6)   ┐ idénticos
Space6  → var(--space-6)   ┘
Space7  → var(--space-8)   ┐ idénticos
Space8  → var(--space-8)   ┘
Space9  → var(--space-12)  ┐
Space10 → var(--space-12)  │ idénticos
Space11 → var(--space-12)  │
Space12 → var(--space-12)  ┘
```

Un junior que cambia `Space5` por `Space6` buscando "un poco más de aire" no ve
ningún cambio, y concluye que no entiende la librería. Una escala con pasos que
no se distinguen es peor que una escala corta.

**Acción**: reducir `Space` a los pasos que existen de verdad (`Space0..Space6`,
seis valores distintos), o añadir los tokens que faltan en `tinywasm/css`.

`Ratio` tiene un problema distinto pero de la misma familia: **el mismo valor
significa dos cosas según dónde se use**. En `Split` describe el reparto de
columnas y los nombres son correctos (`RatioTwoThirds` → `2fr 1fr`). En `Frame`
describe una proporción de aspecto, y ahí los nombres no corresponden:

```go
Frame(RatioHalf)          → aspect-ratio: 1/1    // un cuadrado, no una mitad
Frame(RatioTwoThirds)     → aspect-ratio: 3/2
Frame(RatioThreeQuarters) → aspect-ratio: 4/3
```

Un junior que quiere una imagen apaisada elige `RatioHalf` y obtiene un cuadrado.

**Acción**: separar en dos escalas, `SplitRatio` (reparto) y `AspectRatio`
(proporción, con nombres que digan lo que son: `AspectSquare`, `Aspect3x2`,
`Aspect4x3`, `Aspect16x9`).

---

## C. La API todavía exige criterio de diseño

Éste es el punto central de la pregunta original. La librería dice que un color
es "una decisión completa" y que no hace falta saber de diseño. Es verdad para
*un* elemento en *un* estado. Deja de ser verdad en cuanto hay interacción.

### C.1 El junior tiene que emparejar a mano las superficies de interacción

`Surface` tiene 37 constantes organizadas en familias de cuatro:
`Panel`/`PanelHover`/`PanelFocus`/`PanelPress`. Para un botón hay que escribir:

```go
Part("cta", On(Accent)).
Cue(widget.Hover, "cta", On(AccentHover)).
Cue(widget.Focus, "cta", On(AccentFocus)).
Cue(widget.Press, "cta", On(AccentPress))
```

Cuatro líneas, y nada impide escribir esto, que compila igual:

```go
Part("cta", On(Accent)).
Cue(widget.Hover, "cta", On(DangerHover))   // el botón se vuelve rojo al pasar el ratón
```

Justo la decisión que un junior sin ojo de diseño va a tomar mal — y es la que la
librería le deja tomar. La familia de la superficie ya determina cuáles son sus
estados de interacción; que el usuario los repita es trabajo manual sin criterio
aportado.

**Acción**: añadir un `Opt` que resuelva la familia entera de una vez.

```go
// Interactive aplica la superficie s y deriva automáticamente sus estados
// hover, focus y press. Es la forma normal de estilar algo pulsable.
func Interactive(s Surface) Opt
```

de modo que lo anterior sea:

```go
Part("cta", Interactive(Accent))
```

Internamente marca la regla, y `Stylesheet()` emite las cuatro reglas con la
familia correcta. Las constantes `*Hover`/`*Focus`/`*Press` pueden pasar a
privadas: dejan de ser parte de la API que el junior ve.

Con esto, `On()` queda para superficies estáticas e `Interactive()` para lo
pulsable, y elegir mal deja de ser posible.

### C.2 El foco usa `:focus`, no `:focus-visible`

```go
case widget.Focus: return ":focus"
```

`:focus` dibuja el anillo de foco también al hacer clic con el ratón, que es
justo lo que todo el mundo intenta quitar y acaba quitando mal (rompiendo la
accesibilidad de teclado). `:focus-visible` lo dibuja solo cuando corresponde.

**Acción**: cambiar a `:focus-visible`. Es una línea y es la decisión correcta
para el 100% de los casos; no hay motivo para dejársela al consumidor.

### C.3 Accesibilidad: `Kind` se pide y no se usa

`Widget` obliga a implementar `WidgetKind() Kind`, y `Kind` enumera los patrones
de WAI-ARIA. Pero **nada en el repositorio lee ese valor**. No emite `role`, no
valida qué estados son legales para ese `Kind`, no comprueba nada.

Un junior lo rellena a ojo, no pasa nada, y aprende que da igual lo que ponga.
Igual con `Selectable`/`Dismissible`/`Expandable` en `capability.go`: interfaces
declaradas que nadie consume.

Esto es la mitad de "no hace falta saber de diseño" que la librería aún no
cumple: la parte de accesibilidad. Un junior no sabe qué rol ARIA lleva un menú
ni qué atributos necesita, y aquí es donde más ayuda le haría.

**Acción**, por orden de valor:

1. `func (k Kind) Role() fmt.KeyValue` — devuelve el `role` ARIA del patrón, de
   la misma forma que `State.Attr()` ya devuelve su atributo. Barato y resuelve
   el caso más común.
2. `func (k Kind) Allows(s State) bool` — qué estados tienen sentido para ese
   patrón (`Open` en un `Listbox` no significa nada). Permite que `Validate()`
   avise.
3. Documentar el teclado esperado por `Kind` en el godoc, aunque la librería no
   lo implemente: es lo que el junior no sabe y no va a buscar.

Si se decide que `Kind` y las capacidades no van a usarse, hay que quitarlas. Una
API que pide datos que ignora enseña a no confiar en ella.

### C.4 `Width()` significa dos cosas distintas

```go
Root(Width(Half))            → width: 50%
Root(Center(), Width(Half))  → --max-width: 50%   (y width: 100%)
```

El mismo `Opt` cambia de significado según qué otro `Opt` lo acompañe, y no hay
nada en el nombre ni en el godoc que lo diga. `Center()` sin `Width()` usa
`--max-w-prose` por defecto.

**Acción**: separar en `Width(Size)` y `MaxWidth(Size)`, y que `Center()` lea
solo `MaxWidth`. O renombrar a `Center(Size)` recibiendo el tope directamente,
que es más difícil de usar mal:

```go
Root(Center(Prose))
```

### C.5 Nombres que colisionan entre paquetes

En el test de consumidor, escrito por los propios autores:

```go
When(widget.Selected, "item", style.On(style.Selected))
```

`widget.Selected` es un `State`; `style.Selected` es una `Surface`. Son cosas
distintas con el mismo nombre. Igual `widget.Disabled` vs `style.Disabled`. Y
`style.Overlay` es un nivel de `Elevation`, no tiene relación con `overlay.go`
(donde el concepto se llama `Backdrop` — el propio comentario del código explica
que tuvo que llamarse así para no chocar).

**Acción**: renombrar las `Surface` para que no se confundan con `State`:
`SurfaceSelected` / `SelectedSkin`, o mejor, un prefijo consistente en toda la
familia. Y `Elevation.Overlay` → `Elevation.Popover` o `ElevationHighest`.

---

## D. No hay por dónde empezar

Es, en la práctica, la primera barrera que encuentra un junior.

- **README.md**: 8 líneas, una frase de descripción y un enlace. No hay ni un
  bloque de código.
- **example/main.go**: imprime `"button"` por consola. No construye una hoja, no
  genera CSS, no muestra markup. Es el único ejemplo del repositorio y no enseña
  nada de lo que la librería hace.
- **Sin godoc de paquete**: no hay `doc.go` ni en `widget` ni en `style`. En
  pkg.go.dev la librería aparece como una lista de símbolos sin narrativa.
- **Sin ejemplos ejecutables**: ningún `Example_xxx()`, que es el mecanismo
  estándar de Go para documentar con código verificado por el compilador.
- **No se explica el último paso**: la librería produce un `*css.Stylesheet`. En
  ninguna parte se dice cómo llega eso al navegador, ni cómo se escribe el markup
  con las clases correctas. El junior se queda con un objeto en la mano.
- **Idioma mezclado**: `widget.go`, `kind.go`, `state.go` y todo `style/` están
  comentados en español; `field.go` y `ARCHITECTURE.md` en inglés. Hay que elegir
  uno.

**Acción**: escribir un `docs/GUIDE.md` orientado a tarea, no a concepto, con un
recorrido completo de un widget real de principio a fin:

1. declarar `Name` y `Part`
2. implementar `Widget`
3. escribir el `Style() *style.Sheet`
4. generar el CSS y servirlo
5. escribir el markup usando `Name.Class(Part).AsAttr()` y `State.Attr()`

Y una **tabla de decisión** que es lo que de verdad sustituye al criterio de
diseño — el junior no debe elegir, debe buscar en la tabla:

| Quiero… | Uso |
|---|---|
| una columna de cosas | `Stack(Space2)` |
| una fila de botones | `Row(Space1)` |
| una rejilla que se adapta sola | `Grid(TrackSm, Space2)` |
| lista + detalle | `Split(RatioTwoThirds, Space3)` |
| centrar una columna de texto | `Center(Prose)` |
| el fondo de la página | `On(Page)` |
| una tarjeta o panel | `On(Panel)` + `Round(RadiusMd)` |
| algo pulsable | `Interactive(Accent)` |
| texto secundario | `On(Muted)` |
| un error | `On(Danger)` |
| que ocupe el alto restante | `Fill()` |
| que haga scroll interno | `Scrolls()` |

Complementar con `doc.go` en ambos paquetes y con `Example` funcions para `Of`,
`Stack`, `Split`, `Interactive` y `Backdrop`.

---

## Orden de trabajo propuesto

Por relación valor/esfuerzo, y en este orden:

**Primero — arreglar lo que está roto** (nada de esto es discutible)
1. `Shown()` → `display: revert-layer` (A) + test de regresión
2. `:focus` → `:focus-visible` (C.2)
3. Documentar en godoc qué pasos de `Space` son distintos y qué produce
   `Frame(Ratio…)`, para que deje de engañar mientras no se pueda romper la API (B.3)

**Segundo — que la librería enseñe al que la usa**
4. `Interactive(Surface)` y ocultar las constantes `*Hover`/`*Focus`/`*Press` (C.1)
5. `Sheet.Validate()` con los cuatro chequeos listados (B.1)
6. `Kind.Role()` (C.3)

**Tercero — que se pueda empezar sin preguntar**
7. `docs/GUIDE.md` con el recorrido completo y la tabla de decisión (D)
8. `doc.go` en ambos paquetes + `Example` functions (D)
9. Reescribir `example/main.go` para que genere una hoja de verdad (D)
10. README con un bloque de código que se entienda en diez segundos (D)

**Cuarto — limpieza de API, rompe compatibilidad**
11. `Rule`/`Sheet` a privado, `FlowType` a enum (B.2)
12. Podar `Space` a sus pasos reales y separar `Ratio` en `SplitRatio` /
    `AspectRatio` (B.3)
13. Separar `Width`/`MaxWidth` o `Center(Size)` (C.4)
14. Renombrar `Surface` que colisionan con `State` (C.5)
15. Decidir sobre `Kind` y las capacidades: usarlas o quitarlas (C.3)
16. Unificar el idioma de los comentarios (D)

Los puntos 11–16 conviene agruparlos en una sola versión mayor en vez de ir
rompiendo la API poco a poco.

---

## Criterio de aceptación

La pregunta "¿es intuitiva para un junior?" tiene respuesta objetiva. Se puede
verificar así: dar a alguien que no sepa de diseño la `GUIDE.md` y pedirle que
construya un panel desplegable con una lista seleccionable, sin ayuda y sin
abrir el código de la librería.

Hoy no lo consigue: se topa con A (se le rompe al abrir), con B.1 si comete una
errata, y con C.1 (no sabe qué superficies emparejar).

Con los puntos 1–8 hechos, debería conseguirlo.
