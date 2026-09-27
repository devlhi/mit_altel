# Invoice berbasis referensi eksternal

## Referensi yang benar-benar diperiksa

Pada revisi ini contoh PDF diunduh dari situs resmi, dirender dengan Poppler, dan diperiksa secara visual:

- [Zoho Minimalist](https://www.zoho.com/invoice/invoice-templates/sample/zoho-invoice-minimalist-template.pdf) — acuan utama: penerima di kiri atas, judul kanan, metadata di bawah garis aksen, saldo terhubung garis vertikal, tabel ringan, pembayaran dalam bidang netral, identitas penerbit di footer.
- [Zoho Grand](https://www.zoho.com/invoice/invoice-templates/sample/zoho-invoice-grand-template.pdf) — dibandingkan, tetapi tidak dipilih karena kop gelapnya terlalu dekat dengan desain yang baru ditolak pengguna.
- [Zoho Standard](https://www.zoho.com/invoice/invoice-templates/sample/zoho-invoice-standard-template.pdf) — dibandingkan untuk hierarki informasi, tetapi kop logo/judul juga terlalu dekat dengan versi terdahulu.

## Arah yang dipilih

Adaptasi struktur Minimalist dengan warna biru sesuai preferensi pengguna. Tidak menyalin logo, nama, data, syarat, pajak, atau metode pembayaran contoh. Saldo utama tetap dihitung dari data tersimpan.

- Warna: putih `#ffffff`, biru `#155ac0`, teks `#273344`, sekunder `#687584`, garis abu `#e0e5eb`, bidang pembayaran `#f6f8fb`.
- Tipografi: Arial dengan bobot normal untuk judul 40 px; nama pelanggan 20 px; data 12–13 px; saldo 18–21 px. Hirarki tidak memakai judul poster, badge, atau kotak kartu.
- Layout: penerima kiri dan invoice kanan; metadata kiri dan saldo kanan, dipisahkan garis siku biru; tabel penuh; pembayaran dalam pita netral; identitas perusahaan bawah.

```text
Ditagihkan kepada                         Invoice
Nama / alamat                             Nomor
-------------------------------------------------
Periode         September      | Sisa bayar  Rp…
ID pelanggan    PLG-…          |
Mitra / status                 |

Layanan                Qty       Tarif     Jumlah
-------------------------------------------------
Paket internet

[ Rekening                Total / sudah dibayar  ]
[ Konfirmasi              Sisa pembayaran   Rp…  ]

Catatan
-------------------------------------------------
Logo  Ali Telematika                Kontak / web
```

## Review sebelum implementasi

Perubahan komposisi mengikuti contoh eksternal secara nyata: penerbit berpindah ke bawah, saldo masuk area metadata atas, tabel tanpa header blok, dan tidak ada motif jaringan. Ini berbeda dari versi sebelumnya yang selalu diawali kop biru. Tetap gunakan invoice A4 dan preview lebar yang sudah berfungsi.

## Verifikasi

- `npm run test:invoice`: 6 tes lulus; nominal tersimpan, nomor stabil, status lunas, informasi opsional, dan escaping tetap benar.
- `npm run generate`: lulus; log `.audit/invoice-reference-build.log`.
- Review desktop: kertas 794 × 1123 px tanpa overflow. Mode Pas lebar menampilkan komposisi referensi; Pas layar menampilkan keseluruhan dokumen termasuk identitas penerbit bawah.
- Review ponsel 390 × 844: lebar dokumen aplikasi tetap 390 px, tidak ada gulir bertingkat pada modal. Saldo atas dan bawah invoice lunas sama-sama Rp0.
- Screenshot: `docs/screenshots/invoice-reference.png`, `invoice-reference-full.png`, dan `invoice-reference-mobile.png`.
- Tidak ada penulisan data operasional selama revisi. Dialog cetak/PDF belum diuji untuk desain ini.
