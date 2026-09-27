# Pembaruan neraca, pendapatan mitra, dan MPWA

## Neraca

Buka Pembukuan, tetapkan saldo awal sebenarnya, lalu sinkronkan tagihan. Tab Neraca menunjukkan posisi kumulatif hingga tanggal yang dipilih. Tampilan mengikuti pola laporan [Kledo](https://kledo.com/laporan-neraca/): kelompok akun, subtotal, penyembunyian saldo nol, serta nominal biru untuk membuka rincian jurnal. Perbandingan menggunakan akhir bulan sebelumnya. Cetak/PDF dan CSV mencakup seluruh laporan, termasuk akun dalam kelompok yang ditutup.

Laba tahun berjalan dihitung dari 1 Januari; hasil usaha tahun sebelumnya ditampilkan terpisah dari saldo laba awal. Kelengkapan neraca tetap bergantung pada saldo awal dan jurnal yang benar. Belum ada integrasi akun Kledo, klasifikasi jatuh tempo pinjaman, atau impor data Kledo.

## BHP/USO dari pendapatan mitra

- **Total tagihan mitra** (bawaan): nominal tagihan tersimpan menurut tahun periode layanan, sebelum bagi hasil, termasuk tagihan belum lunas.
- **Pembayaran diterima**: nominal tagihan lunas menurut tahun tanggal pembayaran. Ini pilihan pemantauan internal, bukan penetapan basis pelaporan resmi.
- Rincian sumber tersedia per mitra, per bulan, dan per tagihan. FAB tidak dijumlahkan ulang. Tidak memerlukan jurnal pembukuan.
- Tagihan baru menyimpan mitra saat diterbitkan. Identitas mitra tagihan lama dibekukan saat migrasi dan ditandai sebagai data lama.
- Simpan parameter tahun untuk mengingat pilihan basis, tarif, dan penyesuaian. Tarif bawaan tetap BHP 0,50% serta USO 1,25%; bagian mitra tidak dipotong otomatis.

## Menghubungkan MPWA

1. Buka **Pengaturan → WhatsApp gateway · MPWA**.
2. Isi URL dasar API MPWA (misalnya domain gateway; sertakan `/api` hanya jika instalasi memang memakai awalan tersebut), nomor sender yang terhubung, dan API key. API key kosong saat pengeditan mempertahankan kunci lama.
3. Aktifkan pengiriman dan klik **Simpan MPWA**. Menyimpan tidak melakukan tes kirim atau mengirim pesan apa pun.
4. Tambahkan nomor WhatsApp di **Pelanggan → Edit**. Nomor mitra memakai kontak pada **Mitra**. Nomor lokal `08...` dinormalisasi menjadi `628...`.
5. Buka **Keuangan → Preview invoice → Kirim WhatsApp** atau **Dokumen → Kirim WhatsApp**. Pesan umum tersedia pada menu **WhatsApp → Pesan baru**. Nomor dan pesan bisa ditinjau sebelum menekan Kirim sekarang.

Gateway menggunakan POST form pada `/send-message` serta `/send-media`, dengan `api_key`, `sender`, `number`, dan `message` atau `caption`, sesuai [contoh API resmi MPWA](https://docwa.m-pedia.my.id/docs/wa-gateway-md/v5.5.0/example-api-class?lang=id). Versi/instalasi gateway dapat memiliki prefix endpoint berbeda. TLS tetap diverifikasi dan redirect tidak diikuti agar API key tidak diteruskan ke domain lain.

### Pilihan isi pesan

- **Ringkasan teks**: bisa digunakan dari aplikasi lokal tanpa domain publik.
- **Teks + tautan dokumen**: aplikasi membuat salinan HTML invoice/PKS/FAB dengan tautan acak, berlaku 7 hari. Identitas, nominal, rekening, dan status invoice sesuai saat pengiriman. Perubahan pembayaran setelah pengiriman tidak memperbarui salinan lama; buat kiriman baru bila perlu.
- **Teks + lampiran PDF**: simpan PDF dari preview terlebih dahulu, lalu pilih file pada formulir pengiriman (maksimal 8 MB). MPWA mengambil file melalui tautan acak yang berlaku 24 jam. PDF tidak dibuat otomatis oleh backend.

Untuk tautan dan PDF, isi **URL publik aplikasi ISP** dengan domain HTTPS yang benar-benar mengarah ke aplikasi ini. `localhost` tidak bisa dibuka oleh pelanggan atau MPWA di server lain. Reverse proxy harus meneruskan `/share/*` ke backend tanpa login; rute ini memeriksa token acak dan masa berlaku, bukan token login administrator. Hindari pencatatan token `/share/*` di log reverse proxy. Backend sendiri menyamarkan path tersebut di log. Tautan dapat ditutup dari detail riwayat. File yang sudah diunduh atau pesan yang sudah diterima tidak dapat ditarik melalui fitur ini.

### Riwayat dan hasil kirim

- **Diterima MPWA**: gateway menerima permintaan. Bukan bukti sudah sampai atau dibaca; periksa status di gateway.
- **Ditolak**: gateway menyatakan gagal atau menolak permintaan. Periksa pengaturan, sender, nomor, dan kuota.
- **Belum pasti**: timeout, respons tidak dikenali, gangguan server, atau aplikasi berhenti saat mengirim. Periksa riwayat MPWA sebelum mengirim ulang.

Permintaan yang sama memakai identitas pengiriman yang sama sehingga klik ulang atau pemeriksaan hasil tidak mengirim ganda. Aplikasi tidak mencoba ulang otomatis. Preview invoice/dokumen yang sudah berubah ditolak saat dikirim agar teks dan dokumen tidak berbeda. Riwayat tetap tersedia setelah perubahan data sumber.

API key disimpan terenkripsi di backend dan tidak dikembalikan ke browser. Cadangkan database beserta `meta` karena kunci enkripsi bergantung pada rahasia aplikasi yang disimpan di sana. Lampiran kedaluwarsa tidak bisa diakses; pembersihan data lampiran lama dilakukan saat aplikasi dimulai.

## Validasi pembaruan

Tes Go mencakup batas tanggal neraca, laba ditahan, rincian saldo, sumber pendapatan mitra, perubahan harga/mitra, tahun pembayaran, validasi kontak/URL, enkripsi dan redaksi API key, autentikasi, tautan kedaluwarsa/dicabut, lampiran PDF, perubahan invoice setelah preview, kirim ganda bersamaan, timeout, penolakan gateway, serta pencegahan redirect kredensial. Semua pengiriman pengujian menuju server tiruan lokal.

Build Nuxt dan tes invoice/pagination dijalankan. Uji UI menggunakan salinan database `.audit/kledo-mpwa-qa.db`, termasuk neraca dengan saldo uji serta tampilan 390 px. Saldo uji dan kredensial gateway tidak dipindahkan ke database utama. Cadangan sebelum pembaruan berada di `.audit/before-kledo-mpwa/`; cadangan sebelum aktivasi di `.audit/before-kledo-mpwa-activation.db`.

Aktivasi lokal selesai pada 27 September 2026. Pemeriksaan SQLite setelah migrasi: `ok`; seluruh nilai kolom lama pada 13 tabel yang diperiksa sama dengan cadangan. Database utama memiliki 22 snapshot mitra tagihan; saldo awal/jurnal, pengaturan MPWA, riwayat pesan, serta tautan kirim tetap kosong. Gateway belum diaktifkan karena URL, API key, dan sender belum diberikan. Tidak ada pesan pengujian yang dikirim ke pelanggan asli.
