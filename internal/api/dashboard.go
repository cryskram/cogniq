package api

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var dashboardFS embed.FS

func dashboardHandler() http.Handler {
	sub, err := fs.Sub(dashboardFS, "web")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}
