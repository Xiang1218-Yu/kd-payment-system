// Package handler is the HTTP transport layer. Each handler does only three
// things: decode the request, call one service method, encode the response.
// All business logic lives in the service layer, so handlers stay thin and
// free of pricing/scheduling/store knowledge.
package handler

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"kd-payment-system/backend/internal/service"
)

//go:embed dist/*
var webFS embed.FS

// New builds the mux wiring every API route and falling back to the embedded
// SPA for everything else.
func New(svc *service.Services) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/regions", regionsHandler(svc))
	mux.HandleFunc("GET /api/cabinets", cabinetsHandler(svc))
	mux.HandleFunc("GET /api/cabinets/{id}", cabinetDetailHandler(svc))
	mux.HandleFunc("GET /api/pricing/{cabinetId}", pricingHandler(svc))
	mux.HandleFunc("POST /api/dropoff", dropoffHandler(svc))
	mux.HandleFunc("POST /api/pickup", pickupHandler(svc))
	mux.HandleFunc("GET /api/stats/dashboard", dashboardHandler(svc))
	mux.HandleFunc("GET /api/sim/state", simStateHandler(svc))
	mux.HandleFunc("POST /api/sim/tick", simTickHandler(svc))
	mux.HandleFunc("POST /api/sim/reset", simResetHandler(svc))

	// Serve the embedded frontend. Files under dist/* are served at their
	// path; anything else falls back to index.html so client-side routing
	// (e.g. /pricing) works on refresh.
	mux.Handle("/", spaHandler())

	return mux
}

// spaHandler serves embedded static files with SPA fallback to index.html.
func spaHandler() http.Handler {
	sub, err := fs.Sub(webFS, "dist")
	if err != nil {
		// dist missing means the frontend wasn't built before compiling.
		// Serve a helpful notice instead of crashing.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend not built; run `npm run build` in /frontend", http.StatusServiceUnavailable)
		})
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(sub, path); err != nil {
			// Unknown path → SPA fallback so client routes resolve.
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
