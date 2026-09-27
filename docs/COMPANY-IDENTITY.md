# Identitas PT Ali Telematika Nusantara

Ditinjau pada 27 September 2026. Permintaan: mencari identitas perusahaan dari sumber publik dan menerapkannya pada invoice serta profil aplikasi.

## Data dan sumber

| Data | Nilai yang diterapkan | Sumber |
| --- | --- | --- |
| Nama badan usaha | PT Ali Telematika Nusantara | Situs perusahaan dan APJII |
| Merek | ALTELMATARA NETWORK | Situs perusahaan dan APJII |
| Alamat lengkap | Perumahan Surya Pagi, Desa Tuladenggi, Kec. Telaga Biru, Kab. Gorontalo, Provinsi Gorontalo 96137 | APJII; desa, kecamatan, kabupaten, provinsi dan kode pos juga cocok dengan situs perusahaan |
| Telepon layanan | +62 822-9816-8300 | Situs perusahaan |
| Email layanan | cs@altelmatara.id | Alamat email publik pada situs perusahaan, didekode dari obfuscation Cloudflare |
| Situs web | altelmatara.id | Situs perusahaan, APJII dan PeeringDB |

- Situs perusahaan: https://altelmatara.id/
- Direktori APJII: https://web.apjii.or.id/anggota/pengguna_ip?legality=&name=&page=213
- PeeringDB, pencocokan badan usaha/domain: https://www.peeringdb.com/net/41741
- Aset logo asli: https://altelmatara.id/assets/images/logo.png

APJII mencantumkan nomor berbeda, +62 822-9261-8190. Aplikasi menggunakan nomor layanan dari situs perusahaan; keduanya tidak digabungkan atau dianggap setara tanpa konfirmasi perusahaan.

## Penerapan

- Logo PNG asli disimpan di `frontend/public/branding/altelmatara-network.png`, 3001 × 516 piksel, tanpa perubahan warna atau proporsi. Invoice memuat salinan lokal, sehingga tidak bergantung pada situs eksternal saat preview/cetak.
- Logo dipakai pada invoice, sidebar, login, profil pengaturan, dan miniatur dokumen. Kop invoice menampilkan nama badan usaha, alamat lengkap, telepon dan email; situs terdapat pada footer.
- Profil aktif disimpan melalui API pengaturan. Nilai kota dibuat ringkas (`Gorontalo`); alamat lengkap tetap menyertakan kabupaten, provinsi dan kode pos. Invoice tidak menambahkan kota lagi ketika sudah tercantum dalam alamat.
- DefaultSettings pada source backend disamakan untuk instalasi baru. Profil aktif langsung diperbarui melalui API tanpa restart layanan.
- NPWP semua nol dan rekening BRI contoh dikosongkan karena bukan data resmi terverifikasi. Invoice belum lunas mengarahkan pelanggan menghubungi perusahaan ketika rekening belum diisi.
- Penanda tangan dan jabatan yang sebelumnya tersimpan dipertahankan; keduanya belum diverifikasi dari sumber publik. Template PKS/FAB dan data transaksi tidak diubah. Kop PKS/FAB masih menggunakan renderer backend yang ada; perubahan logo pada revisi ini berfokus pada invoice dan antarmuka aplikasi.
- Cadangan pengaturan sebelumnya: `.audit/settings-before-official-company.json`. Cadangan source: `.audit/before-official-branding/`.

## Pemeriksaan

- Tes invoice mencakup nominal tagihan, status lunas, escaping input, rekening kosong dan penanganan alamat lengkap.
- Build frontend: `npm run generate`. Log: `.audit/official-branding-build.log`.
- Backend: `go test ./...`.
- Preview invoice: logo berhasil dimuat dengan ukuran asli 3001 × 516; dokumen 794 × 1123 piksel tanpa overflow horizontal. Tampilan ponsel diperiksa pada 390 × 844.
- Nominal dan status pada invoice lunas/belum lunas tetap sesuai data tersimpan.
- Screenshot: `docs/screenshots/invoice-official-brand.png`, `docs/screenshots/invoice-official-brand-mobile.png`.
- Dialog cetak dan hasil ekspor PDF tidak diuji pada revisi ini.
