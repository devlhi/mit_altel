# Invoice perusahaan dan preview yang terbaca

**Arahan pengguna yang dikonfirmasi:** Teknologi — biru, lebih berwarna. Arah ini menggantikan navy formal awal. Gunakan kop biru `#165dcc` / `#123e89` dengan ilustrasi jaringan vektor sederhana, teks putih, tabel biru pucat `#eef5ff`, dan total biru gelap. Motif terbatas pada kop dan relevan dengan layanan internet. Isi dokumen tetap bersih dan padat.

## Rencana

Pengguna menolak versi minimalis. Arah kerja awal: invoice perusahaan navy, padat, formal, sambil meminta preferensi spesifik. Perbaikan utama adalah komposisi dokumen dan ruang preview: hilangkan sidebar yang menduplikasi isi, tampilkan kertas dengan lebar sebenarnya sebagai mode awal desktop, dan sisakan mode satu halaman sebagai pilihan.

- Warna: navy `#203650`, putih `#ffffff`, abu data `#5c6978`, garis `#dce3eb`, bidang tabel `#f3f6f9`, lunas `#247053`.
- Tipografi: Arial; merek 20 px, judul invoice 34 px, nama pelanggan 18 px, nominal akhir 30 px, teks isi 12–13 px. Angka tabular.
- Layout: kop perusahaan dengan kontak di kanan; judul dan nomor invoice dalam satu baris; data pelanggan/periode/ID dalam panel tipis; tabel padat; rekening kiri dan total kanan; footer kontak bantuan.

```text
Logo + perusahaan                      Alamat / kontak
-------------------------------------------------------
Invoice                 Status          Nomor invoice

Pelanggan               Periode         ID pelanggan
Alamat                  Mitra

Layanan                   Qty      Tarif       Jumlah
-------------------------------------------------------
Paket internet

Informasi pembayaran                 Total / dibayar
Rekening                             Sisa pembayaran

Catatan / kontak
```

## Review

Versi sebelumnya terlihat kecil karena seluruh kertas A4 diperkecil agar muat pada kanvas sempit. Layout dialog baru memusatkan dokumen dan menyediakan zoom Pas lebar sebagai awal di desktop. Data tidak diulang dalam sidebar, ruang kosong antarbagiannya dipadatkan. Pertahankan detail transaksi dan ukuran A4 cetak; jangan menambahkan tanggal, pajak, rekening, QR, atau tanda tangan fiktif.

## Verifikasi

- `npm run test:invoice`: 6 tes lulus, termasuk nominal tersimpan, status pembayaran, escaping, dan nomor invoice.
- `npm run generate`: lulus. Log `.audit/invoice-technology-build.log`.
- Desktop 1280 × 720: preview terbuka dengan Pas lebar, kertas 794 × 1123 px, tidak ada overflow isi invoice. Sidebar dihapus dan modal menggunakan satu area gulir dokumen.
- Ponsel 390 × 844: preview awal Pas layar; lebar halaman tetap 390 px, tombol cetak terlihat. Invoice Andi Pratama lunas menunjukkan Rp0.
- Tinggi kanvas cocok dengan batas modal pada desktop maupun ponsel; `scrollHeight` dan `clientHeight` modal-body sama, sehingga tidak ada gulir bertingkat.
- Screenshot: `docs/screenshots/invoice-technology.png`, `invoice-technology-payment.png`, dan `invoice-technology-mobile.png`.
- Data operasional tidak diubah. Dialog cetak sistem dan PDF belum diuji untuk revisi ini.
