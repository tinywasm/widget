---
PLAN: "feat(style): apilamiento declarado, chip sin transform y contrato de cromo flotante"
EXECUTOR: jules
REVIEWER: none
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.
>
> Es la **etapa 2 de 4**. Orden obligatorio: **css → widget → components →
> layout**. Requiere el token `--chip-height` publicado en `tinywasm/css`
> (ver su `docs/PLAN.md`). No empezar antes.

# Plan — cerrar los agujeros del DSL de estilos que producen defectos silenciosos

## 1. Diagnóstico: no es un problema de z-index, son tres problemas distintos

Una sesión de depuración sobre una vista CRUD en móvil produjo cuatro defectos
visuales. Agruparlos correctamente es lo que decide qué se puede automatizar:

**Familia A — orden de pintado (z-index de verdad).**
- **A1.** Una leyenda `OnEdge` quedó sepultada bajo el `<input>` hermano en iOS.
  El DSL emitía `z-index: auto`, confiando en que la especificación pinta un
  elemento posicionado por encima de uno que no lo está. Safari compone un control
  de formulario (más aún si está `:disabled`) por encima de un hermano con
  `z-index: auto`. Todos los demás motores lo pintaban bien: un defecto que solo
  existe en un navegador es la firma de un valor **no declarado**.
- **A2.** El `outline` del estado `Locked` pintaba **sobre** la leyenda. Los
  outlines se pintan al final del *stacking context*, por encima de los
  descendientes posicionados.

**Familia B — espacio y geometría (esto NO es z-index).**
- **B1.** El desplegable de una fila desbordaba la pantalla: está anclado a un
  disparador que se mueve ~340px entre los dos estados de un *scroll-snap*, algo
  que CSS no puede observar.
- **B2.** El badge de la última fila se solapaba con un botón de acción flotante.
  `OnEdge` posiciona con `transform`, y un `transform` es **invisible para
  `scrollHeight`** por especificación: ningún padding calculado desde la caja del
  badge podía reservar el hueco donde el badge realmente se pinta.

**Familia C — fallo silencioso del propio DSL.**
- **C1.** `formatRule` ordenaba las declaraciones **alfabéticamente** antes de
  deduplicar. `padding-block-end:` ordena antes que `padding:` (`-` < `:` en
  ASCII), así que `Pad()` seguido de `PadEdge()` — la forma de "valor general más
  una excepción en un borde" — salía con el atajo AL FINAL, ganándole a la
  propiedad específica que debía sobrescribirlo. Escribir la llamada correcta
  producía CSS al revés, sin error ni aviso. **Ya corregido** (dedup que preserva
  el orden), **sin test que lo fije**.

### La respuesta a "¿esto lo elimina el harness?"

**Parcialmente, y lo que no puede cerrar es exactamente lo que necesita un
contrato tipado nuevo, no más automatización.**

- **A1, A2 y C1 sí se cierran**, y deben cerrarse aquí: ocurren dentro de una
  misma hoja, donde el DSL ya conoce todas las Parts y todas las primitivas que
  escapan del flujo. Que hoy no lo haga es un agujero del harness, no una
  limitación de CSS. Violan directamente el principio 6 (*nunca un fallo
  silencioso*) y el 3 (*los estados ilegales no deben poder escribirse*).
- **B2 no se cierra automatizando**, porque el botón flotante y el badge viven en
  **widgets distintos, hojas distintas y repos distintos**. Ninguno sabe que el
  otro existe, y ninguna cantidad de tipos dentro de una hoja puede saberlo. Esto
  es literalmente el *missing contract at a seam* del harness: falta un tipo que
  nombre lo que cruza entre las dos librerías. Hoy se resuelve porque un host
  alcanza el padding de un hijo por coincidencia de cascada — un parche en la
  hoja, el antipatrón que el propio documento señala. La solución es **añadir el
  contrato** (etapa 5), no adivinar.
- **B1 no se cierra en CSS.** CSS no ve la posición de un scroll-snap. Lo honesto
  es que la elección (anclado al disparador vs. anclado al viewport) sea
  **explícita y tipada**, no fingir que se puede derivar.

## 2. Contexto del repo para un agente sin contexto previo

- Módulo: `github.com/tinywasm/widget`. `docs/PLAN.md` va junto a `go.mod`.
- El DSL vive en `style/`. Piezas relevantes:
  - `sheet.go` — el struct `rule` (todos los flags) y los métodos del `Sheet`.
  - `except.go`, `flow.go`, `overlay.go` — las `Option` que llenan `rule`.
  - `emit_decls.go` — `rule.Decls()`, que traduce `rule` a declaraciones CSS,
    y `formatRule`, que las serializa.
  - `emit.go` — el recorrido que agrupa reglas por capa, estado, cue y device.
  - `validate.go` — `Sheet.Validate()`, los diagnósticos que ya existen.
- **No hay escotilla de CSS crudo y no se debe añadir una.** Todo sale de una
  `Option` tipada.
- Nada de librería estándar en paquetes WASM: `tinywasm/fmt`, nunca
  `errors`/`strconv`/`strings`.
- Empotrado por valor: `dom.Element` como valor, nunca `*dom.Element`.
- Prohibidas las cadenas repetidas en la lógica: constante con nombre.
- El emisor ya escribe **doble declaración** para el color (estática primero,
  `light-dark()` después) — ver el comentario en `Decls`. Cualquier propiedad
  nueva que use `color-mix`/`light-dark` sigue ese mismo patrón.

## 3. Etapas

### Etapa 1 — fijar el orden de emisión con un test (cierra C1)

`formatRule` ya deduplica preservando el orden. Falta el test que impida que
alguien "optimice" volviendo a ordenar.

Añadir en `style/` un test que construya una hoja con `Pad(Space1)` **y**
`PadEdge(EdgeBottom, Space12)` sobre la misma Part y verifique que en el bloque
emitido el índice de `padding:` es **menor** que el de `padding-block-end:`.
Comprobar índices, no `Contains`: el defecto era de orden, y un `Contains` pasaba
igual cuando estaba mal.

Añadir un segundo caso que verifique que la deduplicación sigue viva: dos
`Option` que emitan `display: flex;` deben producir **una** sola línea.

**Aceptación:** el test falla si se restaura `sort.Strings(decls)` en
`formatRule`.

### Etapa 2 — `OnEdge` sin `transform` (cierra B2 en origen)

Hoy `OnEdge` emite `transform: translateY(±50%)`, elegido para ser agnóstico a la
altura del chip. Con `--chip-height` publicado (plan de `css`) esa razón
desaparece.

Sustituir el `transform` por un margen negativo calculado:

```
inset-block-start: <space>;
margin-block-start: calc(-0.5 * var(--chip-height, 1.25rem));
```

(y el simétrico con `inset-block-end` / `margin-block-end` para `EdgeBottom`).

Efectos, todos deseados:
- la caja del chip **existe** para `scrollHeight`, así que un ancestro puede
  reservar su hueco;
- desaparece el *stacking context* implícito que creaba el `transform`;
- el fallback `1.25rem` mantiene el comportamiento si alguien usa el DSL con una
  hoja de tokens antigua.

**Aceptación:** ninguna regla emitida por `OnEdge` contiene `transform`;
`grep -rn "translateY" style/` solo aparece, si acaso, en `SlideDeck`.

### Etapa 3 — el apilamiento se declara, nunca se hereda de `auto`

Hoy el z-index sale de **tres sitios sin relación entre sí**: `layerVar(kind)`
para los overlays reales, un `z-index: 1` puesto a mano en `OnEdge`, y `auto`
para todo lo demás. `auto` significa "depende del orden del DOM y de cómo componga
el navegador" — es decir, el fallo silencioso de A1.

1. Unificar en una única función, p. ej. `stackingFor(r rule) string`, que sea el
   **único** sitio del paquete que produce un `z-index`.
2. Toda primitiva que saque un elemento del flujo (`OnEdge`, `Flyout`, `Drawer`,
   `Backdrop`, `Docked`) obtiene de ahí un valor **declarado**, nunca `auto`.
3. Mantener los dos escalones que ya existen y nombrarlos:
   - **local** (`1`): cromo que cabalga sobre el contenido de su propio widget —
     `OnEdge`, `Docked(Parent, …)`. No debe alcanzar la capa de overlay: si lo
     hiciera, empataría con un dropdown real y ganaría por orden de DOM (esto ya
     ocurrió y está documentado en el comentario de `hasDocked`).
   - **overlay** (`var(--z-dropdown)` y superiores, vía `Kind.Layer()`): overlays
     de verdad — `Flyout`, `Drawer`, `Backdrop(Viewport)`, `Docked(Viewport, …)`.
4. Añadir en `Validate()` un diagnóstico si una Part fuera de flujo acabara sin
   posición de apilamiento resoluble.

**Aceptación:** `grep -rn "z-index" style/*.go` (sin tests) solo aparece dentro de
`stackingFor`; los tests existentes de `Docked(Viewport)` siguen exigiendo
`var(--z-dropdown,100)`.

### Etapa 4 — los bordes de estado se pintan con `box-shadow`, no con `outline`

Una regla de estado (`When`/`Cue`) emite hoy su borde como
`outline` + `outline-offset: -1px`, deliberadamente, para que un estado no pueda
cambiar el tamaño de la caja que pinta. Eso resuelve el tamaño y **crea dos
defectos**:

- **A2:** los outlines se pintan al final del stacking context, por encima de los
  descendientes posicionados — por eso el borde del estado `Locked` cruzaba por
  encima de la leyenda que monta sobre esa misma línea.
- **Safari < 16.4 no aplica `border-radius` a un `outline`.** Un iPhone 7 se queda
  en iOS 15: ahí **todos** los bordes de estado del sistema salen cuadrados, no
  solo uno.

`box-shadow: 0 0 0 1px <color>` (o con `inset`) cumple la restricción original —
no afecta al layout — y además respeta el `border-radius` en todos los motores y
se pinta como parte del fondo del propio elemento, no por encima de sus
descendientes.

**Punto de diseño a resolver en la revisión:** `Raise(Elevation)` ya usa
`box-shadow`. Si una Part tiene elevación **y** borde de estado, las dos sombras
deben **componerse en una sola declaración separada por comas**, no pisarse. Hay
que decidir el orden (el anillo primero, la elevación después) y dejarlo en un
único sitio del emisor.

**Aceptación:** un test que construya una Part con `Raise` y con un estado que
lleve borde, y verifique que sale **una** declaración `box-shadow` con las dos
capas; y que ninguna regla de estado emite `outline`.

### Etapa 5 — el contrato que falta en la costura: cromo flotante ↔ región con scroll

Este es el que B2 necesita y ninguna hoja puede resolver sola.

El mecanismo debe cruzar la frontera entre widgets **sin que ninguno conozca al
otro**. Las custom properties se heredan, así que sirven exactamente para eso:

1. `Scroll()` pasa a emitir además:
   ```
   padding-block-end: var(--floating-bottom, 0px);
   ```
   Es decir: *toda* región con scroll reserva por defecto lo que le digan que hay
   flotando encima, y `0px` si nadie dice nada. Sin coste para quien no lo use.

2. Una `Option` nueva en el host, del estilo
   `FloatingChrome(edge Edge, size Size, gap Space)`, emite sobre **su propia
   caja**:
   ```
   --floating-bottom: calc(<size> + 2 * <gap>);
   ```
   Como se hereda, cualquier `Scroll()` descendiente lo recoge — esté en el widget
   que esté, venga del repo que venga.

Así el host declara *"ocupo esta franja de mi borde inferior"* y el componente
hijo declara *"esta es mi región con scroll"*, y ninguno necesita saber el nombre
de clase del otro. Es el tipo que faltaba en la costura.

**Aceptación:** un test con forma de consumidor **dentro de este repo** (regla del
harness: una API no está publicada hasta que un test así la demuestra) que monte
un host con `FloatingChrome` y un hijo con `Scroll()` y verifique que el hijo
emite el `padding-block-end` que consume la variable.

| Etapa | Archivos | Puerta |
|---|---|---|
| 1 | `style/emit_decls.go` (solo test), `style/*_test.go` | — |
| 2 | `style/emit_decls.go` | tras publicar `css` |
| 3 | `style/emit_decls.go`, `style/emit_helpers.go`, `style/validate.go` | tras 2 |
| 4 | `style/emit_decls.go` | tras 3 |
| 5 | `style/except.go`, `style/emit_decls.go`, `style/*_test.go` | tras 3 |

Ejecutar `go build ./... && go test ./... -count=1` en la raíz del módulo. Deben
pasar **todos** los paquetes.

## 4. Lo que este plan NO hace

- **No intenta resolver B1** (desplegable que desborda según el scroll-snap). CSS
  no puede observar esa posición. La decisión seguirá siendo explícita en el
  consumidor: anclado al disparador o anclado al viewport.
- No toca `Interactive()` ni la derivación de hover/focus/press.
- No cambia la escala de tokens `Z*`; solo unifica **quién** los usa.
