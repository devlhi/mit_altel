package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type Levy struct {
	Basis        string `json:"basis"`
	Year         int    `json:"year"`
	BHP          int64  `json:"bhp_rate"`
	USO          int64  `json:"uso_rate"`
	Adjustment   int64  `json:"adjustment"`
	BadDebt      int64  `json:"bad_debt"`
	Interconnect int64  `json:"interconnect"`
	Notes        string `json:"notes"`
	Evidence     string `json:"evidence"`
}

func levyBase(gross int64, l Levy) (int64, error) {
	if l.Year < 2000 || l.Year > 2100 || l.BHP < 0 || l.USO < 0 || l.BHP > 10000 || l.USO > 10000 {
		return 0, fmt.Errorf("Tahun atau tarif tidak valid")
	}
	if l.Adjustment < -maxMoney || l.Adjustment > maxMoney || l.BadDebt < 0 || l.BadDebt > maxMoney || l.Interconnect < 0 || l.Interconnect > maxMoney {
		return 0, fmt.Errorf("Nilai penyesuaian / pengurang tidak valid")
	}
	if (l.Adjustment != 0 || l.BHP != 50 || l.USO != 125) && strings.TrimSpace(l.Notes) == "" {
		return 0, fmt.Errorf("Isi alasan penyesuaian pendapatan atau perubahan tarif")
	}
	if (l.BadDebt > 0 || l.Interconnect > 0) && strings.TrimSpace(l.Evidence) == "" {
		return 0, fmt.Errorf("Isi referensi bukti pengurang yang memenuhi syarat")
	}
	base := gross + l.Adjustment - l.BadDebt - l.Interconnect
	if base < 0 {
		return 0, fmt.Errorf("Pengurang melebihi pendapatan; tinjau kembali dasar perhitungan")
	}
	return base, nil
}
func levyAmount(base, rate int64) int64 { return (base/10000)*rate + ((base%10000)*rate+5000)/10000 }
func telecomRevenue(entries []Journal, year int) (int64, []int64) {
	var gross int64
	months := make([]int64, 12)
	prefix := strconv.Itoa(year) + "-"
	for _, j := range entries {
		if j.Void || !strings.HasPrefix(j.Date, prefix) {
			continue
		}
		m, _ := strconv.Atoi(j.Date[5:7])
		if m < 1 || m > 12 {
			continue
		}
		for _, line := range j.Lines {
			a, ok := account(line.Account)
			if ok && a.Telecom {
				net := line.Credit - line.Debit
				gross += net
				months[m-1] += net
			}
		}
	}
	return gross, months
}
func (h *Handlers) LevyReport(w http.ResponseWriter, r *http.Request) {
	year, e := strconv.Atoi(r.URL.Query().Get("year"))
	if e != nil || year < 2000 || year > 2100 {
		httpErr(w, 400, "Tahun tidak valid")
		return
	}
	l := Levy{Year: year, BHP: 50, USO: 125, Basis: "billed"}
	var raw string
	err := h.DB.QueryRow(`SELECT data FROM levy_config WHERE year=?`, year).Scan(&raw)
	if err == nil {
		if json.Unmarshal([]byte(raw), &l) != nil {
			httpErr(w, 500, "Konfigurasi tidak dapat dibaca")
			return
		}
	}
	if l.Basis == "" {
		l.Basis = "billed"
	}
	savedBasis := l.Basis
	if basis := r.URL.Query().Get("basis"); basis != "" {
		l.Basis = basis
	}
	gross, months, partners, bills, e := partnerRevenue(h.DB, year, l.Basis)
	if e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	base, e := levyBase(gross, l)
	warning := ""
	if e != nil {
		warning = e.Error()
		base = 0
	}
	writeJSON(w, 200, map[string]any{"saved_basis": savedBasis, "config": l, "gross": gross, "months": months, "partners": partners, "bills": bills, "base": base, "bhp": levyAmount(base, l.BHP), "uso": levyAmount(base, l.USO), "due": fmt.Sprintf("%d-04-30", year+1), "warning": warning})
}
func (h *Handlers) SaveLevy(w http.ResponseWriter, r *http.Request) {
	var l Levy
	if readJSON(r, &l) != nil {
		httpErr(w, 400, "Isian tidak valid")
		return
	}
	if len(l.Notes) > 2000 || len(l.Evidence) > 2000 {
		httpErr(w, 400, "Catatan terlalu panjang")
		return
	}
	if l.Basis == "" {
		l.Basis = "billed"
	}
	gross, _, _, _, e := partnerRevenue(h.DB, l.Year, l.Basis)
	if e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	if _, e = levyBase(gross, l); e != nil {
		httpErr(w, 400, e.Error())
		return
	}
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal membuka transaksi")
		return
	}
	defer tx.Rollback()
	_, e = tx.Exec(`INSERT INTO levy_config(year,data) VALUES(?,?) ON CONFLICT(year) DO UPDATE SET data=excluded.data`, l.Year, jsonText(l))
	if e == nil {
		e = audit(tx, "Parameter BHP / USO disimpan", l)
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan parameter")
		return
	}
	writeJSON(w, 200, l)
}

type POPAsset struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Serial string `json:"serial"`
	IP     string `json:"ip"`
	Status string `json:"status"`
	Value  int64  `json:"value"`
}
type POP struct {
	ID        int64      `json:"id"`
	Kode      string     `json:"kode"`
	Nama      string     `json:"nama"`
	Status    string     `json:"status"`
	Alamat    string     `json:"alamat"`
	Latitude  *float64   `json:"latitude"`
	Longitude *float64   `json:"longitude"`
	PIC       string     `json:"pic"`
	Telepon   string     `json:"telepon"`
	Upstream  string     `json:"upstream"`
	Capacity  int64      `json:"capacity"`
	Cost      int64      `json:"cost"`
	Notes     string     `json:"notes"`
	Assets    []POPAsset `json:"assets"`
	Customers []int64    `json:"customers"`
}

func (h *Handlers) ListPOP(w http.ResponseWriter, r *http.Request) {
	out := []POP{}
	rows, e := h.DB.Query(`SELECT id,data FROM pops ORDER BY kode`)
	if e != nil {
		httpErr(w, 500, "Gagal membaca POP")
		return
	}
	for rows.Next() {
		var p POP
		var raw string
		if e = rows.Scan(&p.ID, &raw); e != nil {
			break
		}
		id := p.ID
		if e = json.Unmarshal([]byte(raw), &p); e != nil {
			break
		}
		p.ID = id
		out = append(out, p)
	}
	if e == nil {
		e = rows.Err()
	}
	rows.Close()
	if e != nil {
		httpErr(w, 500, "Gagal membaca POP")
		return
	}
	mapping := map[int64][]int64{}
	rows, e = h.DB.Query(`SELECT pop_id,customer_id FROM pop_customers ORDER BY customer_id`)
	if e != nil {
		httpErr(w, 500, "Gagal membaca pelanggan POP")
		return
	}
	for rows.Next() {
		var p, c int64
		if e = rows.Scan(&p, &c); e != nil {
			break
		}
		mapping[p] = append(mapping[p], c)
	}
	if e == nil {
		e = rows.Err()
	}
	rows.Close()
	if e != nil {
		httpErr(w, 500, "Gagal membaca pelanggan POP")
		return
	}
	for i := range out {
		out[i].Customers = mapping[out[i].ID]
		if out[i].Customers == nil {
			out[i].Customers = []int64{}
		}
		if out[i].Assets == nil {
			out[i].Assets = []POPAsset{}
		}
	}
	writeJSON(w, 200, out)
}
func (h *Handlers) SavePOP(w http.ResponseWriter, r *http.Request) {
	var p POP
	if readJSON(r, &p) != nil {
		httpErr(w, 400, "Isian POP tidak valid")
		return
	}
	p.ID = operationID(r)
	if r.Method == "PUT" && p.ID <= 0 {
		httpErr(w, 400, "ID POP tidak valid")
		return
	}
	p.Kode = strings.TrimSpace(strings.ToUpper(p.Kode))
	p.Nama = strings.TrimSpace(p.Nama)
	if p.Kode == "" || p.Nama == "" || len(p.Kode) > 50 || len(p.Nama) > 150 || len(p.Notes) > 4000 || len(p.Alamat) > 1000 {
		httpErr(w, 400, "Kode dan nama wajib diisi; periksa panjang isian")
		return
	}
	if p.Status != "Aktif" && p.Status != "Pemeliharaan" && p.Status != "Nonaktif" {
		httpErr(w, 400, "Status POP tidak valid")
		return
	}
	if p.Capacity < 0 || p.Capacity > 100000000 || p.Cost < 0 || p.Cost > maxMoney || len(p.Assets) > 1000 || len(p.Customers) > 100000 {
		httpErr(w, 400, "Kapasitas / biaya / inventaris melebihi batas")
		return
	}
	if (p.Latitude == nil) != (p.Longitude == nil) {
		httpErr(w, 400, "Isi kedua koordinat atau kosongkan keduanya")
		return
	}
	if p.Latitude != nil && (math.IsNaN(*p.Latitude) || math.Abs(*p.Latitude) > 90 || math.IsNaN(*p.Longitude) || math.Abs(*p.Longitude) > 180) {
		httpErr(w, 400, "Koordinat tidak valid")
		return
	}
	for _, a := range p.Assets {
		if strings.TrimSpace(a.Name) == "" || a.Value < 0 || a.Value > maxMoney || len(a.Name) > 150 || len(a.Serial) > 150 || len(a.IP) > 100 {
			httpErr(w, 400, "Nama dan nilai perangkat harus valid")
			return
		}
	}
	tx, e := h.DB.Begin()
	if e != nil {
		httpErr(w, 500, "Gagal membuka transaksi")
		return
	}
	defer tx.Rollback()
	seen := map[int64]bool{}
	for _, id := range p.Customers {
		if seen[id] {
			httpErr(w, 400, "Pelanggan terduplikasi")
			return
		}
		seen[id] = true
		var n int
		if e = tx.QueryRow(`SELECT COUNT(*) FROM pelanggan WHERE id=?`, id).Scan(&n); e != nil || n == 0 {
			httpErr(w, 400, "Pelanggan tidak ditemukan")
			return
		}
		var owner int64
		_ = tx.QueryRow(`SELECT pop_id FROM pop_customers WHERE customer_id=?`, id).Scan(&owner)
		if owner != 0 && owner != p.ID {
			httpErr(w, 409, "Pelanggan sudah terhubung ke POP lain. Lepaskan dari POP lama terlebih dahulu.")
			return
		}
	}
	if p.ID == 0 {
		res, err := tx.Exec(`INSERT INTO pops(kode,data) VALUES(?,?)`, p.Kode, jsonText(p))
		e = err
		if e == nil {
			p.ID, e = res.LastInsertId()
		}
	} else {
		res, err := tx.Exec(`UPDATE pops SET kode=?,data=? WHERE id=?`, p.Kode, jsonText(p), p.ID)
		e = err
		if e == nil {
			n, _ := res.RowsAffected()
			if n == 0 {
				httpErr(w, 404, "POP tidak ditemukan")
				return
			}
		}
	}
	if e != nil {
		httpErr(w, 409, "Kode POP sudah dipakai atau data gagal disimpan")
		return
	}
	_, e = tx.Exec(`DELETE FROM pop_customers WHERE pop_id=?`, p.ID)
	if e == nil {
		for _, id := range p.Customers {
			_, e = tx.Exec(`INSERT INTO pop_customers(customer_id,pop_id) VALUES(?,?)`, id, p.ID)
			if e != nil {
				break
			}
		}
	}
	if e == nil {
		e = audit(tx, "Data POP disimpan", p)
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		httpErr(w, 500, "Gagal menyimpan POP")
		return
	}
	writeJSON(w, 200, p)
}
