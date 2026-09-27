package main

import (
	"database/sql"
	"net/http"
	"sync"
)

var operationMu sync.RWMutex

func operationHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			operationMu.RLock()
			defer operationMu.RUnlock()
		} else {
			operationMu.Lock()
			defer operationMu.Unlock()
		}
		fn(w, r)
	}
}

type Handlers struct {
	DB *sql.DB
}

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	h := &Handlers{DB: db}
	registerOperations(mux, h)
	registerMessaging(mux, h)
	registerPOPLetters(mux, h)
	handle := func(pattern string, fn http.HandlerFunc) { mux.HandleFunc(pattern, operationHandler(fn)) }

	// Auth
	handle("POST /api/login", h.Login)
	handle("GET /api/me", h.Me)
	handle("PUT /api/password", h.ChangePassword)

	// Dashboard
	handle("GET /api/dashboard", h.Dashboard)

	// Settings
	handle("GET /api/settings", h.GetSettings)
	handle("PUT /api/settings", h.UpdateSettings)
	handle("GET /api/settings/defaults", h.DefaultTemplates)

	// Mitra
	handle("GET /api/mitra", h.ListMitra)
	handle("POST /api/mitra", h.CreateMitra)
	handle("GET /api/mitra/{id}", h.GetMitraDetail)
	handle("PUT /api/mitra/{id}", h.UpdateMitra)
	handle("DELETE /api/mitra/{id}", h.DeleteMitra)

	// Paket
	handle("GET /api/paket", h.ListPaket)
	handle("POST /api/paket", h.CreatePaket)
	handle("PUT /api/paket/{id}", h.UpdatePaket)
	handle("DELETE /api/paket/{id}", h.DeletePaket)

	// Pelanggan
	handle("GET /api/pelanggan", h.ListPelanggan)
	handle("POST /api/pelanggan", h.CreatePelanggan)
	handle("PUT /api/pelanggan/{id}", h.UpdatePelanggan)
	handle("DELETE /api/pelanggan/{id}", h.DeletePelanggan)

	// Tiket
	handle("GET /api/tiket", h.ListTiket)
	handle("POST /api/tiket", h.CreateTiket)
	handle("PUT /api/tiket/{id}", h.UpdateTiket)
	handle("DELETE /api/tiket/{id}", h.DeleteTiket)

	// Keuangan
	handle("GET /api/keuangan", h.GetKeuangan)
	handle("POST /api/keuangan/generate", h.GenerateTagihan)
	handle("PUT /api/tagihan/{id}", h.UpdateTagihan)
	handle("DELETE /api/tagihan/{id}", h.DeleteTagihan)

	// Dokumen (PKS & FAB)
	handle("GET /api/dokumen", h.ListDokumen)
	handle("POST /api/dokumen", h.CreateDokumen)
	handle("DELETE /api/dokumen/{id}", h.DeleteDokumen)
	handle("GET /api/dokumen/{id}/html", h.DokumenHTML)
	handle("POST /api/dokumen/{id}/kirim", h.DokumenKirim)

	// Utilitas data
	handle("POST /api/seed", h.ResetSeed)
}
