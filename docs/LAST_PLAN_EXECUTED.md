---
PLAN: "fix(style): un Flyout dentro de una región Scroll() se recorta en silencio"
TAG: v0.6.1
EXECUTOR: opencode
REVIEWER: none
---

> **EJECUTADO en v0.6.1 (candidato 1: `Validate()` rechaza la cadena de
> contención que pasa por un `Scroll()`).** Candidato 2 (una primitiva que
> escape del recortador) queda aparcado hasta que un consumidor lo necesite.
>
> Continuación del anterior (`LAST_PLAN_EXECUTED.md`), que cerró el robo del
> contenedor de bloque. Este es el hueco que quedó **debajo** de aquél y que
> sólo se vio al medir con el primero ya arreglado.

# Plan — `widget/style`: el recortador que el `Flyout` no ve

## El hallazgo, medido

Con el `Flyout` ya anclado correctamente a su fila (etapa A/B anteriores),
abriendo el menú de la **última** fila de la lista, a 1440x900:

```
list     top 113.2   bottom 828.4      (.targetlist__list, overflow-y: auto)
options  top 818.4   bottom 903.2      alto real 84.8
                                       → 10px visibles. 74.8 recortados.
```

`.targetlist__list` es una región `Scroll()`. Un descendiente
`position: absolute` queda recortado por ella, por especificación. Anclar bien
el `Flyout` arregló que se pintara encima de su propia fila; lo que hizo fue
dejarlo **atrapado dentro del recortador**.

## Por qué es el mismo defecto otra vez

Dos primitivas se encuentran — `Scroll()` en un ancestro, `Flyout()` en un
descendiente — y **no hay ningún tipo que nombre lo que cruza entre ellas**. El
autor escribe las dos llamadas correctamente, por separado, y el resultado es un
panel a medias. Ni error de compilación, ni diagnóstico: un fallo silencioso,
principio 6.

Es literalmente el patrón que `FloatingChrome`/`Scroll()` ya resolvió para el
otro seam de este mismo par (la reserva de espacio bajo un botón flotante). Falta
la mitad del recorte.

Y tiene el historial que lo confirma: el `Docked(Viewport, …)` que `targetlist`
llevaba en móvil **era** la vía de escape de este recortador. Escapaba saliendo
de la fila entera, aterrizaba a 502px en una esquina de pantalla encima de dos
filas ajenas, y eso obligó a un `Veil()` que desenfocaba la propia fila sobre la
que se estaba actuando. Tres parches encadenados, todos aguas abajo, todos
porque el contrato falta aguas arriba.

## Estado del consumidor

`targetlist` **ya no lo sufre**: sus opciones dejaron de ser un overlay y son un
acordeón en flujo dentro de la fila. Misma última fila, medida después:
**41.6px de 41.6px visibles**. Pero eso es una decisión de ese componente, no un
arreglo del DSL.

`usermenu` sigue teniendo un `Flyout` real. Hoy no está dentro de un scroller y
por eso funciona — por suerte, no por contrato. Nada impide que el próximo
consumidor lo meta en uno y vuelva a descubrir esto en un navegador.

## Candidatos

### 1 — `Validate()` lo rechaza *(recomendada)*

Reutiliza exactamente el árbol de partes que el plan anterior introdujo con
`Within()`: si la cadena de contención declarada de un `Flyout` pasa por una
parte con `Scroll()`, es un error.

Barato, sin CSS nuevo, y convierte el fallo silencioso en un diagnóstico de
desarrollo — que es todo lo que el harness pide. El autor entonces elige a
conciencia: acordeón en flujo, o sacar el panel del scroller.

### 2 — Una primitiva que salga del recortador de verdad

Reposicionamiento con `position: fixed` calculado, o el `top layer` del
navegador (`popover`). Resuelve el caso en lugar de prohibirlo, y cuesta
bastante más. Sólo si aparece un consumidor que de verdad necesite un overlay
dentro de un scroller y no le sirva el acordeón.

## Tests

El guardia del consumidor ya existe y está verde:
`components/targetlist/targetlist_test.go` → `TestOptionsNeverLeaveTheFlow`,
que prohíbe `position: absolute|fixed` en `.targetlist__options` en cualquier
dispositivo, con los tres bugs encadenados documentados en su comentario.

Falta el de esta librería, con forma de consumidor: un sheet que declare
`Part(list, Scroll())` + `Within(list, panel, Flyout(...))` y exija que
`Validate()` lo rechace. Hoy devolvería cero errores.

## Lo que este plan NO hace

- **No toca `components`.** `targetlist` ya está fuera del problema y `usermenu`
  no lo tiene. Cuando la etapa 1 exista, `usermenu` sólo tiene que seguir
  validando.
- **No revierte el acordeón.** Es la solución correcta para una lista de filas
  accionables, independientemente de lo que se decida aquí.
- **No persigue el solape residual de ~4.4px** entre badge y botón flotante en
  móvil, que sigue aparcado.
