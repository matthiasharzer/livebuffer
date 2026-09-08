package ui

import "embed"

//go:embed dist/*
//go:embed index.html
var Content embed.FS
