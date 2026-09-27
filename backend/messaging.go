package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type WAConfig struct {
	Enabled          bool   `json:"enabled"`
	BaseURL          string `json:"base_url"`
	Sender           string `json:"sender"`
	PublicURL        string `json:"public_url"`
	HasKey           bool   `json:"has_api_key"`
	InvoiceTemplate  string `json:"invoice_template"`
	DocumentTemplate string `json:"document_template"`
}

func defaultWAConfig() WAConfig {
	return WAConfig{
		InvoiceTemplate:  "Halo {{nama}},\n\nBerikut invoice {{nomor}} dari {{perusahaan}}.\nPeriode: {{periode}}\nTotal tagihan: {{total}}\nStatus: {{status}}\nSisa pembayaran: {{sisa}}\n\n{{rekening}}\n\nTerima kasih.",
		DocumentTemplate: "Halo {{nama}},\n\nBerikut dokumen {{jenis}} {{nomor}} dari {{perusahaan}}.\nTanggal: {{tanggal}}\n{{ringkasan}}\n\nSilakan periksa dokumen yang kami kirimkan. Terima kasih.",
	}
}
func migrateMessaging(db *sql.DB) error {
	rows, e := db.Query(`PRAGMA table_info(pelanggan)`)
	if e != nil {
		return e
	}
	found := false
	for rows.Next() {
		var id, nn, pk int
		var name, typ string
		var def any
		if e = rows.Scan(&id, &name, &typ, &nn, &def, &pk); e != nil {
			rows.Close()
			return e
		}
		if name == "telepon" {
			found = true
		}
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	if !found {
		if _, e = db.Exec(`ALTER TABLE pelanggan ADD COLUMN telepon TEXT NOT NULL DEFAULT ''`); e != nil {
			return e
		}
	}
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS wa_config(id INTEGER PRIMARY KEY CHECK(id=1),data TEXT NOT NULL,secret TEXT NOT NULL DEFAULT '')`,
		`CREATE TABLE IF NOT EXISTS wa_outbox(id INTEGER PRIMARY KEY AUTOINCREMENT,request_key TEXT NOT NULL UNIQUE,request_hash TEXT NOT NULL,kind TEXT NOT NULL,source_id INTEGER NOT NULL,reference TEXT NOT NULL,recipient TEXT NOT NULL,message TEXT NOT NULL,mode TEXT NOT NULL,status TEXT NOT NULL,provider_id TEXT NOT NULL DEFAULT '',error TEXT NOT NULL DEFAULT '',created_at TEXT NOT NULL,finished_at TEXT NOT NULL DEFAULT '',share_hash TEXT NOT NULL DEFAULT '')`,
		`CREATE TABLE IF NOT EXISTS wa_shares(hash TEXT PRIMARY KEY,content BLOB NOT NULL,mime TEXT NOT NULL,filename TEXT NOT NULL,expires_at INTEGER NOT NULL,revoked INTEGER NOT NULL DEFAULT 0)`,
		`CREATE INDEX IF NOT EXISTS wa_outbox_source ON wa_outbox(kind,source_id)`,
		`UPDATE wa_outbox SET status='unknown',error='Aplikasi berhenti saat mengirim. Periksa MPWA sebelum mengirim ulang.' WHERE status='sending'`,
		`DELETE FROM wa_shares WHERE expires_at < unixepoch()`,
	} {
		if _, e = db.Exec(q); e != nil {
			return e
		}
	}
	return nil
}
func registerMessaging(m *http.ServeMux, h *Handlers) {
	handle := func(pattern string, fn http.HandlerFunc) { m.HandleFunc(pattern, operationHandler(fn)) }
	handle("GET /api/whatsapp/settings", h.WASettings)
	handle("PUT /api/whatsapp/settings", h.WASaveSettings)
	handle("GET /api/whatsapp/preview", h.WAPreview)
	handle("GET /api/whatsapp/history", h.WAHistory)
	handle("POST /api/whatsapp/attachment", h.WAAttachment)
	handle("POST /api/whatsapp/history/{id}/revoke", h.WARevoke)
	// Reserve idempotency under the write lock; release it before network I/O.
	m.HandleFunc("POST /api/whatsapp/send", h.WASend)
	m.HandleFunc("GET /share/{token}", h.WAShare)
}
func normalizePhone(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || strings.ContainsRune(" +-().", c)) {
			return "", fmt.Errorf("Nomor WhatsApp hanya boleh berisi angka dengan kode negara")
		}
	}
	s = strings.NewReplacer(" ", "", "+", "", "-", "", "(", "", ")", "", ".", "").Replace(s)
	if strings.HasPrefix(s, "0") {
		s = "62" + s[1:]
	}
	if len(s) < 10 || len(s) > 15 || s[0] == '0' {
		return "", fmt.Errorf("Nomor WhatsApp harus 10–15 digit, contoh 628123456789")
	}
	return s, nil
}
func validGatewayURL(raw string, public bool) (string, error) {
	if raw == "" {
		return "", nil
	}
	u, e := url.Parse(strings.TrimSpace(raw))
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("URL tidak valid; jangan sertakan kredensial, query, atau fragmen")
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		local := u.Hostname() == "localhost" || ip != nil && (ip.IsLoopback() || ip.IsPrivate())
		if u.Scheme != "http" || !local || public {
			return "", fmt.Errorf("Gunakan HTTPS; HTTP gateway hanya diizinkan untuk localhost atau IP jaringan privat")
		}
	}
	if public && (u.Hostname() == "localhost" || net.ParseIP(u.Hostname()) != nil && (net.ParseIP(u.Hostname()).IsPrivate() || net.ParseIP(u.Hostname()).IsLoopback())) {
		return "", fmt.Errorf("URL publik harus bisa diakses penerima WhatsApp")
	}
	if public && strings.Trim(u.Path, "/") != "" {
		return "", fmt.Errorf("URL publik cukup domain aplikasi, tanpa path")
	}
	return strings.TrimRight(u.String(), "/"), nil
}
func sealWA(db *sql.DB, value string, decrypt bool) (string, error) {
	key := sha256.Sum256([]byte("mpwa-key-v1:" + GetSecret(db)))
	block, e := aes.NewCipher(key[:])
	if e != nil {
		return "", e
	}
	g, e := cipher.NewGCM(block)
	if e != nil {
		return "", e
	}
	if decrypt {
		data, e := base64.StdEncoding.DecodeString(value)
		if e != nil || len(data) < g.NonceSize() {
			return "", fmt.Errorf("Kunci gateway tidak dapat dibaca")
		}
		v, e := g.Open(nil, data[:g.NonceSize()], data[g.NonceSize():], nil)
		return string(v), e
	}
	nonce := make([]byte, g.NonceSize())
	if _, e = rand.Read(nonce); e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(g.Seal(nonce, nonce, []byte(value), nil)), nil
}
func loadWA(db *sql.DB) (WAConfig, string, error) {
	c := defaultWAConfig()
	var raw, key string
	e := db.QueryRow(`SELECT data,secret FROM wa_config WHERE id=1`).Scan(&raw, &key)
	if e == sql.ErrNoRows {
		return c, "", nil
	}
	if e != nil {
		return c, "", e
	}
	if e = json.Unmarshal([]byte(raw), &c); e != nil {
		return c, "", e
	}
	c.HasKey = key != ""
	return c, key, nil
}
func (h *Handlers) WASettings(w http.ResponseWriter, r *http.Request) {
	c, _, e := loadWA(h.DB)
	if e != nil {
		httpErr(w, 500, "Gagal membaca pengaturan MPWA")
		return
	}
	writeJSON(w, 200, c)
}
func (h *Handlers) WASaveSettings(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	var in struct {
		WAConfig
		APIKey   string `json:"api_key"`
		ClearKey bool   `json:"clear_key"`
	}
	if readJSON(r, &in) != nil {
		httpErr(w, 400, "Isian tidak valid")
		return
	}
	c := in.WAConfig
	var e error
	if c.BaseURL, e = validGatewayURL(c.BaseURL, false); e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	if c.PublicURL, e = validGatewayURL(c.PublicURL, true); e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	if c.Sender, e = normalizePhone(c.Sender); e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	if len(in.APIKey) > 2000 || len(c.InvoiceTemplate) > 6000 || len(c.DocumentTemplate) > 6000 || strings.TrimSpace(c.InvoiceTemplate) == "" || strings.TrimSpace(c.DocumentTemplate) == "" {
		httpErr(w, 400, "Isi template pesan; maksimal 6.000 karakter")
		return
	}
	_, key, e := loadWA(h.DB)
	if e != nil {
		httpErr(w, 500, "Gagal membaca pengaturan")
		return
	}
	if in.ClearKey {
		key = ""
	}
	if strings.TrimSpace(in.APIKey) != "" {
		key, e = sealWA(h.DB, strings.TrimSpace(in.APIKey), false)
		if e != nil {
			httpErr(w, 500, "Gagal menyimpan kunci API")
			return
		}
	}
	c.HasKey = key != ""
	if c.Enabled && (c.BaseURL == "" || c.Sender == "" || !c.HasKey) {
		httpErr(w, 400, "Isi URL MPWA, nomor sender, dan API key sebelum mengaktifkan")
		return
	}
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan pengaturan")
		return
	}
	defer tx.Rollback()
	_, e = tx.Exec(`INSERT INTO wa_config(id,data,secret) VALUES(1,?,?) ON CONFLICT(id) DO UPDATE SET data=excluded.data,secret=excluded.secret`, jsonText(c), key)
	if e == nil {
		e = audit(tx, "Pengaturan MPWA diperbarui", map[string]any{"enabled": c.Enabled, "sender": c.Sender, "base_url": c.BaseURL, "key_changed": in.ClearKey || in.APIKey != ""})
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan MPWA")
		return
	}
	writeJSON(w, 200, c)
}

type WASource struct {
	Revision  string `json:"revision"`
	Kind      string `json:"kind"`
	ID        int64  `json:"source_id"`
	Reference string `json:"reference"`
	Recipient string `json:"recipient"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	HTML      string `json:"-"`
}

func replaceWA(template string, values map[string]string) string {
	// One pass: content from a customer's name cannot introduce new placeholders.
	re := regexp.MustCompile(`\{\{([a-z_]+)\}\}`)
	message := re.ReplaceAllStringFunc(template, func(v string) string {
		if value, ok := values[strings.TrimSuffix(strings.TrimPrefix(v, "{{"), "}}")]; ok {
			return value
		}
		return v
	})
	return strings.TrimSpace(regexp.MustCompile(`\n{3,}`).ReplaceAllString(message, "\n\n"))
}
func waSource(db *sql.DB, c WAConfig, kind string, id int64) (WASource, error) {
	s := getSettings(db)
	out := WASource{Kind: kind, ID: id}
	v := map[string]string{"perusahaan": s.NamaISP}
	if kind == "custom" && id == 0 {
		out.Reference = "Pesan manual"
		return out, nil
	}
	if kind == "invoice" {
		var b Tagihan
		var address, phone, packageName string
		e := db.QueryRow(`SELECT g.id,g.periode,g.jumlah,g.status,g.metode,g.tanggal_bayar,COALESCE(p.nama,''),COALESCE(p.alamat,''),COALESCE(p.telepon,''),COALESCE(k.nama,'Layanan internet'),g.pelanggan_id,COALESCE(m.nama,'') FROM tagihan g LEFT JOIN pelanggan p ON p.id=g.pelanggan_id LEFT JOIN paket k ON k.id=p.paket_id LEFT JOIN mitra m ON m.id=p.mitra_id WHERE g.id=?`, id).Scan(&b.ID, &b.Periode, &b.Jumlah, &b.Status, &b.Metode, &b.TanggalBayar, &b.PelangganNama, &address, &phone, &packageName, &b.PelangganID, &b.MitraNama)
		if e != nil {
			return out, fmt.Errorf("Invoice tidak ditemukan")
		}
		out.Reference = fmt.Sprintf("INV/%s/%06d", strings.ReplaceAll(b.Periode, "-", ""), id)
		out.Name = b.PelangganNama
		out.Recipient = phone
		v["nama"] = out.Name
		v["nomor"] = out.Reference
		v["periode"] = periodeLabel(b.Periode)
		v["total"] = rupiah(b.Jumlah)
		v["status"] = b.Status
		v["sisa"] = rupiah(b.Jumlah)
		v["rekening"] = ""
		if b.Status == "Lunas" {
			v["sisa"] = rupiah(0)
		} else if s.Rekening != "" {
			v["rekening"] = "Pembayaran melalui:\n" + s.Rekening
		}
		out.Message = replaceWA(c.InvoiceTemplate, v)
		out.Revision = hashText(jsonText([]any{s, b, address, packageName, phone, out.Message}))
		out.HTML = sharedInvoiceHTML(s, b, out.Reference, address, packageName)
		return out, nil
	}
	if kind == "document" {
		var number, typ, date, period, detail, snapshot string
		var mid int64
		e := db.QueryRow(`SELECT nomor,jenis,tanggal,periode,rincian,snapshot,mitra_id FROM dokumen WHERE id=?`, id).Scan(&number, &typ, &date, &period, &detail, &snapshot, &mid)
		if e != nil {
			return out, fmt.Errorf("Dokumen tidak ditemukan")
		}
		var stored DocSnapshot
		if json.Unmarshal([]byte(snapshot), &stored) != nil {
			return out, fmt.Errorf("Data dokumen tidak dapat dibaca")
		}
		m := stored.Mitra
		if m.ID == 0 {
			m, e = getMitraRaw(db, mid)
			if e != nil {
				return out, fmt.Errorf("Mitra dokumen tidak ditemukan")
			}
		}
		if stored.Company != nil {
			s = *stored.Company
		}
		if stored.MasaBulan <= 0 {
			stored.MasaBulan = s.MasaPKSBulan
		}
		var rc *FABRincian
		if typ == "FAB" {
			var d FABRincian
			if json.Unmarshal([]byte(detail), &d) != nil {
				return out, fmt.Errorf("Rincian FAB tidak valid")
			}
			rc = &d
		}
		out.Reference = number
		out.Name = m.Nama
		out.Recipient = m.Telepon
		// Current contact may be corrected without changing the signed document snapshot.
		if current, e := getMitraRaw(db, mid); e == nil {
			out.Recipient = current.Telepon
		}
		v["nama"] = m.Nama
		v["nomor"] = number
		v["jenis"] = typ
		v["tanggal"] = tanggalIndo(date)
		v["perusahaan"] = s.NamaISP
		v["ringkasan"] = ""
		if rc != nil {
			v["ringkasan"] = "Periode: " + periodeLabel(period) + "\nBagi hasil: " + rupiah(rc.Nilai)
		}
		out.Message = replaceWA(c.DocumentTemplate, v)
		out.Revision = hashText(jsonText([]any{snapshot, detail, s, out.Recipient, out.Message}))
		out.HTML = buildDocHTML(s, m, number, typ, date, rc, stored.MasaBulan, true)
		return out, nil
	}
	return out, fmt.Errorf("Pilih invoice, dokumen, atau pesan manual")
}
func (h *Handlers) WAPreview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	c, _, e := loadWA(h.DB)
	if e != nil {
		httpErr(w, 500, "Gagal membaca MPWA")
		return
	}
	s, e := waSource(h.DB, c, r.URL.Query().Get("kind"), id)
	if e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"source": s, "enabled": c.Enabled && c.HasKey, "can_link": c.PublicURL != "", "sender": c.Sender})
}

type WAOutbox struct {
	ID         int64  `json:"id"`
	Kind       string `json:"kind"`
	SourceID   int64  `json:"source_id"`
	Reference  string `json:"reference"`
	Recipient  string `json:"recipient"`
	Message    string `json:"message"`
	Mode       string `json:"mode"`
	Status     string `json:"status"`
	ProviderID string `json:"provider_id"`
	Error      string `json:"error"`
	Created    string `json:"created_at"`
	Finished   string `json:"finished_at"`
	HasShare   bool   `json:"has_share"`
	Revoked    bool   `json:"revoked"`
}

const waSelect = `SELECT o.id,o.kind,o.source_id,o.reference,o.recipient,o.message,o.mode,o.status,o.provider_id,o.error,o.created_at,o.finished_at,o.share_hash!='',COALESCE(s.revoked,0) FROM wa_outbox o LEFT JOIN wa_shares s ON s.hash=o.share_hash`

func scanWA(scan func(...any) error) (WAOutbox, error) {
	var o WAOutbox
	e := scan(&o.ID, &o.Kind, &o.SourceID, &o.Reference, &o.Recipient, &o.Message, &o.Mode, &o.Status, &o.ProviderID, &o.Error, &o.Created, &o.Finished, &o.HasShare, &o.Revoked)
	return o, e
}
func (h *Handlers) WAHistory(w http.ResponseWriter, r *http.Request) {
	where := ""
	args := []any{}
	if kind := r.URL.Query().Get("kind"); kind != "" {
		where = " WHERE o.kind=? AND o.source_id=?"
		args = append(args, kind, r.URL.Query().Get("id"))
	}
	rows, e := h.DB.Query(waSelect+where+` ORDER BY o.id DESC`, args...)
	if e != nil {
		httpErr(w, 500, "Gagal membaca riwayat WhatsApp")
		return
	}
	defer rows.Close()
	out := []WAOutbox{}
	for rows.Next() {
		o, e := scanWA(rows.Scan)
		if e != nil {
			httpErr(w, 500, "Gagal membaca riwayat WhatsApp")
			return
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		httpErr(w, 500, "Gagal membaca riwayat WhatsApp")
		return
	}
	writeJSON(w, 200, out)
}
func newShareToken() (string, string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", "", e
	}
	token := hex.EncodeToString(b)
	return token, hashText(token), nil
}
func hashText(s string) string { b := sha256.Sum256([]byte(s)); return hex.EncodeToString(b[:]) }
func (h *Handlers) WAAttachment(w http.ResponseWriter, r *http.Request) {
	c, _, e := loadWA(h.DB)
	if e != nil || c.PublicURL == "" {
		httpErr(w, 400, "Isi URL publik aplikasi sebelum mengunggah PDF untuk MPWA")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 9<<20)
	if e = r.ParseMultipartForm(9 << 20); e != nil {
		httpErr(w, 400, "PDF maksimal 8 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()
	f, _, e := r.FormFile("file")
	if e != nil {
		httpErr(w, 400, "Pilih file PDF")
		return
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
	if e != nil || len(b) > 8<<20 || !bytes.HasPrefix(b, []byte("%PDF-")) {
		httpErr(w, 400, "File harus PDF valid, maksimal 8 MB")
		return
	}
	token, hash, e := newShareToken()
	if e != nil {
		httpErr(w, 500, "Gagal menyiapkan lampiran")
		return
	}
	_, e = h.DB.Exec(`INSERT INTO wa_shares(hash,content,mime,filename,expires_at) VALUES(?,?,'application/pdf','dokumen.pdf',?)`, hash, b, time.Now().Add(24*time.Hour).Unix())
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan lampiran")
		return
	}
	writeJSON(w, 201, map[string]string{"attachment_id": token})
}
func (h *Handlers) WAShare(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if len(token) != 64 {
		http.NotFound(w, r)
		return
	}
	var content []byte
	var mime, name string
	e := h.DB.QueryRow(`SELECT content,mime,filename FROM wa_shares WHERE hash=? AND revoked=0 AND expires_at>?`, hashText(token), time.Now().Unix()).Scan(&content, &mime, &name)
	if e != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; sandbox")
	if mime == "application/pdf" {
		w.Header().Set("Content-Disposition", `attachment; filename="dokumen.pdf"`)
	}
	_, _ = w.Write(content)
}
func (h *Handlers) WARevoke(w http.ResponseWriter, r *http.Request) {
	res, e := h.DB.Exec(`UPDATE wa_shares SET revoked=1 WHERE hash=(SELECT share_hash FROM wa_outbox WHERE id=?)`, operationID(r))
	if e != nil {
		httpErr(w, 500, "Gagal menutup tautan")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpErr(w, 404, "Tautan tidak ditemukan atau sudah kedaluwarsa")
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

type WASendInput struct {
	Revision     string `json:"revision"`
	Key          string `json:"request_key"`
	Kind         string `json:"kind"`
	SourceID     int64  `json:"source_id"`
	Recipient    string `json:"recipient"`
	Message      string `json:"message"`
	IncludeLink  bool   `json:"include_link"`
	AttachmentID string `json:"attachment_id"`
}

var requestKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{16,80}$`)

func (h *Handlers) WASend(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	var in WASendInput
	if readJSON(r, &in) != nil || !requestKeyPattern.MatchString(in.Key) || len(in.Message) > 10000 || strings.TrimSpace(in.Message) == "" {
		httpErr(w, 400, "Pesan atau identitas pengiriman tidak valid")
		return
	}
	var e error
	in.Recipient, e = normalizePhone(in.Recipient)
	if e != nil || in.Recipient == "" {
		httpErr(w, 400, "Isi nomor penerima dengan kode negara yang valid")
		return
	}
	operationMu.Lock()
	// The closure always releases the mutation lock, including on validation errors.
	var conf WAConfig
	var apiKey, media string
	var out WAOutbox
	reserved := func() bool {
		defer operationMu.Unlock()
		requestHash := hashText(jsonText(in))
		var oldHash string
		e = h.DB.QueryRow(`SELECT request_hash FROM wa_outbox WHERE request_key=?`, in.Key).Scan(&oldHash)
		if e == nil {
			if oldHash != requestHash {
				httpErr(w, 409, "Pengiriman ini sudah diproses. Mulai pesan baru untuk perubahan.")
				return false
			}
			old, e := scanWA(h.DB.QueryRow(waSelect+` WHERE o.request_key=?`, in.Key).Scan)
			if e != nil {
				httpErr(w, 500, "Gagal membaca hasil pengiriman")
			} else {
				writeJSON(w, 200, old)
			}
			return false
		}
		if e != sql.ErrNoRows {
			httpErr(w, 500, "Gagal memeriksa pengiriman")
			return false
		}
		var encrypted string
		conf, encrypted, e = loadWA(h.DB)
		if e != nil || !conf.Enabled || encrypted == "" {
			httpErr(w, 400, "Lengkapi dan aktifkan MPWA di Pengaturan")
			return false
		}
		apiKey, e = sealWA(h.DB, encrypted, true)
		if e != nil {
			httpErr(w, 500, "Kunci API tidak dapat dibaca; masukkan kembali di Pengaturan")
			return false
		}
		source, e := waSource(h.DB, conf, in.Kind, in.SourceID)
		if e != nil {
			httpErr(w, 400, e.Error())
			return false
		}
		if in.Kind != "custom" && (in.Revision == "" || in.Revision != source.Revision) {
			httpErr(w, 409, "Data dokumen berubah sejak preview. Tutup dan buka kembali pesan sebelum mengirim.")
			return false
		}
		if (in.IncludeLink || in.AttachmentID != "") && conf.PublicURL == "" {
			httpErr(w, 400, "Isi URL publik aplikasi untuk mengirim tautan atau PDF")
			return false
		}
		if in.IncludeLink && (source.HTML == "" || in.AttachmentID != "") {
			httpErr(w, 400, "Pilih satu lampiran: tautan dokumen atau PDF")
			return false
		}
		body := in.Message
		mode := "text"
		shareHash := ""
		token := ""
		if in.IncludeLink {
			token, shareHash, e = newShareToken()
			if e != nil {
				httpErr(w, 500, "Gagal menyiapkan tautan")
				return false
			}
			body += "\n\nLihat " + source.Reference + ":\n" + conf.PublicURL + "/share/" + token + "\nTautan berlaku 7 hari."
			mode = "link"
		}
		if in.AttachmentID != "" {
			if len(in.AttachmentID) != 64 {
				httpErr(w, 400, "Lampiran tidak valid")
				return false
			}
			shareHash = hashText(in.AttachmentID)
			var n int
			e = h.DB.QueryRow(`SELECT COUNT(*) FROM wa_shares WHERE hash=? AND mime='application/pdf' AND revoked=0 AND expires_at>?`, shareHash, time.Now().Add(5*time.Minute).Unix()).Scan(&n)
			if e != nil || n == 0 {
				httpErr(w, 400, "Lampiran kedaluwarsa; unggah kembali PDF")
				return false
			}
			media = conf.PublicURL + "/share/" + in.AttachmentID
			mode = "pdf"
		}
		tx, e := h.DB.Begin()
		if e != nil {
			httpErr(w, 500, "Gagal menyiapkan pengiriman")
			return false
		}
		defer tx.Rollback()
		if in.IncludeLink {
			_, e = tx.Exec(`INSERT INTO wa_shares(hash,content,mime,filename,expires_at) VALUES(?,?,'text/html; charset=utf-8','dokumen.html',?)`, shareHash, []byte(source.HTML), time.Now().Add(7*24*time.Hour).Unix())
		}
		created := time.Now().UTC().Format(time.RFC3339)
		var result sql.Result
		if e == nil {
			result, e = tx.Exec(`INSERT INTO wa_outbox(request_key,request_hash,kind,source_id,reference,recipient,message,mode,status,created_at,share_hash) VALUES(?,?,?,?,?,?,?,?,'sending',?,?)`, in.Key, requestHash, in.Kind, in.SourceID, source.Reference, in.Recipient, body, mode, created, shareHash)
		}
		if e == nil {
			e = tx.Commit()
		}
		if e != nil {
			httpErr(w, 500, "Gagal menyiapkan pengiriman")
			return false
		}
		id, _ := result.LastInsertId()
		out = WAOutbox{ID: id, Kind: in.Kind, SourceID: in.SourceID, Reference: source.Reference, Recipient: in.Recipient, Message: body, Mode: mode, Status: "sending", Created: created, HasShare: shareHash != ""}
		return true
	}()
	if !reserved {
		return
	}
	// Do not retry HTTP requests automatically: a timeout may already be delivered.
	out.Status, out.ProviderID, out.Error = sendMPWA(conf, apiKey, out.Recipient, out.Message, media)
	out.Finished = time.Now().UTC().Format(time.RFC3339)
	_, e = h.DB.Exec(`UPDATE wa_outbox SET status=?,provider_id=?,error=?,finished_at=? WHERE id=?`, out.Status, out.ProviderID, out.Error, out.Finished, out.ID)
	if e != nil {
		httpErr(w, 500, "Hasil gateway belum tersimpan. Periksa riwayat MPWA sebelum mengirim ulang.")
		return
	}
	writeJSON(w, 200, out)
}

var mpwaTimeout = 25 * time.Second

func sendMPWA(c WAConfig, key, recipient, message, media string) (string, string, string) {
	endpoint := "/send-message"
	form := url.Values{"api_key": {key}, "sender": {c.Sender}, "number": {recipient}, "message": {message}}
	if media != "" {
		endpoint = "/send-media"
		form.Set("url", media)
		form.Set("media_type", "document")
		form.Set("caption", message)
	}
	req, e := http.NewRequest(http.MethodPost, c.BaseURL+endpoint, strings.NewReader(form.Encode()))
	if e != nil {
		return "failed", "", "URL gateway tidak valid"
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: mpwaTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, e := client.Do(req)
	if e != nil {
		return "unknown", "", "Tidak ada kepastian hasil dari MPWA. Periksa riwayat MPWA sebelum mengirim ulang."
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode < 500 {
		return "failed", "", fmt.Sprintf("MPWA menolak permintaan (HTTP %d). Periksa URL endpoint, API key, dan sender.", resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		return "unknown", "", fmt.Sprintf("MPWA mengalami gangguan (HTTP %d). Periksa riwayat gateway sebelum mengirim ulang.", resp.StatusCode)
	}
	var result struct {
		Status *bool `json:"status"`
		Data   struct {
			Key struct {
				ID string `json:"id"`
			} `json:"key"`
		} `json:"data"`
	}
	if e = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); e != nil {
		return "unknown", "", "Respons MPWA tidak dikenali. Periksa versi API dan riwayat gateway."
	}
	if result.Status == nil {
		return "unknown", "", "Respons MPWA tidak memuat status. Periksa riwayat gateway sebelum mengirim ulang."
	}
	if !*result.Status {
		return "failed", "", "MPWA menolak pesan. Periksa nomor penerima, status perangkat, kuota, dan API key di gateway."
	}
	id := result.Data.Key.ID
	if len(id) > 128 || strings.Contains(id, key) {
		id = ""
	}
	return "accepted", id, ""
}
