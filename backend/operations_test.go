package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func testOperations(t *testing.T) (*sql.DB, *http.ServeMux) {
	t.Helper()
	db, e := OpenDB(filepath.Join(t.TempDir(), "operations.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	mux := http.NewServeMux()
	RegisterRoutes(mux, db)
	return db, mux
}
func opRequest(t *testing.T, mux *http.ServeMux, method, url string, body any, status int) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(method, url, strings.NewReader(jsonText(body))))
	if w.Code != status {
		t.Fatalf("%s %s: expected %d, got %d: %s", method, url, status, w.Code, w.Body.String())
	}
	return w
}
func execTest(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, e := db.Exec(q, args...); e != nil {
		t.Fatal(e)
	}
}
func TestJournalBalancesAndCashFlow(t *testing.T) {
	entries := []Journal{
		{Date: "2026-01-01", Source: "opening", Flow: "opening", Lines: []JournalLine{{"1102", 1000000, 0}, {"3101", 0, 1000000}}},
		{Date: "2026-01-02", Flow: "operating", Lines: []JournalLine{{"1103", 200000, 0}, {"4101", 0, 200000}}},
		{Date: "2026-01-03", Flow: "operating", Lines: []JournalLine{{"1102", 150000, 0}, {"1103", 0, 150000}}},
		{Date: "2026-01-04", Flow: "operating", Lines: []JournalLine{{"5101", 50000, 0}, {"1102", 0, 50000}}},
		{Date: "2026-01-05", Flow: "investing", Lines: []JournalLine{{"1201", 100000, 0}, {"1102", 0, 100000}}},
		{Date: "2026-01-06", Flow: "operating", Lines: []JournalLine{{"1101", 10000, 0}, {"1102", 0, 10000}}},
		{Date: "2026-01-07", Void: true, Flow: "operating", Lines: []JournalLine{{"5199", 999999, 0}, {"1102", 0, 999999}}},
		{Date: "2026-02-01", Flow: "operating", Lines: []JournalLine{{"1102", 999999, 0}, {"4101", 0, 999999}}},
	}
	_, totals := reportBalances(entries, "2026-01-01", "2026-01-31")
	for key, want := range map[string]int64{"asset": 1150000, "equity": 1150000, "difference": 0, "profit": 150000, "cash_close": 1000000, "opening_cash": 1000000, "cash_open": 0, "operating": 100000, "investing": -100000} {
		if totals[key] != want {
			t.Errorf("%s = %d, want %d", key, totals[key], want)
		}
	}
	_, period := reportBalances(entries, "2026-01-03", "2026-01-31")
	if period["profit"] != -50000 || period["cash_open"] != 1000000 || period["retained"] != 150000 {
		t.Fatalf("period boundaries incorrect: %v", period)
	}
}
func TestOpeningAndJournalValidation(t *testing.T) {
	_, mux := testOperations(t)
	opRequest(t, mux, "POST", "/api/pembukuan/journal", Journal{Date: "2026-01-01", Description: "Before opening", Flow: "operating", Lines: []JournalLine{{"1102", 1, 0}, {"3101", 0, 1}}}, 400)
	for _, body := range []any{map[string]any{"date": "2026-01-01", "lines": []JournalLine{{"1102", 1000, 0}}}, map[string]any{"date": "2026-02-30", "zero": true}, map[string]any{"date": "2026-01-01"}, map[string]any{"date": "2026-01-01", "lines": []JournalLine{{"1102", 1000, 0}, {"4101", 0, 1000}}}} {
		opRequest(t, mux, "POST", "/api/pembukuan/opening", body, 400)
	}
	opRequest(t, mux, "POST", "/api/pembukuan/opening", map[string]any{"date": "2026-01-01", "zero": true}, 201)
	opRequest(t, mux, "POST", "/api/pembukuan/opening", map[string]any{"date": "2026-01-01", "zero": true}, 409)
	for _, j := range []Journal{
		{Date: "2025-12-31", Description: "Old", Flow: "operating", Lines: []JournalLine{{"1102", 1, 0}, {"3101", 0, 1}}},
		{Date: "2026-01-02", Description: "Unbalanced", Flow: "operating", Lines: []JournalLine{{"1102", 2, 0}, {"3101", 0, 1}}},
		{Date: "2026-01-02", Description: "Unknown", Flow: "operating", Lines: []JournalLine{{"9999", 1, 0}, {"3101", 0, 1}}},
		{Date: "2026-01-02", Description: "Negative", Flow: "operating", Lines: []JournalLine{{"1102", -1, 0}, {"3101", 0, -1}}},
		{Date: "2026-01-02", Description: "Both sides", Flow: "operating", Lines: []JournalLine{{"1102", 1, 1}}},
	} {
		opRequest(t, mux, "POST", "/api/pembukuan/journal", j, 400)
	}
	good := Journal{Date: "2026-01-03", Description: "Capital", Flow: "financing", Source: "opening", Void: true, Lines: []JournalLine{{"1102", 100000, 0}, {"3101", 0, 100000}}}
	result := opRequest(t, mux, "POST", "/api/pembukuan/journal", good, 201)
	var saved Journal
	json.Unmarshal(result.Body.Bytes(), &saved)
	if saved.Source != "" || saved.Void {
		t.Fatal("client must not forge system source or void state")
	}
	opRequest(t, mux, "POST", fmt.Sprintf("/api/pembukuan/journal/%d/void", saved.ID), map[string]string{"reason": "Koreksi pencatatan"}, 200)
	opRequest(t, mux, "POST", fmt.Sprintf("/api/pembukuan/journal/%d/void", saved.ID), map[string]string{"reason": "Koreksi pencatatan"}, 409)
	opRequest(t, mux, "POST", "/api/pembukuan/journal/1/void", map[string]string{"reason": "Hapus saldo awal"}, 409)
}
func TestBillSyncIdempotentAndReopenedPayment(t *testing.T) {
	db, mux := testOperations(t)
	execTest(t, db, "DELETE FROM tagihan")
	var customer int64
	db.QueryRow(`SELECT id FROM pelanggan LIMIT 1`).Scan(&customer)
	execTest(t, db, `INSERT INTO tagihan(id,pelanggan_id,periode,jumlah,status,metode,tanggal_bayar) VALUES(100,?,'2026-01',200000,'Lunas','Transfer','2026-01-08')`, customer)
	execTest(t, db, `INSERT INTO tagihan(id,pelanggan_id,periode,jumlah,status,metode,tanggal_bayar) VALUES(101,?,'2025-12',300000,'Lunas','Tunai','2026-01-09')`, customer)
	opRequest(t, mux, "POST", "/api/pembukuan/opening", map[string]any{"date": "2026-01-01", "lines": []JournalLine{{"1103", 300000, 0}, {"3101", 0, 300000}}}, 201)
	opRequest(t, mux, "POST", "/api/pembukuan/sync", nil, 200)
	existing, e := loadJournals(db)
	if e != nil {
		t.Fatal(e)
	}
	if len(existing) != 4 {
		t.Fatalf("expected opening, invoice, two receipts; got %d", len(existing))
	}
	second := opRequest(t, mux, "POST", "/api/pembukuan/sync", nil, 200)
	if !strings.Contains(second.Body.String(), `"changes":0`) {
		t.Fatal("repeated sync must be idempotent", second.Body.String())
	}
	_, totals := reportBalances(existing, "2026-01-01", "2026-01-31")
	if totals["income"] != 200000 || totals["cash_close"] != 500000 || totals["difference"] != 0 {
		t.Fatalf("opening receivable receipts must not duplicate revenue: %v", totals)
	}
	opRequest(t, mux, "DELETE", "/api/tagihan/100", nil, 409)
	opRequest(t, mux, "DELETE", fmt.Sprintf("/api/pelanggan/%d", customer), nil, 409)
	opRequest(t, mux, "PUT", "/api/tagihan/100", map[string]string{"status": "Belum"}, 200)
	desired, _ := desiredBills(db, "2026-01-01")
	if pendingBills(existing, desired) != 1 {
		t.Fatal("reopened bill must require synchronization")
	}
	opRequest(t, mux, "POST", "/api/pembukuan/sync", nil, 200)
	updated, _ := loadJournals(db)
	_, totals = reportBalances(updated, "2026-01-01", "2026-01-31")
	if totals["cash_close"] != 300000 || totals["income"] != 200000 {
		t.Fatalf("reopening should void receipt only: %v", totals)
	}
	if e = ResetSeed(db); e == nil {
		t.Fatal("reset must protect financial history")
	}
	opRequest(t, mux, "POST", "/api/seed", nil, 409)
}
func TestLevyDeductionsAndRevenueScope(t *testing.T) {
	l := Levy{Year: 2026, BHP: 50, USO: 125, BadDebt: 100000, Interconnect: 200000, Evidence: "Bukti A"}
	base, e := levyBase(10000000, l)
	if e != nil || base != 9700000 {
		t.Fatal(base, e)
	}
	if levyAmount(base, l.BHP) != 48500 || levyAmount(base, l.USO) != 121250 {
		t.Fatal("incorrect statutory default estimates")
	}
	if levyAmount(100, 50) != 1 {
		t.Fatal("half rupiah rounds up")
	}
	l.Evidence = ""
	if _, e = levyBase(10000000, l); e == nil {
		t.Fatal("deductions need evidence")
	}
	l.Evidence = "A"
	if _, e = levyBase(1, l); e == nil {
		t.Fatal("negative base rejected")
	}
	l.BHP = 100
	l.BadDebt = 0
	l.Interconnect = 0
	if _, e = levyBase(10000000, l); e == nil {
		t.Fatal("rate override needs explanation")
	}
	entries := []Journal{{Date: "2026-02-01", Lines: []JournalLine{{"4101", 0, 100}, {"4102", 0, 50}, {"4201", 0, 999}}}, {Date: "2026-02-02", Void: true, Lines: []JournalLine{{"4101", 0, 500}}}, {Date: "2025-01-01", Lines: []JournalLine{{"4101", 0, 200}}}}
	gross, months := telecomRevenue(entries, 2026)
	if gross != 150 || months[1] != 150 {
		t.Fatal("nontelecom, void, and prior-year income must be excluded", gross)
	}
}
func TestPOPUniqueAssignmentsAndRollback(t *testing.T) {
	db, mux := testOperations(t)
	var customer int64
	db.QueryRow(`SELECT id FROM pelanggan LIMIT 1`).Scan(&customer)
	p := POP{Kode: "GTO-01", Nama: "POP Test", Status: "Aktif", Customers: []int64{customer}, Assets: []POPAsset{{Name: "Router", Value: 100000}}}
	w := opRequest(t, mux, "POST", "/api/pop", p, 200)
	var saved POP
	json.Unmarshal(w.Body.Bytes(), &saved)
	p.Kode = "GTO-02"
	opRequest(t, mux, "POST", "/api/pop", p, 409)
	p.Customers = []int64{999999}
	opRequest(t, mux, "POST", "/api/pop", p, 400)
	p.Customers = nil
	p.Assets[0].Value = -1
	opRequest(t, mux, "POST", "/api/pop", p, 400)
	saved.Status = "Nonaktif"
	saved.Customers = nil
	opRequest(t, mux, "PUT", fmt.Sprintf("/api/pop/%d", saved.ID), saved, 200)
	p.Assets[0].Value = 100000
	p.Customers = []int64{customer}
	opRequest(t, mux, "POST", "/api/pop", p, 200)
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM pops`).Scan(&n)
	if n != 2 {
		t.Fatal("failed saves must not insert POP")
	}
}
func TestFABStoredPricesAndCompanySnapshot(t *testing.T) {
	db, mux := testOperations(t)
	execTest(t, db, `DELETE FROM tagihan`)
	var customer, mitra, paket int64
	db.QueryRow(`SELECT id,mitra_id,paket_id FROM pelanggan LIMIT 1`).Scan(&customer, &mitra, &paket)
	execTest(t, db, `UPDATE mitra SET rekening='Bank Mitra 123' WHERE id=?`, mitra)
	execTest(t, db, `UPDATE settings SET nama_isp='Perusahaan Sebelum',rekening='Bank Perusahaan 999',tpl_fab=?`, defaultTplFAB)
	execTest(t, db, `INSERT INTO tagihan(pelanggan_id,periode,jumlah,status) VALUES(?,'2025-12',123456,'Lunas')`, customer)
	execTest(t, db, `UPDATE paket SET harga=999999 WHERE id=?`, paket)
	rc := computeFAB(db, mitra, "2025-12")
	if len(rc.Rows) != 1 || rc.Rows[0].Harga != 123456 || rc.Rows[0].Subtotal != 123456 {
		t.Fatal("FAB must use billed amount")
	}
	payload := map[string]any{"jenis": "FAB", "mitra_id": mitra, "tanggal": "2025-12-31", "periode": "2025-12"}
	w := opRequest(t, mux, "POST", "/api/dokumen", payload, 201)
	var d Dokumen
	json.Unmarshal(w.Body.Bytes(), &d)
	if !strings.Contains(d.Nomor, "2025") {
		t.Fatal("number must follow document date", d.Nomor)
	}
	opRequest(t, mux, "POST", "/api/dokumen", payload, 409)
	execTest(t, db, `UPDATE settings SET nama_isp='Perusahaan Sesudah',tpl_fab='<p>Template changed</p>'`)
	execTest(t, db, `UPDATE mitra SET rekening='Bank Berubah' WHERE id=?`, mitra)
	token, e := issueToken(db, 1)
	if e != nil {
		t.Fatal(e)
	}
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/dokumen/%d/html?mode=preview", d.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	preview := httptest.NewRecorder()
	mux.ServeHTTP(preview, req)
	html := preview.Body.String()
	for _, want := range []string{"Perusahaan Sebelum", "Bank Mitra 123", "data:image/png;base64,"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing snapshot %q", want)
		}
	}
	for _, bad := range []string{"Perusahaan Sesudah", "Bank Perusahaan 999", "Bank Berubah", "Template changed"} {
		if strings.Contains(html, bad) {
			t.Errorf("snapshot changed: %q", bad)
		}
	}
}

func TestPaymentDateAndTemplateMigration(t *testing.T) {
	db, mux := testOperations(t)
	var id int64
	db.QueryRow(`SELECT id FROM tagihan LIMIT 1`).Scan(&id)
	url := fmt.Sprintf("/api/tagihan/%d", id)
	opRequest(t, mux, "PUT", url, map[string]string{"status": "Lunas", "metode": "Tunai", "tanggal_bayar": "2026-01-10"}, 200)
	opRequest(t, mux, "PUT", url, map[string]string{"status": "Lunas"}, 200)
	var date, method string
	db.QueryRow(`SELECT tanggal_bayar,metode FROM tagihan WHERE id=?`, id).Scan(&date, &method)
	if date != "2026-01-10" || method != "Tunai" {
		t.Fatal("re-saving a paid bill must retain its payment date and method")
	}
	opRequest(t, mux, "PUT", url, map[string]string{"status": "Lunas", "tanggal_bayar": "2100-01-01"}, 400)
	execTest(t, db, `UPDATE settings SET tpl_fab=?`, legacyTplFAB)
	if e := migrateOperations(db); e != nil {
		t.Fatal(e)
	}
	if s := getSettings(db); s.TplFAB != defaultTplFAB {
		t.Fatal("unchanged legacy FAB must migrate")
	}
	custom := "<p>Template milik pengguna {{rekening}}</p>"
	execTest(t, db, `UPDATE settings SET tpl_fab=?`, custom)
	if e := migrateOperations(db); e != nil {
		t.Fatal(e)
	}
	if s := getSettings(db); s.TplFAB != custom {
		t.Fatal("custom templates must be preserved")
	}
}

func TestPartnerBankRoundTrip(t *testing.T) {
	db, mux := testOperations(t)
	m := Mitra{Nama: "Mitra Baru", ShareMitra: 25, Rekening: "Bank Uji 123 a.n. Mitra"}
	response := opRequest(t, mux, "POST", "/api/mitra", m, 201)
	var created Mitra
	if e := json.Unmarshal(response.Body.Bytes(), &created); e != nil {
		t.Fatal(e)
	}
	saved, e := getMitraRaw(db, created.ID)
	if e != nil || saved.Rekening != m.Rekening {
		t.Fatal("partner bank was not saved", e)
	}
	saved.Rekening = "Bank Uji 456 a.n. Mitra"
	opRequest(t, mux, "PUT", fmt.Sprintf("/api/mitra/%d", saved.ID), saved, 200)
	updated, e := getMitraRaw(db, saved.ID)
	if e != nil || updated.Rekening != saved.Rekening {
		t.Fatal("partner bank was not updated", e)
	}
	opRequest(t, mux, "POST", "/api/dokumen", map[string]any{"jenis": "PKS", "mitra_id": saved.ID, "tanggal": "2026-09-27", "masa_bulan": 12}, 201)
	opRequest(t, mux, "DELETE", fmt.Sprintf("/api/mitra/%d", saved.ID), nil, 409)
}
