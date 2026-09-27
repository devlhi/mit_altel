# Invoice gaya kedua

## Rencana

Invoice layanan internet dengan identitas seperti lembar rekening perusahaan. Dokumen harus terbaca pada preview dan cukup berbeda dari desain sebelumnya: judul besar di kiri, identitas perusahaan di kanan, ringkasan periode/status dalam pita biru penuh, tabel terbuka, dan jumlah akhir lebih tegas.

- Warna: putih `#ffffff`, tinta `#172c42`, biru layanan `#146c94`, biru pucat `#edf5f9`, abu teks `#526775`, hijau lunas `#187454`.
- Tipografi: Trebuchet MS untuk judul dan identitas; Arial untuk data dan angka agar layout offline konsisten. Judul 58 px; total 34 px; data utama 13–16 px.
- Tata letak: teks rata kiri; nominal rata kanan; blok rekening 55% dan total 45%. Jarak vertikal memberi ruang pada rincian tanpa menyisakan sebagian besar halaman kosong.

```text
Invoice                         Logo / perusahaan
Nomor invoice                   Alamat dan kontak

[ Periode layanan                  Status pembayaran ]

Ditagihkan kepada               ID pelanggan / mitra
Nama dan alamat

Layanan               Qty           Tarif     Jumlah
Paket internet

Informasi pembayaran            Total / dibayar
Bank dan konfirmasi             Sisa pembayaran

Terima kasih
-----------------------------------------------------
Kontak bantuan                  Website / nomor
```

## Review sebelum implementasi

Fokus karakter pada pita periode/status yang cocok untuk tagihan bulanan layanan internet. Tidak menambahkan nomor urut dekoratif, tanda tangan palsu, tanggal jatuh tempo, pajak, atau QR pembayaran tanpa data sumber. Kop, nomor, dan total diperbesar untuk mengatasi preview lama yang terasa terlalu kecil. Tabel tanpa header gelap dan kartu saldo tanpa sudut membulat mengubah komposisi, bukan hanya warna.

## Validasi

- `npm run test:invoice`: enam tes lulus; nominal tersimpan, status lunas, escaping, dan nomor invoice tetap benar.
- `npm run generate`: lulus; output aktif pada server lokal. Log `.audit/invoice-style-2-build.log`.
- Review aplikasi desktop: invoice Bpk. Joko Rp150.000, lembar 794 × 1123 px tanpa overflow; invoice Andi Pratama lunas dengan sisa Rp0 dan pita hijau.
- Review ponsel 390 × 844: halaman tidak melebar, kertas utuh dalam mode Pas layar, tombol cetak tetap terlihat.
- Zoom Pas lebar sekarang kembali ke awal dokumen (`scrollTop` dan `scrollLeft` nol); aksen viewer mengikuti warna invoice.
- Screenshot: `docs/screenshots/invoice-style-2.png`, `invoice-style-2-detail.png`, `invoice-style-2-lunas.png`, dan `invoice-style-2-mobile.png`.
- Tidak mengubah data operasional. Dialog cetak sistem dan hasil PDF belum diperiksa pada perubahan gaya ini.
