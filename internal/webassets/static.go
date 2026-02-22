package webassets

import "embed"

// FS contains the embedded frontend assets.
//
//go:embed static/*
var FS embed.FS
