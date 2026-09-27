package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func OpenDB(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite aman dari race tulis bila koneksi diserialisasi.
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		return nil, err
	}
	if err := migrateOperations(db); err != nil {
		return nil, err
	}
	if err := migratePartnerRevenue(db); err != nil {
		return nil, err
	}
	if err := migrateMessaging(db); err != nil {
		return nil, err
	}
	if err := migratePOPLetters(db); err != nil {
		return nil, err
	}
	if err := ensureSecret(db); err != nil {
		return nil, err
	}
	if err := seedIfEmpty(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			pass_hash TEXT NOT NULL,
			nama TEXT NOT NULL DEFAULT '',
			pass_changed INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			nama_isp TEXT, alamat TEXT, kota TEXT, telepon TEXT, email TEXT, website TEXT,
			npwp TEXT, penandatangan TEXT, jabatan TEXT, rekening TEXT,
			masa_pks_bulan INTEGER NOT NULL DEFAULT 12,
			tpl_pks TEXT NOT NULL DEFAULT '', tpl_fab TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS mitra (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kode TEXT NOT NULL DEFAULT '', nama TEXT NOT NULL, jenis TEXT NOT NULL DEFAULT 'Reseller',
			pemilik TEXT NOT NULL DEFAULT '', telepon TEXT NOT NULL DEFAULT '', email TEXT NOT NULL DEFAULT '',
			alamat TEXT NOT NULL DEFAULT '', wilayah TEXT NOT NULL DEFAULT '', npwp TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Aktif', share_mitra REAL NOT NULL DEFAULT 30,
			tanggal_gabung TEXT NOT NULL DEFAULT '', catatan TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS paket (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nama TEXT NOT NULL, kecepatan INTEGER NOT NULL DEFAULT 0,
			harga INTEGER NOT NULL DEFAULT 0, keterangan TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS pelanggan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nama TEXT NOT NULL, alamat TEXT NOT NULL DEFAULT '',
			mitra_id INTEGER NOT NULL DEFAULT 0, paket_id INTEGER NOT NULL DEFAULT 0,
			pppoe TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'Aktif',
			tanggal_pasang TEXT NOT NULL DEFAULT '', catatan TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS tiket (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nomor TEXT NOT NULL, pelanggan_id INTEGER NOT NULL,
			jenis TEXT NOT NULL DEFAULT '', prioritas TEXT NOT NULL DEFAULT 'Sedang',
			status TEXT NOT NULL DEFAULT 'Baru', teknisi TEXT NOT NULL DEFAULT '',
			deskripsi TEXT NOT NULL DEFAULT '', dibuat TEXT NOT NULL DEFAULT '', selesai TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS tagihan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pelanggan_id INTEGER NOT NULL, periode TEXT NOT NULL,
			jumlah INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'Belum',
			metode TEXT NOT NULL DEFAULT '', tanggal_bayar TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS dokumen (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nomor TEXT NOT NULL, jenis TEXT NOT NULL, mitra_id INTEGER NOT NULL,
			tanggal TEXT NOT NULL, periode TEXT NOT NULL DEFAULT '', nilai INTEGER NOT NULL DEFAULT 0,
			rincian TEXT NOT NULL DEFAULT '', snapshot TEXT NOT NULL DEFAULT '',
			kirim TEXT NOT NULL DEFAULT '[]', created_at TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS counters (
			jenis TEXT NOT NULL, tahun INTEGER NOT NULL, last INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (jenis, tahun)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	// Database lama belum punya kolom pass_changed; error duplikat kolom diabaikan.
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN pass_changed INTEGER NOT NULL DEFAULT 0`)
	return nil
}

func ensureSecret(db *sql.DB) error {
	var v string
	err := db.QueryRow(`SELECT value FROM meta WHERE key='jwt_secret'`).Scan(&v)
	if err == nil && v != "" {
		return nil
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	_, err = db.Exec(`INSERT OR REPLACE INTO meta(key, value) VALUES('jwt_secret', ?)`, hex.EncodeToString(buf))
	return err
}

func GetSecret(db *sql.DB) string {
	var v string
	_ = db.QueryRow(`SELECT value FROM meta WHERE key='jwt_secret'`).Scan(&v)
	return v
}

func seedIfEmpty(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err = db.Exec(`INSERT INTO users(username, pass_hash, nama) VALUES('admin', ?, 'Administrator')`, string(hash)); err != nil {
		return err
	}

	s := DefaultSettings()
	_, err = db.Exec(`INSERT INTO settings(id, nama_isp, alamat, kota, telepon, email, website, npwp, penandatangan, jabatan, rekening, masa_pks_bulan, tpl_pks, tpl_fab)
		VALUES(1, ?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.NamaISP, s.Alamat, s.Kota, s.Telepon, s.Email, s.Website, s.NPWP, s.Penandatangan, s.Jabatan, s.Rekening, s.MasaPKSBulan, s.TplPKS, s.TplFAB)
	if err != nil {
		return err
	}
	return seedSampleData(db)
}

// ResetSeed menghapus seluruh data operasional dan mengisi ulang data contoh.
// Akun login tidak dihapus.
func ResetSeed(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM ledger)+(SELECT COUNT(*) FROM pops)+(SELECT COUNT(*) FROM levy_config)+(SELECT COUNT(*) FROM wa_outbox)+(SELECT COUNT(*) FROM wa_config)+(SELECT COUNT(*) FROM pop_letters)+(SELECT COUNT(*) FROM pop_letter_profile)`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("Reset data contoh dinonaktifkan karena sudah ada pembukuan atau data operasional baru")
	}
	for _, t := range []string{"dokumen", "tagihan", "tiket", "pelanggan", "mitra", "paket", "counters"} {
		if _, err := db.Exec(`DELETE FROM ` + t); err != nil {
			return err
		}
	}
	return seedSampleData(db)
}

func seedSampleData(db *sql.DB) error {
	type pRow struct {
		nama, ket string
		cep       int
		harga     int64
	}
	pakets := []pRow{
		{"Home 10", "Internet rumahan, hemat", 10, 150000},
		{"Home 20", "Paling laris untuk keluarga", 20, 220000},
		{"Home 50", "Untuk streaming & WFH", 50, 350000},
		{"Bisnis 100", "Dedicated untuk usaha, IP publik", 100, 750000},
	}
	pketID := make([]int64, len(pakets))
	for i, p := range pakets {
		res, err := db.Exec(`INSERT INTO paket(nama, kecepatan, harga, keterangan) VALUES(?,?,?,?)`, p.nama, p.cep, p.harga, p.ket)
		if err != nil {
			return err
		}
		pketID[i], _ = res.LastInsertId()
	}

	type mRow struct {
		kode, nama, jenis, pemilik, telp, email, alamat, wilayah, npwp, status, gabung, catatan string
		share                                                                                   float64
	}
	mitras := []mRow{
		{"MTR-001", "CV Sinar Data", "Reseller", "Budi Santoso", "0812-3456-7890", "budi@sinardata.id", "Jl. Merdeka No. 45", "Bandung", "01.234.567.8-901.000", "Aktif", "2025-01-15", "", 30},
		{"MTR-002", "RT Net Alam Asri", "Perumahan (RT/RW Net)", "Dedi Kurniawan", "0813-2222-1111", "dedi.alamasri@mail.com", "Blok C2 No. 10 Perum Alam Asri", "Bandung", "", "Aktif", "2025-03-02", "Wilayah perumahan 300 unit", 35},
		{"MTR-003", "Warnet Zona Game", "Warung Internet", "Ujang Hermawan", "0857-9999-8888", "ujang@zonagame.id", "Jl. Cibabat No. 88", "Cimahi", "", "Aktif", "2025-05-20", "", 25},
		{"MTR-004", "Agen Mega Net", "Agen", "Sari Melati", "0821-5555-4444", "sari@meganet.id", "Jl. Rama No. 7", "Kab. Bandung", "", "Pending", time.Now().Format("2006-01") + "-11", "Menunggu konfirmasi dokumen", 20},
	}
	mitraID := make([]int64, len(mitras))
	for i, m := range mitras {
		res, err := db.Exec(`INSERT INTO mitra(kode,nama,jenis,pemilik,telepon,email,alamat,wilayah,npwp,status,share_mitra,tanggal_gabung,catatan)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			m.kode, m.nama, m.jenis, m.pemilik, m.telp, m.email, m.alamat, m.wilayah, m.npwp, m.status, m.share, m.gabung, m.catatan)
		if err != nil {
			return err
		}
		mitraID[i], _ = res.LastInsertId()
	}

	type cRow struct {
		nama, alamat   string
		mitra, paket   int
		pppoe, status  string
		pasang, catatn string
	}
	custs := []cRow{
		{"Andi Pratama", "Jl. Kenanga No. 3", 0, 1, "andi.home", "Aktif", "2025-02-10", ""},
		{"Rina Wulandari", "Jl. Kenanga No. 9", 0, 0, "rina.home", "Aktif", "2025-02-15", ""},
		{"Warung Kopi Barokah", "Jl. Merdeka No. 12", 0, 3, "barokah.biz", "Aktif", "2025-04-01", "Tambah IP publik 1"},
		{"Keluarga Hasan", "Jl. Anggrek No. 5", 0, 1, "hasan.home", "Aktif", "2025-06-20", ""},
		{"Bpk. Suryadi", "Jl. Anggrek No. 11", 0, 0, "suryadi.home", "Isolir", "2025-09-05", "Menunggak 2 bulan"},
		{"Ibu Fatimah", "Blok A1 No. 4", 1, 1, "fatimah.a1", "Aktif", "2025-03-15", ""},
		{"Rahmat Hidayat", "Blok A5 No. 2", 1, 2, "rahmat.a5", "Aktif", "2025-04-10", ""},
		{"Kos Putri Melati", "Blok B3", 1, 2, "kosmelati", "Aktif", "2025-05-01", "Share 1 koneksi ke 8 kamar"},
		{"Bpk. Joko", "Blok B7 No. 8", 1, 0, "joko.b7", "Aktif", "2025-07-22", ""},
		{"Warnet Zona Game", "Jl. Cibabat No. 88", 2, 3, "zonagame", "Aktif", "2025-05-20", "Koneksi utama warnet"},
		{"Icha Nuraini", "Jl. Cibabat No. 90", 2, 0, "icha.home", "Nonaktif", "2025-08-30", "Berhenti langganan"},
		{"Dr. Wawan", "Jl. Melati No. 2", 0, 2, "wawan.clinic", "Aktif", "2026-01-10", "Klinik gigi"},
	}
	custID := make([]int64, len(custs))
	custHarga := make([]int64, len(custs))
	for i, c := range custs {
		res, err := db.Exec(`INSERT INTO pelanggan(nama,alamat,mitra_id,paket_id,pppoe,status,tanggal_pasang,catatan) VALUES(?,?,?,?,?,?,?,?)`,
			c.nama, c.alamat, mitraID[c.mitra], pketID[c.paket], c.pppoe, c.status, c.pasang, c.catatn)
		if err != nil {
			return err
		}
		custID[i], _ = res.LastInsertId()
		custHarga[i] = pakets[c.paket].harga
	}

	now := time.Now()
	tahun := now.Year()
	prevPeriode := now.AddDate(0, -1, 0).Format("2006-01")
	currPeriode := now.Format("2006-01")

	type tRow struct {
		cust                        int
		jenis, prio, status, teknes string
		desk, dibuat, selesai       string
	}
	tikets := []tRow{
		{0, "Mati Total", "Tinggi", "Diproses", "Agus", "Internet mati sejak pagi, indikator LOS merah.", now.AddDate(0, 0, -1).Format("2006-01-02") + " 08:15", ""},
		{6, "Lambat", "Sedang", "Baru", "", "Kecepatan turun drastis di jam malam.", now.Format("2006-01-02") + " 09:40", ""},
		{5, "Pindah Alamat", "Rendah", "Selesai", "Rio", "Pindah titik dari Blok A1 ke A9.", now.AddDate(0, 0, -6).Format("2006-01-02"), now.AddDate(0, 0, -5).Format("2006-01-02")},
		{2, "Perangkat", "Sedang", "Baru", "", "ONT restart terus setiap ±30 menit.", now.AddDate(0, 0, -2).Format("2006-01-02") + " 14:05", ""},
	}
	for i, t := range tikets {
		if _, err := db.Exec(`INSERT INTO tiket(nomor,pelanggan_id,jenis,prioritas,status,teknisi,deskripsi,dibuat,selesai) VALUES(?,?,?,?,?,?,?,?,?)`,
			fmt.Sprintf("TKT/%d/%03d", tahun, i+1), custID[t.cust], t.jenis, t.prio, t.status, t.teknes, t.desk, t.dibuat, t.selesai); err != nil {
			return err
		}
	}

	// Tagihan: periode sebelumnya (mayoritas lunas) & periode berjalan (sebagian lunas).
	prevBelum := map[int]bool{4: true, 8: true}
	currLunas := map[int]bool{0: true, 1: true, 2: true, 5: true, 6: true, 7: true}
	for i := range custs {
		if custs[i].status == "Nonaktif" {
			continue
		}
		if _, err := db.Exec(`INSERT INTO tagihan(pelanggan_id,periode,jumlah,status,metode,tanggal_bayar) VALUES(?,?,?,?,?,?)`,
			custID[i], prevPeriode, custHarga[i], boolStatus(!prevBelum[i]), "Transfer", prevPeriode+"-20"); err != nil {
			return err
		}
		st := "Belum"
		tgl := ""
		if currLunas[i] {
			st, tgl = "Lunas", currPeriode+"-"+fmt.Sprintf("%02d", minInt(now.Day(), 28))
		}
		if _, err := db.Exec(`INSERT INTO tagihan(pelanggan_id,periode,jumlah,status,metode,tanggal_bayar) VALUES(?,?,?,?,?,?)`,
			custID[i], currPeriode, custHarga[i], st, pickMetode(st), tgl); err != nil {
			return err
		}
	}

	// Dokumen contoh: 1 PKS (mitra pertama) + 1 FAB (periode sebelumnya).
	pksTgl := now.AddDate(0, -8, 0).Format("2006-01-02")
	snapMitra, err := getMitraRaw(db, mitraID[0])
	if err != nil {
		return err
	}
	snapJSON, _ := marshalJSON(DocSnapshot{Mitra: snapMitra, MasaBulan: 12})
	fabTanggal := fmt.Sprintf("%s-%02d", currPeriode, minInt(now.Day(), 5))
	rinc := computeFAB(db, mitraID[0], prevPeriode)
	rincJSON, _ := marshalJSON(rinc)
	if _, err := db.Exec(`INSERT INTO dokumen(nomor,jenis,mitra_id,tanggal,periode,nilai,rincian,snapshot,kirim,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`,
		fmt.Sprintf("PKS/%d/001", tahun), "PKS", mitraID[0], pksTgl, "", 0, "", string(snapJSON), "[]", pksTgl+" 10:00"); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT INTO dokumen(nomor,jenis,mitra_id,tanggal,periode,nilai,rincian,snapshot,kirim,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`,
		fmt.Sprintf("FAB/%d/001", tahun), "FAB", mitraID[0], fabTanggal, prevPeriode, rinc.Nilai, string(rincJSON), string(snapJSON),
		`[{"via":"wa","tanggal":"`+fabTanggal+` 11:20"}]`, fabTanggal+" 11:00"); err != nil {
		return err
	}

	// Sinkronkan counter dokumen/tiket dengan nomor yang dipakai data contoh,
	// agar penerbitan berikutnya tidak bentrok nomor.
	for _, js := range []struct {
		jenis string
		n     int
	}{{"TKT", len(tikets)}, {"PKS", 1}, {"FAB", 1}} {
		if _, err := db.Exec(`INSERT INTO counters(jenis,tahun,last) VALUES(?,?,?)
			ON CONFLICT(jenis,tahun) DO UPDATE SET last=excluded.last`, js.jenis, tahun, js.n); err != nil {
			return err
		}
	}
	return nil
}

func boolStatus(lunas bool) string {
	if lunas {
		return "Lunas"
	}
	return "Belum"
}

func pickMetode(status string) string {
	if status == "Lunas" {
		return "Transfer"
	}
	return ""
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
