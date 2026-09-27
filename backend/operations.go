package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Money is stored in whole rupiah; journal lines never use floating point.
const maxMoney int64 = 1_000_000_000_000

type Account struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Cash    bool   `json:"cash"`
	Telecom bool   `json:"telecom"`
}

var accounts = []Account{
	{"1101", "Kas", "asset", true, false}, {"1102", "Bank", "asset", true, false},
	{"1103", "Piutang pelanggan", "asset", false, false}, {"1104", "Uang muka / biaya dibayar di muka", "asset", false, false},
	{"1201", "Peralatan jaringan & aset tetap", "asset", false, false}, {"1202", "Akumulasi penyusutan", "asset", false, false},
	{"2101", "Utang usaha / mitra", "liability", false, false}, {"2102", "Utang BHP", "liability", false, false},
	{"2103", "Utang USO", "liability", false, false}, {"2104", "Pinjaman", "liability", false, false},
	{"3101", "Modal", "equity", false, false}, {"3102", "Saldo laba awal", "equity", false, false},
	{"4101", "Pendapatan internet", "income", false, true}, {"4102", "Pendapatan aktivasi telekomunikasi", "income", false, true},
	{"4201", "Pendapatan nontelekomunikasi", "income", false, false},
	{"5101", "Beban bandwidth", "expense", false, false}, {"5102", "Beban listrik", "expense", false, false},
	{"5103", "Beban sewa", "expense", false, false}, {"5104", "Beban gaji", "expense", false, false},
	{"5105", "Beban pemeliharaan", "expense", false, false}, {"5106", "Beban bagi hasil mitra", "expense", false, false},
	{"5107", "Beban BHP", "expense", false, false}, {"5108", "Beban USO", "expense", false, false},
	{"5109", "Beban penyusutan", "expense", false, false}, {"5199", "Beban operasional lainnya", "expense", false, false},
}

func account(code string) (Account, bool) {
	for _, a := range accounts {
		if a.Code == code {
			return a, true
		}
	}
	return Account{}, false
}

type JournalLine struct {
	Account string `json:"account"`
	Debit   int64  `json:"debit"`
	Credit  int64  `json:"credit"`
}
type Journal struct {
	ID          int64         `json:"id"`
	Date        string        `json:"date"`
	Description string        `json:"description"`
	Reference   string        `json:"reference"`
	Flow        string        `json:"flow"`
	PopID       int64         `json:"pop_id"`
	Source      string        `json:"source"`
	Void        bool          `json:"void"`
	Lines       []JournalLine `json:"lines"`
}
type Audit struct {
	ID     int64  `json:"id"`
	At     string `json:"at"`
	Action string `json:"action"`
	Detail string `json:"detail"`
}

func migrateOperations(db *sql.DB) error {
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS book_config (id INTEGER PRIMARY KEY CHECK(id=1), start_date TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ledger (id INTEGER PRIMARY KEY AUTOINCREMENT, source TEXT UNIQUE, date TEXT NOT NULL, description TEXT NOT NULL, reference TEXT NOT NULL DEFAULT '', flow TEXT NOT NULL, pop_id INTEGER NOT NULL DEFAULT 0, lines TEXT NOT NULL, void INTEGER NOT NULL DEFAULT 0)`,
		`CREATE INDEX IF NOT EXISTS ledger_date ON ledger(date)`,
		`CREATE TABLE IF NOT EXISTS operation_audit (id INTEGER PRIMARY KEY AUTOINCREMENT, at TEXT NOT NULL, action TEXT NOT NULL, detail TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS pops (id INTEGER PRIMARY KEY AUTOINCREMENT, kode TEXT NOT NULL UNIQUE, data TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS pop_customers (customer_id INTEGER PRIMARY KEY, pop_id INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS levy_config (year INTEGER PRIMARY KEY, data TEXT NOT NULL)`,
		`CREATE TRIGGER IF NOT EXISTS pop_customer_cleanup AFTER DELETE ON pelanggan BEGIN DELETE FROM pop_customers WHERE customer_id=OLD.id; END`,
	} {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	rows, err := db.Query(`PRAGMA table_info(mitra)`)
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, nn, pk int
		var name, typ string
		var def any
		if err = rows.Scan(&cid, &name, &typ, &nn, &def, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == "rekening" {
			found = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if !found {
		_, err = db.Exec(`ALTER TABLE mitra ADD COLUMN rekening TEXT NOT NULL DEFAULT ''`)
	}
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE settings SET tpl_fab=? WHERE tpl_fab IN (?,?)`, defaultTplFAB, legacyTplFAB, legacyOperationsTplFAB)
	return err
}
func jsonText(v any) string { b, _ := json.Marshal(v); return string(b) }
func operationID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id
}
func registerOperations(m *http.ServeMux, h *Handlers) {
	handle := func(pattern string, fn http.HandlerFunc) { m.HandleFunc(pattern, operationHandler(fn)) }
	handle("GET /api/pembukuan", h.BookReport)
	handle("GET /api/pembukuan/account", h.BookAccount)
	handle("POST /api/pembukuan/opening", h.BookOpening)
	handle("POST /api/pembukuan/journal", h.BookJournal)
	handle("POST /api/pembukuan/journal/{id}/void", h.BookVoid)
	handle("POST /api/pembukuan/sync", h.BookSync)
	handle("GET /api/bhp-uso", h.LevyReport)
	handle("PUT /api/bhp-uso", h.SaveLevy)
	handle("GET /api/pop", h.ListPOP)
	handle("POST /api/pop", h.SavePOP)
	handle("PUT /api/pop/{id}", h.SavePOP)
}
func validDate(v string) bool {
	t, e := time.Parse("2006-01-02", v)
	return e == nil && t.Format("2006-01-02") == v && t.Year() >= 2000 && t.Year() <= 2100
}
func audit(tx *sql.Tx, action string, detail any) error {
	_, e := tx.Exec(`INSERT INTO operation_audit(at,action,detail) VALUES(?,?,?)`, time.Now().UTC().Format(time.RFC3339), action, jsonText(detail))
	return e
}
func startDate(db *sql.DB) string {
	var s string
	_ = db.QueryRow(`SELECT start_date FROM book_config WHERE id=1`).Scan(&s)
	return s
}
func validateJournal(j Journal, opening bool) error {
	if !validDate(j.Date) || strings.TrimSpace(j.Description) == "" {
		return fmt.Errorf("Tanggal dan uraian jurnal wajib valid")
	}
	if len(j.Description) > 500 || len(j.Reference) > 200 || len(j.Lines) > 100 {
		return fmt.Errorf("Isian jurnal terlalu panjang")
	}
	if !opening && j.Flow != "operating" && j.Flow != "investing" && j.Flow != "financing" {
		return fmt.Errorf("Pilih kelompok arus kas")
	}
	var d, c int64
	for _, l := range j.Lines {
		a, ok := account(l.Account)
		if !ok {
			return fmt.Errorf("Akun %s tidak ditemukan", l.Account)
		}
		if opening && (a.Kind == "income" || a.Kind == "expense") {
			return fmt.Errorf("Saldo awal hanya untuk akun neraca")
		}
		if l.Debit < 0 || l.Credit < 0 || l.Debit > maxMoney || l.Credit > maxMoney || (l.Debit > 0 && l.Credit > 0) {
			return fmt.Errorf("Isi satu sisi debit atau kredit per baris, maksimal Rp 1 triliun")
		}
		d += l.Debit
		c += l.Credit
	}
	if d != c {
		return fmt.Errorf("Debit dan kredit belum seimbang (selisih %s)", rupiah(d-c))
	}
	if !opening && d == 0 {
		return fmt.Errorf("Jurnal harus memiliki nilai")
	}
	return nil
}
func insertJournal(tx *sql.Tx, j Journal) (int64, error) {
	var source any
	if j.Source != "" {
		source = j.Source
	}
	res, e := tx.Exec(`INSERT INTO ledger(source,date,description,reference,flow,pop_id,lines,void) VALUES(?,?,?,?,?,?,?,?)`, source, j.Date, j.Description, j.Reference, j.Flow, j.PopID, jsonText(j.Lines), j.Void)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}
func loadJournals(db *sql.DB) ([]Journal, error) {
	rows, e := db.Query(`SELECT id,COALESCE(source,''),date,description,reference,flow,pop_id,lines,void FROM ledger ORDER BY date DESC,id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Journal{}
	for rows.Next() {
		var j Journal
		var raw string
		if e = rows.Scan(&j.ID, &j.Source, &j.Date, &j.Description, &j.Reference, &j.Flow, &j.PopID, &raw, &j.Void); e != nil {
			return nil, e
		}
		if e = json.Unmarshal([]byte(raw), &j.Lines); e != nil {
			return nil, e
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
func (h *Handlers) BookOpening(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Date  string        `json:"date"`
		Lines []JournalLine `json:"lines"`
		Zero  bool          `json:"zero"`
	}
	if readJSON(r, &in) != nil {
		httpErr(w, 400, "Isian tidak valid")
		return
	}
	j := Journal{Date: in.Date, Description: "Saldo awal", Source: "opening", Flow: "opening", Lines: in.Lines}
	if e := validateJournal(j, true); e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	var total int64
	for _, l := range j.Lines {
		total += l.Debit
	}
	if total == 0 && !in.Zero {
		httpErr(w, 400, "Konfirmasi saldo awal nol atau isi saldo sebenarnya")
		return
	}
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal membuka transaksi")
		return
	}
	defer tx.Rollback()
	if _, e = tx.Exec(`INSERT INTO book_config(id,start_date) VALUES(1,?)`, in.Date); e != nil {
		httpErr(w, 409, "Saldo awal sudah ditetapkan. Koreksi melalui jurnal penyesuaian.")
		return
	}
	if _, e = insertJournal(tx, j); e == nil {
		e = audit(tx, "Saldo awal ditetapkan", j)
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan saldo awal")
		return
	}
	writeJSON(w, 201, map[string]bool{"ok": true})
}
func (h *Handlers) BookJournal(w http.ResponseWriter, r *http.Request) {
	var j Journal
	if readJSON(r, &j) != nil {
		httpErr(w, 400, "Isian tidak valid")
		return
	}
	if e := validateJournal(j, false); e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	start := startDate(h.DB)
	if start == "" || j.Date < start {
		httpErr(w, 400, "Tetapkan saldo awal; tanggal jurnal harus sejak tanggal mulai pembukuan")
		return
	}
	if j.PopID != 0 {
		var n int
		h.DB.QueryRow(`SELECT COUNT(*) FROM pops WHERE id=?`, j.PopID).Scan(&n)
		if n == 0 {
			httpErr(w, 400, "POP tidak ditemukan")
			return
		}
	}
	j.ID = 0
	j.Source = ""
	j.Void = false
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal membuka transaksi")
		return
	}
	defer tx.Rollback()
	j.ID, e = insertJournal(tx, j)
	if e == nil {
		e = audit(tx, "Jurnal dicatat", j)
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan jurnal")
		return
	}
	writeJSON(w, 201, j)
}
func (h *Handlers) BookVoid(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if readJSON(r, &in) != nil || len(strings.TrimSpace(in.Reason)) < 5 || len(in.Reason) > 500 {
		httpErr(w, 400, "Alasan pembatalan minimal 5, maksimal 500 karakter")
		return
	}
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal membuka transaksi")
		return
	}
	defer tx.Rollback()
	res, e := tx.Exec(`UPDATE ledger SET void=1 WHERE id=? AND source IS NULL AND void=0`, operationID(r))
	if e != nil {
		httpErr(w, 500, "Gagal membatalkan jurnal")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpErr(w, 409, "Hanya jurnal manual aktif yang dapat dibatalkan")
		return
	}
	e = audit(tx, "Jurnal dibatalkan", map[string]any{"id": operationID(r), "alasan": in.Reason})
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan pembatalan")
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// Desired bill journals are deterministic. Synchronization is explicit and idempotent.
func desiredBills(db *sql.DB, start string) (map[string]Journal, error) {
	out := map[string]Journal{}
	if start == "" {
		return out, nil
	}
	rows, e := db.Query(`SELECT g.id,g.periode,g.jumlah,g.status,g.metode,g.tanggal_bayar,COALESCE(p.nama,'Pelanggan dihapus') FROM tagihan g LEFT JOIN pelanggan p ON p.id=g.pelanggan_id ORDER BY g.id`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var id, amount int64
		var period, status, method, paid, name string
		if e = rows.Scan(&id, &period, &amount, &status, &method, &paid, &name); e != nil {
			return nil, e
		}
		date := period + "-01"
		if amount < 0 || amount > maxMoney || !validDate(date) {
			return nil, fmt.Errorf("Tagihan #%d memiliki nilai atau periode tidak valid", id)
		}
		if amount == 0 {
			continue
		}
		if date >= start {
			k := fmt.Sprintf("bill:%d:invoice", id)
			out[k] = Journal{Date: date, Source: k, Description: "Tagihan " + name + " · " + period, Reference: fmt.Sprintf("Tagihan #%d", id), Flow: "operating", Lines: []JournalLine{{"1103", amount, 0}, {"4101", 0, amount}}}
		}
		if status == "Lunas" {
			if !validDate(paid) {
				return nil, fmt.Errorf("Tagihan lunas #%d belum memiliki tanggal pembayaran valid", id)
			}
			if paid >= start {
				cash := "1102"
				if strings.EqualFold(method, "Tunai") {
					cash = "1101"
				}
				k := fmt.Sprintf("bill:%d:payment", id)
				out[k] = Journal{Date: paid, Source: k, Description: "Pembayaran " + name + " · " + period, Reference: fmt.Sprintf("Tagihan #%d · %s", id, method), Flow: "operating", Lines: []JournalLine{{cash, amount, 0}, {"1103", 0, amount}}}
			}
		}
	}
	return out, rows.Err()
}
func sameJournal(a, b Journal) bool { a.ID = 0; b.ID = 0; return jsonText(a) == jsonText(b) }
func pendingBills(existing []Journal, desired map[string]Journal) int {
	count := 0
	seen := map[string]bool{}
	for _, j := range existing {
		if !strings.HasPrefix(j.Source, "bill:") {
			continue
		}
		seen[j.Source] = true
		want, ok := desired[j.Source]
		if ok {
			if !sameJournal(j, want) {
				count++
			}
		} else if !j.Void {
			count++
		}
	}
	for k := range desired {
		if !seen[k] {
			count++
		}
	}
	return count
}
func (h *Handlers) BookSync(w http.ResponseWriter, r *http.Request) {
	start := startDate(h.DB)
	if start == "" {
		httpErr(w, 400, "Tetapkan saldo awal terlebih dahulu")
		return
	}
	// operationHandler serializes writes across all API routes while reading sources.
	desired, e := desiredBills(h.DB, start)
	if e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	existing, e := loadJournals(h.DB)
	if e != nil {
		httpErr(w, 500, "Gagal membaca jurnal")
		return
	}
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal membuka transaksi")
		return
	}
	defer tx.Rollback()
	// Serialize all billing and ledger mutations with the operations mutex (see handlers).
	seen := map[string]bool{}
	changes := 0
	for _, j := range existing {
		if !strings.HasPrefix(j.Source, "bill:") {
			continue
		}
		seen[j.Source] = true
		want, ok := desired[j.Source]
		if !ok {
			if j.Void {
				continue
			}
			want = j
			want.Void = true
		}
		if sameJournal(j, want) {
			continue
		}
		want.ID = j.ID
		_, e = tx.Exec(`UPDATE ledger SET date=?,description=?,reference=?,flow=?,lines=?,void=? WHERE id=?`, want.Date, want.Description, want.Reference, want.Flow, jsonText(want.Lines), want.Void, j.ID)
		if e == nil {
			e = audit(tx, "Sinkronisasi tagihan diperbarui", map[string]any{"before": j, "after": want})
		}
		if e != nil {
			break
		}
		changes++
	}
	if e == nil {
		keys := []string{}
		for k := range desired {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if seen[k] {
				continue
			}
			j := desired[k]
			j.ID, e = insertJournal(tx, j)
			if e == nil {
				e = audit(tx, "Tagihan dibukukan", j)
			}
			if e != nil {
				break
			}
			changes++
		}
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Sinkronisasi gagal; tidak ada perubahan disimpan")
		return
	}
	writeJSON(w, 200, map[string]int{"changes": changes})
}

type AccountBalance struct {
	Account
	Debit   int64 `json:"debit"`
	Credit  int64 `json:"credit"`
	Balance int64 `json:"balance"`
	Period  int64 `json:"period"`
}

func reportBalances(entries []Journal, from, to string) ([]AccountBalance, map[string]int64) {
	totals := map[string]int64{"asset": 0, "liability": 0, "equity": 0, "income": 0, "expense": 0, "profit": 0, "retained": 0, "cash_open": 0, "cash_close": 0, "operating": 0, "investing": 0, "financing": 0, "opening_cash": 0}
	balances := make([]AccountBalance, 0, len(accounts))
	for _, a := range accounts {
		b := AccountBalance{Account: a}
		for _, j := range entries {
			if j.Void || j.Date > to {
				continue
			}
			for _, l := range j.Lines {
				if l.Account != a.Code {
					continue
				}
				net := l.Debit - l.Credit
				b.Debit += l.Debit
				b.Credit += l.Credit
				b.Balance += net
				if j.Date >= from {
					b.Period += net
				}
				if a.Cash {
					totals["cash_close"] += net
					if j.Date < from {
						totals["cash_open"] += net
					} else if j.Source == "opening" {
						totals["opening_cash"] += net
					} else {
						totals[j.Flow] += net
					}
				}
			}
		}
		if a.Kind == "income" || a.Kind == "liability" || a.Kind == "equity" {
			b.Balance = -b.Balance
			b.Period = -b.Period
		}
		if a.Kind == "income" || a.Kind == "expense" {
			totals[a.Kind] += b.Period
			if a.Kind == "income" {
				totals["retained"] += b.Balance
			} else {
				totals["retained"] -= b.Balance
			}
		} else {
			totals[a.Kind] += b.Balance
		}
		balances = append(balances, b)
	}
	totals["profit"] = totals["income"] - totals["expense"]
	totals["equity"] += totals["retained"]
	totals["difference"] = totals["asset"] - totals["liability"] - totals["equity"]
	return balances, totals
}
func (h *Handlers) BookReport(w http.ResponseWriter, r *http.Request) {
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	if !validDate(from) || !validDate(to) || from > to {
		httpErr(w, 400, "Rentang tanggal tidak valid")
		return
	}
	entries, e := loadJournals(h.DB)
	if e != nil {
		httpErr(w, 500, "Gagal membaca pembukuan")
		return
	}
	start := startDate(h.DB)
	desired, e := desiredBills(h.DB, start)
	warning := ""
	if e != nil {
		warning = e.Error()
	}
	balances, totals := reportBalances(entries, from, to)
	periodEntries := []Journal{}
	for _, j := range entries {
		if j.Date >= from && j.Date <= to {
			periodEntries = append(periodEntries, j)
		}
	}
	logs := []Audit{}
	rows, e := h.DB.Query(`SELECT id,at,action,detail FROM operation_audit ORDER BY id DESC`)
	if e != nil {
		httpErr(w, 500, "Gagal membaca riwayat")
		return
	}
	for rows.Next() {
		var a Audit
		if e = rows.Scan(&a.ID, &a.At, &a.Action, &a.Detail); e != nil {
			break
		}
		logs = append(logs, a)
	}
	if e == nil {
		e = rows.Err()
	}
	rows.Close()
	if e != nil {
		httpErr(w, 500, "Gagal membaca riwayat")
		return
	}
	writeJSON(w, 200, map[string]any{"balance_sheet": balanceSnapshot(entries, to), "comparison": balanceSnapshot(entries, priorMonthEnd(to)), "start_date": start, "accounts": balances, "totals": totals, "journals": periodEntries, "audit": logs, "pending": pendingBills(entries, desired), "warning": warning, "from": from, "to": to})
}
