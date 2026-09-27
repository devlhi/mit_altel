# Pembukuan internal, BHP/USO, POP, dan dokumen mitra

Diterapkan 27 September 2026 pada aplikasi lokal. Ruang lingkup: operasional internal. Saldo awal, jurnal, POP, dan penyesuaian BHP/USO pada database utama dibiarkan kosong. Pengujian berisi angka contoh hanya menggunakan database salinan.

## Mulai menggunakan

1. Buka **Pembukuan → Isi saldo awal**. Tentukan tanggal mulai dan saldo sebenarnya: kas/bank, piutang, uang muka, peralatan, akumulasi penyusutan, utang, serta modal/saldo laba awal. Total debit harus sama dengan kredit. Konfirmasi saldo nol hanya bila memang benar.
2. Piutang awal mencakup tagihan yang pendapatannya sudah diakui sebelum tanggal mulai. Aplikasi memakai tanggal 1 pada periode tagihan sebagai tanggal pengakuan pendapatan. Bila mulai di tengah bulan, tagihan bulan tersebut termasuk tagihan sebelum tanggal mulai.
3. Klik **Sinkronkan tagihan**. Tagihan sejak tanggal mulai dicatat sebagai debit Piutang dan kredit Pendapatan internet. Pembayaran sejak tanggal mulai masuk Kas (Tunai) atau Bank (Transfer) dengan kredit Piutang. Pembayaran tagihan lama tidak dihitung ulang sebagai pendapatan.
4. Gunakan **Tambah jurnal** untuk biaya, pembelian aset, setoran modal, bagi hasil mitra, pencadangan BHP/USO, dan penyesuaian. Isi debit/kredit dalam rupiah bulat dan referensi bukti. Alokasi POP bersifat opsional.
5. Pilih rentang tanggal dan klik **Tampilkan**. Neraca menunjukkan posisi sampai tanggal akhir; laba-rugi dan arus kas memakai rentang yang dipilih. Hasil usaha kumulatif setelah saldo awal masuk ekuitas neraca.

Saldo awal hanya ditetapkan sekali. Koreksi berikutnya melalui jurnal penyesuaian. Jurnal manual dapat dibatalkan dengan alasan; catatan dan alasan tetap ada di Riwayat. Pembatalan mengeluarkan jurnal tersebut dari seluruh laporan, termasuk periode lalu. Aplikasi belum menerapkan penguncian periode atau pembalikan jurnal pada periode berikutnya.

### Sinkronisasi dan pembayaran

- Sinkronisasi dilakukan secara eksplisit dan aman diulang: jurnal sumber yang sama tidak digandakan. Perubahan tagihan menampilkan jumlah perubahan yang belum dibukukan.
- Membatalkan status lunas pada Keuangan memerlukan sinkronisasi lagi; jurnal penerimaan terkait akan ditandai batal. Jurnal pendapatan tagihan tetap ada.
- Formulir pembayaran di Keuangan menyediakan tanggal sebenarnya dan pilihan Transfer/Tunai. Menyimpan kembali tagihan lunas tanpa tanggal baru mempertahankan tanggal sebelumnya.
- Tagihan yang sudah masuk pembukuan serta pelanggan terkait tidak dapat dihapus. Gunakan penyesuaian dan status nonaktif untuk mempertahankan riwayat.
- Pembelian perangkat pada Data POP tidak otomatis membuat jurnal aset. Estimasi biaya bulanan POP juga bukan transaksi beban.
- Belum ada penyusutan, pajak, pembagian hasil, atau pembayaran regulator otomatis. Catat jurnal sesuai transaksi/bukti sebenarnya. Saldo Bank masih satu akun gabungan, belum dipisahkan per rekening.

### Laporan dan ekspor

- Neraca mengikuti alur laporan Kledo: tanggal posisi, kelompok akun bertingkat, subtotal, sembunyikan saldo nol, perbandingan akhir bulan sebelumnya, serta klik nominal untuk rincian jurnal. Laba ditahan tahun lalu dan laba tahun berjalan dipisahkan. Tersedia CSV serta Cetak / PDF; cetak dan ekspor mencakup seluruh laporan. Pinjaman masih satu kelompok tanpa klasifikasi jatuh tempo.
- Laba-rugi, arus kas, daftar jurnal, dan riwayat perubahan pembukuan/POP/parameter BHP-USO tetap tersedia.
- CSV pembukuan memuat seluruh akun, mutasi, saldo, ringkasan, serta baris jurnal untuk periode yang ditampilkan. Jurnal yang dibatalkan diberi status tersendiri.
- CSV POP dan estimasi BHP/USO tersedia dari halaman masing-masing. Ekspor tidak dibatasi halaman tabel yang sedang terlihat.
- Semua tabel aplikasi baru memakai kontrol Sebelumnya/Berikutnya dan pilihan jumlah baris. Formulir baris jurnal/perangkat adalah editor, bukan tabel laporan.

## BHP penyelenggaraan telekomunikasi dan KPU/USO

Halaman **BHP & USO** mengambil pendapatan langsung dari **tagihan pelanggan per mitra**, tanpa perlu saldo awal atau sinkronisasi jurnal. Basis bawaan adalah total tagihan menurut periode layanan, sebelum bagi hasil, termasuk belum lunas. Pilihan **Pembayaran diterima** memakai tagihan lunas berdasarkan tanggal bayar aktual pada tahun yang dipilih. Pembayaran tahun lalu diterima tahun ini masuk tahun ini pada basis tersebut. Ini alternatif pemantauan internal; basis kas tidak otomatis sama dengan dasar pelaporan regulator.

Rincian per mitra dan per bulan dapat ditelusuri ke setiap tagihan. FAB tidak dihitung lagi karena merupakan rekap tagihan. Harga paket terbaru tidak mengganti nominal sumber. Tagihan baru membekukan identitas mitra saat diterbitkan. Tagihan lama diberi penanda migrasi dan memakai mitra saat migrasi; aplikasi tidak dapat memulihkan penugasan mitra masa lalu yang tidak pernah disimpan. Pendapatan tanpa mitra tetap disertakan dan diberi peringatan. Simpan parameter untuk menetapkan basis tahun tersebut.

Tarif bawaan: BHP 0,50% dan KPU/USO 1,25%. Dasar estimasi = pendapatan tercatat + penyesuaian pendapatan − piutang nyata tidak tertagih yang memenuhi syarat − kewajiban interkoneksi/ketersambungan yang memenuhi syarat. Pengurang perlu referensi bukti; perubahan tarif atau pendapatan perlu alasan. Biaya bandwidth dan bagi hasil tidak otomatis menjadi pengurang. Hasil dibulatkan ke rupiah terdekat; batas pembayaran tahunan ditampilkan 30 April tahun berikutnya.

Sumber resmi yang ditinjau:

- [Permenkominfo 5/2021, terutama Pasal 188–192](https://jdih.komdigi.go.id/produk_hukum/view/id/768/t/peraturan%2Bmenteri%2Bkomunikasi%2Bdan%2Binformatika%2Bnomor%2B5%2Btahun%2B2021).
- [PP 43/2023](https://peraturan.go.id/files/pp-no-43-tahun-2023.pdf).

Ini kalkulasi internal, bukan pelaporan resmi, pembayaran, atau perhitungan BHP frekuensi radio. Menyimpan parameter tidak membukukan beban/utang. Gunakan preset jurnal pencadangan BHP/USO bila diperlukan. Pembayaran dicatat terpisah, misalnya debit Utang BHP dan kredit Bank. Jika ada pendapatan di luar tagihan mitra yang belum tercakup, masukkan penyesuaian dengan alasan dan referensi bukti.

## Data POP

Catat kode unik, nama, alamat, koordinat opsional, status, PIC, kontak, upstream, kapasitas, estimasi biaya, catatan, perangkat, dan pelanggan. Status bersifat manual, tidak menghubungi atau memantau router. Satu pelanggan hanya dapat berada pada satu POP. Untuk memindahkan, lepaskan dari POP lama lalu tambahkan ke POP baru. POP lama dapat diberi status Nonaktif; tidak ada penghapusan otomatis riwayatnya.

## PKS dan FAB

- Logo perusahaan asli disertakan langsung dalam HTML dokumen sehingga preview/cetak tidak bergantung pada unduhan logo dari situs luar.
- Dokumen baru menyimpan salinan identitas perusahaan, data mitra, rekening, masa PKS, dan template ketika diterbitkan. Mengubah Pengaturan/Mitra tidak mengubah isi salinan dokumen tersebut.
- Dokumen lama yang belum mempunyai salinan profil perusahaan masih memakai pengaturan perusahaan saat dibuka. Nilai/rincian serta data mitra yang sudah tersimpan tetap dipertahankan; aplikasi tidak dapat merekonstruksi profil historis yang dahulu belum disimpan.
- Tahun nomor dokumen mengikuti tanggal dokumen. FAB ganda untuk mitra dan periode yang sama ditolak.
- FAB baru mengelompokkan berdasarkan nominal tagihan yang tersimpan, bukan harga paket terbaru. Persentase serta nominal bagi hasil disimpan saat penerbitan.
- **Mitra → Edit → Rekening penerima bagi hasil** untuk rekening tujuan FAB. **Pengaturan → Rekening perusahaan** untuk pembayaran invoice pelanggan.
- FAB memakai susunan ringkas: identitas penerima, periode, rincian tagihan, penerimaan, hak mitra, rekening, dan tanda tangan. Dokumen dengan penanda tangan/rekening penting yang belum lengkap diberi penanda DRAF. Rekap bukan bukti transfer.
- Template FAB bawaan lama dimigrasikan hanya bila masih sama persis dengan template bawaan yang dikenal. Template kustom tetap dipertahankan; gunakan placeholder `{{rekening_mitra}}` untuk tujuan pembayaran bagi hasil.
- Mitra yang memiliki arsip PKS/FAB tidak dapat dihapus otomatis bersama arsipnya; gunakan status Nonaktif.

## Verifikasi dan pemulihan

- `go test ./... -count=1` dan `go vet ./...` berhasil. Tes mencakup debit/kredit, batas tanggal, arus kas, sinkronisasi berulang, pembayaran piutang lama, pembatalan lunas, validasi pengurang, pemisahan pendapatan nontelekomunikasi, penetapan POP unik, nominal FAB, profil dokumen tersimpan, tanggal pembayaran, dan pelestarian template kustom.
- 7 tes invoice dan 6 tes pagination berhasil. `npm run generate` berhasil; ada peringatan internal Nitro mengenai `cache-driver` yang tidak menggagalkan build statis.
- Uji browser pada salinan database: saldo awal, sinkronisasi, jurnal biaya, laporan, estimasi BHP/USO, pembuatan POP/perangkat/relasi pelanggan. Preview dokumen diperiksa pada versi aktif; formulir saldo awal diperiksa pada lebar 390 px.
- Salinan konsisten SQLite sebelum aktivasi: `.audit/before-operations-activation.db`. Binary sebelumnya: `.audit/isp-api-before-operations.exe`. Cadangan source awal: `.audit/before-operations-finance/`.
- Pemeriksaan sesudah aktivasi: semua nilai kolom lama pada users/meta/mitra/paket/pelanggan/tagihan/dokumen/tiket/counters sama dengan cadangan, integritas SQLite `ok`. Database utama tidak diisi saldo atau POP contoh.
- Pemulihan database harus dilakukan saat server berhenti dan hanya bila benar-benar diperlukan; memulihkan cadangan dapat menghilangkan transaksi yang dibuat setelah cadangan.
