package widget

import "github.com/tinywasm/fmt"

// Name identifica un widget. Es el prefijo de TODA clase que emite, así que dos
// widgets no pueden colisionar aunque elijan el mismo nombre de parte.
type Name string

// Part es una ranura nombrada de la anatomía de un widget (vocabulario Open UI).
// Es local al widget: "row", "menu", "header". Nunca lleva prefijo.
type Part string

// Class es un nombre de clase CSS. NO tiene constructor público: la única forma
// de obtener una es derivarla de un Name. Escribir Class("lo-que-sea") no compila
// fuera de este paquete.
type Class string

func (c Class) String() string {
	return string(c)
}

func (c Class) AsAttr() fmt.KeyValue {
	return fmt.KeyValue{Key: "class", Value: string(c)}
}

// Root es la clase exterior del widget.
func (n Name) Root() Class {
	return Class(n)
}

// Class deriva la clase de una parte: "targetlist__row".
func (n Name) Class(p Part) Class {
	return Class(string(n) + "__" + string(p))
}
