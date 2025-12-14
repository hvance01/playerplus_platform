package handler

import (
	"embed"
)

//go:embed all:dist
var frontendFS embed.FS
