module github.com/tinywasm/widget

go 1.25.2

require github.com/tinywasm/fmt v0.25.5

require github.com/tinywasm/css v0.4.10

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
)

// TEMP local dev only — do not publish/commit: picks up uncommitted
// css changes (light-dark()/color-mix() Safari-legacy fallback) so the
// running demo reflects them immediately.
