package main

import (
	"database/sql"
	"net/http"
	"strings"
)

func getSettings(db *sql.DB) Settings {
	s := Settings{}
	err := db.QueryRow(`SELECT nama_isp, alamat, kota, telepon, email, website, npwp, penandatangan, jabatan, rekening, masa_pks_bulan, tpl_pks, tpl_fab
		FROM settings WHERE id=1`).
		Scan(&s.NamaISP, &s.Alamat, &s.Kota, &s.Telepon, &s.Email, &s.Website, &s.NPWP, &s.Penandatangan, &s.Jabatan, &s.Rekening, &s.MasaPKSBulan, &s.TplPKS, &s.TplFAB)
	if err == sql.ErrNoRows {
		def := DefaultSettings()
		_, _ = db.Exec(`INSERT INTO settings(id, nama_isp, alamat, kota, telepon, email, website, npwp, penandatangan, jabatan, rekening, masa_pks_bulan, tpl_pks, tpl_fab)
			VALUES(1,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			def.NamaISP, def.Alamat, def.Kota, def.Telepon, def.Email, def.Website, def.NPWP, def.Penandatangan, def.Jabatan, def.Rekening, def.MasaPKSBulan, def.TplPKS, def.TplFAB)
		return def
	}
	return s
}

func (h *Handlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, getSettings(h.DB))
}

func (h *Handlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if err := readJSON(r, &s); err != nil || strings.TrimSpace(s.NamaISP) == "" {
		httpErr(w, 400, "Nama ISP wajib diisi")
		return
	}
	if s.MasaPKSBulan <= 0 {
		s.MasaPKSBulan = 12
	}
	_, err := h.DB.Exec(`INSERT INTO settings(id, nama_isp, alamat, kota, telepon, email, website, npwp, penandatangan, jabatan, rekening, masa_pks_bulan, tpl_pks, tpl_fab)
		VALUES(1,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			nama_isp=excluded.nama_isp, alamat=excluded.alamat, kota=excluded.kota, telepon=excluded.telepon,
			email=excluded.email, website=excluded.website, npwp=excluded.npwp, penandatangan=excluded.penandatangan,
			jabatan=excluded.jabatan, rekening=excluded.rekening, masa_pks_bulan=excluded.masa_pks_bulan,
			tpl_pks=excluded.tpl_pks, tpl_fab=excluded.tpl_fab`,
		s.NamaISP, s.Alamat, s.Kota, s.Telepon, s.Email, s.Website, s.NPWP, s.Penandatangan, s.Jabatan, s.Rekening, s.MasaPKSBulan, s.TplPKS, s.TplFAB)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan pengaturan")
		return
	}
	writeJSON(w, 200, s)
}

// DefaultTemplates mengembalikan template bawaan agar bisa dipulihkan dari UI.
func (h *Handlers) DefaultTemplates(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"tpl_pks": defaultTplPKS, "tpl_fab": defaultTplFAB})
}

func (h *Handlers) ResetSeed(w http.ResponseWriter, r *http.Request) {
	var recorded int
	if err := h.DB.QueryRow(`SELECT (SELECT COUNT(*) FROM ledger)+(SELECT COUNT(*) FROM pops)+(SELECT COUNT(*) FROM levy_config)+(SELECT COUNT(*) FROM wa_outbox)+(SELECT COUNT(*) FROM wa_config)+(SELECT COUNT(*) FROM pop_letters)+(SELECT COUNT(*) FROM pop_letter_profile)`).Scan(&recorded); err != nil {
		httpErr(w, 500, "Gagal memeriksa data operasional")
		return
	}
	if recorded > 0 {
		httpErr(w, 409, "Reset data contoh dinonaktifkan karena sudah ada pembukuan atau data POP.")
		return
	}
	if err := ResetSeed(h.DB); err != nil {
		httpErr(w, 500, "Gagal memuat ulang data contoh")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
