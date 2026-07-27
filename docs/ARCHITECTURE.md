# Arquitectura de `tinywasm/widget`

Este documento define el **Qué** y el **Por Qué** del contrato visual, la anatomía, los estados, la disposición y el sistema de estilos de los componentes en la arquitectura de `tinywasm/widget`.

---

## Qué es `tinywasm/widget`

`tinywasm/widget` es el módulo que gobierna la estructura de los componentes visuales de la aplicación. Nombra y unifica de qué piezas está hecho un componente visual, en qué estados puede estar y cómo se disponen esas piezas sin acoplarse a datos, transporte, DOM, o texto CSS libre.

El sistema está dividido estrictamente en dos paquetes con objetivos y restricciones opuestas:

1.  **`github.com/tinywasm/widget` (Raíz, compatible con WASM):**
    *   **Propósito:** Define la identidad, anatomía del componente (Open UI), estados, cues del navegador y tipos ARIA.
    *   **Restricción:** Debe ser extremadamente ligero (sin lógica de aspecto ni generación de estilos) porque **viaja al binario WASM**.

2.  **`github.com/tinywasm/widget/style` (Hoja de estilos, exento de WASM):**
    *   **Propósito:** Define escalas cerradas de tamaño, tipografía, elevación y color (Surface), primitivas de flujo (Flow) y excepciones de diseño. Genera hojas de estilo CSS deterministas en capas fijas.
    *   **Restricción:** Excluido de compilar en WebAssembly usando la etiqueta de compilación `//go:build !wasm` para evitar que viaje al binario cliente.

---

## Estructura y Componentes Core

### 1. El Contrato de Anatomía y Estados (Paquete Raíz)

*   **`Name`:** Identificador único de un widget, usado como prefijo para todas sus clases y selectores. Previene colisiones de estilos.
*   **`Part`:** Ranura nombrada local al widget (por ejemplo, `"row"`, `"menu"`, `"header"`).
*   **`Class`:** Nombre de clase CSS derivado de forma determinista desde un `Name` y un `Part` (por ejemplo, `.targetlist__row`). Es un tipo opaco sin constructor público fuera del paquete para garantizar la coherencia de nombres por construcción.
*   **`State` vs `Cue`:**
    *   `State` representa los estados lógicos que posee el widget en Go (como `Selected`, `Disabled`, `Open`, `Invalid`). Genera atributos de datos del DOM de forma coincidente por construcción (`data-selected="true"`).
    *   `Cue` representa los estados interactivos propios del navegador (como `Hover`, `Focus`, `Press`, `Target`) que solo existen en la hoja de estilos como pseudo-clases CSS (`:hover`, `:focus`).

---

## El Sistema de Estilos y Disposición (`widget/style`)

La hoja de estilos está diseñada bajo el principio de **cero escape**: no acepta strings libres, unidades de viewport (`vw`/`vh`) ni herramental de estilización arbitrario.

### 1. Escalas Cerradas y Enums
Todas las propiedades visuales se especifican mediante escalas estrictas y tipadas:
*   `Space`: Escala de espaciado del sistema (de 0 a 12).
*   `Radius`: Curvatura de bordes (`RadiusSm`, `RadiusMd`, `RadiusLg`, `RadiusFull`).
*   `TextSize` & `Weight`: Sistema tipográfico cerrado.
*   `Elevation`: Sombras de elevación (`Flat`, `Raised`, `Floating`, `Overlay`).
*   `Size`: Ancho relativo exclusivo del contenedor (`Content`, `Prose`, `Third`, `Half`, `TwoThirds`, `Full`). No se permite la declaración directa de alto (`height`), promoviendo el flujo automático de CSS.

### 2. El Color como una Decisión Completa: `Surface`
No existe la selección libre de colores de fondo o texto. `Surface` unifica fondo, texto y bordes en una decisión visual coherente (por ejemplo, `Page`, `Panel`, `Sunken`, `Accent`), garantizando que siempre se mantenga el contraste adecuado.

### 3. Primitivas de Flujo Responsivo (`Flow`)
Inspirado en *Every Layout*, define primitivas como `Stack`, `Row`, `Split` y `Grid`. Son intrínsecamente responsivas y utilizan container queries (`@container`) en lugar de media queries (`@media`), reaccionando al ancho de su propio contenedor.

---

## Garantías de Emisión de CSS

El motor de emisión de hojas de estilo (`emit.go`) produce CSS que cumple con las siguientes garantías:

1.  **Orden de Capas Fijo (Cascade Layers):**
    ```css
    @layer tokens, primitives, widgets, states;
    ```
    Previene problemas de especificidad y elimina la necesidad de usar `!important`.
2.  **Salida Determinista:** La emisión para una misma definición produce byte a byte el mismo texto exacto, facilitando las revisiones de cambios (diffs).
3.  **Especificidad Plana:** Todos los selectores generados son de un único nivel de profundidad (por ejemplo, `.clase` o `.clase[data-state]`), evitando el acoplamiento al árbol del DOM.
