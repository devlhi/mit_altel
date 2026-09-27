# Invoice pelanggan dan preview

Gaya visual terbaru mengikuti [penyempurnaan berdasarkan screenshot pengguna](INVOICE-SCREENSHOT-REFINEMENT.md): kop perusahaan di atas, total tagihan sebagai fokus, dan metadata ringkas. Catatan berikut mendokumentasikan implementasi awal fitur dan aturan data invoice.

Tagihan pelanggan kini memiliki invoice tersendiri di Keuangan dengan tombol **Preview invoice** yang terlihat. Halaman Dokumen juga memberi jalan masuk ke invoice pelanggan.

## Arah desain

- Warna: kertas `#ffffff`, tinta `#202b4b`, biru `#405de6`, garis `#e0e5f0`, teks sekunder `#6d7890`, lunas `#168461`.
- Arial untuk dokumen agar ukuran cetak konsisten; angka besar hanya pada sisa pembayaran. Nama pelanggan dan periode langsung terbaca.
- Kop perusahaan dan nomor invoice berdampingan, dua kolom untuk penerima/periode, tabel layanan, blok total di kanan, instruksi pembayaran di kiri. Penutup tenang dengan kontak ISP. Alignment kiri; angka rata kanan.
- Nominal berasal dari tagihan tersimpan, bukan harga paket saat ini. Nomor stabil berdasarkan periode + ID tagihan. Status lunas menampilkan sisa Rp0 dan tanggal/metode pembayaran yang tercatat.

```text
Logo + perusahaan              Invoice + nomor
Alamat / kontak                Status
------------------------------------------------
Ditagihkan kepada              Periode layanan
Nama, alamat, ID pelanggan      Bulan, mitra
------------------------------------------------
Layanan             Kuantitas     Harga    Jumlah
------------------------------------------------
Informasi pembayaran           Total tagihan
Rekening / konfirmasi          Dibayar
                               Sisa pembayaran
------------------------------------------------
Terima kasih                   Kontak ISP
```

Preview dan cetak memakai HTML invoice yang sama. Saat dibuka, aplikasi mengambil ulang data melalui API Keuangan, Pelanggan, dan Pengaturan yang sudah dilindungi autentikasi. Fitur aktif melalui build frontend dan tidak memerlukan pergantian backend. Tidak membuat tagihan tambahan saat dibuka, atau menambahkan tanggal jatuh tempo, pajak, dan biaya yang belum tersimpan di sistem.

Viewer menyediakan Pas layar (satu halaman), Pas lebar, dan 100%. Pada layar sempit, ringkasan pindah ke atas lembar dokumen; ukuran kertas tetap A4. Nomor invoice dibuat konsisten dari periode dan ID tagihan. Status, identitas, dan rekening mengikuti data terbaru saat preview dibuka.

Implementasi: `frontend/lib/invoice.ts`, `frontend/assets/css/invoice.css`, `frontend/components/InvoicePreview.vue`; tombol pada `frontend/pages/keuangan.vue` dan akses dari `frontend/pages/dokumen.vue`. Semua nilai API di-escape sebelum dimasukkan ke HTML; iframe membatasi skrip dan memakai HTML lokal melalui srcdoc.

## Review

Invoice harus membantu pelanggan memahami apa yang dibayar. Hilangkan kartu promosi, grafik, dan dekorasi lembar kertas dari dokumen invoice. Gunakan nilai sisa pembayaran sebagai penekanan utama; status lunas jelas agar invoice tidak terlihat sebagai tagihan baru.

## Verifikasi

- `npm run test:invoice`: 6 tes lulus. Mencakup nominal tagihan tersimpan, lunas/sisa Rp0, pembatalan pembayaran, escaping, informasi opsional, dan konsistensi nomor.
- `npm run generate`: lulus. Log: `.audit/invoice-build.log`.
- `go test ./...`: lulus; sumber backend kembali sama seperti sebelum pengerjaan invoice.
- Browser aplikasi utama `http://localhost:8080/keuangan`: tombol Preview invoice membuka dokumen Bpk. Joko dengan nominal Rp150.000; invoice Andi Pratama menampilkan Lunas dan sisa Rp0. Tombol cetak aktif setelah iframe selesai dimuat. Mode 100% memakai skala 1 dan Pas lebar menyesuaikan lebar viewer.
- Preview default menampilkan satu halaman lengkap. Screenshot: `docs/screenshots/invoice-preview.png` dan `invoice-lunas.png`.
- Tujuh respons API operasional identik dengan baseline, tercatat di `.audit/invoice-data-integrity.json`. Tidak membuat tagihan, mengubah pembayaran, atau mengirim invoice saat pengujian.
- Layar ponsel 390 × 844: lembar A4 tampil utuh pada Pas layar, tombol cetak terlihat, dan lebar halaman tetap 390 px. Zoom 100% menggulir di dalam viewer tanpa overflow halaman. Screenshot: `docs/screenshots/invoice-mobile.png`.
- Tautan Invoice pelanggan dari Dokumen membuka Keuangan. Pencarian `INV/202609/000018` menampilkan hanya tagihan Bpk. Joko; preview berhasil dibuka dari hasil pencarian.
- Dialog cetak sistem dan hasil PDF fisik belum diverifikasi.
