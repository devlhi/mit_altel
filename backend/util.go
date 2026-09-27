package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(dst)
}

var bulanID = map[time.Month]string{
	time.January: "Januari", time.February: "Februari", time.March: "Maret",
	time.April: "April", time.May: "Mei", time.June: "Juni",
	time.July: "Juli", time.August: "Agustus", time.September: "September",
	time.October: "Oktober", time.November: "November", time.December: "Desember",
}

var hariID = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

// tanggalIndo mengubah "2006-01-02" menjadi "26 September 2026".
func tanggalIndo(iso string) string {
	iso = strings.TrimSpace(iso)
	if t, err := time.Parse("2006-01-02", iso); err == nil {
		return fmt.Sprintf("%d %s %d", t.Day(), bulanID[t.Month()], t.Year())
	}
	if t, err := time.Parse(time.RFC3339, iso); err == nil {
		return fmt.Sprintf("%d %s %d", t.Day(), bulanID[t.Month()], t.Year())
	}
	return iso
}

func hariIndo(iso string) string {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(iso))
	if err != nil {
		t = time.Now()
	}
	return hariID[t.Weekday()]
}

// periodeLabel mengubah "2026-09" menjadi "September 2026".
func periodeLabel(p string) string {
	if t, err := time.Parse("2006-01", p); err == nil {
		return fmt.Sprintf("%s %d", bulanID[t.Month()], t.Year())
	}
	return p
}

func addMonthsISO(iso string, months int) string {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(iso))
	if err != nil {
		t = time.Now()
	}
	y, m, d := t.Date()
	total := int(m) - 1 + months
	ny := y + total/12
	nm := total%12 + 1
	last := time.Date(ny, time.Month(nm+1), 0, 0, 0, 0, 0, time.UTC).Day()
	if d > last {
		d = last
	}
	return time.Date(ny, time.Month(nm), d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

// rupiah memformat 1250000 menjadi "Rp 1.250.000".
func rupiah(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(s[i : i+3])
	}
	if neg {
		return "-Rp " + b.String()
	}
	return "Rp " + b.String()
}

func angkaKata(n int64) string {
	a := []string{"", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh", "sebelas"}
	var t func(x int64) string
	t = func(x int64) string {
		switch {
		case x < 12:
			return a[x]
		case x < 20:
			return t(x-10) + " belas"
		case x < 100:
			return t(x/10) + " puluh " + t(x%10)
		case x < 200:
			return "seratus " + t(x-100)
		case x < 1000:
			return t(x/100) + " ratus " + t(x%100)
		case x < 2000:
			return "seribu " + t(x-1000)
		case x < 1000000:
			return t(x/1000) + " ribu " + t(x%1000)
		case x < 1000000000:
			return t(x/1000000) + " juta " + t(x%1000000)
		case x < 1000000000000:
			return t(x/1000000000) + " miliar " + t(x%1000000000)
		default:
			return t(x/1000000000000) + " triliun " + t(x%1000000000000)
		}
	}
	return strings.Join(strings.Fields(t(n)), " ")
}

func terbilang(n int64) string {
	if n == 0 {
		return "nol rupiah"
	}
	return angkaKata(n) + " rupiah"
}

// waNormalize mengubah 0812.../62.../812... menjadi format 62812...
func waNormalize(t string) string {
	s := strings.TrimSpace(t)
	var b strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	d := b.String()
	switch {
	case strings.HasPrefix(d, "62"):
		return d
	case strings.HasPrefix(d, "0"):
		return "62" + d[1:]
	case strings.HasPrefix(d, "8"):
		return "62" + d
	default:
		return d
	}
}

func escHTML(s string) string { return html.EscapeString(s) }

var numMu sync.Mutex

// nextNumber menghasilkan nomor dokumen berurutan, mis. PKS/2026/003.
func nextNumber(db *sql.DB, jenis string, now time.Time) (string, error) {
	numMu.Lock()
	defer numMu.Unlock()
	tahun := now.Year()
	var last int
	err := db.QueryRow(`SELECT last FROM counters WHERE jenis=? AND tahun=?`, jenis, tahun).Scan(&last)
	switch {
	case err == sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO counters(jenis,tahun,last) VALUES(?,?,1)`, jenis, tahun); err != nil {
			return "", err
		}
		last = 1
	case err != nil:
		return "", err
	default:
		last++
		if _, err = db.Exec(`UPDATE counters SET last=? WHERE jenis=? AND tahun=?`, last, jenis, tahun); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s/%d/%03d", jenis, tahun, last), nil
}

func nowISO() string { return time.Now().Format("2006-01-02") }

func nowClock() string { return time.Now().Format("15:04") }

func nowPeriode() string { return time.Now().Format("2006-01") }
