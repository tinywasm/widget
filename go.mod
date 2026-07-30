module github.com/tinywasm/widget

go 1.25.2

require github.com/tinywasm/fmt v0.25.5

require github.com/tinywasm/css v0.4.1

// TEMP: local css checkout — accent + control-height tokens (pending release).
replace github.com/tinywasm/css => ../css
