# Instalasi & Operasional — Ali Telematika ISP

Panduan lengkap memasang, memperbarui, membackup, dan menelusuri masalah.

- Panduan pemakaian aplikasi: [PANDUAN-APLIKASI.md](PANDUAN-APLIKASI.md)
- Referensi API: [API.md](API.md)

---

## A. Instalasi

### A0. ⚡ Satu perintah (cara tercepat — VPS Ubuntu/Debian)

Salin-tempel **satu baris** ini di VPS:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/devlhi/mit_altel/main/quick-install.sh)
```

Yang terjadi otomatis:
1. Deteksi arsitektur server (amd64/ARM64) → unduh installer yang tepat dari repo GitHub.
2. Jalankan installer sebagai root (minta `sudo` bila perlu).
3. Installer bertanya: **domain**, **email SSL**, **aktifkan SSL (y/n)** — isi langsung.
4. Terpasang: aplikasi + systemd + nginx + (opsional) SSL Let's Encrypt + auto-renew.

Mode non-interaktif (tanpa pertanyaan):

```bash
sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=admin@contoh.id \
  bash <(curl -fsSL https://raw.githubusercontent.com/devlhi/mit_altel/main/quick-install.sh)
```

> Catatan: pakai `bash <(curl …)` — BUKAN `curl … | bash` — agar pertanyaan
> interaktif tetap bisa dijawab. Prasyarat SSL: DNS domain → IP VPS.
> Koneksi VPS ke GitHub diperlukan (raw.githubusercontent.com port 443).

### A1. Installer satu file manual (offline/alternatif)

`install-alitelematika.sh` (±7 MB, di root proyek) memuat binary backend Linux
+ frontend jadi + semua konfigurasi. Server **tidak perlu Go/Node**.

```bash
scp install-alitelematika.sh root@IP_VPS:/root/
ssh root@IP_VPS
chmod +x install-alitelematika.sh
sudo ./install-alitelematika.sh
```

Jawab pertanyaannya: **domain**, **email SSL**, **aktifkan SSL (y/n)**.
Prasyarat SSL: DNS `A` record domain sudah mengarah ke IP VPS
(installer memeriksa dan berhenti dengan pesan bila belum).

Opsi: `--info` (versi/checksum), `--extract-only DIR` (periksa isi tanpa
menginstal), non-interaktif `sudo APP_DOMAIN=d SSL_EMAIL=e ./install-...sh`.

Membuat ulang installer setelah mengubah kode (di komputer development):

```bash
bash deploy/build-installer.sh              # linux/amd64
ARCH=arm64 bash deploy/build-installer.sh   # VPS ARM
```

### A2. Installer dari source (VPS)

```bash
sudo bash deploy/install.sh     # prompt domain/SSL sama seperti A1
```
Berbeda dari A1: VPS memasang Go 1.26.6 + Node 20 dan membangun dari source
(lambat pertama kali, cocok bila ingin build sendiri).

### A3. Komputer lokal dengan Laragon (Windows)

1. Proyek di `C:\laragon\www\alitelematika`.
2. Klik dua kali **`start.bat`** → aplikasi di **http://localhost:8080**
   (hanya bisa diakses dari komputer ini).
3. Opsional domain indah `http://alitelematika.test`: vhost sudah terpasang di
   `C:\laragon\etc\nginx\sites-enabled\alitelematika.test.conf`; tambahkan baris
   `127.0.0.1 alitelematika.test` ke `C:\Windows\System32\drivers\etc\hosts`
   (Notepad *as Administrator*), pastikan nginx Laragon aktif.
4. **`stop.bat`** untuk menghentikan.

### A4. Development

```bash
# Terminal 1 — API :8080
cd backend && go run . -addr :8080
# Terminal 2 — Nuxt dev :3000 (proxy /api)
cd frontend && npm install && npm run dev
```
Untuk dev lintas-port, jalankan API dengan `APP_ORIGIN=http://localhost:3000`
(CORS hanya aktif bila env ini diisi).

---

## B. Update aplikasi

**VPS (dari installer satu file):** unggah `install-alitelematika.sh` baru,
jalankan lagi — service, nginx, dan **data tidak tersentuh**; hanya binary &
file web yang diganti.

**Laragon:** build binary baru (`cd backend && go build -o isp-api.exe .`),
build frontend (`cd frontend && npm run generate`), lalu `stop.bat` → `start.bat`.

---

## C. Data & Backup

- Semua data ada di **satu file SQLite**: `backend/data/isp.db` (Laragon) atau
  `/opt/isp-app/data/isp.db` (VPS). Salin file itu = backup penuh.
- Aplikasi **harus dihentikan** (atau gunakan SQLite online-backup) saat menyalin
  agar hasil konsisten; cara aman di VPS:

```bash
systemctl stop isp-api
cp /opt/isp-app/data/isp.db /root/backup/isp-$(date +%F).db
systemctl start isp-api
```

- Restore: hentikan aplikasi → timpa `isp.db` dengan file backup → jalankan lagi.
- Cron harian di VPS:
```bash
0 2 * * * systemctl stop isp-api && cp /opt/isp-app/data/isp.db /root/backup/isp-$(date +\%F).db && systemctl start isp-api
```
- Pindah server: salin `isp.db` ke VPS baru **sebelum** menjalankan installer.

---

## D. Sertifikat SSL

- Diterbitkan Let's Encrypt (gratis) oleh installer; diperpanjang otomatis
  (`certbot.timer`). Cek manual: `sudo certbot renew --dry-run`.
- Ganti domain: arahkan DNS domain baru → jalankan ulang installer dengan
  domain baru (sertifikat lama diganti).
- Log certbot: `/var/log/letsencrypt/letsencrypt.log`.

---

## E. Perintah operasional (VPS)

```bash
systemctl status isp-api        # status aplikasi
journalctl -u isp-api -f        # log aplikasi realtime
systemctl restart isp-api       # restart
systemctl reload nginx          # setelah edit /etc/nginx/sites-available/isp-app.conf
nginx -t                        # uji konfigurasi nginx sebelum reload
```

---

## F. Troubleshooting

| Gejala | Kemungkinan & solusi |
|---|---|
| `502 Bad Gateway` | Service aplikasi mati → `systemctl status isp-api`, lihat `journalctl -u isp-api -n 50` |
| Installer berhenti "DNS ... bukan IP server" | Domain belum diarahkan (A record) ke IP VPS → perbaiki DNS, jalankan ulang |
| Gagal login terus padahal benar | Terkunci rate-limit 60 detik setelah 5x salah — tunggu lalu coba |
| "Sesi berakhir" di pratinjau dokumen | Token kadaluarsa/validasi ulang — login ulang lalu buka ulang |
| WhatsApp tidak terkirim | Cek Pengaturan MPWA (URL HTTPS, sender, key), status perangkat gateway; lihat Riwayat WhatsApp untuk pesan galat spesifik |
| Sinkron pembukuan gagal "nilai/periode tidak valid" | Ada tagihan tidak wajar (mis. harga paket minus dimasukkan sebelum validasi ada) — perbaiki datanya di Keuangan lalu sinkron ulang |
| Lupa password admin | Belum ada reset mandiri — minta bantuan teknis untuk reset langsung di database (bcrypt) |
| Upload PDF ditolak | File harus PDF asli (diawali `%PDF-`) maks 8 MB |
| Aplikasi tidak bisa diakses dari komputer lain (Laragon) | Memang dirancang begitu (binding 127.0.0.1 demi keamanan) — gunakan domain `alitelematika.test`/localhost |

---

## G. Keamanan — catatan penting

- Ganti password default **segera** setelah instalasi (Pengaturan → Keamanan).
- Aplikasi hanya mendengarkan `127.0.0.1:8080`; satu-satunya pintu publik
  adalah nginx (disarankan aktifkan SSL).
- Login dibatasi 5 percobaan/5 menit; penggantian password memutus semua sesi.
- API key gateway WhatsApp disimpan terenkripsi dan tidak pernah dikirim ke
  browser.
- Sebelum membagikan/mempublikasikan source: folder `.audit/`, `backend/data/`,
  dan binary sudah masuk `.gitignore` — jangan sertakan `isp.db` di repo publik.
