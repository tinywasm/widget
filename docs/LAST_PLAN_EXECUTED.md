---
PLAN: "fix(style): el Flyout debe colgar de su Anchor, y el DSL debe decirlo cuando no puede"
TAG: v0.6.0
EXECUTOR: opencode
REVIEWER: none
---

> **EJECUTADO en v0.6.0 (etapa 1: `Within()` + `Validate()` + firmas honestas).**
> La etapa 2 (la construcción legal para `components`) se decide en el plan B.
>
> **Etapa A de un cambio en 2 repos.**
>
> | | Repo | Plan | Qué |
> |---|---|---|---|
> | **A** | **`widget`** | **este plan** | cierra el hueco: el sheet aprende el árbol de partes y `Validate()` deja de callarse |
> | B | `components` | `components/docs/PLAN.md` | `targetlist` adopta la construcción legal; `usermenu` gana el test que nunca tuvo |

# Plan — `widget/style`: el contenedor de bloque del `Flyout`

## El bug, medido

En desktop (1440x900), el menú `⋮` de una fila de `targetlist` abre su desplegable
**encima de su propia fila**, tapando la etiqueta que acababa de tocar.

Medido en el navegador **antes** de escribir una sola línea de este plan:

```
row      top 113.2   bottom 163.2   (alto 50)
summary  top 118.0   bottom 142.0   (alto 24)
options  top 142.0                  → 21.2px POR ENCIMA del fondo de su fila
label    top 126.2   bottom 150.2   → el desplegable le come 8.2px de texto

options.offsetParent === .targetlist__menu     ← la prueba
```

`offsetParent` es el dato que cierra el caso: el contenedor de bloque del
desplegable **no es la fila**, es el `<details>` de 24px.

## Por qué pasa

Tres opciones del DSL, escritas exactamente como las documenta su firma:

```go
Part(PartRow,     style.Anchor())                                        // position: relative
Part(PartMenu,    style.Docked(style.Parent, EdgeTop, SideStart, Space1)) // position: absolute  ← roba
Part(PartOptions, style.Flyout(style.SideStart))                          // position: absolute
                                                                          // inset-block-start: 100%
```

`Flyout` emite `inset-block-start: 100%`, y CSS resuelve ese `100%` contra el
**ancestro posicionado más cercano**, no contra el que declaró `Anchor()`. El
`<details>` intermedio también es `position: absolute`, así que se interpone: el
`100%` vale 24px (el alto del disparador) en vez de 50px (el alto de la fila).

El `Anchor()` de la fila es **código muerto**. Nada cuelga nunca de él.

## Por qué esto es un defecto de `widget`, no de `components`

Tres cosas fallan a la vez, y las tres viven aquí.

### 1. Las firmas prometen algo que la emisión no cumple

En `overlay.go`, hoy:

```go
// Flyout … El ancestro Anchor() más cercano es de donde cuelga.
// Docked … Parent lo fija a la esquina del Anchor más cercano.
```

Las dos frases son **falsas** en cuanto algo intermedio esté posicionado, y el
autor no tiene forma de verlo desde el punto de llamada. Es exactamente el
"fallo silencioso" que el principio 6 del harness existe para eliminar: ni error
de compilación, ni diagnóstico, sólo un desplegable en el sitio equivocado.

### 2. `Validate()` ya persigue este error, pero sólo la mitad fácil

`validate.go` ya tiene `checkPosition`, con este comentario:

> *Anchor() junto a Docked()/Flyout() es el error fácil: los dos ya son
> contenedores de bloque, así que el Anchor sobra además de ser destructivo.*

O sea: **el problema ya está identificado en el código**, pero `checkPosition`
sólo mira *dentro de una regla*. La versión **entre partes** — la que sí se
publicó rota — es estructuralmente invisible, porque `s.partRules` es un mapa
plano: el sheet no sabe qué parte se renderiza dentro de cuál.

### 3. El consumidor ya carga la regla memorizada

`targetlist/css.go` lleva escrito, a mano:

> *No `Anchor()` aquí — `Docked` ya lo hace contenedor de bloque, y los dos se
> pelean por `position`.*

Un comentario así **es** el síntoma. La checklist del harness lo nombra:
*"Cosas que hay que recordar. Cualquier paso obligatorio que el autor deba
acordarse de hacer → eso es un agujero en el harness; se cierra con tipos o con
un único camino, no con prosa."*

## Los tests ya están escritos y en rojo

Antes de este plan, y para que quede de resguardo permanente.

**`widget/style/anchor_contract_test.go`** (nuevo, ya en el repo):

| Test | Hoy | Qué fija |
|---|---|---|
| `TestAnchorAndFlyoutWithNothingBetween` | 🟢 **verde** | El caso de control: la forma de `usermenu` (`Anchor > Flyout`, nada en medio) es la que el DSL documenta y **debe seguir funcionando**. Cualquier diagnóstico que también dispare aquí se pasó de largo. |
| `TestDockedPartBetweenAnchorAndFlyoutIsDiagnosed` | 🔴 **rojo** | Reproduce la composición de `targetlist` por la API pública, tal cual la escribe un consumidor, y exige que la librería **diga algo**. Hoy `Validate()` devuelve cero errores. |

El test rojo asserta sobre **el diagnóstico**, no sobre el CSS emitido, a
propósito: sea cual sea la construcción que elija la etapa 2, lo que no puede
volver a pasar es que un consumidor escriba esto y reciba silencio.

---

## Etapa 1 — que el sheet conozca el árbol de partes

**Esta etapa es la que cierra el harness. Sin ella la etapa 2 es otro parche.**

### 1.1 El precedente que ya existe

No hay que inventar un concepto nuevo. `Sheet` **ya** tiene la relación
contenedor→parte, sólo que limitada a los cues:

```go
CueWithin(container, part widget.Part, …)
CueWithinHover(container, part widget.Part, …)
```

y `validate.go` **ya** valida los dos extremos (`checkCueWithin`: que ambos sean
partes declaradas, y que no sean la misma). Lo que falta es promover esa relación
de "detalle de un selector de cue" a **dato estructural del sheet**.

### 1.2 La declaración

```go
// Within declara que part se renderiza DENTRO de container, y aplica las
// opciones a part. Es la misma parte que Part() declararía; lo que añade es
// la relación de contención, que el sheet necesita para razonar sobre
// posicionamiento (quién es el contenedor de bloque de quién).
func (s *Sheet) Within(container, part widget.Part, opts ...Option) *Sheet
```

Se lee como el DOM: `Within(PartMenu, PartOptions, style.Flyout(...))` es
"options, dentro de menu".

**No es una migración masiva.** `Part()` sigue existiendo y sigue siendo lo
normal. `Within()` sólo hace falta donde la contención cambia el resultado —
y la etapa 1.3 se encarga de que en esos casos no sea opcional.

### 1.3 El diagnóstico, cerrado por defecto

Regla nueva en `Validate()`:

- Si el sheet declara **a la vez** una parte con `Flyout(...)` y una parte que
  es contenedor de bloque (`Docked(Parent, …)`, `OnEdge`, `Backdrop(Parent)`),
  y **no** ha declarado la contención entre ellas con `Within`:
  → **error**: composición ambigua, declara la anidación.

- Si **sí** la ha declarado, y en la cadena entre el `Flyout` y su `Anchor` hay
  una parte posicionada:
  → **error preciso**, nombrando las dos partes y qué está robando qué.

Esto es el principio 8 (*cerrado por defecto*): el silencio se rechaza; abrir
cuesta una línea explícita y greppable. El autor que no sabe nada de esto
escribe el código, compila, y **el sheet le dice** — que es la definición de
harness cerrado del propio documento.

> ⚠️ El caso de control (`usermenu`: `Anchor > Flyout`, ningún `Docked` en el
> sheet) **no debe disparar**. `TestAnchorAndFlyoutWithNothingBetween` es el
> guardia de eso y ya está verde: si se pone rojo, el diagnóstico se pasó.

### 1.4 Arreglar las firmas que mienten

En `overlay.go`, reescribir los comentarios de `Anchor`, `Docked` y `Flyout`
para que digan **lo que la emisión hace de verdad**: el contenedor de bloque es
el ancestro posicionado más cercano, y `Anchor()` sólo gana si nada se
interpone. Principio 7: *si usar la API exige leer un documento largo, la API
está incompleta* — pero un comentario que afirma algo falso es peor que no tener
comentario.

---

## Etapa 2 — una forma legal de escribirlo

La etapa 1 convierte el fallo silencioso en un error. Falta que exista una
llamada **correcta** que el autor pueda escribir en su lugar; si no, sólo
habríamos cambiado un bug por un callejón sin salida.

Tres candidatas. **La decisión se toma con medidas reales en la etapa B**, no
aquí; este plan las ordena por coste creciente para `widget`.

### 2a — El disparador vuelve al flujo *(recomendada: no añade API, la quita)*

Preguntarse por qué el `<details>` está `Docked`. El comentario del consumidor
dice: *"tanto el menú como el badge salen del flujo, así que la etiqueta es lo
único que dimensiona la fila"*. Pero la fila ya tiene
`min-height: var(--control-height)` = 50px, y el menú mide 24px: **un menú en
flujo no dimensiona nada**. El que sí necesita salir del flujo es el badge, que
envuelve.

Si el disparador va en flujo como primer hijo flex de la fila:

- se coloca solo en el borde de entrada, sin `Docked`;
- el ancestro posicionado más cercano del `Flyout` pasa a ser **la fila**;
- `Flyout` cumple su documentación **por construcción**, no por vigilancia;
- sobra el `PadInline(Space8)` de la etiqueta, que existía sólo para esquivarlo.

**Coste, y hay que medirlo, no suponerlo:** la fila tiene `Pad(Space3)` = 12px,
así que en flujo el icono arranca a 12px en vez de los 4px del `Docked(Space1)`.
En el *sliver* móvil de ~37.5px que deja `MasterDetail(Most)`, eso son 12+24=36px
contra 4+24=28px. Entra, pero sin margen. **Verificar en dispositivo antes de
cerrar esta opción** — es exactamente el presupuesto de píxeles que el
comentario de `targetlist` dice que está ajustado.

Para `widget` esta opción cuesta **cero API nueva**. Es la que prefiere la
doctrina: *colapsar caminos redundantes, superficie mínima*.

### 2b — `Docked` con modo "abarca el Anchor"

Un `Docked` que se fija a **los dos** bordes de bloque en vez de a una esquina:

```
inset-block-start: <gap>;  inset-block-end: <gap>;
```

La caja del `<details>` pasa a medir lo que la fila, y el `100%` del `Flyout`
llega abajo.

**Costes, los dos reales:**
1. El desplegable cuelga de `anchor.bottom − gap`, no de `anchor.bottom`: con
   `Space1` se mete 4px dentro de la fila. Casi exacto, no exacto.
2. La caja del `<details>` pasa a ser una franja de 24×42px que **intercepta
   clics** destinados a la fila (que es seleccionable). Se mitiga con
   `pointer-events: none` en el contenedor y `auto` en el `<summary>` y el
   panel — pero eso es **otra primitiva más** que añadir aquí.

### 2c — El panel deja de ser descendiente del disparador

El `<div class="options">` pasa a ser hermano del `<details>` e hijo directo de
la fila. El contenedor de bloque pasa a ser **exactamente** el `Anchor`.

Exige una primitiva de estado por descendencia — `.row:has(.menu[open]) .options`
— es decir un `RevealedByWithin(container, state)` en este repo. `:has()` está
soportado en todos los navegadores modernos desde 2023.

Es la más correcta estructuralmente y la más cara. Queda documentada como la
salida si 2a no pasa la medida del *sliver* y 2b no convence.

---

## Orden de ejecución

| # | Etapa | Entregable | Verde cuando |
|---|---|---|---|
| 1 | 1.1–1.3 | `Within()` + reglas nuevas en `Validate()` | `TestDockedPartBetweenAnchorAndFlyoutIsDiagnosed` pasa **y** `TestAnchorAndFlyoutWithNothingBetween` sigue pasando |
| 2 | 1.4 | comentarios de `Anchor`/`Docked`/`Flyout` corregidos | revisión |
| 3 | 2a/2b/2c | la construcción elegida | el test de `components` (etapa B) pasa |

Las etapas 1 y 2 se pueden publicar sin la 3: dejan el bug **visible** aunque
todavía no arreglado, que ya es mejor que hoy.

---

## Lo que este plan NO hace

- **No toca `tinywasm/css`.** No hace falta ningún token nuevo. Verificado.
- **No toca `tinywasm/layout`.** Se buscaron todas las llamadas a
  `Anchor`/`Docked(Parent)`/`Flyout` en los cuatro repos: en `layout` hay
  `Docked` (crudview, rightpanel, platformd) pero **ningún `Flyout`**, así que
  la composición que rompe no existe ahí. No es un olvido, es una comprobación.
- **No toca `components`.** Ese es el plan B, y depende de que esto se publique.
- **No cambia `checkPosition`.** La regla dentro-de-una-regla que ya existe es
  correcta; lo que se añade es la versión entre partes.
- **No adopta CSS Anchor Positioning** (`anchor-name`/`position-anchor`).
  Resolvería esto de raíz, pero el soporte no da para un sistema de diseño en
  producción. Se descarta explícitamente para que nadie lo re-proponga sin datos
  nuevos.

## Hallazgos aparcados (no son de este plan)

- **`usermenu` no tiene ni un fichero de test.** El caso que funciona es el que
  nadie vigila. Se recoge en el plan de `components`.
- **El solape residual de ~4.4px entre badge y botón flotante** en móvil sigue
  abierto; es interno de `targetlist` y no tiene que ver con este contenedor de
  bloque.
- **El solape hamburguesa/searchbar en `platformd`** queda fuera del alcance de
  `FloatingChrome` (la searchbar no es una región `Scroll()`).
