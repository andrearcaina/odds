package web

import (
	"embed"
	"io/fs"
)

//go:embed static/index.html
var indexHTML string

//go:embed static
var staticFiles embed.FS

//go:embed static/index.js
var indexJS []byte

//go:embed static/index.css
var indexCSS []byte

var staticAssets fs.FS

func init() {
	staticAssets, _ = fs.Sub(staticFiles, "static")
}
