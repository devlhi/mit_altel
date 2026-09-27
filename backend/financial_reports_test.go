package main

import (
	"encoding/json"
	"testing"
)

func TestPartnerRevenueUsesStoredBillsAndFrozenAttribution(t *testing.T) {
	db, mux := testOperations(t)
	execTest(t, db, `DELETE FROM tagihan`)
	var customer, partner int64
	var name string
	if e := db.QueryRow(`SELECT p.id,m.id,m.nama FROM pelanggan p JOIN mitra m ON m.id=p.mitra_id LIMIT 1`).Scan(&customer, &partner, &name); e != nil {
		t.Fatal(e)
	}
	execTest(t, db, `INSERT INTO tagihan(pelanggan_id,periode,jumlah,status,metode,tanggal_bayar) VALUES(?, '2026-12', 100001,'Lunas','Transfer','2027-01-02'),(?,'2026-11',200003,'Belum','','')`, customer, customer)
	execTest(t, db, `UPDATE pelanggan SET mitra_id=0 WHERE id=?`, customer)
	execTest(t, db, `UPDATE paket SET harga=999999`)
	gross, months, partners, bills, e := partnerRevenue(db, 2026, "billed")
	if e != nil || gross != 300004 || months[10] != 200003 || months[11] != 100001 || len(partners) != 1 || partners[0].ID != partner || partners[0].Name != name || len(bills) != 2 || bills[0].Legacy {
		t.Fatalf("unexpected source: %d %v %+v %+v %v", gross, months, partners, bills, e)
	}
	paid, _, _, _, e := partnerRevenue(db, 2026, "paid")
	if e != nil || paid != 0 {
		t.Fatalf("receipt counted in invoice year: %d %v", paid, e)
	}
	paid, _, _, _, e = partnerRevenue(db, 2027, "paid")
	if e != nil || paid != 100001 {
		t.Fatalf("receipt omitted: %d %v", paid, e)
	}
	// Ledger entries and FAB never add to the partner revenue source.
	execTest(t, db, `INSERT INTO ledger(date,description,flow,lines) VALUES('2026-12-01','Unrelated ledger','operating',?)`, jsonText([]JournalLine{{"1102", 9999999, 0}, {"4101", 0, 9999999}}))
	w := opRequest(t, mux, "GET", "/api/bhp-uso?year=2026", nil, 200)
	var report struct {
		Gross, BHP, USO int64
		Config          Levy
	}
	if e = json.Unmarshal(w.Body.Bytes(), &report); e != nil {
		t.Fatal(e)
	}
	if report.Gross != 300004 || report.BHP != 1500 || report.USO != 3750 || report.Config.Basis != "billed" {
		t.Fatalf("wrong levy source: %+v", report)
	}
	opRequest(t, mux, "PUT", "/api/bhp-uso", Levy{Year: 2027, BHP: 50, USO: 125, Basis: "paid"}, 200)
	opRequest(t, mux, "GET", "/api/bhp-uso?year=2026&basis=untrusted", nil, 400)
}

func TestBalanceSheetCutoffProfitAndDrilldown(t *testing.T) {
	j := []Journal{
		{ID: 1, Date: "2025-12-01", Lines: []JournalLine{{"1102", 1000, 0}, {"3101", 0, 1000}}},
		{ID: 2, Date: "2025-12-31", Lines: []JournalLine{{"1102", 400, 0}, {"4101", 0, 400}}},
		{ID: 3, Date: "2026-01-20", Lines: []JournalLine{{"5101", 70, 0}, {"1102", 0, 70}}},
		{ID: 4, Date: "2026-01-31", Lines: []JournalLine{{"1103", 800, 0}, {"4101", 0, 800}}},
		{ID: 5, Date: "2026-02-01", Lines: []JournalLine{{"1102", 999, 0}, {"4101", 0, 999}}},
		{ID: 6, Date: "2026-01-25", Void: true, Lines: []JournalLine{{"1102", 999, 0}, {"4101", 0, 999}}},
	}
	snap := balanceSnapshot(j, "2026-01-31")
	totals := snap["totals"].(map[string]int64)
	if snap["current_profit"].(int64) != 730 || snap["prior_profit"].(int64) != 400 || totals["difference"] != 0 || totals["asset"] != 2130 {
		t.Fatalf("wrong snapshot: %+v", snap)
	}
	for code, want := range map[string]int64{"1102": 1330, "1103": 800, "3101": 1000, "profit_prior": 400, "profit_current": 730} {
		_, balance, e := accountMovements(j, code, "2026-01-31")
		if e != nil || balance != want {
			t.Fatalf("%s detail %d want %d: %v", code, balance, want, e)
		}
	}
	for input, want := range map[string]string{"2026-01-31": "2025-12-31", "2024-03-15": "2024-02-29", "2026-03-31": "2026-02-28"} {
		if got := priorMonthEnd(input); got != want {
			t.Errorf("%s comparison %s", input, got)
		}
	}
}
