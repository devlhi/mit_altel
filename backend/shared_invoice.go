package main

import (
	_ "embed"
	"fmt"
	"strings"
	"time"
)

//go:embed assets/invoice.css
var sharedInvoiceCSS string

func sharedInvoiceHTML(s Settings, b Tagihan, number, address, service string) string {
	e := escHTML
	paid := b.Status == "Lunas"
	balance := b.Jumlah
	paidAmount := int64(0)
	status := "Belum dibayar"
	class := ""
	payment := `<h2>Informasi pembayaran</h2><p>Hubungi ` + e(s.NamaISP) + ` untuk informasi rekening dan konfirmasi pembayaran.</p>`
	if s.Rekening != "" {
		payment = `<h2>Informasi pembayaran</h2><p class="bank-account">` + e(s.Rekening) + `</p><p>Cantumkan nomor invoice pada keterangan pembayaran.</p>`
	}
	if paid {
		balance = 0
		paidAmount = b.Jumlah
		status = "Lunas"
		class = "is-paid"
		payment = `<h2>Pembayaran diterima</h2><p class="paid-copy">Terima kasih. Tagihan ini sudah lunas.</p><dl><dt>Tanggal</dt><dd>` + e(tanggalIndo(b.TanggalBayar)) + `</dd><dt>Metode</dt><dd>` + e(b.Metode) + `</dd></dl>`
	}
	addr := e(s.Alamat)
	if s.Kota != "" && !strings.Contains(strings.ToLower(s.Alamat), strings.ToLower(s.Kota)) {
		addr += "<br>" + e(s.Kota)
	}
	return `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` + e(number) + `</title><style>` + sharedInvoiceCSS + `
 .shared-note{padding:12px 24px;background:#eff5ff;color:#365576;font:12px/1.6 Arial;text-align:center}@media screen and (max-width:700px){.invoice-paper{width:100%;padding:24px}.invoice-heading,.invoice-meta,.invoice-settlement{display:block}.invoice-identity{text-align:left;margin-top:24px}.invoice-settlement>dl{margin-top:24px}.invoice-heading .issuer{max-width:100%}.invoice-paper .company img{max-width:100%}}@media print{.shared-note{display:none}}
 </style></head><body class="` + class + `"><div class="shared-note">Salinan invoice per ` + e(time.Now().Format("02 January 2006 15:04")) + `. Gunakan menu Cetak pada browser untuk menyimpan PDF.</div><main class="invoice-paper">
 <header class="invoice-heading"><div class="issuer"><div class="company"><img src="data:image/png;base64,` + companyLogoBase64 + `" width="3001" height="516" alt="ALTELMATARA NETWORK"><b>` + e(s.NamaISP) + `</b></div><address>` + addr + `</address><p class="issuer-contact"><span>` + e(s.Telepon) + `</span><span>` + e(s.Email) + `</span></p></div><div class="invoice-identity"><h1>Invoice</h1><p class="invoice-no">` + e(number) + `</p></div></header>
 <section class="invoice-overview"><div class="invoice-amount"><span>Total tagihan</span><strong>` + rupiah(b.Jumlah) + `</strong></div><div class="invoice-period"><span class="payment-status">` + status + `</span><p>Periode layanan <b>` + e(periodeLabel(b.Periode)) + `</b></p></div></section>
 <section class="invoice-meta"><div class="recipient"><span class="section-label">Ditagihkan kepada</span><h2>` + e(b.PelangganNama) + `</h2><p>` + e(address) + `</p></div><dl><dt>ID pelanggan</dt><dd>` + fmt.Sprintf("PLG-%05d", b.PelangganID) + `</dd><dt>Mitra layanan</dt><dd>` + e(b.MitraNama) + `</dd></dl></section>
 <table class="service-table"><caption>Rincian tagihan</caption><thead><tr><th>Layanan &amp; keterangan</th><th>Qty</th><th>Tarif</th><th>Jumlah</th></tr></thead><tbody><tr><td><b>` + e(service) + `</b><small>Langganan internet<br>` + e(periodeLabel(b.Periode)) + `</small></td><td>1</td><td>` + rupiah(b.Jumlah) + `</td><td>` + rupiah(b.Jumlah) + `</td></tr></tbody></table>
 <section class="invoice-settlement"><div class="payment-note">` + payment + `</div><dl class="invoice-totals"><div><dt>Total tagihan</dt><dd>` + rupiah(b.Jumlah) + `</dd></div><div><dt>Sudah dibayar</dt><dd>` + rupiah(paidAmount) + `</dd></div><div class="balance"><dt>Sisa pembayaran</dt><dd>` + rupiah(balance) + `</dd></div></dl></section>
 <div class="invoice-message"><h2>Terima kasih telah menggunakan layanan kami.</h2><p>Untuk pertanyaan terkait tagihan, hubungi ` + e(s.NamaISP) + ` dengan menyebutkan nomor invoice ini.</p></div><footer class="invoice-footer"><div><b>` + e(s.NamaISP) + `</b><span>` + e(s.Website) + `</span></div><div><span>` + e(number) + `</span><span>Tagihan layanan internet</span></div></footer></main></body></html>`
}
