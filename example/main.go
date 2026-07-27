package main

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func main() {
	name := widget.Name("button")
	fmt.Println(name.Root().String())
}
