package main

// DefaultSettings mengembalikan pengaturan awal beserta template dokumen bawaan.
func DefaultSettings() Settings {
	return Settings{
		NamaISP:       "PT Ali Telematika Nusantara",
		Alamat:        "Perumahan Surya Pagi, Desa Tuladenggi, Kec. Telaga Biru, Kab. Gorontalo, Provinsi Gorontalo 96137",
		Kota:          "Gorontalo",
		Telepon:       "+62 822-9816-8300",
		Email:         "cs@altelmatara.id",
		Website:       "altelmatara.id",
		NPWP:          "",
		Penandatangan: "",
		Jabatan:       "",
		Rekening:      "",
		MasaPKSBulan:  12,
		TplPKS:        defaultTplPKS,
		TplFAB:        defaultTplFAB,
	}
}

const docCSS = `
*{box-sizing:border-box}
:root{--ink:#202b4b;--blue:#405de6;--muted:#788199;--line:#e0e5f0}
body{margin:0;background:#edf0f6;font-family:Georgia,'Times New Roman',serif;color:#303b50;font-size:10.5pt;line-height:1.75}
.page{background:white;width:210mm;max-width:100%;min-height:297mm;margin:24px auto;padding:17mm 18mm;box-shadow:0 5px 25px #202b4b0d;border-top:4px solid var(--ink)}
.toolbar{position:sticky;top:0;z-index:5;display:flex;gap:18px;align-items:center;background:var(--ink);color:white;padding:12px 24px;font:13px Arial,sans-serif}
.toolbar button{background:var(--blue);color:white;border:0;padding:10px 18px;border-radius:6px;font:600 13px Arial,sans-serif;cursor:pointer}
.letterhead{display:flex;justify-content:space-between;align-items:flex-start;gap:24px;border-bottom:1px solid var(--line);padding-bottom:23px;margin-bottom:27px;font-family:Arial,sans-serif}
.lh-identity{min-width:0;flex:1}.company-logo{display:block;width:235px;max-width:100%;height:auto;margin-bottom:13px}
.letterhead h1{margin:0;color:var(--ink);font-size:12pt;letter-spacing:-.2px;font-weight:700;line-height:1.2;overflow-wrap:anywhere}
.letterhead .lh-sub{font:8pt Arial,sans-serif;letter-spacing:.12em;color:#929bb0;margin-top:7px}
.letterhead p{margin:0;font-size:7.5pt;color:var(--muted);line-height:1.65;text-align:right;max-width:43%}
.lh-mark{flex:none;width:43px;height:43px;color:var(--blue);display:flex;align-items:center;justify-content:center}
.lh-mark svg{width:43px;height:43px}
.doc-label{display:flex;align-items:center;gap:9px;font:700 8pt Arial,sans-serif;letter-spacing:.09em;color:var(--blue);margin:0 0 13px}
.doc-label span{font-weight:400;color:#8993aa;letter-spacing:.06em;border-left:1px solid var(--line);padding-left:9px;text-transform:uppercase}
.document-note{font:9pt/1.6 Arial,sans-serif;padding:9px 12px;margin:12px 0 20px;border:1px solid #e2d2a4;background:#fffaf0;color:#7b5b20}
.doc-title{font-family:Arial,sans-serif;color:var(--ink);font-size:20pt;font-weight:600;line-height:1.28;letter-spacing:-.6px;text-align:left;margin:0 0 10px;max-width:97%}
.doc-no{font:10pt Arial,sans-serif;color:var(--muted);text-align:left;margin:0 0 22px}
.doc-no-sub{text-align:left;margin:-12px 0 24px;color:var(--blue);font:600 10pt Arial,sans-serif}
h3{font:700 10.5pt Arial,sans-serif;color:var(--ink);margin:23px 0 7px;break-after:avoid;page-break-after:avoid}
p{margin:9px 0;text-align:justify;orphans:3;widows:3;overflow-wrap:anywhere}
ol,ul{padding-left:22px} li{margin:5px 0;text-align:justify;orphans:3;widows:3} .parties{background:#f6f8fd;border-left:2px solid #bdcaed;padding:13px 20px 13px 34px}.parties li{padding-left:3px}
table.meta{font-family:Arial,sans-serif;width:100%;background:#f6f8fd;border-collapse:collapse;margin:24px 0} table.meta td{padding:7px 15px;font-size:9pt;vertical-align:top;text-align:left} table.meta tr:first-child td{padding-top:15px} table.meta tr:last-child td{padding-bottom:15px} table.meta td:first-child{color:var(--muted)}
table.rincian{width:100%;border-collapse:collapse;margin:20px 0;font:9pt Arial,sans-serif;table-layout:auto}
table.rincian th,table.rincian td{border:0;border-bottom:1px solid var(--line);padding:11px 9px;vertical-align:top}
table.rincian th{background:var(--ink);color:white;text-align:left;font-size:8pt;line-height:1.4;font-weight:600}
table.rincian td:nth-last-child(-n+2){text-align:right;white-space:nowrap;font-variant-numeric:tabular-nums} table.rincian th:nth-last-child(-n+2){text-align:right}
table.rincian tr{break-inside:avoid;page-break-inside:avoid} table.rincian thead{display:table-header-group} table.rincian .total-row td{background:#f6f8fd;font-weight:600} table.rincian tbody tr:last-child td{background:#edf1ff;color:#304780;font-weight:700}
.rekening{white-space:pre-line;border-left:2px solid #bdcaed;background:#f6f8fd;padding:14px 18px;display:block;font:10pt/1.7 Arial,sans-serif;text-align:left;break-inside:avoid}
.ttd{margin-top:30px;break-inside:avoid;page-break-inside:avoid}.ttd table{width:100%;border-collapse:collapse}.ttd td{width:50%;text-align:center;vertical-align:top;padding:10px;font:9pt/1.7 Arial,sans-serif}.sign-space{height:75px}
.doc-footer{margin-top:32px;padding-top:12px;border-top:1px solid var(--line);display:flex;justify-content:space-between;gap:20px;color:var(--muted);font:8pt/1.5 Arial,sans-serif;overflow-wrap:anywhere}
.document-fab{font-family:Arial,sans-serif;font-size:10pt;line-height:1.65}.document-fab p{text-align:left}.document-fab .letterhead p{text-align:right}.document-fab .doc-title{font-size:27pt;letter-spacing:-.8px}.document-fab .ttd{margin-top:24px}
.document-fab{font-size:9.5pt;line-height:1.5}.document-fab .letterhead{margin-bottom:17px;padding-bottom:16px}.document-fab .company-logo{width:210px;margin-bottom:9px}.document-fab .doc-label{margin-bottom:10px}.document-fab .doc-title{font-size:24pt;line-height:1.2;margin-bottom:8px}.document-fab .doc-no{margin-bottom:14px;font-size:9pt}.document-fab table.meta{margin:14px 0}.document-fab table.meta td{padding:5px 10px;font-size:8.5pt}.document-fab table.meta tr:first-child td{padding-top:10px}.document-fab table.meta tr:last-child td{padding-bottom:10px}.document-fab table.rincian{margin:16px 0 12px}.document-fab table.rincian th,.document-fab table.rincian td{padding:8px 7px;font-size:8.5pt;line-height:1.4}.document-fab p{margin:7px 0}.document-fab .basis-note{font-size:8.5pt;color:#596478;line-height:1.5}.document-fab .payment-block{break-inside:avoid}.document-fab .payment-block h3{font-size:9pt;margin:15px 0 6px}.document-fab .rekening{font-size:9pt;line-height:1.5;padding:9px 13px}.document-fab .ttd{margin-top:18px}.document-fab .ttd td{font-size:8.5pt;line-height:1.5;padding:6px}.document-fab .sign-space{height:48px}.document-fab .doc-footer{margin-top:20px}.document-fab .document-note{font-size:8pt;padding:6px 9px;margin:9px 0 14px;line-height:1.5}
@media screen and (max-width:600px){.page{padding:25px 20px;margin:0;min-height:100vh}.letterhead{flex-wrap:wrap;gap:15px}.letterhead h1{font-size:17pt}.letterhead p,.document-fab .letterhead p{font-size:8pt;text-align:left;max-width:none}.doc-title{font-size:17pt}.document-fab .doc-title{font-size:21pt}.lh-mark{width:37px;height:37px}.rincian th,.rincian td{padding:8px 4px!important;font-size:8pt}table.rincian th{width:auto!important}.doc-footer{flex-wrap:wrap}}
@media print{
 @page{size:A4;margin:16mm 16mm 18mm}
 body{background:white;font-size:10.5pt;-webkit-print-color-adjust:exact;print-color-adjust:exact}
 .page{box-shadow:none;border-top:0;margin:0;padding:0;width:auto;max-width:none;min-height:0}
 .no-print{display:none!important}.letterhead{break-inside:avoid;page-break-inside:avoid}
 a{color:inherit;text-decoration:none}.doc-footer{break-inside:avoid}
}
`
const defaultTplPKS = `<h2 class="doc-title">Perjanjian Kerja Sama Penyediaan Layanan Internet</h2>
<p class="doc-no">Nomor: {{nomor}}</p>

<p>Pada hari ini {{hari}}, tanggal {{tanggal}}, yang bertanda tangan di bawah ini:</p>
<ol class="parties">
<li><b>{{nama_isp}}</b>, berkedudukan di {{alamat_isp}}, {{kota_isp}}, dalam hal ini diwakili oleh <b>{{penandatangan}}</b> selaku {{jabatan}}, bertindak untuk dan atas nama {{nama_isp}} &mdash; selanjutnya disebut <b>PIHAK PERTAMA</b>;</li>
<li><b>{{nama_mitra}}</b> ({{jenis_mitra}}), beralamat di {{alamat_mitra}}, {{wilayah_mitra}}, dalam hal ini diwakili oleh <b>{{pemilik_mitra}}</b> selaku Pemilik/Penanggung Jawab, bertindak untuk dan atas nama {{nama_mitra}} &mdash; selanjutnya disebut <b>PIHAK KEDUA</b>.</li>
</ol>

<p>Pihak Pertama dan Pihak Kedua secara bersama-sama selanjutnya disebut <b>PARA PIHAK</b>. PARA PIHAK terlebih dahulu menerangkan hal-hal sebagai berikut:</p>
<ol type="a">
<li>bahwa PIHAK PERTAMA merupakan perusahaan penyedia jasa internet dan jaringan komunikasi data yang memiliki izin penyelenggaraan yang berlaku;</li>
<li>bahwa PIHAK KEDUA merupakan {{jenis_mitra}} yang memiliki kemampuan, jaringan, dan akses pasar di wilayah {{wilayah_mitra}};</li>
<li>bahwa PARA PIHAK bermaksud mengadakan kerja sama penyediaan dan pemasaran layanan internet sebagaimana diatur dalam perjanjian ini.</li>
</ol>

<p>Berdasarkan hal-hal tersebut di atas, PARA PIHAK sepakat untuk mengadakan Perjanjian Kerja Sama dengan syarat-syarat sebagai berikut:</p>

<h3>Pasal 1 &mdash; Dasar Kerja Sama</h3>
<p>PARA PIHAK sepakat mengadakan kerja sama penyediaan layanan internet, dengan PIHAK PERTAMA bertindak sebagai penyedia layanan internet (bandwidth, backbone, dan perangkat inti jaringan), dan PIHAK KEDUA bertindak sebagai mitra yang memasarkan, menjual, dan mengelola layanan kepada pelanggan di wilayah kerja sama.</p>

<h3>Pasal 2 &mdash; Lingkup Kerja Sama</h3>
<ol type="a">
<li>PIHAK PERTAMA menyediakan layanan internet berlangganan (rumah/bisnis) maupun dedicated beserta dukungan teknis tingkat jaringan (NOC).</li>
<li>PIHAK KEDUA melakukan pemasaran, pendaftaran pelanggan, instalasi jaringan lokal (last mile), penagihan, dan pelayanan pelanggan tingkat pertama di wilayah {{wilayah_mitra}}.</li>
<li>Pelanggan yang didaftarkan oleh PIHAK KEDUA tercatat atas nama PIHAK PERTAMA sebagai penyedia layanan resmi.</li>
<li>Rincian harga jual, paket layanan, dan wilayah kerja sama mengikuti lampiran/daftar harga yang diterbitkan PIHAK PERTAMA dan dapat berubah dengan pemberitahuan tertulis.</li>
</ol>

<h3>Pasal 3 &mdash; Hak dan Kewajiban PIHAK PERTAMA</h3>
<ol type="a">
<li>Menyediakan bandwidth, IP publik (bila diperlukan), dan perangkat inti jaringan;</li>
<li>Memberikan dukungan teknis dan monitoring jaringan melalui tim NOC;</li>
<li>Memberikan materi promosi, harga jual standar, dan pelatihan teknis kepada PIHAK KEDUA;</li>
<li>Melakukan pembayaran bagi hasil sesuai Pasal 5;</li>
<li>Memberikan laporan rekapitulasi pelanggan dan pendapatan kepada PIHAK KEDUA setiap bulan.</li>
</ol>

<h3>Pasal 4 &mdash; Hak dan Kewajiban PIHAK KEDUA</h3>
<ol type="a">
<li>Memasarkan layanan dan mengembangkan pelanggan di wilayah kerja sama;</li>
<li>Melakukan instalasi jaringan lokal, perawatan, dan penanganan gangguan tingkat pertama;</li>
<li>Melakukan penagihan pembayaran langganan kepada pelanggan secara tertib dan tertulis;</li>
<li>Menjaga nama baik PIHAK PERTAMA dan tidak melakukan perbuatan yang merugikan PIHAK PERTAMA maupun pelanggan;</li>
<li>Memberikan laporan pelanggan, pendapatan, dan penyelesaian gangguan secara berkala sesuai format yang disepakati.</li>
</ol>

<h3>Pasal 5 &mdash; Skema Bagi Hasil</h3>
<ol type="a">
<li>Hasil kerja sama dibagi dengan komposisi <b>{{share_mitra}}% ({{share_mitra_terbilang}} persen) untuk PIHAK KEDUA</b> dan <b>{{share_isp}}% ({{share_isp_terbilang}} persen) untuk PIHAK PERTAMA</b> dari pendapatan langganan yang telah diterima (tertagih lunas) di wilayah kerja sama pada periode berjalan.</li>
<li>Perhitungan dan pembayaran bagi hasil dilakukan setiap bulan berdasarkan rekapitulasi yang diterbitkan PIHAK PERTAMA dalam bentuk Faktur Bisnis (FAB).</li>
<li>Pajak-pajak yang timbul atas masing-masing pihak menjadi tanggung jawab masing-masing pihak.</li>
</ol>

<h3>Pasal 6 &mdash; Masa Berlaku</h3>
<p>Perjanjian ini berlaku untuk jangka waktu <b>{{masa_bulan}} ({{masa_terbilang}}) bulan</b> terhitung sejak {{tanggal_mulai}} sampai dengan {{tanggal_akhir}}, dan diperpanjang secara otomatis untuk periode yang sama apabila tidak ada pihak yang mengajukan pengakhiran secara tertulis paling lambat 30 (tiga puluh) hari sebelum berakhirnya perjanjian.</p>

<h3>Pasal 7 &mdash; Kerahasiaan</h3>
<p>PARA PIHAK wajib menjaga kerahasiaan seluruh data, informasi teknis, komersial, dan keuangan milik masing-masing pihak yang diperoleh selama berlangsungnya kerja sama ini, baik selama maupun setelah berakhirnya perjanjian.</p>

<h3>Pasal 8 &mdash; Keadaan Kahar (Force Majeure)</h3>
<p>PARA PIHAK tidak bertanggung jawab atas kegagalan pelaksanaan kewajiban yang disebabkan oleh keadaan kahar, antara lain bencana alam, kebakaran, huru-hara, perang, pemogokan, gangguan jaringan listrik/telekomunikasi dari penyedia hulu (upstream), atau peraturan pemerintah.</p>

<h3>Pasal 9 &mdash; Penyelesaian Perselisihan</h3>
<p>Segala perselisihan yang timbul dari dan sehubungan dengan perjanjian ini akan diselesaikan terlebih dahulu secara musyawarah untuk mufakat. Apabila tidak tercapai, PARA PIHAK sepakat menyelesaikannya melalui Pengadilan Negeri {{kota}}.</p>

<h3>Pasal 10 &mdash; Penutup</h3>
<p>Perjanjian ini dibuat dalam rangkap 2 (dua) asli yang masing-masing mempunyai kekuatan hukum yang sama, ditandatangani PARA PIHAK dalam keadaan sadar tanpa paksaan dari pihak manapun, di {{kota}}, pada tanggal {{tanggal}}.</p>

<div class="ttd">
<table>
<tr>
<td>PIHAK PERTAMA<br><b>{{nama_isp}}</b><div class="sign-space"></div>( {{penandatangan}} )<br>{{jabatan}}</td>
<td>PIHAK KEDUA<br><b>{{nama_mitra}}</b><div class="sign-space"></div>( {{pemilik_mitra}} )<br>Pemilik / Penanggung Jawab</td>
</tr>
</table>
</div>`

const defaultTplFAB = `<h2 class="doc-title">Rekapitulasi bagi hasil</h2>
<p class="doc-no">{{nomor}} &nbsp; / &nbsp; {{tanggal}}</p>
<table class="meta"><tr><td>Mitra penerima</td><td><b>{{nama_mitra}}</b> · {{jenis_mitra}}</td><td>Periode</td><td><b>{{periode_label}}</b></td></tr><tr><td>Alamat</td><td colspan="3">{{alamat_mitra}}, {{wilayah_mitra}}</td></tr><tr><td>Kontak</td><td>{{telepon_mitra}}</td><td>Bagi hasil mitra</td><td>{{share_mitra}}%</td></tr></table>
<table class="rincian"><thead><tr><th>No</th><th>Layanan</th><th style="text-align:center">Tagihan</th><th>Tarif tersimpan</th><th>Jumlah</th></tr></thead><tbody>
{{rincian_rows}}
<tr class="total-row"><td colspan="4">Total tagihan · {{jml_tagihan}} tagihan</td><td>{{total_tagihan}}</td></tr>
<tr class="total-row"><td colspan="4">Pembayaran diterima · {{jml_lunas_tagihan}} tagihan lunas</td><td>{{total_lunas}}</td></tr>
<tr><td colspan="4">Hak bagi hasil mitra · {{share_mitra}}%</td><td>{{nilai_bagihasil}}</td></tr>
</tbody></table>
<p class="amount-words">Terbilang: <b>{{nilai_terbilang}}.</b></p>
<p class="basis-note">Dasar bagi hasil adalah tagihan berstatus lunas pada saat penerbitan, sesuai kesepakatan kerja sama. Pelanggan aktif saat penerbitan: {{jml_pelanggan}}.</p>
<div class="payment-block"><h3>Rekening penerima mitra</h3><p class="rekening">{{rekening_mitra}}</p><p class="basis-note">Pembayaran mengikuti jadwal yang disepakati dalam PKS. Rekap ini bukan bukti transfer.</p></div>
<div class="ttd"><table><tr><td>Diterbitkan oleh<br><b>{{nama_isp}}</b><div class="sign-space"></div><b>{{penandatangan}}</b><br>{{jabatan}}</td><td>Diperiksa oleh<br><b>{{nama_mitra}}</b><div class="sign-space"></div><b>{{pemilik_mitra}}</b><br>Penanggung jawab mitra</td></tr></table></div>`

const legacyTplFAB = `<h2 class="doc-title">Faktur Bisnis (FAB) &mdash; Rekapitulasi Bagi Hasil</h2>
<p class="doc-no">Nomor: {{nomor}}</p>
<p class="doc-no-sub">Periode: {{periode_label}}</p>

<table class="meta">
<tr><td width="90">Kepada</td><td>: <b>{{nama_mitra}}</b> ({{jenis_mitra}})</td></tr>
<tr><td>Alamat</td><td>: {{alamat_mitra}}, {{wilayah_mitra}}</td></tr>
<tr><td>Telepon</td><td>: {{telepon_mitra}}</td></tr>
</table>

<p>Berikut rekapitulasi tagihan dan perhitungan bagi hasil kerja sama layanan internet untuk periode <b>{{periode_label}}</b>:</p>

<table class="rincian">
<thead>
<tr><th width="36">No</th><th>Paket Layanan</th><th width="70" style="text-align:center">Pelanggan</th><th width="120">Harga</th><th width="140">Jumlah</th></tr>
</thead>
<tbody>
{{rincian_rows}}
<tr class="total-row"><td colspan="4" style="text-align:right">Subtotal Tagihan ({{jml_tagihan}} tagihan)</td><td style="text-align:right">{{total_tagihan}}</td></tr>
<tr><td colspan="4" style="text-align:right">Diterima / Lunas ({{jml_lunas_tagihan}} tagihan)</td><td style="text-align:right">{{total_lunas}}</td></tr>
<tr><td colspan="4" style="text-align:right">Bagi Hasil PIHAK KEDUA ({{share_mitra}}%)</td><td style="text-align:right"><b>{{nilai_bagihasil}}</b></td></tr>
</tbody>
</table>

<p><b>Dengan huruf:</b> {{nilai_terbilang}}.</p>
<p>Pelanggan aktif mitra pada periode ini: {{jml_pelanggan}} pelanggan. Bagi hasil dihitung dari tagihan berstatus lunas sesuai Pasal 5 Perjanjian Kerja Sama.</p>

<p>Pembayaran bagi hasil ditransfer ke rekening berikut paling lambat 10 (sepuluh) hari kerja setelah dokumen ini diterbitkan:</p>
<p class="rekening">{{rekening}}</p>

<p>Demikian Faktur Bisnis ini dibuat dengan sebenarnya untuk dipergunakan sebagaimana mestinya.</p>

<div class="ttd">
<table>
<tr>
<td>Hormat kami,<br><b>{{nama_isp}}</b><div class="sign-space"></div>( {{penandatangan}} )<br>{{jabatan}}</td>
<td>Diterima oleh,<br><b>{{nama_mitra}}</b><div class="sign-space"></div>( {{pemilik_mitra}} )</td>
</tr>
</table>
</div>`

const legacyOperationsTplFAB = `<h2 class="doc-title">Faktur Bisnis (FAB) &mdash; Rekapitulasi Bagi Hasil</h2>
<p class="doc-no">Nomor: {{nomor}}</p>
<p class="doc-no-sub">Periode: {{periode_label}}</p>

<table class="meta">
<tr><td width="90">Kepada</td><td>: <b>{{nama_mitra}}</b> ({{jenis_mitra}})</td></tr>
<tr><td>Alamat</td><td>: {{alamat_mitra}}, {{wilayah_mitra}}</td></tr>
<tr><td>Telepon</td><td>: {{telepon_mitra}}</td></tr>
</table>

<p>Berikut rekapitulasi tagihan dan perhitungan bagi hasil kerja sama layanan internet untuk periode <b>{{periode_label}}</b>:</p>

<table class="rincian">
<thead>
<tr><th width="36">No</th><th>Paket Layanan</th><th width="70" style="text-align:center">Pelanggan</th><th width="120">Harga</th><th width="140">Jumlah</th></tr>
</thead>
<tbody>
{{rincian_rows}}
<tr class="total-row"><td colspan="4" style="text-align:right">Subtotal Tagihan ({{jml_tagihan}} tagihan)</td><td style="text-align:right">{{total_tagihan}}</td></tr>
<tr><td colspan="4" style="text-align:right">Diterima / Lunas ({{jml_lunas_tagihan}} tagihan)</td><td style="text-align:right">{{total_lunas}}</td></tr>
<tr><td colspan="4" style="text-align:right">Bagi Hasil PIHAK KEDUA ({{share_mitra}}%)</td><td style="text-align:right"><b>{{nilai_bagihasil}}</b></td></tr>
</tbody>
</table>

<p><b>Dengan huruf:</b> {{nilai_terbilang}}.</p>
<p>Pelanggan aktif saat penerbitan: {{jml_pelanggan}} pelanggan. Bagi hasil dihitung dari tagihan berstatus lunas sesuai Pasal 5 Perjanjian Kerja Sama.</p>

<p>Pembayaran bagi hasil ditujukan ke rekening mitra berikut. Jadwal pembayaran mengikuti kesepakatan kerja sama:</p>
<p class="rekening">{{rekening_mitra}}</p>

<p>Demikian Faktur Bisnis ini dibuat dengan sebenarnya untuk dipergunakan sebagaimana mestinya.</p>

<div class="ttd">
<table>
<tr>
<td>Hormat kami,<br><b>{{nama_isp}}</b><div class="sign-space"></div>( {{penandatangan}} )<br>{{jabatan}}</td>
<td>Diterima oleh,<br><b>{{nama_mitra}}</b><div class="sign-space"></div>( {{pemilik_mitra}} )</td>
</tr>
</table>
</div>`
