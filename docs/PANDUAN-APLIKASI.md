# Panduan Penggunaan — Ali Telematika ISP

Panduan operasional aplikasi untuk operator/admin. Untuk instalasi lihat
[INSTALASI-OPERASIONAL.md](INSTALASI-OPERASIONAL.md), untuk referensi teknis API lihat
[API.md](API.md).

Login: `admin` / `admin123` (wajib diganti setelah login pertama di
**Pengaturan → Keamanan**).

---

## Alur Kerja Ringkas

### Setiap hari
1. Buka **Dashboard** — cek *Tiket Terbuka* dan pendapatan bulan berjalan.
2. **Tiket Gangguan** — kerjakan tiket *Baru* → klik ✓ untuk majukan status
   (Baru → Diproses → Selesai), atau edit untuk isi teknisi/catatan.
3. Tandai pembayaran pelanggan yang masuk di **Keuangan** (tombol *Lunasi*).

### Awal bulan (penagihan)
1. **Keuangan** → pilih periode bulan baru → **Generate Tagihan**
   (membuat tagihan untuk semua pelanggan Aktif yang belum ditagih).
2. Kirim invoice ke pelanggan lewat **WhatsApp** (dari halaman Keuangan/
   WhatsApp) — lampiran tautan atau PDF.
3. Sepanjang bulan: tandai *Lunas* saat pembayaran diterima.

### Akhir bulan (bagi hasil mitra)
1. **Keuangan** → pastikan tagihan periode tertutup → lihat **Rekap per Mitra**.
2. **Dokumen** → **Terbitkan FAB** per mitra (rincian dihitung otomatis dari
   tagihan *lunas* × persentase bagi hasil mitra).
3. Cetak/simpan PDF, kirim ke mitra via WhatsApp/Email (riwayat pengiriman
   tercatat).
4. **Pembukuan** → **Sinkronkan Tagihan** → jurnal otomatis rapi → cek Neraca.

### Tahunan (regulasi)
- **BHP-USO** → hitung dan simpan parameter BHP/USO setahun — dasar laporan
  ke Komdigi.

---

## Modul

### 1. Dashboard
Ringkasan: mitra aktif, pelanggan aktif, tiket terbuka, pendapatan bulan ini
(dari tagihan lunas) + persentase kolektibilitas, grafik pendapatan 6 bulan,
tiket yang perlu ditangani, rekap bagi hasil mitra, dokumen terbaru.

### 2. Mitra
Mitra kerja sama (reseller, agen, warnet, RT/RW net).
- **Tambah Mitra**: isi nama, jenis, pemilik, kontak (telepon WA & email —
  dipakai untuk kirim dokumen), wilayah, **% bagi hasil** (0–100), status.
- Kode mitra terisi otomatis (`MTR-001`, dst) bila dikosongkan.
- **Detail** (ikon mata): rekap pelanggan & pendapatan, riwayat dokumen, plus
  tombol pintas **Terbitkan PKS / FAB**.
- Mitra tidak bisa dihapus selama masih memiliki pelanggan — pindahkan
  pelanggannya dulu.

### 3. Pelanggan
Data pelanggan per mitra & paket.
- Kolom penting: **User PPPoE/IP** (identitas teknis), **Status**
  (Aktif / Isolir / Nonaktif — hanya *Aktif* ditagih otomatis).
- Menghapus pelanggan juga menghapus tagihan & tiketnya (ada konfirmasi).

### 4. Paket Layanan
Daftar paket + harga bulanan. Harga dipakai saat *Generate Tagihan*.
Validasi: harga & kecepatan tidak boleh negatif/ekstrem. Paket yang sedang
dipakai pelanggan tidak bisa dihapus.

### 5. Tiket Gangguan
- Buat tiket: pilih pelanggan, jenis gangguan, prioritas, teknisi, uraian.
- Nomor otomatis `TKT/TAHUN/NNN`.
- Tombol ✓ memajukan status; tiket *Selesai* mencatat waktu penyelesaian.

### 6. Keuangan
- Pilih **periode** (bulan) → **Generate Tagihan** → tabel tagihan pelanggan.
- **Lunasi / Buka**: toggle status bayar (metode bisa diisi saat pelunasan).
- **Rekap per Mitra**: total tagihan, lunas, share %, **bagi hasil mitra** —
  dan tombol **Terbitkan FAB** langsung dari baris mitra.

### 7. Dokumen (PKS & FAB)
- **PKS** — Perjanjian Kerja Sama: untuk mitra baru/perpanjangan. Pilih mitra,
  tanggal, masa berlaku (default dari Pengaturan). Isi dokumen memuat data
  mitra, skema bagi hasil, pasal-pasal lengkap.
- **FAB** — Faktur Bisnis (rekap bagi hasil): pilih mitra + periode; rincian
  per paket, total lunas, nilai bagi hasil (dengan terbilang) dihitung otomatis.
- Nomor otomatis: `PKS/2026/003`, `FAB/2026/002`.
- Aksi: **Pratinjau**, **Cetak/Simpan PDF** (pilih *Save as PDF* di dialog
  cetak), **Kirim WhatsApp / Email**, hapus. Status pengiriman terlihat di
  daftar (mis. "WhatsApp · 05 Sep 2026").
- Isi template dapat diubah di **Pengaturan → Template Dokumen**.

### 8. Pembukuan (akuntansi)
Pembukuan berbasis jurnal sederhana dengan akun standar Indonesia.
- **Saldo Awal**: ditetapkan sekali (hanya akun neraca: kas, bank, piutang,
  aset, utang, modal).
- **Jurnal Manual**: catat pengeluaran/pendapatan lain; debit=kredit divalidasi.
- **Sinkronkan Tagihan**: jurnal otomatis dari tagihan (pendapatan saat
  ditagih, kas/piutang saat lunas). Idempaten — jalankan ulang kapan saja,
  jurnal lama dikoreksi bukan dobel.
- **Batalkan** jurnal manual (bukan hasil sinkronisasi) dengan alasan.
- Laporan: **Neraca** (dengan pembanding bulan lalu), Laba Rugi & Arus Kas
  (dari halaman), buku besar per akun (klik akun), riwayat audit.
- Uang disimpan dalam rupiah utuh — tidak ada pembulatan desimal.

### 9. BHP-USO
Perhitungan kontribusi BHP & USO tahunan berdasar pendapatan telekomunikasi
(dari pembukuan / tagihan):
- Basis: *total tagihan* atau *pembayaran diterima*.
- Tarif default: BHP 0,5% & USO 1,25% (dapat diubah + wajib alasan).
- Penyesuaian pendapatan, piutang tak tertagih, interkoneksi — dengan bukti.
- Jatuh tempo: 30 April tahun berikutnya.

### 10. POP (Point of Presence)
Inventaris lokasi/aset jaringan: kode, nama, status, alamat, koordinat, PIC,
kapasitas, biaya, daftar perangkat (tipe/serial/IP/nilai), dan pelanggan yang
terhubung ke POP tersebut. Data ini dipakai modul Surat POP.

### 11. Surat POP (Komdigi)
Surat permohonan verifikasi izin di luar komitmen untuk lokasi POP:
- **Profil** disimpan sekali (kota, NIB, penanda tangan, penerima).
- Tiga mode: *permohonan awal*, *lanjutan* (rujuk surat sebelumnya), dan
  *template* cepat (isi nomor sendiri).
- **Pratinjau** dulu (area kuning = isian wajib belum lengkap) → **Terbitkan**
  (nomor otomatis `NNN/POP/ATN-KOMDIGI/BULAN/TAHUN`, atau nomor manual).
- Surat terbit **tidak dapat diubah** — hanya dapat dibatalkan (diberi
  watermark DIBATALKAN + alasan). Lihat detail di [SURAT-POP.md](SURAT-POP.md).

### 12. WhatsApp (MPWA)
Kirim pesan otomatis lewat gateway WhatsApp (MPWA) milik sendiri:
- **Pengaturan**: URL gateway (wajib HTTPS kecuali localhost/IP privat), nomor
  sender, API key (disimpan terenkripsi), **URL publik aplikasi** (untuk
  tautan/lampiran), dan template pesan invoice/dokumen.
- Kirim: **invoice pelanggan**, **dokumen (PKS/FAB)**, atau pesan manual —
  lengkap pratinjau sebelum kirim.
- Lampiran: tautan dokumen online (kedaluwarsa 7 hari) atau PDF (24 jam) —
  semua tautan bisa **dicabut** dari Riwayat.
- Riwayat pengiriman per tujuan; pengiriman ber-idempotensi (tidak dobel kirim
  saat tombol diklik dua kali).
- Nomor HP pelanggan diisi di data **Pelanggan**; nomor mitra di data **Mitra**.

### 13. Pengaturan
- **Profil ISP**: identitas untuk kop dokumen (nama, alamat, rekening, dsb.).
- **Template Dokumen**: sunting isi PKS & FAB (HTML sederhana + placeholder),
  ada tombol kembalikan ke bawaan.
- **Keamanan**: ganti password (semua sesi lama otomatis berakhir).
- **Data**: muat ulang data contoh (menghapus data operasional!), info backup.
