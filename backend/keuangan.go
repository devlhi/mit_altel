package main

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"time"
)

type RekapRow struct {
	MitraID        int64   `json:"mitra_id"`
	MitraNama      string  `json:"mitra_nama"`
	SharePct       float64 `json:"share_pct"`
	PelangganAktif int     `json:"pelanggan_aktif"`
	TotalTagihan   int64   `json:"total_tagihan"`
	TotalLunas     int64   `json:"total_lunas"`
	Nilai          int64   `json:"nilai"`
}

// rekapAll merangkum tagihan satu periode per mitra (dasar penerbitan FAB).
func rekapAll(db *sql.DB, periode string) []RekapRow {
	rows, err := db.Query(`SELECT m.id, m.nama, IFNULL(m.share_mitra,0),
		(SELECT COUNT(*) FROM pelanggan p WHERE p.mitra_id=m.id AND p.status='Aktif'),
		IFNULL(t.total,0), IFNULL(t.lunas,0)
		FROM mitra m
		LEFT JOIN (
			SELECT p2.mitra_id AS mid, SUM(g.jumlah) AS total,
				SUM(CASE WHEN g.status='Lunas' THEN g.jumlah ELSE 0 END) AS lunas
			FROM tagihan g JOIN pelanggan p2 ON p2.id=g.pelanggan_id
			WHERE g.periode=? GROUP BY p2.mitra_id
		) t ON t.mid=m.id
		ORDER BY m.nama`, periode)
	if err != nil {
		return []RekapRow{}
	}
	defer rows.Close()
	out := []RekapRow{}
	for rows.Next() {
		var r RekapRow
		if err := rows.Scan(&r.MitraID, &r.MitraNama, &r.SharePct, &r.PelangganAktif, &r.TotalTagihan, &r.TotalLunas); err != nil {
			continue
		}
		r.Nilai = int64(math.Round(float64(r.TotalLunas) * r.SharePct / 100.0))
		out = append(out, r)
	}
	return out
}

func rekapMitraPeriode(db *sql.DB, mitraID int64, periode string) RekapRow {
	for _, r := range rekapAll(db, periode) {
		if r.MitraID == mitraID {
			return r
		}
	}
	return RekapRow{}
}

// computeFAB menghitung rincian Faktur Bisnis untuk satu mitra & periode.
func computeFAB(db *sql.DB, mitraID int64, periode string) FABRincian {
	r := FABRincian{Periode: periode, Rows: []FABRow{}}
	if m, err := getMitraRaw(db, mitraID); err == nil {
		r.SharePct = m.ShareMitra
	}
	_ = db.QueryRow(`SELECT COUNT(*) FROM pelanggan WHERE mitra_id=? AND status='Aktif'`, mitraID).Scan(&r.JmlPelanggan)
	_ = db.QueryRow(`SELECT COUNT(*),
		IFNULL(SUM(CASE WHEN g.status='Lunas' THEN 1 ELSE 0 END),0),
		IFNULL(SUM(g.jumlah),0),
		IFNULL(SUM(CASE WHEN g.status='Lunas' THEN g.jumlah ELSE 0 END),0)
		FROM tagihan g JOIN pelanggan p ON p.id=g.pelanggan_id
		WHERE p.mitra_id=? AND g.periode=?`, mitraID, periode).
		Scan(&r.JmlTagihan, &r.JmlLunasTagihan, &r.TotalTagihan, &r.TotalLunas)
	rows, err := db.Query(`SELECT COALESCE(k.nama,'Paket?'), COUNT(*), g.jumlah, SUM(g.jumlah)
		FROM tagihan g
		JOIN pelanggan p ON p.id=g.pelanggan_id
		LEFT JOIN paket k ON k.id=p.paket_id
		WHERE p.mitra_id=? AND g.periode=?
		GROUP BY k.id, k.nama, g.jumlah ORDER BY g.jumlah`, mitraID, periode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var fr FABRow
			if err := rows.Scan(&fr.Paket, &fr.Jumlah, &fr.Harga, &fr.Subtotal); err == nil {
				r.Rows = append(r.Rows, fr)
			}
		}
	}
	r.Nilai = int64(math.Round(float64(r.TotalLunas) * r.SharePct / 100.0))
	return r
}

func (h *Handlers) GetKeuangan(w http.ResponseWriter, r *http.Request) {
	periode := r.URL.Query().Get("periode")
	if periode == "" {
		periode = nowPeriode()
	}
	rows, err := h.DB.Query(`SELECT g.id, g.pelanggan_id, g.periode, g.jumlah, g.status, g.metode, g.tanggal_bayar,
		COALESCE(p.nama,''), COALESCE(k.nama,''), p.mitra_id, COALESCE(m.nama,'')
		FROM tagihan g
		JOIN pelanggan p ON p.id=g.pelanggan_id
		LEFT JOIN paket k ON k.id=p.paket_id
		LEFT JOIN mitra m ON m.id=p.mitra_id
		WHERE g.periode=? ORDER BY p.nama`, periode)
	if err != nil {
		httpErr(w, 500, "Gagal memuat tagihan")
		return
	}
	defer rows.Close()
	tagihan := []Tagihan{}
	var total, lunas int64
	for rows.Next() {
		var t Tagihan
		if err := rows.Scan(&t.ID, &t.PelangganID, &t.Periode, &t.Jumlah, &t.Status, &t.Metode, &t.TanggalBayar, &t.PelangganNama, &t.PaketNama, &t.MitraID, &t.MitraNama); err != nil {
			continue
		}
		total += t.Jumlah
		if t.Status == "Lunas" {
			lunas += t.Jumlah
		}
		tagihan = append(tagihan, t)
	}
	pct := 0.0
	if total > 0 {
		pct = math.Round(float64(lunas)/float64(total)*1000) / 10
	}
	writeJSON(w, 200, map[string]any{
		"periode": periode,
		"tagihan": tagihan,
		"summary": map[string]any{
			"total": total, "lunas": lunas, "belum": total - lunas, "pct": pct, "jumlah_tagihan": len(tagihan),
		},
		"rekap": rekapAll(h.DB, periode),
	})
}

func (h *Handlers) GenerateTagihan(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Periode string `json:"periode"`
	}
	if err := readJSON(r, &in); err != nil || in.Periode == "" {
		in.Periode = nowPeriode()
	}
	if _, err := time.Parse("2006-01", in.Periode); err != nil {
		httpErr(w, 400, "Format periode tidak valid (YYYY-MM)")
		return
	}
	rows, err := h.DB.Query(`SELECT p.id, COALESCE(k.harga,0)
		FROM pelanggan p JOIN paket k ON k.id=p.paket_id
		WHERE p.status='Aktif'
		AND NOT EXISTS(SELECT 1 FROM tagihan g WHERE g.pelanggan_id=p.id AND g.periode=?)`, in.Periode)
	if err != nil {
		httpErr(w, 500, "Gagal menyiapkan tagihan")
		return
	}
	type item struct {
		id    int64
		harga int64
	}
	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.harga); err == nil {
			items = append(items, it)
		}
	}
	rows.Close()
	for _, it := range items {
		if _, err := h.DB.Exec(`INSERT INTO tagihan(pelanggan_id,periode,jumlah,status) VALUES(?,?,?,'Belum')`, it.id, in.Periode, it.harga); err != nil {
			httpErr(w, 500, "Gagal membuat tagihan")
			return
		}
	}
	writeJSON(w, 200, map[string]any{"periode": in.Periode, "dibuat": len(items)})
}

func (h *Handlers) UpdateTagihan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var in struct {
		Status       string `json:"status"`
		Metode       string `json:"metode"`
		TanggalBayar string `json:"tanggal_bayar"`
	}
	if err := readJSON(r, &in); err != nil || (in.Status != "Lunas" && in.Status != "Belum") {
		httpErr(w, 400, "Status harus Lunas atau Belum")
		return
	}
	tgl, metode := "", ""
	if in.Status == "Lunas" {
		tgl = in.TanggalBayar
		if tgl == "" {
			_ = h.DB.QueryRow(`SELECT tanggal_bayar FROM tagihan WHERE id=? AND status='Lunas'`, id).Scan(&tgl)
		}
		if tgl == "" {
			tgl = nowISO()
		}
		if !validDate(tgl) || tgl > nowISO() {
			httpErr(w, 400, "Tanggal pembayaran harus valid dan tidak melebihi hari ini")
			return
		}
		if in.Metode != "" && in.Metode != "Transfer" && in.Metode != "Tunai" {
			httpErr(w, 400, "Metode pembayaran harus Transfer atau Tunai")
			return
		}
		if in.Metode == "" {
			_ = h.DB.QueryRow(`SELECT metode FROM tagihan WHERE id=? AND status='Lunas'`, id).Scan(&metode)
			if metode == "" {
				metode = "Transfer"
			}
		} else {
			metode = in.Metode
		}
	}
	res, err := h.DB.Exec(`UPDATE tagihan SET status=?, metode=?, tanggal_bayar=? WHERE id=?`, in.Status, metode, tgl, id)
	if err != nil {
		httpErr(w, 500, "Gagal memperbarui tagihan")
		return
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		httpErr(w, 404, "Tagihan tidak ditemukan")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "status": in.Status, "metode": metode, "tanggal_bayar": tgl})
}

func (h *Handlers) DeleteTagihan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var posted int
	h.DB.QueryRow(`SELECT COUNT(*) FROM ledger WHERE source=? OR source=?`, fmt.Sprintf("bill:%d:invoice", id), fmt.Sprintf("bill:%d:payment", id)).Scan(&posted)
	if posted > 0 {
		httpErr(w, 409, "Tagihan telah dibukukan dan tidak dapat dihapus. Gunakan jurnal penyesuaian untuk koreksi keuangan.")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM tagihan WHERE id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus tagihan")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
