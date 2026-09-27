# Penyempurnaan berdasarkan screenshot pengguna

## Temuan dan arah

Screenshot pengguna memperlihatkan invoice lunas dengan dua penekanan Rp0, penerbit tidak terlihat di atas, garis metadata berbentuk siku yang memakan ruang, dan teks yang terlalu seragam. Revisi ini memperbaiki hierarki informasi tersebut dengan format dokumen perusahaan.

Acuan sebelumnya tetap tersedia pada [Zoho Standard](https://www.zoho.com/invoice/invoice-templates/sample/zoho-invoice-standard-template.pdf), yang PDF-nya sudah diperiksa. Gunakan posisi penerbit di atas dan data transaksi ringkas. Jangan mengulang pola panel dekoratif, motif jaringan, atau garis vertikal panjang.

- Warna: putih `#ffffff`, navy `#172f50`, biru `#2357ad`, teks sekunder `#62738a`, garis `#dce4ee`, hijau status `#227553`.
- Tipografi: Segoe UI mengikuti aplikasi, fallback Arial. Nama perusahaan 23 px, invoice 28 px, total utama 32 px, data 12–14 px. Bobot 600 hanya untuk informasi utama.
- Layout: kop penerbit dan nomor; ringkasan total/status dalam satu bidang horizontal; penerima dan metadata sejajar; tabel; informasi pembayaran kiri, perhitungan kanan; kontak bantuan di bawah.

```text
Logo  Ali Telematika                      Invoice
Alamat / kontak                          Nomor
-------------------------------------------------
Total tagihan                   Lunas / Belum bayar
Rp220.000                       Periode layanan

Pelanggan                       ID / mitra
Nama, alamat

Layanan                    Qty    Tarif    Jumlah
-------------------------------------------------
Paket

Pembayaran diterima              Total tagihan
Tanggal / metode                 Sudah dibayar
                                 Sisa pembayaran Rp0

Kontak perusahaan
```

## Review sebelum implementasi

Saldo nol bukan informasi utama untuk invoice yang sudah lunas. Total asli tetap prominen dan saldo tetap hadir pada rincian untuk menjaga kebenaran finansial. Pemisah hanya mengelompokkan data; tidak ada panel kosong setinggi metadata. Identitas ISP tidak lagi tersembunyi di footer.

## Validasi

- `npm run test:invoice`: 6 tes lulus. `npm run generate`: lulus; log `.audit/invoice-refined-build.log`.
- Invoice Andi Pratama: total utama Rp220.000, status Lunas, sisa di rincian Rp0. Kop perusahaan langsung terlihat di atas.
- Invoice Bpk. Joko: total Rp150.000, status Belum dibayar, sisa Rp150.000.
- Desktop: dokumen 794 × 1123 px, tidak ada overflow. Ponsel 390 × 844: lebar halaman tetap 390 px, preview dan tombol cetak terlihat.
- Screenshot: `docs/screenshots/invoice-refined-paid.png`, `invoice-refined-full.png`, `invoice-refined-mobile.png`.
- Tidak mengubah data operasional. Dialog cetak dan PDF belum diverifikasi pada revisi ini.
