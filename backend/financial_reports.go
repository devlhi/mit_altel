package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Freeze partner attribution when an invoice is created. Legacy invoices are
// labelled because their original partner assignment cannot be reconstructed.
func migratePartnerRevenue(db *sql.DB) error {
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS bill_partner (bill_id INTEGER PRIMARY KEY, mitra_id INTEGER NOT NULL, name TEXT NOT NULL, share_bps INTEGER NOT NULL, legacy INTEGER NOT NULL DEFAULT 0)`,
		`INSERT OR IGNORE INTO bill_partner(bill_id,mitra_id,name,share_bps,legacy) SELECT g.id,COALESCE(m.id,0),COALESCE(m.nama,'Tanpa mitra'),CAST(ROUND(COALESCE(m.share_mitra,0)*100) AS INTEGER),1 FROM tagihan g LEFT JOIN pelanggan p ON p.id=g.pelanggan_id LEFT JOIN mitra m ON m.id=p.mitra_id`,
		`CREATE TRIGGER IF NOT EXISTS bill_partner_capture AFTER INSERT ON tagihan BEGIN INSERT INTO bill_partner(bill_id,mitra_id,name,share_bps) SELECT NEW.id,COALESCE(m.id,0),COALESCE(m.nama,'Tanpa mitra'),CAST(ROUND(COALESCE(m.share_mitra,0)*100) AS INTEGER) FROM pelanggan p LEFT JOIN mitra m ON m.id=p.mitra_id WHERE p.id=NEW.pelanggan_id; END`,
		`CREATE TRIGGER IF NOT EXISTS bill_partner_cleanup AFTER DELETE ON tagihan BEGIN DELETE FROM bill_partner WHERE bill_id=OLD.id; END`,
	} {
		if _, e := db.Exec(q); e != nil {
			return e
		}
	}
	return nil
}

type PartnerRevenue struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Bills  int     `json:"bills"`
	Gross  int64   `json:"gross"`
	Months []int64 `json:"months"`
	Legacy int     `json:"legacy"`
}
type RevenueBill struct {
	ID        int64  `json:"id"`
	PartnerID int64  `json:"partner_id"`
	Partner   string `json:"partner"`
	Customer  string `json:"customer"`
	Period    string `json:"period"`
	Date      string `json:"date"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	Legacy    bool   `json:"legacy"`
}

func partnerRevenue(db *sql.DB, year int, basis string) (int64, []int64, []PartnerRevenue, []RevenueBill, error) {
	months := make([]int64, 12)
	partners := []PartnerRevenue{}
	bills := []RevenueBill{}
	if basis != "billed" && basis != "paid" {
		return 0, months, partners, bills, fmt.Errorf("Pilih basis total tagihan atau pembayaran diterima")
	}
	rows, e := db.Query(`SELECT g.id,COALESCE(bp.mitra_id,0),COALESCE(bp.name,'Tanpa mitra'),COALESCE(p.nama,'Pelanggan dihapus'),g.periode,g.jumlah,g.status,g.tanggal_bayar,COALESCE(bp.legacy,1) FROM tagihan g LEFT JOIN bill_partner bp ON bp.bill_id=g.id LEFT JOIN pelanggan p ON p.id=g.pelanggan_id ORDER BY g.periode,g.id`)
	if e != nil {
		return 0, months, partners, bills, e
	}
	defer rows.Close()
	grouped := map[int64]*PartnerRevenue{}
	var gross int64
	prefix := strconv.Itoa(year) + "-"
	for rows.Next() {
		var b RevenueBill
		var paid string
		if e = rows.Scan(&b.ID, &b.PartnerID, &b.Partner, &b.Customer, &b.Period, &b.Amount, &b.Status, &paid, &b.Legacy); e != nil {
			return 0, months, partners, bills, e
		}
		b.Date = b.Period + "-01"
		if basis == "paid" {
			if b.Status != "Lunas" {
				continue
			}
			b.Date = paid
		}
		if !strings.HasPrefix(b.Date, prefix) {
			continue
		}
		if !validDate(b.Date) || b.Amount < 0 || b.Amount > maxMoney {
			return 0, months, partners, bills, fmt.Errorf("Tanggal atau nominal tagihan #%d tidak valid", b.ID)
		}
		month, _ := strconv.Atoi(b.Date[5:7])
		gross += b.Amount
		months[month-1] += b.Amount
		p := grouped[b.PartnerID]
		if p == nil {
			p = &PartnerRevenue{ID: b.PartnerID, Name: b.Partner, Months: make([]int64, 12)}
			grouped[b.PartnerID] = p
		}
		p.Gross += b.Amount
		p.Months[month-1] += b.Amount
		p.Bills++
		if b.Legacy {
			p.Legacy++
		}
		bills = append(bills, b)
	}
	for _, p := range grouped {
		partners = append(partners, *p)
	}
	sort.Slice(partners, func(i, j int) bool { return partners[i].Name < partners[j].Name })
	return gross, months, partners, bills, rows.Err()
}

func priorMonthEnd(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1).Format("2006-01-02")
}
func balanceSnapshot(entries []Journal, date string) map[string]any {
	a, t := reportBalances(entries, date[:4]+"-01-01", date)
	return map[string]any{"date": date, "accounts": a, "totals": t, "current_profit": t["profit"], "prior_profit": t["retained"] - t["profit"]}
}

type AccountMovement struct {
	ID          int64  `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Reference   string `json:"reference"`
	Account     string `json:"account"`
	Debit       int64  `json:"debit"`
	Credit      int64  `json:"credit"`
	Balance     int64  `json:"balance"`
}

func accountMovements(entries []Journal, code, date string) ([]AccountMovement, int64, error) {
	a, ok := account(code)
	profit := code == "profit_current" || code == "profit_prior"
	if !ok && !profit {
		return nil, 0, fmt.Errorf("Akun tidak ditemukan")
	}
	copyEntries := append([]Journal(nil), entries...)
	sort.Slice(copyEntries, func(i, j int) bool {
		if copyEntries[i].Date == copyEntries[j].Date {
			return copyEntries[i].ID < copyEntries[j].ID
		}
		return copyEntries[i].Date < copyEntries[j].Date
	})
	out := []AccountMovement{}
	var balance int64
	for _, j := range copyEntries {
		if j.Void || j.Date > date {
			continue
		}
		if code == "profit_current" && j.Date < date[:4]+"-01-01" {
			continue
		}
		if code == "profit_prior" && j.Date >= date[:4]+"-01-01" {
			continue
		}
		for _, l := range j.Lines {
			la, exists := account(l.Account)
			if !exists {
				continue
			}
			if profit {
				if la.Kind != "income" && la.Kind != "expense" {
					continue
				}
			} else if l.Account != code {
				continue
			}
			n := l.Debit - l.Credit
			if profit || a.Kind == "liability" || a.Kind == "equity" || a.Kind == "income" {
				n = -n
			}
			balance += n
			out = append(out, AccountMovement{j.ID, j.Date, j.Description, j.Reference, l.Account, l.Debit, l.Credit, balance})
		}
	}
	return out, balance, nil
}
func (h *Handlers) BookAccount(w http.ResponseWriter, r *http.Request) {
	code, date := r.URL.Query().Get("account"), r.URL.Query().Get("to")
	if !validDate(date) {
		httpErr(w, 400, "Tanggal posisi tidak valid")
		return
	}
	entries, e := loadJournals(h.DB)
	if e != nil {
		httpErr(w, 500, "Gagal membaca jurnal")
		return
	}
	rows, balance, e := accountMovements(entries, code, date)
	if e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"rows": rows, "balance": balance, "account": code, "to": date})
}
