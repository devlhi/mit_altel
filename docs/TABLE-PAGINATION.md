# Pagination semua tabel aplikasi

Kontrol bersama `TablePagination.vue` menyediakan Sebelumnya, Berikutnya, posisi halaman, rentang data, total hasil dan pilihan jumlah baris. Logika pemotongan data berada di `usePagination.ts`.

## Cakupan

1. Dashboard: kontribusi mitra.
2. Mitra: daftar mitra, pelanggan dalam detail mitra, dokumen dalam detail mitra.
3. Pelanggan, paket layanan, tiket gangguan.
4. Keuangan: invoice pelanggan dan rekap mitra, dengan halaman yang independen.
5. Dokumen: mode kartu dan daftar menggunakan state halaman yang sama.
6. Pengaturan: tabel placeholder.

Total: 11 tabel interaktif. Default 10 baris dengan pilihan 5/10/25/50. Arsip dokumen mempertahankan 9 item untuk grid, dengan pilihan 9/18/36. Kontrol tetap ditampilkan saat hasil kosong atau hanya satu halaman; tombol dinonaktifkan jika tidak ada halaman tujuan.

Pencarian dan filter diterapkan ke seluruh koleksi sebelum pagination. Perubahan filter, periode atau mitra detail kembali ke halaman pertama. Perubahan jumlah baris juga kembali ke halaman pertama. Jika data terakhir dihapus, posisi halaman menyesuaikan jumlah halaman yang tersisa.

Footer berada di luar area scroll horizontal tabel. Nilai ringkasan, rekap keuangan, dokumen cetak dan invoice tetap menggunakan seluruh data yang relevan. Implementasi ini membagi tampilan di frontend; API masih mengambil seluruh koleksi seperti sebelumnya.

## Validasi

- 6 tes pagination: urutan/kelengkapan data antarahalaman, filter dengan jumlah hasil sama, penghapusan halaman terakhir, perubahan ukuran, hasil kosong, dan state antartabel/detail yang independen.
- 7 tes invoice tetap lulus. `npm run generate` lulus; log `.audit/table-pagination-build.log`.
- Browser pelanggan: 1–10 dari 12 → 11–12 dari 12, kembali ke 1–10, ukuran 5, pencarian dari halaman kedua kembali ke 1–1, hasil kosong 0–0 dengan kedua tombol nonaktif.
- Browser keuangan: halaman kedua berisi tagihan ke-11; rekap mitra tetap pada halaman pertama. Filter Belum kembali ke halaman pertama.
- Browser detail mitra: ukuran 5 dan halaman kedua pelanggan menampilkan data ke-6; halaman dokumen tidak ikut berubah.
- Paket, tiket, mitra, arsip daftar, dan placeholder diperiksa melalui UI. Placeholder ukuran 5 dapat maju ke data ke-6–10.
- Ponsel 390 px: footer tidak overflow dan kontrol berada di luar tabel yang bisa digeser. Screenshot `docs/screenshots/table-pagination-mobile.png`; desktop `docs/screenshots/table-pagination-desktop.png`.
- Cadangan source sebelum perubahan: `.audit/before-table-pagination/`.
