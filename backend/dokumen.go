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

const dokSelect = `SELECT d.id, d.nomor, d.jenis, d.mitra_id, d.tanggal, d.periode, d.nilai, d.rincian, d.kirim, d.created_at, COALESCE(m.nama,'')
	FROM dokumen d LEFT JOIN mitra m ON m.id=d.mitra_id`

func scanDokumen(scan func(...any) error) (Dokumen, error) {
	var (
		d                  Dokumen
		rincian, kirimJSON string
	)
	if err := scan(&d.ID, &d.Nomor, &d.Jenis, &d.MitraID, &d.Tanggal, &d.Periode, &d.Nilai, &rincian, &kirimJSON, &d.CreatedAt, &d.MitraNama); err != nil {
		return d, err
	}
	d.Kirim = []KirimLog{}
	_ = json.Unmarshal([]byte(kirimJSON), &d.Kirim)
	if d.Jenis == "FAB" && rincian != "" {
		var rc FABRincian
		if json.Unmarshal([]byte(rincian), &rc) == nil {
			d.Rincian = &rc
		}
	}
	return d, nil
}

func (h *Handlers) ListDokumen(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	jenis := r.URL.Query().Get("jenis")
	mitra := r.URL.Query().Get("mitra")
	where := "1=1"
	args := []any{}
	if q != "" {
		where += ` AND (LOWER(d.nomor) LIKE ? OR LOWER(m.nama) LIKE ?)`
		lq := likeQ(q)
		args = append(args, lq, lq)
	}
	if jenis != "" {
		where += ` AND d.jenis=?`
		args = append(args, jenis)
	}
	if mitra != "" {
		where += ` AND d.mitra_id=?`
		args = append(args, mitra)
	}
	rows, err := h.DB.Query(dokSelect+` WHERE `+where+` ORDER BY d.id DESC`, args...)
	if err != nil {
		httpErr(w, 500, "Gagal memuat dokumen")
		return
	}
	defer rows.Close()
	out := []Dokumen{}
	for rows.Next() {
		if d, err := scanDokumen(rows.Scan); err == nil {
			out = append(out, d)
		}
	}
	writeJSON(w, 200, out)
}

func listDokumenByMitra(db *sql.DB, mitraID int64, limit int) ([]Dokumen, error) {
	rows, err := db.Query(dokSelect+` WHERE d.mitra_id=? ORDER BY d.id DESC LIMIT ?`, mitraID, limit)
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

func (h *Handlers) CreateDokumen(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Jenis     string `json:"jenis"`
		MitraID   int64  `json:"mitra_id"`
		Tanggal   string `json:"tanggal"`
		Periode   string `json:"periode"`
		MasaBulan int    `json:"masa_bulan"`
	}
	if err := readJSON(r, &in); err != nil || (in.Jenis != "PKS" && in.Jenis != "FAB") {
		httpErr(w, 400, "Jenis dokumen harus PKS atau FAB")
		return
	}
	m, err := getMitraRaw(h.DB, in.MitraID)
	if err == sql.ErrNoRows {
		httpErr(w, 400, "Mitra tidak ditemukan")
		return
	}
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	s := getSettings(h.DB)
	if in.Tanggal == "" {
		in.Tanggal = nowISO()
	}
	if _, err := time.Parse("2006-01-02", in.Tanggal); err != nil {
		httpErr(w, 400, "Tanggal dokumen tidak valid (YYYY-MM-DD)")
		return
	}
	if in.Jenis == "PKS" && (in.MasaBulan < 0 || in.MasaBulan > 120) {
		httpErr(w, 400, "Masa kerja sama harus antara 1 dan 120 bulan")
		return
	}
	masa := in.MasaBulan
	if masa <= 0 {
		masa = s.MasaPKSBulan
		if masa <= 0 {
			masa = 12
		}
	}
	var (
		rinc     *FABRincian
		nilai    int64
		rincJSON = ""
		periode  = ""
	)
	if in.Jenis == "FAB" {
		if in.Periode == "" {
			in.Periode = nowPeriode()
		}
		if _, err := time.Parse("2006-01", in.Periode); err != nil {
			httpErr(w, 400, "Periode tagihan tidak valid (YYYY-MM)")
			return
		}
		var duplicate int
		h.DB.QueryRow(`SELECT COUNT(*) FROM dokumen WHERE jenis='FAB' AND mitra_id=? AND periode=?`, in.MitraID, in.Periode).Scan(&duplicate)
		if duplicate > 0 {
			httpErr(w, 409, "FAB mitra untuk periode ini sudah ada. Periksa dokumen sebelumnya.")
			return
		}
		rc := computeFAB(h.DB, in.MitraID, in.Periode)
		rinc = &rc
		nilai = rc.Nilai
		periode = in.Periode
		if b, err := marshalJSON(rc); err == nil {
			rincJSON = string(b)
		}
	}
	docDate, _ := time.Parse("2006-01-02", in.Tanggal)
	nomor, err := nextNumber(h.DB, in.Jenis, docDate)
	if err != nil {
		httpErr(w, 500, "Gagal membuat nomor dokumen")
		return
	}
	snapJSON, _ := marshalJSON(DocSnapshot{Mitra: m, MasaBulan: masa, Company: &s})
	res, err := h.DB.Exec(`INSERT INTO dokumen(nomor,jenis,mitra_id,tanggal,periode,nilai,rincian,snapshot,kirim,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`,
		nomor, in.Jenis, in.MitraID, in.Tanggal, periode, nilai, rincJSON, string(snapJSON), "[]", nowISO()+" "+nowClock())
	if err != nil {
		httpErr(w, 500, "Gagal menyimpan dokumen")
		return
	}
	id, _ := res.LastInsertId()
	d := Dokumen{
		ID: id, Nomor: nomor, Jenis: in.Jenis, MitraID: in.MitraID, Tanggal: in.Tanggal,
		Periode: periode, Nilai: nilai, Rincian: rinc, Kirim: []KirimLog{},
		MitraNama: m.Nama, CreatedAt: nowISO(),
	}
	writeJSON(w, 201, d)
}

func (h *Handlers) DeleteDokumen(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM dokumen WHERE id=?`, id); err != nil {
		httpErr(w, 500, "Gagal menghapus dokumen")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DokumenHTML menampilkan dokumen sebagai HTML siap cetak.
// Token diterima lewat query string karena iframe pratinjau tidak bisa
// mengirim header Authorization.
func (h *Handlers) DokumenHTML(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 404, "Dokumen tidak ditemukan")
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		token = bearerToken(r)
	}
	if !checkToken(h.DB, token) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`<html><body style="font-family:sans-serif;padding:40px;text-align:center"><h2>Sesi berakhir</h2><p>Silakan login ulang lalu buka kembali dokumen ini.</p></body></html>`))
		return
	}
	var nomor, jenis, tanggal, periode, rincian, snapshot string
	var mitraID int64
	err = h.DB.QueryRow(`SELECT nomor, jenis, mitra_id, tanggal, periode, rincian, snapshot FROM dokumen WHERE id=?`, id).
		Scan(&nomor, &jenis, &mitraID, &tanggal, &periode, &rincian, &snapshot)
	if err == sql.ErrNoRows {
		httpErr(w, 404, "Dokumen tidak ditemukan")
		return
	}
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	var m Mitra
	var masa int
	if snapshot != "" && json.Unmarshal([]byte(snapshot), &struct {
		Mitra     *Mitra `json:"mitra"`
		MasaBulan *int   `json:"masa_bulan"`
	}{&m, &masa}) == nil {
		// snapshot terbaca
	} else if cur, err := getMitraRaw(h.DB, mitraID); err == nil {
		m = cur
	}
	s := getSettings(h.DB)
	var stored DocSnapshot
	if json.Unmarshal([]byte(snapshot), &stored) == nil && stored.Company != nil {
		s = *stored.Company
	}
	if masa <= 0 {
		masa = s.MasaPKSBulan
	}
	var rinc *FABRincian
	if jenis == "FAB" && rincian != "" {
		var rc FABRincian
		if json.Unmarshal([]byte(rincian), &rc) == nil {
			rinc = &rc
		}
	}
	preview := r.URL.Query().Get("mode") == "preview"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	_, _ = w.Write([]byte(buildDocHTML(s, m, nomor, jenis, tanggal, rinc, masa, preview)))
}

func (h *Handlers) DokumenKirim(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		httpErr(w, 400, "ID tidak valid")
		return
	}
	var in struct {
		Via string `json:"via"`
	}
	if err := readJSON(r, &in); err != nil || (in.Via != "wa" && in.Via != "email") {
		httpErr(w, 400, "Via harus wa atau email")
		return
	}
	var kirimJSON string
	err = h.DB.QueryRow(`SELECT kirim FROM dokumen WHERE id=?`, id).Scan(&kirimJSON)
	if err == sql.ErrNoRows {
		httpErr(w, 404, "Dokumen tidak ditemukan")
		return
	}
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	logs := []KirimLog{}
	_ = json.Unmarshal([]byte(kirimJSON), &logs)
	logs = append(logs, KirimLog{Via: in.Via, Tanggal: nowISO() + " " + nowClock()})
	b, _ := marshalJSON(logs)
	if _, err := h.DB.Exec(`UPDATE dokumen SET kirim=? WHERE id=?`, string(b), id); err != nil {
		httpErr(w, 500, "Gagal mencatat pengiriman")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "kirim": logs})
}

// ============ Generator HTML dokumen ============

func docReplacements(s Settings, m Mitra, nomor, tanggal string, rinc *FABRincian, masa int) []string {
	shareM := m.ShareMitra
	if rinc != nil {
		shareM = rinc.SharePct
	}
	if shareM < 0 {
		shareM = 0
	}
	if shareM > 100 {
		shareM = 100
	}
	shareI := 100 - shareM
	formatPct := func(value float64) string {
		return strings.ReplaceAll(strconv.FormatFloat(value, 'f', -1, 64), ".", ",")
	}
	percentageWords := func(value float64) string {
		parts := strings.Split(strconv.FormatFloat(value, 'f', -1, 64), ".")
		whole, _ := strconv.ParseInt(parts[0], 10, 64)
		words := angkaKata(whole)
		if len(parts) > 1 {
			words += " koma"
			for _, digit := range parts[1] {
				words += " " + angkaKata(int64(digit-'0'))
			}
		}
		return words
	}
	pairs := []string{
		"{{nomor}}", nomor,
		"{{hari}}", hariIndo(tanggal),
		"{{tanggal}}", tanggalIndo(tanggal),
		"{{tanggal_mulai}}", tanggalIndo(tanggal),
		"{{tanggal_akhir}}", tanggalIndo(addMonthsISO(tanggal, masa)),
		"{{masa_bulan}}", strconv.Itoa(masa),
		"{{masa_terbilang}}", angkaKata(int64(masa)),
		"{{nama_isp}}", escHTML(s.NamaISP),
		"{{alamat_isp}}", escHTML(s.Alamat),
		"{{kota_isp}}", escHTML(s.Kota),
		"{{telepon_isp}}", escHTML(s.Telepon),
		"{{email_isp}}", escHTML(s.Email),
		"{{website_isp}}", escHTML(s.Website),
		"{{npwp_isp}}", escHTML(s.NPWP),
		"{{penandatangan}}", escHTML(s.Penandatangan),
		"{{jabatan}}", escHTML(s.Jabatan),
		"{{rekening}}", escHTML(s.Rekening),
		"{{rekening_mitra}}", escHTML(partnerAccount(m)),
		"{{kota}}", escHTML(s.Kota),
		"{{nama_mitra}}", escHTML(m.Nama),
		"{{jenis_mitra}}", escHTML(m.Jenis),
		"{{pemilik_mitra}}", escHTML(m.Pemilik),
		"{{telepon_mitra}}", escHTML(m.Telepon),
		"{{email_mitra}}", escHTML(m.Email),
		"{{alamat_mitra}}", escHTML(m.Alamat),
		"{{wilayah_mitra}}", escHTML(m.Wilayah),
		"{{npwp_mitra}}", escHTML(m.NPWP),
		"{{share_mitra}}", formatPct(shareM),
		"{{share_isp}}", formatPct(shareI),
		"{{share_mitra_terbilang}}", percentageWords(shareM),
		"{{share_isp_terbilang}}", percentageWords(shareI),
	}
	if rinc != nil {
		var sb strings.Builder
		for i, row := range rinc.Rows {
			sb.WriteString(fmt.Sprintf(`<tr><td>%d</td><td>%s</td><td style="text-align:center">%d</td><td>%s</td><td>%s</td></tr>`,
				i+1, escHTML(row.Paket), row.Jumlah, rupiah(row.Harga), rupiah(row.Subtotal)))
		}
		if len(rinc.Rows) == 0 {
			sb.WriteString(`<tr><td colspan="5" style="text-align:center">Tidak ada tagihan pada periode ini.</td></tr>`)
		}
		pairs = append(pairs,
			"{{periode_label}}", periodeLabel(rinc.Periode),
			"{{jml_pelanggan}}", strconv.Itoa(rinc.JmlPelanggan),
			"{{jml_tagihan}}", strconv.Itoa(rinc.JmlTagihan),
			"{{jml_lunas_tagihan}}", strconv.Itoa(rinc.JmlLunasTagihan),
			"{{total_tagihan}}", rupiah(rinc.TotalTagihan),
			"{{total_lunas}}", rupiah(rinc.TotalLunas),
			"{{nilai_bagihasil}}", rupiah(rinc.Nilai),
			"{{nilai_terbilang}}", terbilang(rinc.Nilai),
			"{{rincian_rows}}", sb.String(),
		)
	}
	return pairs
}

func buildDocHTML(s Settings, m Mitra, nomor, jenis, tanggal string, rinc *FABRincian, masa int, preview bool) string {
	repl := strings.NewReplacer(docReplacements(s, m, nomor, tanggal, rinc, masa)...)
	tpl := s.TplPKS
	if jenis == "FAB" {
		tpl = s.TplFAB
	}
	if strings.TrimSpace(s.Kota) != "" && strings.Contains(strings.ToLower(s.Alamat), strings.ToLower(s.Kota)) {
		tpl = strings.ReplaceAll(tpl, "{{alamat_isp}}, {{kota_isp}}", "{{alamat_isp}}")
	}
	if strings.TrimSpace(tpl) == "" {
		tpl = `<p>Template dokumen masih kosong. Isi terlebih dahulu di menu Pengaturan &raquo; Template Dokumen.</p>`
	}
	letterhead := repl.Replace(`<header class="letterhead"><div class="lh-identity"><img class="company-logo" src="data:image/png;base64,` + companyLogoBase64 + `" alt="ALTELMATARA NETWORK"><h1>{{nama_isp}}</h1><div class="lh-sub">LAYANAN INTERNET</div></div><p>{{alamat_isp}}<br>{{telepon_isp}} &middot; {{email_isp}}<br>{{website_isp}}</p></header>`)
	if jenis == "FAB" {
		letterhead += `<div class="doc-label">FAB<span>Rekap bagi hasil mitra</span></div>`
	} else {
		letterhead += `<div class="doc-label">PKS<span>Dokumen kerja sama</span></div>`
	}
	footer := repl.Replace(`<footer class="doc-footer"><span>{{nama_isp}}</span><span>{{nomor}}</span></footer>`)
	missing := []string{}
	if strings.TrimSpace(s.Penandatangan) == "" || strings.TrimSpace(s.Jabatan) == "" {
		missing = append(missing, "penanda tangan perusahaan")
	}
	if strings.TrimSpace(m.Pemilik) == "" {
		missing = append(missing, "penanggung jawab mitra")
	}
	if jenis == "FAB" && strings.TrimSpace(m.Rekening) == "" {
		missing = append(missing, "rekening penerima mitra")
	}
	if len(missing) > 0 {
		letterhead += `<div class="document-note">DRAF — lengkapi ` + escHTML(strings.Join(missing, ", ")) + ` sebelum digunakan sebagai dokumen final.</div>`
	}
	toolbar, script := "", ""
	if !preview {
		toolbar = `<div class="toolbar no-print"><button onclick="window.print()">Cetak / Simpan PDF</button><span>Tekan Ctrl+P jika dialog cetak tidak muncul &middot; pilih &quot;Save as PDF&quot; untuk menyimpan PDF.</span></div>`
		script = `<script>setTimeout(function(){window.print()},600)</script>`
	}
	title := escHTML(nomor)
	docClass := "document-pks"
	if jenis == "FAB" {
		docClass = "document-fab"
	}
	return `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` + title + `</title><style>` + docCSS + `</style></head><body class="` + docClass + `">` + toolbar + `<main class="page">` + letterhead + repl.Replace(tpl) + footer + `</main>` + script + `</body></html>`
}
