package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func marshalJSON(v any) ([]byte, error) { return json.Marshal(v) }

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func likeQ(q string) string { return "%" + strings.ToLower(strings.TrimSpace(q)) + "%" }

// ============ MITRA ============

const mitraCols = `m.id, m.kode, m.nama, m.jenis, m.pemilik, m.telepon, m.email, m.alamat, m.wilayah, m.npwp, m.status, m.share_mitra, m.tanggal_gabung, m.catatan, m.rekening`

func scanMitra(scan func(...any) error, withCounts bool) (Mitra, error) {
	var m Mitra
	var err error
	if withCounts {
		err = scan(&m.ID, &m.Kode, &m.Nama, &m.Jenis, &m.Pemilik, &m.Telepon, &m.Email, &m.Alamat, &m.Wilayah, &m.NPWP, &m.Status, &m.ShareMitra, &m.TanggalGabung, &m.Catatan, &m.Rekening, &m.PelangganAktif, &m.TotalPelanggan)
	} else {
		err = scan(&m.ID, &m.Kode, &m.Nama, &m.Jenis, &m.Pemilik, &m.Telepon, &m.Email, &m.Alamat, &m.Wilayah, &m.NPWP, &m.Status, &m.ShareMitra, &m.TanggalGabung, &m.Catatan, &m.Rekening)
	}
	return m, err
}

func getMitraRaw(db *sql.DB, id int64) (Mitra, error) {
	row := db.QueryRow(`SELECT `+mitraCols+` FROM mitra m WHERE m.id=?`, id)
	return scanMitra(row.Scan, false)
}

func (h *Handlers) ListMitra(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	status := r.URL.Query().Get("status")
	where := "1=1"
	args := []any{}
	if q != "" {
		where += ` AND (LOWER(m.nama) LIKE ? OR LOWER(m.kode) LIKE ? OR LOWER(m.wilayah) LIKE ? OR LOWER(m.pemilik) LIKE ?)`
		lq := likeQ(q)
		args = append(args, lq, lq, lq, lq)
	}
	if status != "" {
		where += ` AND m.status=?`
		args = append(args, status)
	}
	rows, err := h.DB.Query(`SELECT `+mitraCols+`,
		(SELECT COUNT(*) FROM pelanggan p WHERE p.mitra_id=m.id AND p.status='Aktif'),
		(SELECT COUNT(*) FROM pelanggan p WHERE p.mitra_id=m.id)
		FROM mitra m WHERE `+where+` ORDER BY m.nama`, args...)
	if err != nil {
		httpErr(w, 500, "Gagal memuat mitra")
		return
	}
	defer rows.Close()
	out := []Mitra{}
	for rows.Next() {
		m, err := scanMitra(rows.Scan, true)
		if err != nil {
			httpErr(w, 500, "Gagal membaca data mitra")
			return
		}
		out = append(out, m)
	}
	writeJSON(w, 200, out)
}

func (h *Handlers) CreateMitra(w http.ResponseWriter, r *http.Request) {
	var m Mitra
	if err := readJSON(r, &m); err != nil || strings.TrimSpace(m.Nama) == "" {
		httpErr(w, 400, "Nama mitra wajib diisi")
		return
	}
	if m.ShareMitra < 0 || m.ShareMitra > 100 {
		httpErr(w, 400, "Bagi hasil harus 0–100%")
		return
	}
	if m.Status == "" {
		m.Status = "Aktif"
	}
	if m.Jenis == "" {
		m.Jenis = "Reseller"
	}
	res, err := h.DB.Exec(`INSERT INTO mitra(kode,nama,jenis,pemilik,telepon,email,alamat,wilayah,npwp,status,share_mitra,tanggal_gabung,catatan,rekening)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		m.Kode, m.Nama, m.Jenis, m.Pemilik, m.Telepon, m.Email, m.Alamat, m.Wilayah, m.NPWP, m.Status, m.ShareMitra, m.TanggalGabung, m.Catatan, m.Rekening)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan mitra")
		return
	}
	id, _ := res.LastInsertId()
	if strings.TrimSpace(m.Kode) == "" {
		if _, err := h.DB.Exec(`UPDATE mitra SET kode=? WHERE id=?`, fmt.Sprintf("MTR-%03d", id), id); err == nil {
			m.Kode = fmt.Sprintf("MTR-%03d", id)
		}
	}
	m.ID = id
	writeJSON(w, 201, m)
}

func (h *Handlers) UpdateMitra(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var m Mitra
	if err := readJSON(r, &m); err != nil || strings.TrimSpace(m.Nama) == "" {
		httpErr(w, 400, "Nama mitra wajib diisi")
		return
	}
	if m.ShareMitra < 0 || m.ShareMitra > 100 {
		httpErr(w, 400, "Bagi hasil harus 0–100%")
		return
	}
	res, err := h.DB.Exec(`UPDATE mitra SET kode=?,nama=?,jenis=?,pemilik=?,telepon=?,email=?,alamat=?,wilayah=?,npwp=?,status=?,share_mitra=?,tanggal_gabung=?,catatan=?,rekening=? WHERE id=?`,
		m.Kode, m.Nama, m.Jenis, m.Pemilik, m.Telepon, m.Email, m.Alamat, m.Wilayah, m.NPWP, m.Status, m.ShareMitra, m.TanggalGabung, m.Catatan, m.Rekening, id)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan mitra")
		return
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		httpErr(w, 404, "Mitra tidak ditemukan")
		return
	}
	m.ID = id
	writeJSON(w, 200, m)
}

func (h *Handlers) DeleteMitra(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var n int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM pelanggan WHERE mitra_id=?`, id).Scan(&n); err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	if n > 0 {
		httpErr(w, 409, fmt.Sprintf("Mitra masih memiliki %d pelanggan. Pindahkan atau hapus pelanggannya terlebih dahulu.", n))
		return
	}
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM dokumen WHERE mitra_id=?`, id).Scan(&n); err != nil {
		httpErr(w, 500, "Gagal memeriksa arsip mitra")
		return
	}
	if n > 0 {
		httpErr(w, 409, "Mitra memiliki arsip PKS/FAB. Ubah status menjadi Nonaktif untuk mempertahankan dokumen.")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM mitra WHERE id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus mitra")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (h *Handlers) GetMitraDetail(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	m, err := getMitraRaw(h.DB, id)
	if err == sql.ErrNoRows {
		httpErr(w, 404, "Mitra tidak ditemukan")
		return
	}
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}

	pel, err := h.DB.Query(`SELECT p.id, p.nama, p.status, p.pppoe, COALESCE(k.nama,''), COALESCE(k.harga,0), COALESCE(k.kecepatan,0)
		FROM pelanggan p LEFT JOIN paket k ON k.id=p.paket_id WHERE p.mitra_id=? ORDER BY p.nama`, id)
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	type plg struct {
		ID     int64  `json:"id"`
		Nama   string `json:"nama"`
		Status string `json:"status"`
		PPPoE  string `json:"pppoe"`
		Paket  string `json:"paket"`
		Harga  int64  `json:"harga"`
		Cepat  int    `json:"kecepatan"`
	}
	pelanggan := []plg{}
	for pel.Next() {
		var p plg
		_ = pel.Scan(&p.ID, &p.Nama, &p.Status, &p.PPPoE, &p.Paket, &p.Harga, &p.Cepat)
		pelanggan = append(pelanggan, p)
	}
	pel.Close()

	doks, err := listDokumenByMitra(h.DB, id, 10)
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	writeJSON(w, 200, map[string]any{
		"mitra":     m,
		"pelanggan": pelanggan,
		"dokumen":   doks,
		"rekap":     rekapMitraPeriode(h.DB, id, nowPeriode()),
	})
}

// ============ PAKET ============

func (h *Handlers) ListPaket(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT k.id, k.nama, k.kecepatan, k.harga, k.keterangan,
		(SELECT COUNT(*) FROM pelanggan p WHERE p.paket_id=k.id) FROM paket k ORDER BY k.harga`)
	if err != nil {
		httpErr(w, 500, "Gagal memuat paket")
		return
	}
	defer rows.Close()
	out := []Paket{}
	for rows.Next() {
		var p Paket
		_ = rows.Scan(&p.ID, &p.Nama, &p.Kecepatan, &p.Harga, &p.Keterangan, &p.Dipakai)
		out = append(out, p)
	}
	writeJSON(w, 200, out)
}

func (h *Handlers) CreatePaket(w http.ResponseWriter, r *http.Request) {
	var p Paket
	if err := readJSON(r, &p); err != nil || strings.TrimSpace(p.Nama) == "" {
		httpErr(w, 400, "Nama paket wajib diisi")
		return
	}
	if p.Harga < 0 || p.Harga > maxMoney || p.Kecepatan < 0 || p.Kecepatan > 1000000 {
		httpErr(w, 400, "Harga (0–Rp 1 triliun) dan kecepatan (0–1.000.000) tidak valid")
		return
	}
	res, err := h.DB.Exec(`INSERT INTO paket(nama,kecepatan,harga,keterangan) VALUES(?,?,?,?)`, p.Nama, p.Kecepatan, p.Harga, p.Keterangan)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan paket")
		return
	}
	p.ID, _ = res.LastInsertId()
	writeJSON(w, 201, p)
}

func (h *Handlers) UpdatePaket(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var p Paket
	if err := readJSON(r, &p); err != nil || strings.TrimSpace(p.Nama) == "" {
		httpErr(w, 400, "Nama paket wajib diisi")
		return
	}
	if p.Harga < 0 || p.Harga > maxMoney || p.Kecepatan < 0 || p.Kecepatan > 1000000 {
		httpErr(w, 400, "Harga (0–Rp 1 triliun) dan kecepatan (0–1.000.000) tidak valid")
		return
	}
	res, err := h.DB.Exec(`UPDATE paket SET nama=?,kecepatan=?,harga=?,keterangan=? WHERE id=?`, p.Nama, p.Kecepatan, p.Harga, p.Keterangan, id)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan paket")
		return
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		httpErr(w, 404, "Paket tidak ditemukan")
		return
	}
	p.ID = id
	writeJSON(w, 200, p)
}

func (h *Handlers) DeletePaket(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var n int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM pelanggan WHERE paket_id=?`, id).Scan(&n); err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	if n > 0 {
		httpErr(w, 409, fmt.Sprintf("Paket masih dipakai %d pelanggan. Pindahkan pelanggan ke paket lain dulu.", n))
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM paket WHERE id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus paket")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ============ PELANGGAN ============

func (h *Handlers) ListPelanggan(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	mitraID := r.URL.Query().Get("mitra")
	status := r.URL.Query().Get("status")
	where := "1=1"
	args := []any{}
	if q != "" {
		where += ` AND (LOWER(p.nama) LIKE ? OR LOWER(p.alamat) LIKE ? OR LOWER(p.pppoe) LIKE ?)`
		lq := likeQ(q)
		args = append(args, lq, lq, lq)
	}
	if mitraID != "" {
		where += ` AND p.mitra_id=?`
		args = append(args, mitraID)
	}
	if status != "" {
		where += ` AND p.status=?`
		args = append(args, status)
	}
	rows, err := h.DB.Query(`SELECT p.id, p.nama, p.alamat, p.mitra_id, p.paket_id, p.pppoe, p.status, p.tanggal_pasang, p.catatan, p.telepon,
		COALESCE(m.nama,''), COALESCE(k.nama,''), COALESCE(k.kecepatan,0), COALESCE(k.harga,0)
		FROM pelanggan p
		LEFT JOIN mitra m ON m.id=p.mitra_id
		LEFT JOIN paket k ON k.id=p.paket_id
		WHERE `+where+` ORDER BY p.nama`, args...)
	if err != nil {
		httpErr(w, 500, "Gagal memuat pelanggan")
		return
	}
	defer rows.Close()
	out := []Pelanggan{}
	for rows.Next() {
		var p Pelanggan
		if err := rows.Scan(&p.ID, &p.Nama, &p.Alamat, &p.MitraID, &p.PaketID, &p.PPPoE, &p.Status, &p.TanggalPasang, &p.Catatan, &p.Telepon, &p.MitraNama, &p.PaketNama, &p.Kecepatan, &p.Harga); err != nil {
			httpErr(w, 500, "Gagal membaca data pelanggan")
			return
		}
		out = append(out, p)
	}
	writeJSON(w, 200, out)
}

func (h *Handlers) CreatePelanggan(w http.ResponseWriter, r *http.Request) {
	var p Pelanggan
	if err := readJSON(r, &p); err != nil || strings.TrimSpace(p.Nama) == "" || p.MitraID == 0 || p.PaketID == 0 {
		httpErr(w, 400, "Nama, mitra, dan paket wajib diisi")
		return
	}
	phone, phoneErr := normalizePhone(p.Telepon)
	if phoneErr != nil {
		httpErr(w, 400, phoneErr.Error())
		return
	}
	p.Telepon = phone
	if p.Status == "" {
		p.Status = "Aktif"
	}
	if p.TanggalPasang == "" {
		p.TanggalPasang = nowISO()
	}
	res, err := h.DB.Exec(`INSERT INTO pelanggan(nama,alamat,mitra_id,paket_id,pppoe,status,tanggal_pasang,catatan,telepon) VALUES(?,?,?,?,?,?,?,?,?)`,
		p.Nama, p.Alamat, p.MitraID, p.PaketID, p.PPPoE, p.Status, p.TanggalPasang, p.Catatan, p.Telepon)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan pelanggan")
		return
	}
	p.ID, _ = res.LastInsertId()
	writeJSON(w, 201, p)
}

func (h *Handlers) UpdatePelanggan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var p Pelanggan
	if err := readJSON(r, &p); err != nil || strings.TrimSpace(p.Nama) == "" || p.MitraID == 0 || p.PaketID == 0 {
		httpErr(w, 400, "Nama, mitra, dan paket wajib diisi")
		return
	}
	phone, phoneErr := normalizePhone(p.Telepon)
	if phoneErr != nil {
		httpErr(w, 400, phoneErr.Error())
		return
	}
	p.Telepon = phone
	res, err := h.DB.Exec(`UPDATE pelanggan SET nama=?,alamat=?,mitra_id=?,paket_id=?,pppoe=?,status=?,tanggal_pasang=?,catatan=?,telepon=? WHERE id=?`,
		p.Nama, p.Alamat, p.MitraID, p.PaketID, p.PPPoE, p.Status, p.TanggalPasang, p.Catatan, p.Telepon, id)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan pelanggan")
		return
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		httpErr(w, 404, "Pelanggan tidak ditemukan")
		return
	}
	p.ID = id
	writeJSON(w, 200, p)
}

func (h *Handlers) DeletePelanggan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var posted int
	h.DB.QueryRow(`SELECT COUNT(*) FROM ledger WHERE source IN (SELECT 'bill:'||id||':invoice' FROM tagihan WHERE pelanggan_id=?) OR source IN (SELECT 'bill:'||id||':payment' FROM tagihan WHERE pelanggan_id=?)`, id, id).Scan(&posted)
	if posted > 0 {
		httpErr(w, 409, "Pelanggan memiliki riwayat pembukuan. Ubah status menjadi nonaktif untuk mempertahankan riwayat.")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM tagihan WHERE pelanggan_id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus tagihan pelanggan")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM tiket WHERE pelanggan_id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus tiket pelanggan")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM pelanggan WHERE id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus pelanggan")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ============ TIKET ============

func (h *Handlers) ListTiket(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	status := r.URL.Query().Get("status")
	where := "1=1"
	args := []any{}
	if q != "" {
		where += ` AND (LOWER(t.nomor) LIKE ? OR LOWER(p.nama) LIKE ?)`
		lq := likeQ(q)
		args = append(args, lq, lq)
	}
	if status != "" {
		where += ` AND t.status=?`
		args = append(args, status)
	}
	rows, err := h.DB.Query(`SELECT t.id, t.nomor, t.pelanggan_id, t.jenis, t.prioritas, t.status, t.teknisi, t.deskripsi, t.dibuat, t.selesai,
		COALESCE(p.nama,''), COALESCE(m.nama,'')
		FROM tiket t
		LEFT JOIN pelanggan p ON p.id=t.pelanggan_id
		LEFT JOIN mitra m ON m.id=p.mitra_id
		WHERE `+where+` ORDER BY t.id DESC`, args...)
	if err != nil {
		httpErr(w, 500, "Gagal memuat tiket")
		return
	}
	defer rows.Close()
	out := []Tiket{}
	for rows.Next() {
		var t Tiket
		if err := rows.Scan(&t.ID, &t.Nomor, &t.PelangganID, &t.Jenis, &t.Prioritas, &t.Status, &t.Teknisi, &t.Deskripsi, &t.Dibuat, &t.Selesai, &t.PelangganNama, &t.MitraNama); err != nil {
			httpErr(w, 500, "Gagal membaca data tiket")
			return
		}
		out = append(out, t)
	}
	writeJSON(w, 200, out)
}

func (h *Handlers) CreateTiket(w http.ResponseWriter, r *http.Request) {
	var t Tiket
	if err := readJSON(r, &t); err != nil || t.PelangganID == 0 {
		httpErr(w, 400, "Pelanggan wajib dipilih")
		return
	}
	if t.Status == "" {
		t.Status = "Baru"
	}
	if t.Prioritas == "" {
		t.Prioritas = "Sedang"
	}
	if t.Jenis == "" {
		t.Jenis = "Lainnya"
	}
	nomor, err := nextNumber(h.DB, "TKT", time.Now())
	if err != nil {
		httpErr(w, 500, "Gagal membuat nomor tiket")
		return
	}
	stamp := nowISO() + " " + nowClock()
	if t.Status == "Selesai" {
		t.Selesai = stamp
	}
	res, err := h.DB.Exec(`INSERT INTO tiket(nomor,pelanggan_id,jenis,prioritas,status,teknisi,deskripsi,dibuat,selesai) VALUES(?,?,?,?,?,?,?,?,?)`,
		nomor, t.PelangganID, t.Jenis, t.Prioritas, t.Status, t.Teknisi, t.Deskripsi, stamp, t.Selesai)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan tiket")
		return
	}
	t.ID, _ = res.LastInsertId()
	t.Nomor = nomor
	t.Dibuat = stamp
	writeJSON(w, 201, t)
}

func (h *Handlers) UpdateTiket(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var t Tiket
	if err := readJSON(r, &t); err != nil || t.PelangganID == 0 {
		httpErr(w, 400, "Pelanggan wajib dipilih")
		return
	}
	var oldStatus string
	if err := h.DB.QueryRow(`SELECT status FROM tiket WHERE id=?`, id).Scan(&oldStatus); err == sql.ErrNoRows {
		httpErr(w, 404, "Tiket tidak ditemukan")
		return
	}
	now := nowISO() + " " + nowClock()
	switch {
	case t.Status == "Selesai" && oldStatus != "Selesai":
		t.Selesai = now
	case t.Status != "Selesai":
		t.Selesai = ""
	}
	res, err := h.DB.Exec(`UPDATE tiket SET pelanggan_id=?,jenis=?,prioritas=?,status=?,teknisi=?,deskripsi=?,selesai=? WHERE id=?`,
		t.PelangganID, t.Jenis, t.Prioritas, t.Status, t.Teknisi, t.Deskripsi, t.Selesai, id)
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan tiket")
		return
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		httpErr(w, 404, "Tiket tidak ditemukan")
		return
	}
	t.ID = id
	writeJSON(w, 200, t)
}

func (h *Handlers) DeleteTiket(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM tiket WHERE id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus tiket")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
