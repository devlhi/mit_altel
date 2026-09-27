# Referensi API — Ali Telematika ISP

Base URL: `http(s)://<host>/api` · Format: JSON · Server: Golang + SQLite

## Autentikasi

Hampir semua endpoint membutuhkan **Bearer token** (JWT HS256, masa berlaku 7 hari):

```bash
# 1. Login — dapatkan token
curl -X POST https://domain-anda/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"********"}'
# → {"token":"eyJ...","nama":"Administrator","username":"admin"}

# 2. Pakai token
curl https://domain-anda/api/mitra -H "Authorization: Bearer eyJ..."
```

Catatan keamanan:
- 5x gagal login dalam 5 menit (per IP+username) → terkunci 60 detik (`429` + header `Retry-After`).
- Ganti password (`PUT /api/password`) **mematikan semua token lama**.
- Kecuali `/api/login`, dua pengecualian publik: `GET /api/dokumen/{id}/html`
  (terproteksi token query `?token=` untuk iframe) dan `GET /share/{token}`
  (tautan lampiran WhatsApp).

## Konvensi respons

- Sukses: `200`/`201` + data. Gagal: `4xx/5xx` + `{"error":"pesan dalam Bahasa Indonesia"}`.
- Uang selalu angka bulat rupiah (`"jumlah": 150000`).

---

## Auth

| Metode & Path | Fungsi | Body / Query |
|---|---|---|
| `POST /api/login` | Login | `{username, password}` |
| `GET /api/me` | Cek token aktif | — |
| `PUT /api/password` | Ganti password (sesi lama hangus) | `{lama, baru}` (baru ≥ 6 karakter) |

## Dashboard & Pengaturan

| Metode & Path | Fungsi |
|---|---|
| `GET /api/dashboard` | Statistik, grafik 6 bulan, tiket & dokumen terbaru, rekap mitra |
| `GET /api/settings` / `PUT /api/settings` | Baca/simpan profil ISP + template dokumen |
| `GET /api/settings/defaults` | Template PKS/FAB bawaan (untuk tombol reset) |
| `POST /api/seed` | Muat ulang data contoh (menghapus data operasional!) |

## Mitra

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/mitra` | Daftar (+jumlah pelanggan) | `?q=` cari, `?status=` |
| `POST /api/mitra` | Tambah | share 0–100 tervalidasi; kode otomatis `MTR-NNN` |
| `GET /api/mitra/{id}` | Detail: pelanggan, dokumen, rekap periode berjalan | |
| `PUT /api/mitra/{id}` / `DELETE /api/mitra/{id}` | Ubah / hapus | hapus ditolak (`409`) jika masih punya pelanggan |

## Paket & Pelanggan

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET/POST /api/paket`, `PUT/DELETE /api/paket/{id}` | CRUD paket | harga 0–1 triliun, kecepatan ≥ 0; `409` bila dipakai pelanggan |
| `GET /api/pelanggan` | Daftar (+paket, mitra, harga) | `?q=&mitra=&status=` |
| `POST /api/pelanggan`, `PUT/DELETE /api/pelanggan/{id}` | CRUD | hapus ikut menghapus tagihan & tiket pelanggan |

## Tiket

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/tiket` | Daftar | `?q=&status=` |
| `POST /api/tiket` | Buat — nomor otomatis `TKT/TAHUN/NNN` | `{pelanggan_id, jenis, prioritas, status, teknisi, deskripsi}` |
| `PUT /api/tiket/{id}` / `DELETE /api/tiket/{id}` | Ubah (status Selesai mencatat waktu) / hapus | |

## Keuangan

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/keuangan?periode=YYYY-MM` | Tagihan periode + summary + rekap per mitra | |
| `POST /api/keuangan/generate` | Buat tagihan pelanggan Aktif yg belum ditagih | `{periode}`; idempoten |
| `PUT /api/tagihan/{id}` | Lunasi / buka | `{status: "Lunas"\|"Belum", metode}` |
| `DELETE /api/tagihan/{id}` | Hapus tagihan | |

## Dokumen (PKS & FAB)

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/dokumen` | Daftar + status pengiriman | `?q=&jenis=&mitra=` |
| `POST /api/dokumen` | Terbitkan | `{jenis:"PKS"\|"FAB", mitra_id, tanggal, masa_bulan, periode}`; nomor otomatis |
| `GET /api/dokumen/{id}/html` | HTML siap cetak | auth via `?token=`; `&mode=preview` tanpa auto-cetak |
| `POST /api/dokumen/{id}/kirim` | Catat pengiriman | `{via:"wa"\|"email"}` |
| `DELETE /api/dokumen/{id}` | Hapus | |

## Pembukuan

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/pembukuan?from&to` | Neraca (+pembanding bln lalu), laba rugi, arus kas, jurnal periode, audit, jumlah tagihan belum dibukukan | |
| `GET /api/pembukuan/account?account&to` | Buku besar per akun | `account` = kode akun / `profit_current` / `profit_prior` |
| `POST /api/pembukuan/opening` | Saldo awal (sekali) | `{date, lines, zero}`; hanya akun neraca |
| `POST /api/pembukuan/journal` | Jurnal manual | debit=kredit divalidasi |
| `POST /api/pembukuan/journal/{id}/void` | Batalkan jurnal manual | `{reason}` ≥ 5 karakter |
| `POST /api/pembukuan/sync` | Sinkron jurnal dari tagihan | idempoten; mengoreksi bukan mendobel |

## BHP-USO & POP

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/bhp-uso?year&basis` | Hitung BHP/USO | basis `billed`/`paid`; `?basis=` untuk simulasi tanpa menyimpan |
| `PUT /api/bhp-uso` | Simpan parameter | penyesuaian/tarif non-default wajib notes |
| `GET /api/pop` · `POST /api/pop` · `PUT /api/pop/{id}` | CRUD POP + aset + pelanggan terhubung | pelanggan 1 POP saja; koordinat divalidasi |

## Surat POP (Komdigi)

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/surat-pop/options` | Layanan KBLI, profil, kop | |
| `GET /api/surat-pop/sample?service=` | Contoh redaksi | tidak memakan nomor |
| `PUT /api/surat-pop/profile` | Simpan profil surat | NIB 13 digit bila diisi |
| `GET /api/surat-pop` | Arsip surat | |
| `POST /api/surat-pop/preview` | Pratinjau + daftar isian kurang | draft boleh belum lengkap |
| `POST /api/surat-pop` | Terbitkan | `request_key` dari preview (idempoten); nomor otomatis/manual |
| `GET /api/surat-pop/{id}/html` | HTML surat (+watermark bila batal) | Cache-Control no-store |
| `POST /api/surat-pop/{id}/void` | Batalkan dengan alasan | surat terbit tak dapat diubah |

## WhatsApp (MPWA)

| Metode & Path | Fungsi | Catatan |
|---|---|---|
| `GET /api/whatsapp/settings` | Baca konfigurasi | API key tidak pernah dikembalikan (hanya `has_api_key`) |
| `PUT /api/whatsapp/settings` | Simpan + API key (terenkripsi AES-GCM) | gateway wajib HTTPS (kecuali localhost/privat) |
| `GET /api/whatsapp/preview?kind=&id=` | Susun pesan + revisi | kind: `invoice`/`document`/`custom` |
| `POST /api/whatsapp/attachment` | Unggah PDF (maks 8 MB, magic %PDF-) | multipart `file` → `{attachment_id}` |
| `POST /api/whatsapp/send` | Kirim via gateway | `{request_key, kind, source_id, recipient, message, include_link, attachment_id, revision}`; idempoten |
| `GET /api/whatsapp/history?kind=&id=` | Riwayat pengiriman | |
| `POST /api/whatsapp/history/{id}/revoke` | Cabut tautan lampiran | |

## Publik

| Metode & Path | Fungsi |
|---|---|
| `GET /share/{token}` | Tautan lampiran/invoice WhatsApp — token 64 hex acak, kedaluwarsa 24 jam (PDF) / 7 hari (dokumen), bisa dicabut; dilengkapi CSP ketat |

## Kode status umum

`400` isian tidak valid · `401` tanpa/token salah · `404` tidak ditemukan ·
`409` konflik (dipakai/duplikat/sudah diproses) · `429` login terkunci ·
`500` kesalahan server
