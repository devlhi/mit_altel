package main

import (
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// These user-provided assets are intentionally embedded only in authenticated letters.
//
//go:embed assets/pop-letterhead.png
var popLetterhead []byte

//go:embed assets/pop-signature.png
var popSignature []byte

type POPLetterService struct {
	ID          string `json:"id"`
	KBLI        string `json:"kbli"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope,omitempty"`
	SourceURL   string `json:"source_url"`
}

var popLetterServices = []POPLetterService{
	{ID: "isp", KBLI: "61104", Name: "Penyelenggaraan Jasa Akses Internet (Internet Service Provider/ISP)", Description: "Aktivitas Jasa Akses Internet (Internet Service Provider)", SourceURL: "https://oss.go.id/en/kbli/detail/5c4bfd55-4ab5-5290-a110-977bf8922aa1"},
	{ID: "local_fiber", KBLI: "61101", Name: "Penyelenggaraan Jaringan Tetap Lokal Berbasis Packet Switched melalui Media Fiber Optik", Description: "Aktivitas Telekomunikasi dengan Kabel", Scope: "Jaringan tetap lokal", SourceURL: "https://oss.go.id/en/kbli/detail/36273dfa-4e91-5587-acec-509799b2a3fe"},
	{ID: "closed_fiber", KBLI: "61101", Name: "Penyelenggaraan Jaringan Tetap Tertutup melalui Media Fiber Optik Terestrial", Description: "Aktivitas Telekomunikasi dengan Kabel", Scope: "Jaringan tetap tertutup", SourceURL: "https://oss.go.id/en/kbli/detail/36273dfa-4e91-5587-acec-509799b2a3fe"},
}

const popCompany = "PT. ALI TELEMATIKA NUSANTARA"
const popLetterCompanyName = "PT Ali Telematika Nusantara"
const popLetterSubject = "Permohonan Verifikasi Izin Penyelenggaraan Telekomunikasi di Luar Komitmen " + popLetterCompanyName
const popOSSHistory = "Proses verifikasi dan persetujuan tersebut telah dicoba dilakukan bersama antara PT ATN dan Direktorat Layanan Ekosistem Digital melalui akun Online Single Submission (OSS) PT ATN."
const popAddress = "Dusun III Perum Surya Pagi Lokasi V, Desa Tuladenggi, Kec. Telaga Biru, Kabupaten Gorontalo, Gorontalo, 96137"

type POPLetterProfile struct {
	City             string `json:"city"`
	NIB              string `json:"nib"`
	Signer           string `json:"signer"`
	Job              string `json:"job"`
	Recipient        string `json:"recipient"`
	RecipientName    string `json:"recipient_name,omitempty"`
	RecipientAddress string `json:"recipient_address"`
}
type POPLetterLocation struct {
	POPID   int64  `json:"pop_id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}
type POPLetterInput struct {
	RequestKey        string              `json:"request_key"`
	Date              string              `json:"date"`
	Mode              string              `json:"mode"`
	ServiceID         string              `json:"service_id"`
	KBLI              string              `json:"kbli"`
	KBLIVersion       string              `json:"kbli_version"`
	KBLIDescription   string              `json:"kbli_description,omitempty"`
	Service           string              `json:"service"`
	ManualNumber      string              `json:"manual_number"`
	PreviousNumber    string              `json:"previous_number"`
	PreviousDate      string              `json:"previous_date"`
	PreviousSubject   string              `json:"previous_subject,omitempty"`
	PreviousRecipient string              `json:"previous_recipient,omitempty"`
	ProcessHistory    string              `json:"process_history,omitempty"`
	License           string              `json:"license"`
	Note              string              `json:"note"`
	Profile           POPLetterProfile    `json:"profile"`
	Locations         []POPLetterLocation `json:"locations"`
}
type POPLetterRecord struct {
	ID         int64          `json:"id"`
	Number     string         `json:"number"`
	IssuedAt   string         `json:"issued_at"`
	VoidReason string         `json:"void_reason"`
	Data       POPLetterInput `json:"data"`
}

func migratePOPLetters(db *sql.DB) error {
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS pop_letter_profile (id INTEGER PRIMARY KEY CHECK(id=1), data TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS pop_letters (id INTEGER PRIMARY KEY AUTOINCREMENT, number TEXT NOT NULL UNIQUE, issued_at TEXT NOT NULL, request_key TEXT NOT NULL UNIQUE, request_hash TEXT NOT NULL, data TEXT NOT NULL, html TEXT NOT NULL, void_reason TEXT NOT NULL DEFAULT '')`,
		`CREATE TABLE IF NOT EXISTS pop_letter_sequence (year INTEGER PRIMARY KEY, last INTEGER NOT NULL)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
func registerPOPLetters(mux *http.ServeMux, h *Handlers) {
	for pattern, fn := range map[string]http.HandlerFunc{
		"GET /api/surat-pop/options":    h.POPLetterOptions,
		"GET /api/surat-pop/sample":     h.POPLetterSample,
		"PUT /api/surat-pop/profile":    h.SavePOPLetterProfile,
		"GET /api/surat-pop":            h.ListPOPLetters,
		"POST /api/surat-pop/preview":   h.PreviewPOPLetter,
		"POST /api/surat-pop":           h.IssuePOPLetter,
		"GET /api/surat-pop/{id}/html":  h.POPLetterHTML,
		"POST /api/surat-pop/{id}/void": h.VoidPOPLetter,
	} {
		mux.HandleFunc(pattern, operationHandler(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Cache-Control", "no-store"); fn(w, r) }))
	}
}
func getPOPLetterProfile(db *sql.DB) (POPLetterProfile, error) {
	s := getSettings(db)
	p := POPLetterProfile{City: "Gorontalo", Signer: s.Penandatangan, Job: s.Jabatan, Recipient: "Direktur Layanan Ekosistem Digital\nDirektorat Jenderal Ekosistem Digital\nKementerian Komunikasi dan Digital RI", RecipientAddress: "Menara Danareksa Lt. 8\nJl. Medan Merdeka Selatan No. 14, Gambir, Jakarta"}
	var raw string
	err := db.QueryRow(`SELECT data FROM pop_letter_profile WHERE id=1`).Scan(&raw)
	if err == nil {
		err = json.Unmarshal([]byte(raw), &p)
	}
	if err != nil && err != sql.ErrNoRows {
		return p, err
	}
	return p, nil
}
func (h *Handlers) POPLetterOptions(w http.ResponseWriter, r *http.Request) {
	p, err := getPOPLetterProfile(h.DB)
	if err != nil {
		httpErr(w, 500, "Profil surat tidak dapat dibaca")
		return
	}
	writeJSON(w, 200, map[string]any{"simple_template": true, "services": popLetterServices, "profile": p, "company": popCompany, "subject": popLetterSubject, "oss_history": popOSSHistory, "address": popAddress, "phone": "+62 822-9261-8190", "email": "info@altelmatara.id", "website": "www.altelmatara.id", "letterhead": "data:image/png;base64," + base64.StdEncoding.EncodeToString(popLetterhead)})
}

// Sample placeholders never pass through issuance validation and consume no number.
// They intentionally represent missing facts rather than inventing previous letters.
func (h *Handlers) POPLetterSample(w http.ResponseWriter, r *http.Request) {
	p, err := getPOPLetterProfile(h.DB)
	if err != nil {
		httpErr(w, 500, "Profil surat tidak dapat dibaca")
		return
	}
	if p.RecipientName == "" {
		p.RecipientName = "[Nama pejabat penerima]"
	}
	service := popLetterServices[1]
	for _, s := range popLetterServices {
		if s.ID == r.URL.Query().Get("service") {
			service = s
		}
	}
	in := POPLetterInput{Date: time.Now().Format("2006-01-02"), Mode: "followup", ServiceID: service.ID, KBLI: service.KBLI, KBLIVersion: "2025", KBLIDescription: service.Description, Service: service.Name, Profile: p,
		PreviousNumber: "[nomor surat sebelumnya]", PreviousDate: "[tanggal surat sebelumnya]", PreviousSubject: popLetterSubject, PreviousRecipient: strings.Join(strings.Fields(p.Recipient), " "), ProcessHistory: popOSSHistory,
		Locations: []POPLetterLocation{{Name: "[Nama POP 1]", Address: "[Jalan / dusun, desa / kelurahan, kecamatan, kabupaten / kota, provinsi]"}, {Name: "[Nama POP 2]", Address: "[Alamat lengkap lokasi POP kedua]"}, {Name: "[Nama POP 3]", Address: "[Alamat lengkap lokasi POP ketiga]"}},
	}
	body := buildPOPLetterHTML(in, "[nomor surat baru]", true)
	body = strings.Replace(body, ">DRAF</div>", ">CONTOH</div>", 1)
	body = strings.Replace(body, "Pratinjau — nomor final diberikan saat diterbitkan", "Contoh redaksi surat lanjutan — nomor, tanggal rujukan, lokasi, dan riwayat OSS harus disesuaikan", 1)
	writeJSON(w, 200, map[string]string{"html": body})
}
func validatePOPLetterProfile(p *POPLetterProfile) error {
	for _, f := range []struct {
		value    *string
		label    string
		max      int
		required bool
	}{
		{&p.City, "Kota penerbitan", 100, true}, {&p.NIB, "NIB", 30, false}, {&p.Signer, "Nama penandatangan", 150, true}, {&p.Job, "Jabatan", 150, true}, {&p.Recipient, "Penerima surat", 700, true}, {&p.RecipientName, "Nama pejabat penerima", 150, false}, {&p.RecipientAddress, "Alamat penerima", 700, true},
	} {
		*f.value = strings.TrimSpace(*f.value)
		if (f.required && *f.value == "") || len([]rune(*f.value)) > f.max {
			return fmt.Errorf("%s wajib sesuai batas %d karakter", f.label, f.max)
		}
	}
	if p.NIB != "" && !regexp.MustCompile(`^[0-9]{13}$`).MatchString(p.NIB) {
		return fmt.Errorf("NIB harus terdiri dari 13 digit, atau kosongkan jika belum diisi")
	}
	return nil
}
func (h *Handlers) SavePOPLetterProfile(w http.ResponseWriter, r *http.Request) {
	var p POPLetterProfile
	if readJSON(r, &p) != nil {
		httpErr(w, 400, "Data profil tidak valid")
		return
	}
	if err := validatePOPLetterProfile(&p); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	if _, err := h.DB.Exec(`INSERT INTO pop_letter_profile(id,data) VALUES(1,?) ON CONFLICT(id) DO UPDATE SET data=excluded.data`, jsonText(p)); err != nil {
		httpErr(w, 500, "Profil gagal disimpan")
		return
	}
	writeJSON(w, 200, p)
}
func validatePOPLetter(in *POPLetterInput) error {
	return validatePOPLetterInput(in, false)
}

// The everyday form only supplies the new letter number, KBLI and POP addresses.
// Return the resolved data with the preview so issuing/retrying keeps its date and profile.
func preparePOPLetterTemplate(db *sql.DB, in *POPLetterInput) error {
	if in.Mode != "template" {
		return nil
	}
	if in.Date == "" {
		in.Date = time.Now().In(time.FixedZone("WITA", 8*60*60)).Format("2006-01-02")
	}
	if in.Profile == (POPLetterProfile{}) {
		p, err := getPOPLetterProfile(db)
		if err != nil {
			return fmt.Errorf("Profil surat tidak dapat dibaca")
		}
		in.Profile = p
	}
	in.KBLIVersion = "2025"
	// A manually typed known KBLI uses the same official label as the select option.
	if in.ServiceID == "manual" {
		for _, s := range popLetterServices {
			if s.KBLI == strings.TrimSpace(in.KBLI) {
				in.ServiceID = s.ID
				break
			}
		}
		if in.ServiceID == "manual" {
			in.Service = "Penyelenggaraan Telekomunikasi"
			in.KBLIDescription = ""
		}
	}
	in.PreviousNumber, in.PreviousDate, in.PreviousSubject, in.PreviousRecipient = "", "", "", ""
	in.ProcessHistory = popOSSHistory
	in.License, in.Note = "", ""
	for i := range in.Locations {
		in.Locations[i].Name = ""
	}
	return nil
}

// Drafts may leave the follow-up facts blank; issuance always requires real values.
// Invalid non-empty values and all other letter fields have the same validation.
func validatePOPLetterInput(in *POPLetterInput, draft bool) error {
	if !validDate(in.Date) {
		return fmt.Errorf("Tanggal surat tidak valid")
	}
	if in.Mode != "initial" && in.Mode != "followup" && in.Mode != "template" {
		return fmt.Errorf("Pilih permohonan awal atau lanjutan")
	}
	if in.KBLIVersion != "2025" && in.KBLIVersion != "2020" {
		return fmt.Errorf("Pilih versi KBLI")
	}
	if in.ServiceID != "manual" {
		found := false
		for _, s := range popLetterServices {
			if s.ID == in.ServiceID {
				in.KBLI = s.KBLI
				in.Service = s.Name
				in.KBLIDescription = s.Description
				found = true
			}
		}
		if !found || in.KBLIVersion != "2025" {
			return fmt.Errorf("Pilih layanan KBLI 2025, atau gunakan isian manual sesuai NIB/OSS")
		}
	}
	in.KBLI = strings.TrimSpace(in.KBLI)
	in.Service = strings.TrimSpace(in.Service)
	in.KBLIDescription = strings.TrimSpace(in.KBLIDescription)
	if !regexp.MustCompile(`^[0-9]{5}$`).MatchString(in.KBLI) || in.Service == "" || len([]rune(in.Service)) > 350 {
		return fmt.Errorf("Isi KBLI 5 digit dan jenis layanan (maksimal 350 karakter)")
	}
	if (in.KBLIDescription == "" && !(in.Mode == "template" && in.ServiceID == "manual")) || len([]rune(in.KBLIDescription)) > 350 {
		return fmt.Errorf("Isi uraian KBLI sesuai NIB/OSS (maksimal 350 karakter)")
	}
	for _, f := range []struct {
		value *string
		max   int
	}{{&in.ManualNumber, 120}, {&in.PreviousNumber, 120}, {&in.PreviousSubject, 500}, {&in.PreviousRecipient, 700}, {&in.ProcessHistory, 1800}, {&in.License, 250}, {&in.Note, 1800}} {
		*f.value = strings.TrimSpace(*f.value)
		if len([]rune(*f.value)) > f.max {
			return fmt.Errorf("Isian surat melebihi batas karakter")
		}
	}
	if in.Mode == "template" && in.ManualNumber == "" {
		return fmt.Errorf("Isi nomor surat")
	}
	if in.Mode == "followup" {
		if (!draft && (in.PreviousNumber == "" || in.PreviousDate == "")) || (in.PreviousDate != "" && (!validDate(in.PreviousDate) || in.PreviousDate > in.Date)) {
			return fmt.Errorf("Surat lanjutan memerlukan nomor dan tanggal surat sebelumnya, paling lambat tanggal surat ini")
		}
		if in.PreviousSubject == "" {
			in.PreviousSubject = popLetterSubject
		}
		if in.PreviousRecipient == "" {
			in.PreviousRecipient = strings.Join(strings.Fields(in.Profile.Recipient), " ")
		}
		if !draft && in.ProcessHistory == "" {
			return fmt.Errorf("Isi riwayat proses verifikasi untuk surat lanjutan, atau pilih redaksi OSS jika sesuai")
		}
	} else if in.Mode == "initial" {
		in.PreviousDate = ""
		in.PreviousNumber = ""
		in.PreviousSubject = ""
		in.PreviousRecipient = ""
		in.ProcessHistory = ""
	}
	if err := validatePOPLetterProfile(&in.Profile); err != nil {
		return err
	}
	if len(in.Locations) < 1 || len(in.Locations) > 30 {
		return fmt.Errorf("Isi 1 hingga 30 lokasi POP")
	}
	seen := map[string]bool{}
	for i := range in.Locations {
		p := &in.Locations[i]
		p.Name = strings.TrimSpace(p.Name)
		p.Address = strings.TrimSpace(p.Address)
		if (p.Name == "" && in.Mode != "template") || p.Address == "" || len([]rune(p.Name)) > 150 || len([]rune(p.Address)) > 700 || p.POPID < 0 {
			return fmt.Errorf("Lokasi %d: isi nama (maks. 150) dan alamat lengkap (maks. 700 karakter)", i+1)
		}
		key := strings.ToLower(p.Name + "\n" + p.Address)
		if seen[key] {
			return fmt.Errorf("Lokasi POP yang sama tidak boleh dicantumkan dua kali")
		}
		seen[key] = true
	}
	return nil
}

func missingPOPLetterFollowup(in POPLetterInput) []string {
	missing := []string{}
	if in.Mode == "followup" {
		for _, f := range []struct{ value, label string }{
			{in.PreviousNumber, "Nomor surat sebelumnya"},
			{in.PreviousDate, "Tanggal surat sebelumnya"},
			{in.ProcessHistory, "Riwayat proses verifikasi / OSS"},
		} {
			if strings.TrimSpace(f.value) == "" {
				missing = append(missing, f.label)
			}
		}
	}
	return missing
}

type POPLetterPreview struct {
	HTML     string          `json:"html"`
	CanIssue bool            `json:"can_issue"`
	Missing  []string        `json:"missing"`
	Data     *POPLetterInput `json:"data,omitempty"`
}

func (h *Handlers) PreviewPOPLetter(w http.ResponseWriter, r *http.Request) {
	var in POPLetterInput
	if readJSON(r, &in) != nil {
		httpErr(w, 400, "Data surat tidak valid")
		return
	}
	if err := preparePOPLetterTemplate(h.DB, &in); err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	if err := validatePOPLetterInput(&in, true); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	missing := missingPOPLetterFollowup(in)
	number := "DRAF / belum diterbitkan"
	if in.ManualNumber != "" {
		number = in.ManualNumber
	}
	writeJSON(w, 200, POPLetterPreview{HTML: buildPOPLetterHTML(in, number, true), CanIssue: len(missing) == 0, Missing: missing, Data: &in})
}
func (h *Handlers) IssuePOPLetter(w http.ResponseWriter, r *http.Request) {
	var in POPLetterInput
	if readJSON(r, &in) != nil {
		httpErr(w, 400, "Data surat tidak valid")
		return
	}
	if err := preparePOPLetterTemplate(h.DB, &in); err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	if err := validatePOPLetter(&in); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9-]{16,80}$`).MatchString(in.RequestKey) {
		httpErr(w, 400, "Kunci penerbitan tidak valid. Buka ulang pratinjau.")
		return
	}
	hash := sha256.Sum256([]byte(jsonText(in)))
	digest := hex.EncodeToString(hash[:])
	tx, err := h.DB.Begin()
	if err != nil {
		httpErr(w, 500, "Arsip tidak dapat dibuka")
		return
	}
	defer tx.Rollback()
	var existing POPLetterRecord
	var oldHash string
	err = tx.QueryRow(`SELECT id,number,issued_at,request_hash,void_reason FROM pop_letters WHERE request_key=?`, in.RequestKey).Scan(&existing.ID, &existing.Number, &existing.IssuedAt, &oldHash, &existing.VoidReason)
	if err == nil {
		if oldHash != digest {
			httpErr(w, 409, "Permintaan ini sudah diterbitkan dengan isi berbeda. Buat surat baru.")
			return
		}
		existing.Data = in
		writeJSON(w, 200, existing)
		return
	}
	if err != sql.ErrNoRows {
		httpErr(w, 500, "Arsip tidak dapat diperiksa")
		return
	}
	number := in.ManualNumber
	if number == "" {
		t, _ := time.Parse("2006-01-02", in.Date)
		months := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
		for {
			var seq int
			err = tx.QueryRow(`INSERT INTO pop_letter_sequence(year,last) VALUES(?,1) ON CONFLICT(year) DO UPDATE SET last=last+1 RETURNING last`, t.Year()).Scan(&seq)
			if err != nil {
				httpErr(w, 500, "Nomor surat gagal dibuat")
				return
			}
			number = fmt.Sprintf("%03d/POP/ATN-KOMDIGI/%s/%d", seq, months[int(t.Month())-1], t.Year())
			var n int
			err = tx.QueryRow(`SELECT COUNT(*) FROM pop_letters WHERE number=?`, number).Scan(&n)
			if err != nil {
				httpErr(w, 500, "Nomor surat gagal diperiksa")
				return
			}
			if n == 0 {
				break
			}
		}
	} else {
		var n int
		if err = tx.QueryRow(`SELECT COUNT(*) FROM pop_letters WHERE number=?`, number).Scan(&n); err != nil {
			httpErr(w, 500, "Nomor surat gagal diperiksa")
			return
		}
		if n > 0 {
			httpErr(w, 409, "Nomor surat sudah dipakai, termasuk pada surat yang dibatalkan")
			return
		}
	}
	record := POPLetterRecord{Number: number, IssuedAt: time.Now().UTC().Format(time.RFC3339), Data: in}
	res, err := tx.Exec(`INSERT INTO pop_letters(number,issued_at,request_key,request_hash,data,html) VALUES(?,?,?,?,?,?)`, number, record.IssuedAt, in.RequestKey, digest, jsonText(in), buildPOPLetterHTML(in, number, false))
	if err == nil {
		record.ID, err = res.LastInsertId()
	}
	if err == nil {
		err = audit(tx, "pop_letter.issue", map[string]any{"id": record.ID, "number": number, "kbli": in.KBLI, "locations": len(in.Locations)})
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		httpErr(w, 500, "Surat gagal diterbitkan")
		return
	}
	writeJSON(w, 201, record)
}
func (h *Handlers) ListPOPLetters(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id,number,issued_at,void_reason,data FROM pop_letters ORDER BY id DESC`)
	if err != nil {
		httpErr(w, 500, "Arsip gagal dimuat")
		return
	}
	defer rows.Close()
	list := []POPLetterRecord{}
	for rows.Next() {
		var v POPLetterRecord
		var raw string
		if err = rows.Scan(&v.ID, &v.Number, &v.IssuedAt, &v.VoidReason, &raw); err != nil {
			break
		}
		if err = json.Unmarshal([]byte(raw), &v.Data); err != nil {
			break
		}
		v.Data.RequestKey = ""
		list = append(list, v)
	}
	if err != nil || rows.Err() != nil {
		httpErr(w, 500, "Arsip gagal dibaca")
		return
	}
	writeJSON(w, 200, list)
}
func (h *Handlers) POPLetterHTML(w http.ResponseWriter, r *http.Request) {
	var body, reason string
	err := h.DB.QueryRow(`SELECT html,void_reason FROM pop_letters WHERE id=?`, operationID(r)).Scan(&body, &reason)
	if err == sql.ErrNoRows {
		httpErr(w, 404, "Surat tidak ditemukan")
		return
	}
	if err != nil {
		httpErr(w, 500, "Surat tidak dapat dibaca")
		return
	}
	if reason != "" {
		body = strings.Replace(body, "<body>", "<body><div class=watermark>DIBATALKAN</div><div class=void-note>Surat dibatalkan: "+escHTML(reason)+"</div>", 1)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'; frame-ancestors 'self'")
	fmt.Fprint(w, body)
}
func (h *Handlers) VoidPOPLetter(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Reason string `json:"reason"`
	}
	if readJSON(r, &v) != nil {
		httpErr(w, 400, "Data pembatalan tidak valid")
		return
	}
	v.Reason = strings.TrimSpace(v.Reason)
	if v.Reason == "" || len([]rune(v.Reason)) > 500 {
		httpErr(w, 400, "Isi alasan pembatalan (maksimal 500 karakter)")
		return
	}
	tx, err := h.DB.Begin()
	if err != nil {
		httpErr(w, 500, "Arsip tidak dapat dibuka")
		return
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE pop_letters SET void_reason=? WHERE id=? AND void_reason=''`, v.Reason, operationID(r))
	if err != nil {
		httpErr(w, 500, "Pembatalan gagal")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpErr(w, 409, "Surat tidak ditemukan atau sudah dibatalkan")
		return
	}
	if err = audit(tx, "pop_letter.void", map[string]any{"id": operationID(r), "reason": v.Reason}); err == nil {
		err = tx.Commit()
	}
	if err != nil {
		httpErr(w, 500, "Pembatalan gagal")
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func buildPOPLetterHTML(in POPLetterInput, number string, draft bool) string {
	e := escHTML
	lines := func(s string) string { return strings.ReplaceAll(e(s), "\n", "<br>") }
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` + e(number) + `</title><style>
@page{size:A4;margin:12mm 18mm 16mm}*{box-sizing:border-box}html{overflow-x:hidden}body{margin:0;color:#111;background:#edf0f5;font:12pt/1.25 "Times New Roman",Times,serif}.paper{width:210mm;min-height:297mm;margin:0 auto;padding:12mm 18mm 16mm;background:white}.letterhead-block{border-bottom:3px double #111;padding-bottom:2mm;margin-bottom:4mm;break-inside:avoid}.letterhead{display:block;width:100%;max-width:166mm;height:auto;margin:0 auto}.meta{display:flex;justify-content:space-between;gap:8mm;margin-bottom:5mm}.meta span{overflow-wrap:anywhere}.recipient{margin-bottom:5mm}.subject{display:flex;gap:3mm;margin-bottom:5mm}.subject strong{flex:1;font-weight:400}p{margin:0 0 4mm;overflow-wrap:anywhere;text-align:justify;orphans:3;widows:3}ol{padding-left:7mm;margin:3mm 0 5mm}li{padding-left:1mm;margin:0 0 1.5mm;break-inside:avoid;overflow-wrap:anywhere}.closing{break-inside:avoid}.sign{break-inside:avoid;margin-top:4mm}.signature{display:block;width:58mm;height:auto;margin:1mm 0}.signer{text-decoration:underline;font-weight:bold}.watermark{position:fixed;top:44%;left:0;right:0;text-align:center;color:rgba(100,116,139,.15);font:bold 56pt Arial,sans-serif;transform:rotate(-28deg);pointer-events:none;z-index:2}.draft-note,.void-note{font:10pt Arial,sans-serif;text-align:center;padding:3mm;color:#8a3e13;background:#fff3df}.void-note{color:#a82020;background:#ffecec}strong{font-weight:700}.identity{margin-bottom:4mm;font-size:11pt} .recipient,.subject,.sign{overflow-wrap:anywhere}@media print{body{background:white}.paper{width:auto;min-height:0;margin:0;padding:0}.draft-note{display:none}.watermark{position:fixed}}</style></head><body>`)
	if draft {
		note := "Pratinjau — nomor final diberikan saat diterbitkan"
		if in.ManualNumber != "" {
			note = "Pratinjau — surat belum diterbitkan"
		}
		if len(missingPOPLetterFollowup(in)) > 0 {
			note = "Draf lanjutan — bagian kuning perlu dilengkapi sebelum surat diterbitkan"
		}
		b.WriteString(`<div class="draft-note">` + note + `</div><div class="watermark">DRAF</div>`)
	}
	b.WriteString(`<main class="paper"><header class="letterhead-block"><img class="letterhead" alt="` + e(popCompany+". "+popAddress+". Phone: +62 822-9261-8190. Email: info@altelmatara.id. Website: www.altelmatara.id") + `" src="data:image/png;base64,` + base64.StdEncoding.EncodeToString(popLetterhead) + `"></header><div class="meta"><span>No.: ` + e(number) + `</span><span>` + e(in.Profile.City) + `, ` + e(tanggalIndo(in.Date)) + `</span></div><div class="recipient">Kepada Yth.<br><strong>`)
	if in.Profile.RecipientName != "" {
		b.WriteString(e(in.Profile.RecipientName) + `<br>`)
	}
	b.WriteString(lines(in.Profile.Recipient) + `</strong><br>` + lines(in.Profile.RecipientAddress) + `</div>`)
	subject := popLetterSubject + ` (“PT ATN”)`
	if in.Mode == "followup" || in.Mode == "template" {
		subject = "Lanjutan " + subject
	}
	b.WriteString(`<div class="subject"><span>Perihal:</span><strong>` + e(subject) + `</strong></div><p>Dengan hormat,</p>`)
	service := in.Service
	if !strings.HasPrefix(strings.ToLower(service), "izin ") {
		service = "Izin " + service
	}
	classification := e(in.KBLI)
	if in.KBLIDescription != "" {
		classification += ` – ` + e(in.KBLIDescription)
	}
	requestedLicense := `<strong>` + e(service) + ` (KBLI ` + classification + `)</strong>`
	if in.Mode == "template" {
		history := strings.ReplaceAll(e(strings.TrimSuffix(in.ProcessHistory, ".")), "Online Single Submission", `<em>Online Single Submission</em>`)
		history = strings.TrimPrefix(history, "Proses ")
		b.WriteString(`<p>Menindaklanjuti permohonan dari PT ATN kepada ` + e(strings.Join(strings.Fields(in.Profile.Recipient), " ")) + ` perihal ` + e(popLetterSubject) + `, yang mana proses ` + history + `, oleh sebab itu sebagai tindak lanjutnya bersama surat ini <strong>PT ATN</strong> mengajukan permohonan untuk dilakukan verifikasi dan persetujuan atas ` + requestedLicense + ` untuk lokasi-lokasi sebagai berikut:</p>`)
	} else if in.Mode == "followup" {
		previousNumber, previousDate := e(in.PreviousNumber), e(tanggalIndo(in.PreviousDate))
		history := lines(strings.TrimRight(in.ProcessHistory, ". \t\r\n"))
		placeholder := func(s string) string {
			return `<mark class="draft-field" style="background:#fff3cd;color:#704d0c">[` + e(s) + `]</mark>`
		}
		if draft {
			if in.PreviousNumber == "" {
				previousNumber = placeholder("nomor surat sebelumnya")
			}
			if in.PreviousDate == "" {
				previousDate = placeholder("tanggal surat sebelumnya")
			}
			if in.ProcessHistory == "" {
				history = placeholder("Contoh riwayat — sesuaikan: " + strings.TrimSuffix(popOSSHistory, "."))
			}
		}
		history = strings.ReplaceAll(history, "Online Single Submission", `<em>Online Single Submission</em>`)
		b.WriteString(`<p>Menindaklanjuti surat dari PT ATN kepada ` + e(in.PreviousRecipient) + ` Nomor: <strong>` + previousNumber + `</strong> tanggal <strong>` + previousDate + `</strong> perihal ` + e(strings.TrimRight(in.PreviousSubject, ". \t\r\n")) + `. ` + history + `. Oleh sebab itu, sebagai tindak lanjutnya bersama surat ini <strong>PT ATN</strong> mengajukan permohonan untuk dilakukan verifikasi dan persetujuan atas ` + requestedLicense + ` untuk lokasi-lokasi sebagai berikut:</p>`)
	} else {
		b.WriteString(`<p>Sehubungan dengan penyelenggaraan layanan telekomunikasi PT ATN, bersama surat ini <strong>` + e(popLetterCompanyName) + `</strong> mengajukan permohonan untuk dilakukan verifikasi dan persetujuan atas ` + requestedLicense + ` di luar komitmen, untuk lokasi Point of Presence (POP) sebagai berikut:</p>`)
	}
	if in.Profile.NIB != "" || in.License != "" {
		b.WriteString(`<div class="identity">`)
		if in.Profile.NIB != "" {
			b.WriteString(`NIB: ` + e(in.Profile.NIB) + `<br>`)
		}
		if in.License != "" {
			b.WriteString(`Referensi perizinan: ` + e(in.License))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`<ol>`)
	for i, p := range in.Locations {
		ending := ";"
		if i == len(in.Locations)-1 {
			ending = "."
		}
		name := ""
		if p.Name != "" {
			name = e(p.Name) + ", "
		}
		b.WriteString(`<li>` + name + lines(strings.TrimRight(p.Address, ".; \t\r\n")) + ending + `</li>`)
	}
	b.WriteString(`</ol>`)
	if in.Note != "" {
		b.WriteString(`<p>` + lines(in.Note) + `</p>`)
	}
	b.WriteString(`<div class="closing"><p>Demikian surat ini kami sampaikan. Atas perhatian dan kerja sama yang baik kami ucapkan terima kasih.</p><div class="sign">Hormat kami,<br><strong>` + e(popCompany) + `</strong><img class="signature" alt="Tanda tangan dan stempel perusahaan" src="data:image/png;base64,` + base64.StdEncoding.EncodeToString(popSignature) + `"><span class="signer">` + e(in.Profile.Signer) + `</span><br>` + e(in.Profile.Job) + `</div></div></main></body></html>`)
	return b.String()
}
