# Invoice minimalis untuk perusahaan

## Rencana desain

Permintaan terbaru: minimalis, modern, dan layak dipakai perusahaan. Pilih komposisi dokumen bisnis yang tenang: identitas penerbit kiri atas, identitas invoice kanan atas, informasi pelanggan/periode dua kolom, tabel tanpa bidang warna, dan total rata kanan.

- Palet: kertas `#ffffff`, tinta `#263241`, teks sekunder `#637080`, garis `#dce1e7`, logo `#334d69`, status lunas `#287355`.
- Tipografi: Arial untuk dokumen offline yang konsisten; judul 32 px, nama penerbit 20 px, nama pelanggan 18 px, data 12–13 px, saldo 29 px. Hierarki melalui ukuran dan ketebalan, bukan warna blok.
- Layout: margin 54 px, kolom penerima dan periode sejajar, nominal tabel rata kanan, pemisah tipis, informasi rekening tanpa kotak dekoratif.

```text
Logo  Perusahaan                    Invoice
Alamat / kontak                     Nomor / status
--------------------------------------------------
Ditagihkan kepada                   Periode layanan
Nama / alamat / ID                  Mitra

Layanan                Qty       Tarif      Jumlah
--------------------------------------------------
Paket internet
--------------------------------------------------
Informasi pembayaran                Total tagihan
Rekening                            Sudah dibayar
                                    -------------
                                    Sisa pembayaran

Terima kasih

--------------------------------------------------
Bantuan                             Website
```

## Review sebelum implementasi

Hapus pita biru besar, judul berukuran poster, titik dekoratif, panel saldo berwarna, dan border rekening. Pertahankan hanya logo kecil dan status lunas sebagai aksen; penekanan utama ada pada total dan data invoice. Ini mengikuti arahan minimalis perusahaan pengguna. Data, nomor invoice, dan keamanan escaping tetap sama.

## Validasi

- `npm run test:invoice`: enam tes lulus; nominal, nomor invoice, status lunas, informasi opsional, dan escaping tetap benar.
- `npm run generate`: lulus, log `.audit/invoice-minimal-build.log`.
- Desktop: preview Bpk. Joko Rp150.000, kertas 794 × 1123 px tanpa overflow; mode Pas lebar menampilkan detail dengan jelas.
- Invoice Andi Pratama lunas menampilkan sisa Rp0.
- Ponsel 390 × 844: dokumen tetap dalam layar, tidak ada overflow halaman, tombol cetak terlihat.
- Screenshot: `docs/screenshots/invoice-minimal.png`, `invoice-minimal-detail.png`, `invoice-minimal-mobile.png`.
- Perubahan hanya desain frontend; tidak mengubah data pelanggan/tagihan. Dialog cetak dan hasil PDF belum diverifikasi pada revisi ini.
