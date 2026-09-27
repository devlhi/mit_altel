package main

type Mitra struct {
	Rekening      string  `json:"rekening"`
	ID            int64   `json:"id"`
	Kode          string  `json:"kode"`
	Nama          string  `json:"nama"`
	Jenis         string  `json:"jenis"`
	Pemilik       string  `json:"pemilik"`
	Telepon       string  `json:"telepon"`
	Email         string  `json:"email"`
	Alamat        string  `json:"alamat"`
	Wilayah       string  `json:"wilayah"`
	NPWP          string  `json:"npwp"`
	Status        string  `json:"status"`
	ShareMitra    float64 `json:"share_mitra"`
	TanggalGabung string  `json:"tanggal_gabung"`
	Catatan       string  `json:"catatan"`
	// computed (list)
	PelangganAktif int `json:"pelanggan_aktif"`
	TotalPelanggan int `json:"total_pelanggan"`
}

type Paket struct {
	ID         int64  `json:"id"`
	Nama       string `json:"nama"`
	Kecepatan  int    `json:"kecepatan"`
	Harga      int64  `json:"harga"`
	Keterangan string `json:"keterangan"`
	Dipakai    int    `json:"dipakai"`
}

type Pelanggan struct {
	Telepon       string `json:"telepon"`
	ID            int64  `json:"id"`
	Nama          string `json:"nama"`
	Alamat        string `json:"alamat"`
	MitraID       int64  `json:"mitra_id"`
	PaketID       int64  `json:"paket_id"`
	PPPoE         string `json:"pppoe"`
	Status        string `json:"status"`
	TanggalPasang string `json:"tanggal_pasang"`
	Catatan       string `json:"catatan"`
	// computed (list)
	MitraNama string `json:"mitra_nama"`
	PaketNama string `json:"paket_nama"`
	Kecepatan int    `json:"kecepatan"`
	Harga     int64  `json:"harga"`
}

type Tiket struct {
	ID          int64  `json:"id"`
	Nomor       string `json:"nomor"`
	PelangganID int64  `json:"pelanggan_id"`
	Jenis       string `json:"jenis"`
	Prioritas   string `json:"prioritas"`
	Status      string `json:"status"`
	Teknisi     string `json:"teknisi"`
	Deskripsi   string `json:"deskripsi"`
	Dibuat      string `json:"dibuat"`
	Selesai     string `json:"selesai"`
	// computed (list)
	PelangganNama string `json:"pelanggan_nama"`
	MitraNama     string `json:"mitra_nama"`
}

type Tagihan struct {
	ID           int64  `json:"id"`
	PelangganID  int64  `json:"pelanggan_id"`
	Periode      string `json:"periode"`
	Jumlah       int64  `json:"jumlah"`
	Status       string `json:"status"`
	Metode       string `json:"metode"`
	TanggalBayar string `json:"tanggal_bayar"`
	// computed (list)
	PelangganNama string `json:"pelanggan_nama"`
	PaketNama     string `json:"paket_nama"`
	MitraID       int64  `json:"mitra_id"`
	MitraNama     string `json:"mitra_nama"`
}

type KirimLog struct {
	Via     string `json:"via"`
	Tanggal string `json:"tanggal"`
}

type FABRow struct {
	Paket    string `json:"paket"`
	Jumlah   int    `json:"jumlah"`
	Harga    int64  `json:"harga"`
	Subtotal int64  `json:"subtotal"`
}

type FABRincian struct {
	Periode         string   `json:"periode"`
	JmlPelanggan    int      `json:"jml_pelanggan"`
	JmlTagihan      int      `json:"jml_tagihan"`
	JmlLunasTagihan int      `json:"jml_lunas_tagihan"`
	TotalTagihan    int64    `json:"total_tagihan"`
	TotalLunas      int64    `json:"total_lunas"`
	SharePct        float64  `json:"share_pct"`
	Nilai           int64    `json:"nilai"`
	Rows            []FABRow `json:"rows"`
}

type DocSnapshot struct {
	Company   *Settings `json:"company,omitempty"`
	Mitra     Mitra     `json:"mitra"`
	MasaBulan int       `json:"masa_bulan"`
}

type Dokumen struct {
	ID      int64       `json:"id"`
	Nomor   string      `json:"nomor"`
	Jenis   string      `json:"jenis"`
	MitraID int64       `json:"mitra_id"`
	Tanggal string      `json:"tanggal"`
	Periode string      `json:"periode"`
	Nilai   int64       `json:"nilai"`
	Rincian *FABRincian `json:"rincian,omitempty"`
	Kirim   []KirimLog  `json:"kirim"`
	// computed (list)
	MitraNama string `json:"mitra_nama"`
	CreatedAt string `json:"created_at"`
}

type Settings struct {
	NamaISP       string `json:"nama_isp"`
	Alamat        string `json:"alamat"`
	Kota          string `json:"kota"`
	Telepon       string `json:"telepon"`
	Email         string `json:"email"`
	Website       string `json:"website"`
	NPWP          string `json:"npwp"`
	Penandatangan string `json:"penandatangan"`
	Jabatan       string `json:"jabatan"`
	Rekening      string `json:"rekening"`
	MasaPKSBulan  int    `json:"masa_pks_bulan"`
	TplPKS        string `json:"tpl_pks"`
	TplFAB        string `json:"tpl_fab"`
}
