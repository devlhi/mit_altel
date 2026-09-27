# Ali Telematika ISP — Aplikasi Manajemen ISP

Aplikasi manajemen ISP berbasis web untuk mengelola **mitra**, **operasional pelanggan**, **keuangan/tagihan**, serta **penerbitan dokumen PKS & FAB** ke mitra.

- **Backend**: Golang (REST API + SQLite, login JWT) → hasil build **satu file binary**
- **Frontend**: Nuxt 3 (SPA) → hasil build **file statis** yang disajikan langsung oleh binary backend
- **Database**: SQLite (tanpa server database terpisah — cukup satu file `data/isp.db`)
- **Lokasi proyek**: `C:\laragon\www\alitelematika`

## ⚡ Instalasi SATU PERINTAH (VPS Ubuntu/Debian)

Salin-tempel satu baris ini di VPS — installer akan menanyakan **domain**, **email SSL**, dan **aktifkan SSL?**, lalu memasang semuanya (tanpa perlu Go/Node di server):

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/devlhi/mit_altel/main/quick-install.sh)
```

> Prasyarat SSL: DNS `A` record domain sudah mengarah ke IP VPS.
> Detail lengkap: [docs/INSTALASI-OPERASIONAL.md](docs/INSTALASI-OPERASIONAL.md)

## 📚 Dokumentasi

| Dokumen | Isi |
|---|---|
| [docs/PANDUAN-APLIKASI.md](docs/PANDUAN-APLIKASI.md) | Panduan pemakaian semua modul + alur kerja harian/bulanan/tahunan ISP |
| [docs/API.md](docs/API.md) | Referensi lengkap seluruh endpoint API (auth, contoh, kode status) |
| [docs/INSTALASI-OPERASIONAL.md](docs/INSTALASI-OPERASIONAL.md) | Instalasi (satu file / source / Laragon), update, backup-restore, SSL, troubleshooting |
| [docs/SURAT-POP.md](docs/SURAT-POP.md) | Khusus modul Surat POP Komdigi |

> File lain di `docs/` (UX-AUDIT, OPERASIONAL-KEUANGAN, MPWA-DAN-NERACA, catatan
> desain invoice, dsb.) adalah catatan proses pengembangan — dokumentasi resmi
> adalah empat dokumen di atas.

## Menjalankan di Laragon (Windows)

1. Jalankan **`start.bat`** (klik dua kali) — aplikasi berjalan di **http://localhost:8080**
2. (Opsional) Akses dengan domain indah **http://alitelematika.test** lewat nginx Laragon:
   - Vhost sudah dipasang di `C:\laragon\etc\nginx\sites-enabled\alitelematika.test.conf` (proxy ke :8080)
   - Tambahkan baris berikut ke file hosts (`C:\Windows\System32\drivers\etc\hosts`, butuh hak admin — buka Notepad as Administrator):
     ```
     127.0.0.1 alitelematika.test
     ```
   - Pastikan nginx Laragon aktif, lalu buka http://alitelematika.test
3. Hentikan aplikasi dengan **`stop.bat`**

Login pertama: `admin` / `admin123` (segera ganti di menu Pengaturan → Keamanan).

## Fitur

| Modul | Isi |
|---|---|
| Dashboard | Mitra aktif, pelanggan aktif, tiket terbuka, pendapatan bulan ini, grafik 6 bulan, rekap bagi hasil |
| Mitra | CRUD mitra (reseller/agen/warnet/RT-net), kontak, wilayah, % bagi hasil, detail rekap pelanggan & dokumen |
| Pelanggan | CRUD pelanggan per mitra & paket, status Aktif/Isolir/Nonaktif, user PPPoE |
| Paket | Daftar paket layanan + harga bulanan |
| Tiket | Laporan gangguan: prioritas, teknisi, alur status Baru → Diproses → Selesai |
| Keuangan | Generate tagihan bulanan massal, tandai lunas, rekap per mitra |
| Dokumen | **Terbitkan PKS** (perjanjian kerja sama) & **FAB** (faktur bisnis rekap bagi hasil) dengan nomor otomatis, pratinjau, **cetak/simpan PDF**, kirim via **WhatsApp/Email** |
| Pengaturan | Profil ISP (kop dokumen), template dokumen PKS/FAB yang dapat diedit, ganti password, muat data contoh |

> **Tentang FAB**: di aplikasi ini FAB = *Faktur Bisnis* — rekapitulasi tagihan & bagi hasil mitra per bulan (dasar perhitungan pembayaran). Jika di perusahaan Anda FAB berarti dokumen lain, template-nya bisa diubah bebas di menu **Pengaturan → Template Dokumen**.

## Login pertama

```
Username : admin
Password : admin123
```

**Segera ganti password** di menu Pengaturan → Keamanan.

---

## Menjalankan di kom lokal (development)

Prasyarat: Go ≥ 1.22 dan Node.js ≥ 18.

```bash
# Terminal 1 — backend (API di :8080)
cd backend
go mod tidy
go run . -addr :8080

# Terminal 2 — frontend (dev server :3000, proxy /api ke :8080)
cd frontend
npm install
npm run dev
```

Buka http://localhost:3000

## Menjalankan mode produksi di lokal (single binary)

```bash
# Build backend
cd backend
go build -o isp-api .

# Build frontend menjadi file statis
cd ../frontend
npm run generate

# Jalankan semuanya dalam satu proses di port 8080
cd ../backend
./isp-api -addr :8080 -db data/isp.db -static ../frontend/.output/public
```

Buka http://localhost:8080

---

## Install di VPS (Ubuntu/Debian) — installer interaktif dengan SSL

### Cara paling mudah: SATU FILE installer (disarankan)

`install-alitelematika.sh` (± 7 MB, di root proyek) berisi **semua** yang dibutuhkan:
binary backend Linux + frontend jadi + service systemd + konfigurasi nginx + SSL —
server **tidak perlu Go/Node sama sekali**.

```bash
# di komputer Anda: unggah satu file ini ke VPS
scp install-alitelematika.sh root@IP_VPS:/root/

# di VPS:
chmod +x install-alitelematika.sh
sudo ./install-alitelematika.sh
```

Installer akan **bertanya** (tinggal diisi):

1. **Domain aplikasi** — mis. `isp.alitelmatara.id`. Kosongkan jika hanya pakai IP.
2. **Email untuk SSL** — untuk notifikasi Let's Encrypt.
3. **Aktifkan SSL?** — `y`/`n`.

Selesai → aplikasi jalan di `https://domain-anda/` (atau `http://IP/`) dengan nginx,
systemd, redirect HTTPS, HSTS, dan perpanjangan sertifikat otomatis.

Opsi tambahan: `./install-alitelematika.sh --info` (lihat versi/checksum) dan
`./install-alitelematika.sh --extract-only DIR` (periksa isi paket tanpa menginstal).

**Membuat ulang file installer** (setelah mengubah kode, di komputer development):

```bash
bash deploy/build-installer.sh              # hasil: install-alitelematika.sh (linux/amd64)
ARCH=arm64 bash deploy/build-installer.sh   # varian ARM64
```

Mode non-interaktif: `sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=a@b.id ./install-alitelematika.sh`

> **Prasyarat SSL**: arahkan DNS `A` record domain ke IP VPS **sebelum** menjalankan installer
> (installer memeriksa DNS dan berhenti dengan pesan jelas bila belum mengarah).

### Cara alternatif: installer dari source

Unggah folder proyek ini ke VPS (mis. dengan `scp` atau `git clone`), lalu:

```bash
cd isp-app
sudo bash deploy/install.sh
```

Installer akan **bertanya** (Anda tinggal mengisi):

1. **Domain aplikasi** — mis. `isp.alitelmatara.id`. Kosongkan jika hanya pakai IP (SSL tidak bisa untuk IP).
2. **Email untuk SSL** — dipakai Let's Encrypt untuk notifikasi sertifikat.
3. **Aktifkan SSL?** — `y`/`n`. Otomatis disarankan `y` bila domain & email sudah diisi.

Berdasarkan jawaban tersebut, installer otomatis: memasang Go 1.26.6/Node 20/nginx/certbot bila belum ada, build backend & frontend, memasang service systemd `isp-api` (binding `127.0.0.1:8080`), membuat konfigurasi nginx sesuai domain, lalu **menerbitkan sertifikat SSL Let's Encrypt gratis**, mengaktifkan redirect HTTP→HTTPS, header keamanan (HSTS, nosniff, X-Frame-Options, Referrer-Policy), dan pembaruan sertifikat otomatis (`certbot.timer`). Sebelum SSL diterbitkan, installer juga memeriksa DNS domain sudah mengarah ke IP server.

> **Prasyarat SSL**: Arahkan DNS `A` record domain ke IP VPS **sebelum** menjalankan installer.

### Mode non-interaktif (script/CI)

```bash
sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=admin@contoh.id bash deploy/install.sh
# tambahkan SSL=0 untuk memaksa tanpa SSL meski domain terisi
```

### Cara manual

```bash
sudo mkdir -p /opt/isp-app/data

# 1. Build backend (di komputer mana pun yang punya Go)
cd backend
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /opt/isp-app/isp-api .

# 2. Build frontend
cd ../frontend
npm ci && npm run generate
cp -r .output/public /opt/isp-app/web

# 3. Service systemd
sudo cp deploy/isp-api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now isp-api

# 4. Nginx
sudo cp deploy/nginx-isp-app.conf /etc/nginx/sites-available/isp-app
sudo ln -s /etc/nginx/sites-available/isp-app /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx
```

### HTTPS (domain sendiri)

Cara termudah: jalankan ulang installer dan isi domain — SSL dikonfigurasi otomatis:

```bash
sudo APP_DOMAIN=isp.alitelmatara.id SSL_EMAIL=email@anda.com bash deploy/install.sh
```

Atau manual dengan certbot (konfigurasi nginx di `/etc/nginx/sites-available/isp-app.conf` sudah menyediakan jalur `/.well-known/acme-challenge/`):

```bash
sudo apt install certbot
sudo certbot certonly --webroot -w /var/www/letsencrypt -d isp.alitelmatara.id -d www.isp.alitelmatara.id -m email@anda.com --agree-tos -n
# lalu rujuk sertifikat di blok 443 konfigurasi nginx dan reload
```

---

## Operasional & perawatan

```bash
systemctl status isp-api        # status aplikasi
journalctl -u isp-api -f        # lihat log realtime
sudo systemctl restart isp-api  # restart setelah update
```

**Backup**: cukup salin file `/opt/isp-app/data/isp.db` (mis. cron harian ke server lain):
```bash
0 2 * * * cp /opt/isp-app/data/isp.db /root/backup/isp-$(date +\%F).db
```

**Update aplikasi**: tarik kode terbaru → jalankan ulang `install.sh` (aman, data tidak tersentuh).

**Pindah server**: salin `data/isp.db` ke VPS baru sebelum menjalankan aplikasi.

## Struktur proyek

```
isp-app/
├── backend/            # API Golang
│   ├── main.go         # entry point + static file server
│   ├── routes.go       # daftar endpoint
│   ├── auth.go         # login JWT + middleware
│   ├── db.go           # schema SQLite + seed data contoh
│   ├── handlers.go     # CRUD mitra/paket/pelanggan/tiket
│   ├── keuangan.go     # tagihan + rekap bagi hasil
│   ├── dokumen.go      # penerbitan PKS/FAB + generator HTML
│   ├── dashboard.go, settings.go, templates.go, models.go, util.go
│   └── data/           # file database (dibuat otomatis)
├── frontend/           # Nuxt 3 (SPA)
│   ├── pages/          # dashboard, mitra, pelanggan, paket, tiket, keuangan, dokumen, pengaturan, login
│   ├── components/     # modal, ikon, toast, konfirmasi
│   ├── composables/    # API client, format helper, toast/confirm
│   └── assets/css/     # design system
└── deploy/
    ├── install.sh      # installer otomatis VPS
    ├── isp-api.service # unit systemd
    └── nginx-isp-app.conf
```

## Catatan keamanan

- Aplikasi ini dirancang untuk jaringan internal / dipublikasikan di belakang nginx + HTTPS.
- Token login berlaku 7 hari; semua endpoint API (kecuali login) membutuhkan token.
- Data mitra & pelanggan tersimpan di SQLite pada server — lakukan backup rutin.
