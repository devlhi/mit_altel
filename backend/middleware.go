package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path != "/api/dashboard" { // kurangi noise polling
			path := r.URL.Path
			if strings.HasPrefix(path, "/share/") {
				path = "/share/[private]"
			}
			log.Printf("[api] %s %s (%s)", r.Method, path, time.Since(start).Round(time.Millisecond))
		}
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	// CORS hanya aktif bila APP_ORIGIN diisi (mis. saat dev: APP_ORIGIN=http://localhost:3000).
	// Produksi same-origin tidak membutuhkan header CORS sama sekali.
	origin := envOr("APP_ORIGIN", "")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
