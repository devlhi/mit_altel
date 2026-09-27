# Redesign ketiga — ruang kerja Ali Telematika

## Diagnosis visual

Versi sebelumnya masih didominasi kotak seragam, sidebar gelap yang berat, teks kecil, dan jarak yang tidak memberi hierarki kuat. Dashboard belum memiliki pusat perhatian yang membantu operator menentukan prioritas. Dokumen tampak seperti daftar administrasi umum.

## Rencana desain

- Kanvas `#f6f7fb`, permukaan `#ffffff`, tinta `#18223d`, biru utama `#405de6`, garis `#e6e9f1`, hijau status `#168461`.
- Tipografi: Aptos/Segoe UI lokal; judul 30–34 px, angka utama 32 px, isi 13–14 px. Angka memakai tabular numerals. PKS tetap memakai serif untuk isi pasal.
- Navigasi putih yang tenang, identitas ringkas, pencarian halaman dengan Ctrl+K, profil di dasar sidebar.
- Dashboard: empat metrik ringkas, grafik pendapatan lebar, progres penagihan di sebelahnya, lalu pekerjaan terbuka dan rekap mitra. Semua grafik memakai nilai API.
- Dokumen: pemilihan format memakai visual lembar PKS/FAB; arsip tetap menjadi ruang kerja utama, tersedia tampilan daftar/kartu. Pratinjau dengan konteks penerima dan tindakan yang jelas.
- Responsif: sidebar drawer, kartu dokumen satu kolom, tabel mempunyai area scroll sendiri, modal tetap terjangkau.

```text
Sidebar putih | Pencarian halaman                    Tanggal
              | Judul halaman                  Tindakan utama
Dashboard     | Metrik operasional dalam satu baris
              | Grafik pendapatan (lebar) | Progres penagihan
              | Pekerjaan terbuka        | Dokumen terbaru
Dokumen       | Format PKS + lembar       | Format FAB + lembar
              | Filter, pencarian, pilihan daftar/kartu
              | Arsip dokumen
```

## Review arah

Warna saja tidak cukup untuk menanggapi penolakan desain sebelumnya. Struktur dashboard dan halaman dokumen harus berubah. Hindari banner promosi dan dekorasi pada tiap kartu; gunakan visual kertas hanya pada fitur dokumen. Pertahankan perilaku validasi, perhitungan, dan konfirmasi pengiriman yang telah diperbaiki.

## Verifikasi

- `npm run generate`: lulus; 12 rute berhasil dibangun. Ada pemberitahuan SPA (`ssr: false`) dan peringatan bundler Nitro yang tidak menghentikan build. Log tersimpan di `.audit/frontend-build-v3-final.log`.
- `go test ./...`: lulus. Mencakup render PKS/FAB, escaping, persentase pecahan, template khusus, validasi penerbitan, dan autentikasi pratinjau. Tidak ada transaksi atau dokumen uji yang dibuat di database aktif.
- Browser: dashboard, mitra, pelanggan, paket, tiket, keuangan, pengaturan, dan dokumen diperiksa. Grafik dapat memilih bulan dan menampilkan nominal yang sesuai. Pintasan dokumen terbaru membuka dokumen yang benar.
- Arsip: perpindahan kartu/daftar dan pencarian nomor dokumen bekerja. Preferensi tampilan disimpan pada browser. Tombol Buat FAB membuka formulir; memilih CV Sinar Data menampilkan bagi hasil September sebesar Rp336.000.
- Keyboard: Ctrl+K membuka pencarian dan langsung memfokuskan input; Enter membuka hasil; Escape menutup dialog. Label pencarian tetap terbaca teknologi bantu pada layar kecil.
- Responsif: diperiksa pada lebar 1440, 390, dan 320 px. Tidak ada overflow horizontal pada halaman dashboard. Drawer ponsel terbuka penuh; footer formulir tetap berada dalam layar. Tabel dan dokumen lebar memakai scroll di area masing-masing.
- PKS/FAB: kop, label jenis, warna tabel, dan tipografi baru terlihat di pratinjau HTML A4. Isi template tersimpan dan nilai dokumen tetap digunakan. Dialog cetak/PDF sistem tidak diuji ulang pada iterasi ini.
- Integritas: tujuh respons API (mitra, pelanggan, paket, tiket, dokumen, pengaturan, keuangan September) identik sebelum dan sesudah pembaruan. Hasil: `.audit/data-integrity-v3.json`.
- Pemulihan: sumber sebelumnya ada di `.audit/before-redesign-v3.zip`, binary sebelumnya di `.audit/isp-api-before-v3.exe`, database sebelum restart di `.audit/database-before-v3/`.
- Aplikasi aktif: `http://localhost:8080`, menyajikan `frontend/.output/public` dengan backend baru. Screenshot akhir ada di `docs/screenshots/v3-*.png`.
