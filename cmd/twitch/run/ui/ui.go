package ui

import (
	"embed"
	"io/fs"
)

//go:embed public/*
var uiRoot embed.FS

var Content fs.FS

func init() {
	var err error
	Content, err = fs.Sub(uiRoot, "public")
	if err != nil {
		panic(err)
	}
}
