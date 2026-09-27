# Profil kanan atas dan aksen sudut kartu

Permintaan pengguna: pindahkan blok Administrator beserta tombol keluar dari sidebar ke sudut kanan; tambahkan gradient sudut pada seluruh kartu.

## Arah visual

- Palet: putih `#ffffff`, kanvas `#f6f7fb`, teks navy `#18223d`, biru `#2785d9`, hijau `#329f9b`, garis `#e6e9f1`.
- Tipografi: tetap Aptos/Segoe UI; nama akun 12 px, peran 10 px.
- Susunan: `[navigasi / halaman] ... [pencarian] [tanggal] [avatar + nama + keluar]`.
- Kartu: biru-hijau di kanan atas, memudar ke warna dasar; posisi dan intensitas konsisten, tanpa animasi.

Tinjauan brief: gradient diminta secara eksplisit. Warna mengikuti logo ALTELMATARA dan dibatasi di sudut, sehingga informasi utama tetap mudah dibaca. Latar gelap PKS memakai tingkat terang yang sesuai.

## Perubahan

- Profil menjadi satu blok di ujung kanan header. Avatar/nama membuka pengaturan, tombol keluar tetap tersendiri.
- Pada ponsel, tampil avatar dan tombol keluar; nama tetap tersedia dalam label aksesibilitas.
- Gradient dibagikan ke kartu umum, metrik dashboard, kartu/preview dokumen, tautan invoice, ringkasan formulir, pilihan jenis dokumen, kartu workspace dan panel login.
- Header tabel dibuat sedikit transparan agar gradient kartu tidak tertutup. Aksen berupa CSS background dan tidak menangkap klik atau mengubah urutan fokus.
- Dokumen invoice/cetak memiliki stylesheet tersendiri dan tidak terkena aturan kartu aplikasi.

Cadangan sebelum perubahan: `.audit/before-account-corner-gradient/`.

## Validasi

- `npm run generate` lulus; log `.audit/account-corner-gradient-build.log`.
- Browser desktop: satu blok profil dalam header, tidak ada profil di sidebar, tepat satu tombol keluar.
- Seluruh 9 kartu/metrik dashboard memakai gradient. Kartu PKS, FAB, preview arsip, tautan invoice dan kartu Pengaturan juga diperiksa.
- Tautan avatar/nama berhasil membuka `/pengaturan`.
- Viewport 320, 390 dan 901 px: tidak ada overflow horizontal atau tumpang tindih profil dan navigasi. Override viewport dikembalikan setelah pemeriksaan.
- Screenshot: `docs/screenshots/dashboard-account-gradient.png` dan `docs/screenshots/account-gradient-mobile.png`.
