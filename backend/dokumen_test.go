package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDocumentRendering(t *testing.T) {
	s := DefaultSettings()
	s.NamaISP = "ISP <script>alert(1)</script>"
	m := Mitra{Nama: "Mitra & Kawan", Pemilik: "Pemilik", ShareMitra: 32.5}
	rc := &FABRincian{Periode: "2026-09", SharePct: 32.5, TotalLunas: 1000000, Nilai: 325000, Rows: []FABRow{}}
	for _, kind := range []string{"PKS", "FAB"} {
		t.Run(kind, func(t *testing.T) {
			var details *FABRincian
			if kind == "FAB" {
				details = rc
			}
			html := buildDocHTML(s, m, "TEST/2026/001", kind, "2026-09-27", details, 12, true)
			for _, want := range []string{"32,5%", "Mitra &amp; Kawan", "&lt;script&gt;", "doc-footer", "@page{size:A4", "break-inside:avoid"} {
				if !strings.Contains(html, want) {
					t.Errorf("missing %q", want)
				}
			}
			if strings.Contains(html, "{{") {
				t.Error("unresolved placeholders")
			}
			if strings.Contains(html, "<script>") {
				t.Error("preview must not execute print or data scripts")
			}
			if kind == "FAB" && !strings.Contains(html, "Tidak ada tagihan pada periode ini.") {
				t.Error("missing empty invoice state")
			}
		})
	}
	// Published FAB must use its recorded percentage, even if a different mitra value is supplied.
	m.ShareMitra = 40
	html := buildDocHTML(s, m, "FAB/TEST", "FAB", "2026-09-27", rc, 12, false)
	if !strings.Contains(html, "32,5%") || !strings.Contains(html, "window.print()") {
		t.Error("snapshot percentage or print action missing")
	}
	// Custom templates remain the source of the document body.
	s.TplPKS = "<p>Ketentuan khusus untuk {{nama_mitra}}</p>"
	html = buildDocHTML(s, m, "PKS/TEST", "PKS", "2026-09-27", nil, 12, true)
	if !strings.Contains(html, "Ketentuan khusus untuk Mitra &amp; Kawan") || strings.Contains(html, "Pasal 1") {
		t.Error("custom template was replaced")
	}
}

func TestDocumentValidationAndPreview(t *testing.T) {
	db, err := OpenDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{DB: db}
	var mitraID int64
	if err := db.QueryRow("SELECT id FROM mitra LIMIT 1").Scan(&mitraID); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, kind, date, period string
		months, status           int
	}{
		{"invalid date", "PKS", "2026-02-30", "", 12, 400},
		{"invalid period", "FAB", "2026-09-27", "2026-13", 12, 400},
		{"negative duration", "PKS", "2026-09-27", "", -1, 400},
		{"excessive duration", "PKS", "2026-09-27", "", 121, 400},
		{"valid PKS", "PKS", "2026-09-27", "", 24, 201},
		{"valid zero FAB", "FAB", "2026-09-27", "2099-01", 12, 201},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"jenis": tc.kind, "mitra_id": mitraID, "tanggal": tc.date, "periode": tc.period, "masa_bulan": tc.months})
			w := httptest.NewRecorder()
			h.CreateDokumen(w, httptest.NewRequest(http.MethodPost, "/api/dokumen", strings.NewReader(string(body))))
			if w.Code != tc.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if tc.status != 201 {
				return
			}
			var d Dokumen
			if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
				t.Fatal(err)
			}
			token, err := issueToken(db, 1)
			if err != nil {
				t.Fatal(err)
			}
			mux := http.NewServeMux()
			mux.HandleFunc("GET /api/dokumen/{id}/html", h.DokumenHTML)
			request := httptest.NewRequest(http.MethodGet, "/api/dokumen/"+strconv.FormatInt(d.ID, 10)+"/html?mode=preview", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			preview := httptest.NewRecorder()
			mux.ServeHTTP(preview, request)
			if preview.Code != 200 || !strings.Contains(preview.Body.String(), d.Nomor) {
				t.Errorf("preview failed: %d", preview.Code)
			}
			if preview.Header().Get("Cache-Control") != "no-store" {
				t.Error("document preview must not be cached")
			}
			unauth := httptest.NewRecorder()
			mux.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, request.URL.String(), nil))
			if unauth.Code != http.StatusUnauthorized {
				t.Error("preview must require authentication")
			}
		})
	}
}
