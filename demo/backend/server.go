package main

import "net/http"

func newRouter(h *Handler, staticDir string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir(staticDir)))
	mux.HandleFunc("GET /visit", h.handleVisit)
	mux.HandleFunc("POST /convert", h.handleConvert)
	mux.HandleFunc("GET /flag", h.handleFlag)
	mux.HandleFunc("GET /health", h.handleHealth)
	return mux
}
