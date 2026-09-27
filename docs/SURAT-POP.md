# Surat POP Komdigi

## Penggunaan

Buka **Surat POP**, isi **nomor surat**, pilih **KBLI**, lalu isi **alamat POP**. Hanya tiga isian tersebut yang diperlukan. **Tambah alamat POP** menambahkan lokasi lain dalam surat yang sama, hingga 30 alamat.

Pilihan: ISP 61104, jaringan tetap lokal fiber optik 61101, jaringan tetap tertutup fiber optik 61101, serta kode KBLI manual lima digit. Kode yang dikenal memakai nama izin dan uraian katalog server. Kode lain dicantumkan dengan redaksi umum penyelenggaraan telekomunikasi tanpa mengarang uraian kegiatan.

Tanggal otomatis mengikuti hari ini di WITA. Kota, tujuan, alamat tujuan, nama penandatangan, dan jabatan menggunakan profil surat atau pengaturan perusahaan. Nomor yang diisi adalah nomor surat baru dan wajib unik. Tidak perlu mengisi nama POP, riwayat OSS, atau rujukan surat terdahulu.

**Pratinjau surat** menampilkan isi lengkap, kop, dan TTD/stempel dengan penanda DRAF. **Terbitkan surat** menyimpan nomor serta snapshot surat. **Simpan PDF / cetak** pada arsip membuka dialog cetak browser: pilih Save as PDF dan matikan header/footer.

Arsip mendukung pencarian, paginasi, pratinjau/cetak, salin alamat dan KBLI ke surat baru, serta pembatalan dengan alasan. Salin isi mengosongkan nomor agar diisi nomor baru. Surat batal tetap tersimpan dan nomornya tidak dipakai ulang. Arsip tidak berubah ketika profil diperbarui.

## Isi dan rancangan

Satu bidang untuk tiga isian; kop dan ringkasan otomatis menjadi informasi baca saja di sampingnya. Warna: biru #2064b6, teal #299c87, tinta #182338, putih #ffffff, bidang #f3f6fa. Tipografi formulir mengikuti aplikasi; surat memakai Times New Roman 12 pt rata kiri/kanan. Kop asli menjadi identitas utama dan aksen gradient sudut mengikuti permintaan pengguna.

Redaksi mengikuti contoh pengguna: tujuan lengkap, perihal lanjutan verifikasi izin di luar komitmen PT Ali Telematika Nusantara, tindak lanjut permohonan, proses verifikasi dan persetujuan melalui OSS, jenis izin dan KBLI, daftar alamat bernomor, penutup, serta TTD/stempel. Kalimat OSS otomatis mengikuti contoh atas permintaan pengguna. Pembuka mengacu pada permohonan tanpa mengarang nomor/tanggal surat terdahulu. Identitas dan fakta perusahaan lain pada contoh tidak disalin.

Kop dan TTD/stempel asli disimpan sebagai aset privat backend yang memerlukan login, tanpa regenerasi atau pengolahan gambar. Fitur membuat dokumen perusahaan; tidak mengirim ke Komdigi atau menerbitkan izin pemerintah.

## Implementasi dan verifikasi

Mode template menerima tiga isian, menyelesaikan tanggal/profil/redaksi di server, lalu mengembalikan snapshot bersama HTML pratinjau. Penerbitan memakai snapshot tersebut agar profil/tanggal tetap sesuai yang diperiksa pengguna. Nomor unik dan request key idempoten mencegah surat ganda. API mode awal/lanjutan lama tetap didukung; HTML arsip lama tidak diubah.

Tes mencakup payload minimal, otomatisasi tanggal/profil/redaksi, ketiadaan rujukan palsu dan placeholder, input wajib, KBLI manual, escaping alamat, alamat ganda, snapshot setelah profil berubah, nomor unik, idempotensi, dan regresi sebelumnya. Seluruh tes Go, go vet, serta build frontend lulus. Pemeriksaan visual memakai salinan database di port 8087.

## Aktivasi

Aplikasi utama port 8080 sudah memakai backend terbaru pada 27 September 2026. Error formulir baru dengan backend lama diselesaikan dengan mengaktifkan binary yang sesuai. Pratinjau tiga isian diperiksa langsung pada database operasional tanpa menerbitkan surat uji; seluruh isi 23 tabel tetap sama.

Nama pilihan KBLI berasal dari description katalog server yang sama dengan isi surat, mengikuti judul OSS KBLI 2025. Lingkup jaringan tetap lokal/tertutup ditandai tersendiri agar dua pilihan 61101 tidak tertukar. Tautan sumber OSS tersedia di bawah pilihan. Salinan database uji .audit/pop-draft-preview-qa.db dipertahankan; gunakan aplikasi utama 8080 untuk operasional.

## Referensi

Pengguna mengonfirmasi KBLI 61104, 61101, serta pilihan manual. Katalog diperiksa 27 September 2026:
- https://e-telekomunikasi.komdigi.go.id/etelekomunikasi/layanan
- https://oss.go.id/en/kbli/detail/36273dfa-4e91-5587-acec-509799b2a3fe
- https://oss.go.id/en/kbli/detail/5c4bfd55-4ab5-5290-a110-977bf8922aa1
- https://jdih-storage.bkpm.go.id/jdih/jdih/Booklet-Klasifikasi-Baku-Lapangan-Usaha-Indonesia-Tahun-2025-.pdf
