package main

import (
	"database/sql"
	"net/http"
	"time"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	periode := nowPeriode()

	var mitraAktif, plgAktif, tiketTerbuka int
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM mitra WHERE status='Aktif'`).Scan(&mitraAktif)
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM pelanggan WHERE status='Aktif'`).Scan(&plgAktif)
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM tiket WHERE status!='Selesai'`).Scan(&tiketTerbuka)

	var total, lunas int64
	_ = h.DB.QueryRow(`SELECT IFNULL(SUM(jumlah),0), IFNULL(SUM(CASE WHEN status='Lunas' THEN jumlah ELSE 0 END),0)
		FROM tagihan WHERE periode=?`, periode).Scan(&total, &lunas)

	// Pendapatan 6 bulan terakhir (tagihan lunas per periode).
	start := time.Now().AddDate(0, -5, 0).Format("2006-01")
	rows, err := h.DB.Query(`SELECT periode, SUM(jumlah) FROM tagihan WHERE status='Lunas' AND periode>=? GROUP BY periode`, start)
	byPeriode := map[string]int64{}
	if err == nil {
		for rows.Next() {
			var p string
			var v int64
			if rows.Scan(&p, &v) == nil {
				byPeriode[p] = v
			}
		}
		rows.Close()
	}
	type monthVal struct {
		Label string `json:"label"`
		Value int64  `json:"value"`
	}
	months := []monthVal{}
	for i := 5; i >= 0; i-- {
		t := time.Now().AddDate(0, -i, 0)
		months = append(months, monthVal{Label: bulan3(t.Month()), Value: byPeriode[t.Format("2006-01")]})
	}

	tiketRows, err := h.DB.Query(`SELECT t.id, t.nomor, t.pelanggan_id, t.jenis, t.prioritas, t.status, t.teknisi, t.deskripsi, t.dibuat, t.selesai,
		COALESCE(p.nama,''), COALESCE(m.nama,'')
		FROM tiket t
		LEFT JOIN pelanggan p ON p.id=t.pelanggan_id
		LEFT JOIN mitra m ON m.id=p.mitra_id
		WHERE t.status!='Selesai' ORDER BY t.id DESC LIMIT 5`)
	tiketBaru := []Tiket{}
	if err == nil {
		for tiketRows.Next() {
			if t, e := scanTiketRow(tiketRows.Scan); e == nil {
				tiketBaru = append(tiketBaru, t)
			}
		}
		tiketRows.Close()
	}

	doks, _ := listDokumenRecent(h.DB, 5)

	writeJSON(w, 200, map[string]any{
		"mitra_aktif":      mitraAktif,
		"pelanggan_aktif":  plgAktif,
		"tiket_terbuka":    tiketTerbuka,
		"pendapatan_bulan": lunas,
		"tagihan_bulan":    total,
		"revenue_months":   months,
		"tiket_terbaru":    tiketBaru,
		"dokumen_terbaru":  doks,
		"rekap_mitra":      rekapAll(h.DB, periode),
	})
}

func bulan3(m time.Month) string {
	short := map[time.Month]string{
		time.January: "Jan", time.February: "Feb", time.March: "Mar",
		time.April: "Apr", time.May: "Mei", time.June: "Jun",
		time.July: "Jul", time.August: "Agu", time.September: "Sep",
		time.October: "Okt", time.November: "Nov", time.December: "Des",
	}
	return short[m]
}

func scanTiketRow(scan func(...any) error) (Tiket, error) {
	var t Tiket
	err := scan(&t.ID, &t.Nomor, &t.PelangganID, &t.Jenis, &t.Prioritas, &t.Status, &t.Teknisi, &t.Deskripsi, &t.Dibuat, &t.Selesai, &t.PelangganNama, &t.MitraNama)
	return t, err
}

func listDokumenRecent(db *sql.DB, limit int) ([]Dokumen, error) {
	rows, err := db.Query(dokSelect+` ORDER BY d.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Dokumen{}
	for rows.Next() {
		if d, err := scanDokumen(rows.Scan); err == nil {
			out = append(out, d)
		}
	}
	return out, nil
}
