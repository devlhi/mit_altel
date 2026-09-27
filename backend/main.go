package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// serveStatic menyajikan build SPA (frontend/.output/public) dari binary yang
// sama, dengan fallback ke index.html untuk rute sisi klien. Rute /api/* tidak
// pernah sampai ke handler ini karena pattern mux yang lebih spesifik menang.
func serveStatic(mux *http.ServeMux, dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("path static tidak valid: %v", err)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		clean := filepath.Clean(r.URL.Path)
		full := filepath.Join(abs, clean)
		if st, err := os.Stat(full); err == nil && !st.IsDir() {
			http.ServeFile(w, r, full)
			return
		}
		http.ServeFile(w, r, filepath.Join(abs, "index.html"))
	})
	log.Printf("[isp-api] menyajikan frontend dari %s", abs)
}

func main() {
	addr := flag.String("addr", envOr("APP_ADDR", ":8080"), "alamat listen server")
	dbPath := flag.String("db", envOr("APP_DB", "data/isp.db"), "lokasi file database SQLite")
	staticDir := flag.String("static", envOr("APP_STATIC", ""), "direktori build frontend (opsional)")
	flag.Parse()

	db, err := OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	RegisterRoutes(mux, db)
	if *staticDir != "" {
		serveStatic(mux, *staticDir)
	}

	handler := LogMiddleware(CORSMiddleware(AuthMiddleware(mux, db)))
	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("[isp-api] berjalan di http://%s (db: %s)", strings.TrimPrefix(*addr, ":"), *dbPath)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
