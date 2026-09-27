# Audit UI/UX Ali Telematika

Tanggal: 27 September 2026. Cakupan: kode frontend Nuxt, alur operasional, generator HTML PKS/FAB dan pengujian lokal. Ini bukan audit keamanan menyeluruh atau penilaian isi hukum PKS.

## Arah desain

Pengguna: operator ISP yang mengurus pelanggan, mitra, tagihan dan dokumen. Prioritas: menemukan pekerjaan berikutnya, memahami nominal dan status, serta menerbitkan dokumen dengan yakin.

- Palet: tinta jaringan `#16324f`, teal `#087f83`, kanvas `#f3f6f8`, putih `#ffffff`, teks sekunder `#586b7d`, garis `#dce5eb`.
- Tipografi: Segoe UI untuk antarmuka dan angka operasional (tersedia lokal); Georgia untuk isi pasal PKS; Arial untuk FAB dan identitas dokumen. Tidak memerlukan font eksternal.
- Tata letak: navigasi tetap di kiri, judul halaman dan tindakan utama di atas, area kerja lebar dengan tabel rata kiri dan nominal rata kanan. Ponsel memakai drawer dan tabel yang dapat digeser.
- Ciri utama: pusat dokumen dengan dua jalur jelas, PKS dan FAB, ringkasan data sebelum terbit, lalu pratinjau lembar A4.

```text
Navigasi | Judul halaman                     Akun
         | PKS: kerja sama     FAB: bagi hasil
         | Jenis dokumen | Pencarian | Periode
         | Arsip dokumen dan tindakan pratinjau
```

Review rencana: hilangkan kartu bergradasi dan garis dekoratif berulang. Gunakan warna untuk membedakan jenis dokumen dan status, bukan menghias setiap permukaan. Ringkasan dashboard harus mengantar ke pekerjaan operasional.

## Temuan yang ditangani

1. Dekorasi, bayangan dan gradasi berulang mengganggu hierarki informasi.
2. Alur terbit dokumen tidak merangkum mitra, periode, dan dasar nominal FAB.
3. Membuka aplikasi pesan otomatis dicatat sebagai pengiriman berhasil.
4. Modal belum mengelola fokus keyboard, penguncian scroll, dan modal bertumpuk.
5. Tanggal bawaan memakai UTC; berpotensi berbeda dengan tanggal lokal operator.
6. Form CRUD menempatkan tombol simpan di luar form sehingga validasi HTML dapat terlewati.
7. Filter pelanggan membandingkan ID angka dengan string secara tidak konsisten.
8. PKS/FAB memerlukan kop, tabel nominal, ruang tanda tangan, dan aturan pemisahan halaman cetak yang lebih rapi.
9. Persentase di dokumen dibulatkan ke bilangan bulat meskipun perhitungan FAB mendukung pecahan.
10. Nilai kosong, gagal memuat dan tindakan berikutnya belum dibedakan dengan baik.
11. Penjagaan halaman login sekarang memakai middleware global Nuxt, sehingga dijalankan sebelum halaman operasional dirender.

## Implementasi

- Sistem tampilan bersama untuk login, navigasi, dashboard, mitra, pelanggan, paket, tiket, keuangan, dokumen dan pengaturan.
- Pusat dokumen dengan pilihan PKS/FAB, pencarian, filter bulan, jumlah arsip dan paginasi 10 dokumen.
- Form terbit menampilkan mitra, tanggal, masa PKS, rekap tagihan lunas, nilai FAB dan peringatan dokumen FAB untuk periode yang sama.
- Pratinjau memuat HTML melalui header autentikasi, lalu menampilkannya dalam iframe `srcdoc` tanpa eksekusi skrip template. Token tidak lagi ditempelkan ke URL dari antarmuka.
- Pengiriman menggunakan draf; pencatatan hanya dilakukan setelah operator memilih “Saya sudah mengirim”. Tidak ada pesan dikirim selama pengujian.
- Validasi form, asosiasi label, status gagal memuat, tanggal lokal, fokus dialog, scroll lock dan penanganan Escape untuk dialog bertumpuk.
- Kop dokumen dan footer, tipografi berbeda untuk PKS/FAB, nominal rata kanan, header tabel berulang dan blok tanda tangan yang tidak dipisah saat cetak.

## Catatan lanjutan

- Template tersimpan tetap dipertahankan. Gaya cetak bersama diperbarui; operator dapat memilih template bawaan melalui Pengaturan.
- Pengiriman WhatsApp/email merupakan pembukaan draf. PDF masih perlu disimpan dan dilampirkan oleh operator.
- Data historis pengiriman sebelumnya tidak membuktikan pesan telah diterima.
- Perhitungan tagihan lama masih mengacu relasi pelanggan/paket saat ini. Snapshot invoice yang lebih lengkap merupakan pekerjaan backend lanjutan.
- Snapshot dokumen menyimpan mitra tetapi profil ISP/template mengikuti pengaturan saat dokumen dibuka. Pengarsipan HTML final perlu pekerjaan terpisah bila diperlukan dokumen yang sepenuhnya tidak berubah.

## Validasi

| Pemeriksaan | Hasil |
|---|---|
| `npm run generate` | Lulus, 12 rute statis dihasilkan |
| `go test ./... -v` | Lulus: render PKS/FAB, persentase pecahan, escape data, template khusus, dokumen kosong, validasi tanggal/periode/durasi, penerbitan valid, autentikasi pratinjau |
| `go vet ./...` dan `go build` | Lulus |
| Login dan navigasi aplikasi | Berfungsi pada build lokal |
| FAB pada database uji | Ringkasan Rp336.000 sesuai dokumen yang diterbitkan |
| Arsip 12 dokumen | Halaman kedua menampilkan arsip 11–12; pencarian kembali ke halaman pertama |
| Dialog bertumpuk | Escape menutup pengiriman dan mempertahankan pratinjau |
| Form mitra | Email tidak valid menahan penyimpanan; data mitra uji tetap 4 |
| Filter mitra pelanggan | CV Sinar Data menampilkan 6 pelanggan yang sesuai |
| Filter pembayaran | Pilihan belum dibayar menampilkan 5 tagihan pada data uji |
| Ponsel 390 × 844 | Dashboard, mitra, pelanggan, paket, tiket, keuangan, dan dokumen tanpa luapan horizontal halaman; tabel memiliki scroll sendiri |
| Konsol browser | Tidak ditemukan galat pada pemeriksaan akhir |
| Integritas data setelah restart | Data mitra, pelanggan, paket, tiket, dokumen, dan seluruh pengaturan identik dengan snapshot sebelum restart |

Build akhir aktif di http://localhost:8080. Pengujian penerbitan memakai database terpisah `.audit/ui-test.db`; server uji dihentikan setelah selesai. Cadangan sumber dan binary lama serta salinan database sebelum restart ada di `.audit/`.

Build Nuxt masih mengeluarkan peringatan dependensi internal `cache-driver.js` saat prerender; proses selesai dengan exit code 0 dan aplikasi statis telah diperiksa di browser. Tidak dilakukan upgrade dependensi.

Pratinjau HTML PKS/FAB dan pemanggilan aksi cetak diperiksa. Dialog cetak sistem dan berkas PDF akhir tidak diekspor oleh pengujian browser tertanam; paginasi PDF fisik masih perlu diperiksa saat operator menyimpan PDF.

Bukti visual: `docs/screenshots/pusat-dokumen.png`, `docs/screenshots/fab-preview.png`, `docs/screenshots/pks-preview.png`, dan `docs/screenshots/mobile-dokumen.png`.
