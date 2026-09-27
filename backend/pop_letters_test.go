package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func examplePOPLetter() POPLetterInput {
	return POPLetterInput{RequestKey: "pop-letter-test-00001", Date: "2026-09-27", Mode: "initial", ServiceID: "isp", KBLIVersion: "2025", Profile: POPLetterProfile{City: "Gorontalo", Signer: "Penandatangan Uji", Job: "Direktur", Recipient: "Direktur Layanan Ekosistem Digital", RecipientAddress: "Jakarta"}, Locations: []POPLetterLocation{{Name: "POP Uji", Address: "Alamat uji, Gorontalo"}}}
}
func TestPOPLetterValidationAndPreview(t *testing.T) {
	db, mux := testOperations(t)
	in := examplePOPLetter()
	for _, change := range []func(*POPLetterInput){
		func(v *POPLetterInput) { v.Date = "2026-02-30" },
		func(v *POPLetterInput) { v.ServiceID = "" },
		func(v *POPLetterInput) { v.KBLIVersion = "2020" },
		func(v *POPLetterInput) { v.Mode = "followup"; v.PreviousDate = "2026-09-28"; v.PreviousNumber = "LAMA" },
		func(v *POPLetterInput) { v.Locations = nil },
		func(v *POPLetterInput) { v.Locations = []POPLetterLocation{{Name: "X"}} },
		func(v *POPLetterInput) { v.Locations = append(v.Locations, v.Locations[0]) },
		func(v *POPLetterInput) { v.Profile.Signer = "" },
		func(v *POPLetterInput) { v.Profile.NIB = "not-a-nib" },
		func(v *POPLetterInput) { v.ServiceID = "manual"; v.KBLI = "123"; v.Service = "Layanan" },
	} {
		bad := examplePOPLetter()
		change(&bad)
		opRequest(t, mux, "POST", "/api/surat-pop/preview", bad, 400)
		opRequest(t, mux, "POST", "/api/surat-pop", bad, 400)
	}
	in.KBLI = "99999"
	in.Service = "Wrong client label"
	in.Note = `<script>alert(1)</script>`
	preview := opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
	var result POPLetterPreview
	json.Unmarshal(preview.Body.Bytes(), &result)
	for _, want := range []string{"DRAF", "KBLI 61104", "&lt;script&gt;", "data:image/png;base64,", "Penandatangan Uji"} {
		if !strings.Contains(result.HTML, want) {
			t.Fatalf("Missing %s", want)
		}
	}
	if strings.Contains(result.HTML, "KBLI 99999") || strings.Contains(result.HTML, "<script>") || strings.Contains(result.HTML, "Menindaklanjuti surat") {
		t.Fatal("Preview contains client override, unsafe HTML or invented history")
	}
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM pop_letters`).Scan(&n)
	if n != 0 {
		t.Fatal("Preview created archive")
	}
	db.QueryRow(`SELECT COUNT(*) FROM pop_letter_sequence`).Scan(&n)
	if n != 0 {
		t.Fatal("Preview consumed sequence")
	}
	manual := examplePOPLetter()
	manual.ServiceID = "manual"
	manual.KBLI = "61101"
	manual.Service = "Jaringan kabel sesuai NIB"
	manual.KBLIDescription = "Aktivitas Telekomunikasi dengan Kabel"
	manual.Mode = "followup"
	manual.PreviousNumber = "005/ATN/2026"
	manual.PreviousDate = "2026-09-01"
	manual.ProcessHistory = "Dokumen telah disampaikan untuk pemeriksaan awal."
	opRequest(t, mux, "POST", "/api/surat-pop/preview", manual, 200)
}
func TestPOPLetterIssuanceSnapshotNumberAndVoid(t *testing.T) {
	db, mux := testOperations(t)
	in := examplePOPLetter()
	res := opRequest(t, mux, "POST", "/api/surat-pop", in, 201)
	var record POPLetterRecord
	json.Unmarshal(res.Body.Bytes(), &record)
	if record.Number != "001/POP/ATN-KOMDIGI/IX/2026" || record.Data.KBLI != "61104" {
		t.Fatalf("Bad record %+v", record)
	}
	opRequest(t, mux, "POST", "/api/surat-pop", in, 200)
	changed := in
	changed.Note = "Different body"
	opRequest(t, mux, "POST", "/api/surat-pop", changed, 409)
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM pop_letters`).Scan(&n)
	if n != 1 {
		t.Fatalf("Duplicate issuance: %d", n)
	}
	before := opRequest(t, mux, "GET", "/api/surat-pop/1/html", nil, 200).Body.String()
	p := in.Profile
	p.Signer = "New signer"
	opRequest(t, mux, "PUT", "/api/surat-pop/profile", p, 200)
	after := opRequest(t, mux, "GET", "/api/surat-pop/1/html", nil, 200).Body.String()
	if before != after {
		t.Fatal("Issued HTML changed after profile update")
	}
	in.RequestKey = "pop-letter-test-00002"
	in.ManualNumber = record.Number
	opRequest(t, mux, "POST", "/api/surat-pop", in, 409)
	in.ManualNumber = ""
	in.ServiceID = "local_fiber"
	res = opRequest(t, mux, "POST", "/api/surat-pop", in, 201)
	json.Unmarshal(res.Body.Bytes(), &record)
	if record.Number != "002/POP/ATN-KOMDIGI/IX/2026" || record.Data.KBLI != "61101" {
		t.Fatal("Sequence or KBLI incorrect")
	}
	opRequest(t, mux, "POST", "/api/surat-pop/1/void", map[string]string{"reason": "Salah lokasi"}, 200)
	void := opRequest(t, mux, "GET", "/api/surat-pop/1/html", nil, 200).Body.String()
	if !strings.Contains(void, "DIBATALKAN") || !strings.Contains(void, "Salah lokasi") {
		t.Fatal("Void status missing")
	}
	opRequest(t, mux, "POST", "/api/surat-pop/1/void", map[string]string{"reason": "Again"}, 409)
	opRequest(t, mux, "GET", "/api/surat-pop/999/html", nil, 404)
	opRequest(t, mux, "POST", "/api/seed", nil, 409)
}
func TestPOPLetterPrivateAssetsAndMultiLocation(t *testing.T) {
	db, mux := testOperations(t)
	// Production middleware must protect both the image-bearing preview and archived HTML.
	handler := AuthMiddleware(mux, db)
	for _, path := range []string{"/api/surat-pop/options", "/api/surat-pop/sample", "/api/surat-pop/1/html", "/api/surat-pop"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 401 {
			t.Fatalf("Anonymous %s returned %d", path, w.Code)
		}
	}
	in := examplePOPLetter()
	in.Locations = []POPLetterLocation{{Name: "POP A", Address: "Address A"}, {Name: "POP B", Address: "Address B"}}
	validatePOPLetter(&in)
	html := buildPOPLetterHTML(in, "A/2026", false)
	if strings.Count(html, "<li>") != 2 || !strings.Contains(html, "break-inside:avoid") || strings.Contains(html, ">DRAF</div>") {
		t.Fatal("Invalid print output")
	}
}

func TestPOPLetterCompleteFollowupAndRealHistory(t *testing.T) {
	_, mux := testOperations(t)
	in := examplePOPLetter()
	in.Mode = "followup"
	in.ServiceID = "local_fiber"
	in.PreviousNumber = "012/ATN/IX/2026"
	in.PreviousDate = "2026-09-15"
	// A previous letter alone must not assert that a joint OSS verification happened.
	opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
	opRequest(t, mux, "POST", "/api/surat-pop", in, 400)
	in.Profile.RecipientName = "Bapak Penerima Uji"
	in.PreviousRecipient = "Plt. Direktur Layanan Ekosistem Digital Direktorat Jenderal Ekosistem Digital Kementerian Komunikasi dan Digital RI"
	in.PreviousSubject = popLetterSubject
	in.ProcessHistory = popOSSHistory
	in.KBLIDescription = "Incorrect client classification"
	r := opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
	var body POPLetterPreview
	json.Unmarshal(r.Body.Bytes(), &body)
	for _, want := range []string{"Lanjutan " + popLetterSubject, "Bapak Penerima Uji", "Menindaklanjuti surat dari PT ATN kepada " + in.PreviousRecipient, "<strong>012/ATN/IX/2026</strong>", "<strong>15 September 2026</strong>", "perihal " + popLetterSubject, "Proses verifikasi dan persetujuan tersebut telah dicoba dilakukan bersama", "<em>Online Single Submission</em> (OSS)", "Oleh sebab itu, sebagai tindak lanjutnya bersama surat ini", "Izin Penyelenggaraan Jaringan Tetap Lokal Berbasis Packet Switched melalui Media Fiber Optik", "KBLI 61101 – Aktivitas Telekomunikasi dengan Kabel", "Demikian surat ini kami sampaikan"} {
		if !strings.Contains(body.HTML, want) {
			t.Fatalf("Missing full reference content: %s", want)
		}
	}
	if strings.Contains(body.HTML, "Incorrect client classification") {
		t.Fatal("Trusted client KBLI description for a known service")
	}
	in.ProcessHistory = `Dokumen <script>alert(1)</script> sudah disampaikan.`
	in.PreviousSubject = `<img src=x onerror=alert(1)>`
	r = opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
	json.Unmarshal(r.Body.Bytes(), &body)
	if strings.Contains(body.HTML, "<script>") || strings.Contains(body.HTML, "<img src=x") || !strings.Contains(body.HTML, "&lt;script&gt;") {
		t.Fatal("History or prior subject not escaped")
	}
	in.Mode = "initial"
	r = opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
	json.Unmarshal(r.Body.Bytes(), &body)
	if strings.Contains(body.HTML, "Menindaklanjuti") || strings.Contains(body.HTML, "sudah disampaikan") || strings.Contains(body.HTML, "012/ATN/IX/2026") {
		t.Fatal("Stale follow-up details leaked into initial letter")
	}
}

func TestPOPLetterReferenceSampleDoesNotIssue(t *testing.T) {
	db, mux := testOperations(t)
	r := opRequest(t, mux, "GET", "/api/surat-pop/sample", nil, 200)
	var body map[string]string
	json.Unmarshal(r.Body.Bytes(), &body)
	for _, want := range []string{">CONTOH</div>", "[nomor surat sebelumnya]", "[tanggal surat sebelumnya]", "[Nama pejabat penerima]", "KBLI 61101 – Aktivitas Telekomunikasi dengan Kabel", "Online Single Submission"} {
		if !strings.Contains(body["html"], want) {
			t.Fatalf("Sample missing %s", want)
		}
	}
	if strings.Count(body["html"], "<li>") != 3 {
		t.Fatal("Sample must demonstrate multiple POP locations")
	}
	for _, old := range []string{"Integrasi Jaringan Ekosistem", "132/SK/IJE", "Aryo Pamoragung", "29 Juni 2026", "Purwokerto"} {
		if strings.Contains(body["html"], old) {
			t.Fatalf("Copied another company's fact: %s", old)
		}
	}
	r = opRequest(t, mux, "GET", "/api/surat-pop/sample?service=isp", nil, 200)
	json.Unmarshal(r.Body.Bytes(), &body)
	if !strings.Contains(body["html"], "KBLI 61104 – Aktivitas Jasa Akses Internet") {
		t.Fatal("ISP sample did not use correct classification")
	}
	var n int
	db.QueryRow(`SELECT (SELECT COUNT(*) FROM pop_letters)+(SELECT COUNT(*) FROM pop_letter_sequence)`).Scan(&n)
	if n != 0 {
		t.Fatal("Sample created an archive or reserved a number")
	}
}

func TestPOPLetterIncompleteFollowupPreview(t *testing.T) {
	db, mux := testOperations(t)
	in := examplePOPLetter()
	in.Mode = "followup"
	in.PreviousNumber = "   "
	in.ProcessHistory = "   "
	readPreview := func(in POPLetterInput) POPLetterPreview {
		t.Helper()
		r := opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
		var p POPLetterPreview
		if err := json.Unmarshal(r.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	p := readPreview(in)
	if p.CanIssue || len(p.Missing) != 3 {
		t.Fatalf("Incomplete draft can issue: %+v", p.Missing)
	}
	for _, want := range []string{"Lanjutan " + popLetterSubject, "Menindaklanjuti surat", "[nomor surat sebelumnya]", "[tanggal surat sebelumnya]", "[Contoh riwayat — sesuaikan:", "Online Single Submission", "KBLI 61104", "POP Uji, Alamat uji, Gorontalo", `class="draft-field"`} {
		if !strings.Contains(p.HTML, want) {
			t.Fatalf("Missing draft content: %s", want)
		}
	}
	opRequest(t, mux, "POST", "/api/surat-pop", in, 400)
	for _, date := range []string{"invalid", "2026-02-30", "2026-09-28"} {
		bad := in
		bad.PreviousDate = date
		opRequest(t, mux, "POST", "/api/surat-pop/preview", bad, 400)
	}
	in.PreviousNumber = "011/ATN/IX/2026"
	in.PreviousDate = "2026-09-01"
	p = readPreview(in)
	if p.CanIssue || len(p.Missing) != 1 {
		t.Fatal("Missing history must block issuance")
	}
	if strings.Contains(p.HTML, "[nomor surat sebelumnya]") || !strings.Contains(p.HTML, "1 September 2026") {
		t.Fatal("Draft did not keep entered reference")
	}
	opRequest(t, mux, "POST", "/api/surat-pop", in, 400)
	in.ProcessHistory = "Dokumen sedang diperiksa oleh tim perusahaan."
	p = readPreview(in)
	if !p.CanIssue || len(p.Missing) != 0 || strings.Contains(p.HTML, `class="draft-field"`) || strings.Contains(p.HTML, "Online Single Submission") {
		t.Fatal("Completed draft still has suggested OSS history or missing fields")
	}
	var n int
	if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM pop_letters)+(SELECT COUNT(*) FROM pop_letter_sequence)`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("Preview wrote an archive or sequence: %d %v", n, err)
	}
	// Issuance uses exactly the real history, never the preview-only suggestion.
	r := opRequest(t, mux, "POST", "/api/surat-pop", in, 201)
	var record POPLetterRecord
	if err := json.Unmarshal(r.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Data.ProcessHistory != in.ProcessHistory || record.Data.PreviousNumber != in.PreviousNumber {
		t.Fatal("Preview placeholders entered the archive")
	}
	body := opRequest(t, mux, "GET", "/api/surat-pop/1/html", nil, 200).Body.String()
	if strings.Contains(body, "draft-field") || strings.Contains(body, "Contoh riwayat") || strings.Contains(body, ">DRAF</div>") {
		t.Fatal("Draft-only content entered issued letter")
	}
}

func TestPOPLetterThreeFieldTemplate(t *testing.T) {
	db, mux := testOperations(t)
	opRequest(t, mux, "PUT", "/api/surat-pop/profile", examplePOPLetter().Profile, 200)
	in := POPLetterInput{Mode: "template", ManualNumber: "017/POP/ATN/IX/2026", ServiceID: "local_fiber", RequestKey: "simple-letter-test-0001", Locations: []POPLetterLocation{{Address: "Jl. Uji, Desa Tuladenggi, Telaga Biru, Gorontalo"}}}
	r := opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
	var preview POPLetterPreview
	if err := json.Unmarshal(r.Body.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	if !preview.CanIssue || len(preview.Missing) != 0 || preview.Data == nil {
		t.Fatal("Three fields did not produce an issuable preview")
	}
	p := preview.Data
	if p.Date != time.Now().In(time.FixedZone("WITA", 8*60*60)).Format("2006-01-02") || p.Profile.Signer == "" || p.Profile.Recipient == "" || p.ProcessHistory != popOSSHistory {
		t.Fatal("Automatic date, profile or body missing")
	}
	if p.PreviousNumber != "" || p.PreviousDate != "" {
		t.Fatal("Invented previous letter reference")
	}
	for _, want := range []string{"No.: " + in.ManualNumber, "Lanjutan " + popLetterSubject, "Menindaklanjuti permohonan dari PT ATN", "yang mana proses verifikasi dan persetujuan", "Online Single Submission", "KBLI 61101 – Aktivitas Telekomunikasi dengan Kabel", "<li>" + in.Locations[0].Address + ".</li>", "Tanda tangan dan stempel perusahaan"} {
		if !strings.Contains(preview.HTML, want) {
			t.Fatalf("Template missing: %s", want)
		}
	}
	for _, unwanted := range []string{"class=\"draft-field\"", "[nomor surat sebelumnya]", "Contoh riwayat", "132/SK/IJE"} {
		if strings.Contains(preview.HTML, unwanted) {
			t.Fatalf("Unexpected placeholder/reference: %s", unwanted)
		}
	}
	// A profile edit after preview cannot silently change the signed letter or retry hash.
	updated := p.Profile
	updated.Signer = "Nama lain setelah pratinjau"
	opRequest(t, mux, "PUT", "/api/surat-pop/profile", updated, 200)
	r = opRequest(t, mux, "POST", "/api/surat-pop", p, 201)
	var issued POPLetterRecord
	if err := json.Unmarshal(r.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}
	if issued.Number != in.ManualNumber || issued.Data.Profile.Signer != p.Profile.Signer || issued.Data.Date != p.Date {
		t.Fatal("Issued letter differs from preview snapshot")
	}
	opRequest(t, mux, "POST", "/api/surat-pop", p, 200)
	other := *p
	other.RequestKey = "simple-letter-test-0002"
	opRequest(t, mux, "POST", "/api/surat-pop", other, 409)
	var count, sequence int
	db.QueryRow(`SELECT COUNT(*) FROM pop_letters`).Scan(&count)
	db.QueryRow(`SELECT COUNT(*) FROM pop_letter_sequence`).Scan(&sequence)
	if count != 1 || sequence != 0 {
		t.Fatal("Manual numbering made a duplicate or consumed sequence")
	}
}

func TestPOPLetterTemplateValidationAndManualKBLI(t *testing.T) {
	_, mux := testOperations(t)
	opRequest(t, mux, "PUT", "/api/surat-pop/profile", examplePOPLetter().Profile, 200)
	fresh := func() POPLetterInput {
		return POPLetterInput{Mode: "template", ManualNumber: "UJI/2026", ServiceID: "isp", RequestKey: "simple-letter-test-0003", Locations: []POPLetterLocation{{Address: "Alamat POP uji"}}}
	}
	for _, change := range []func(*POPLetterInput){
		func(p *POPLetterInput) { p.ManualNumber = "  " },
		func(p *POPLetterInput) { p.ServiceID = "" },
		func(p *POPLetterInput) { p.ServiceID = "manual"; p.KBLI = "x123" },
		func(p *POPLetterInput) { p.Locations = nil },
		func(p *POPLetterInput) { p.Locations[0].Address = " " },
		func(p *POPLetterInput) {
			p.Locations = append(p.Locations, POPLetterLocation{Name: "Different name", Address: p.Locations[0].Address})
		},
	} {
		in := fresh()
		change(&in)
		opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 400)
		opRequest(t, mux, "POST", "/api/surat-pop", in, 400)
	}
	for _, code := range []string{"61104", "61101", "61999"} {
		in := fresh()
		in.ServiceID = "manual"
		in.KBLI = code
		in.Locations[0].Address = `<script>alert(1)</script>`
		r := opRequest(t, mux, "POST", "/api/surat-pop/preview", in, 200)
		var p POPLetterPreview
		if err := json.Unmarshal(r.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		if !p.CanIssue || !strings.Contains(p.HTML, "KBLI "+code) || strings.Contains(p.HTML, "<script>") || !strings.Contains(p.HTML, "&lt;script&gt;") {
			t.Fatal("Invalid manual KBLI output")
		}
		if code == "61999" && (p.Data.Service != "Penyelenggaraan Telekomunikasi" || p.Data.KBLIDescription != "") {
			t.Fatal("Invented classification for unknown code")
		}
		if code != "61999" && p.Data.KBLIDescription == "" {
			t.Fatal("Known manual code did not resolve official name")
		}
	}
}
