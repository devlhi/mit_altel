export interface InvoiceBill {
  id: number
  pelanggan_id: number
  pelanggan_nama: string
  paket_nama: string
  mitra_nama: string
  periode: string
  jumlah: number
  status: string
  tanggal_bayar?: string
  metode?: string
}

export function invoiceNumber(bill: Pick<InvoiceBill, 'id' | 'periode'>): string {
  return `INV/${bill.periode.replaceAll('-', '')}/${String(bill.id).padStart(6, '0')}`
}

function escapeHtml(value: unknown): string {
  return String(value ?? '').replace(/[&<>"']/g, char => ({'&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;'}[char]!))
}

export function renderInvoice(bill: InvoiceBill, customer: { alamat?: string }, settings: Record<string, any>, css: string): string {
  const e = escapeHtml
  const money = (n: number) => e(new Intl.NumberFormat('id-ID', {style:'currency', currency:'IDR', maximumFractionDigits:0}).format(n))
  const monthNames = ['Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember']
  const [year, month] = bill.periode.split('-')
  const period = e(monthNames[Number(month)-1] ? `${monthNames[Number(month)-1]} ${year}` : bill.periode)
  const paid = bill.status === 'Lunas'
  const number = e(invoiceNumber(bill))
  const amount = Number(bill.jumlah)
  const paymentDate = bill.tanggal_bayar ? new Date(`${bill.tanggal_bayar}T00:00:00`) : null
  const dateLabel = paymentDate && !Number.isNaN(paymentDate.getTime()) ? e(paymentDate.toLocaleDateString('id-ID',{day:'numeric',month:'long',year:'numeric'})) : ''
  const company = e(settings.nama_isp || 'Penyedia layanan internet')
  const address = String(settings.alamat || '')
  const city = String(settings.kota || '').trim()
  const companyAddress = e(address) + (city && !address.toLocaleLowerCase('id-ID').includes(city.toLocaleLowerCase('id-ID')) ? `<br>${e(city)}` : '')
  const service = e(bill.paket_nama || 'Layanan internet')
  const payment = paid
    ? `<h2>Pembayaran diterima</h2><p class="paid-copy">Terima kasih. Tagihan ini sudah lunas.</p><dl>${dateLabel ? `<dt>Tanggal</dt><dd>${dateLabel}</dd>` : ''}${bill.metode ? `<dt>Metode</dt><dd>${e(bill.metode)}</dd>` : ''}</dl>`
    : `<h2>Informasi pembayaran</h2>${settings.rekening ? `<p class="bank-account">${e(settings.rekening)}</p><p>Cantumkan nomor invoice pada keterangan pembayaran. Hubungi kami untuk konfirmasi setelah membayar.</p>` : `<p>Hubungi ${company} untuk informasi rekening dan konfirmasi pembayaran.</p>`}`

  // All business content is escaped before rendering the shared preview/print document.
  return `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${number} — ${e(bill.pelanggan_nama)}</title><style>${css}</style></head>
<body class="${paid ? 'is-paid' : ''}"><main class="invoice-paper">
 <header class="invoice-heading">
  <div class="issuer"><div class="company"><img src="/branding/altelmatara-network.png" width="3001" height="516" alt="ALTELMATARA NETWORK" /><b>${company}</b></div>${address ? `<address>${companyAddress}</address>` : ''}<p class="issuer-contact">${settings.telepon ? `<span>${e(settings.telepon)}</span>` : ''}${settings.email ? `<span>${e(settings.email)}</span>` : ''}</p></div>
  <div class="invoice-identity"><h1>Invoice</h1><p class="invoice-no">${number}</p></div>
 </header>
 <section class="invoice-overview" aria-label="Ringkasan tagihan"><div class="invoice-amount"><span>Total tagihan</span><strong>${money(amount)}</strong></div><div class="invoice-period"><span class="payment-status">${paid ? '<svg viewBox="0 0 16 16" fill="none" aria-hidden="true"><path d="m3 8 3 3 7-7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>Lunas' : 'Belum dibayar'}</span><p>Periode layanan <b>${period}</b></p></div></section>
 <section class="invoice-meta" aria-label="Pelanggan dan mitra layanan">
  <div class="recipient"><span class="section-label">Ditagihkan kepada</span><h2>${e(bill.pelanggan_nama)}</h2>${customer.alamat ? `<p>${e(customer.alamat)}</p>` : ''}</div>
  <dl><dt>ID pelanggan</dt><dd>PLG-${e(String(bill.pelanggan_id).padStart(5,'0'))}</dd>${bill.mitra_nama ? `<dt>Mitra layanan</dt><dd>${e(bill.mitra_nama)}</dd>` : ''}</dl>
 </section>
 <table class="service-table"><caption>Rincian tagihan</caption><thead><tr><th scope="col">Layanan &amp; keterangan</th><th scope="col">Qty</th><th scope="col">Tarif</th><th scope="col">Jumlah</th></tr></thead><tbody><tr><td><b>${service}</b><small>Langganan internet<br>${period}</small></td><td>1</td><td>${money(amount)}</td><td>${money(amount)}</td></tr></tbody></table>
 <section class="invoice-settlement" aria-label="Pembayaran"><div class="payment-note">${payment}</div><dl class="invoice-totals"><div><dt>Total tagihan</dt><dd>${money(amount)}</dd></div><div><dt>Sudah dibayar</dt><dd>${money(paid ? amount : 0)}</dd></div><div class="balance"><dt>Sisa pembayaran</dt><dd>${money(paid ? 0 : amount)}</dd></div></dl></section>
 <div class="invoice-message"><h2>Terima kasih telah menggunakan layanan kami.</h2><p>Untuk pertanyaan terkait tagihan, hubungi ${company} dengan menyebutkan nomor invoice ini.</p></div>
 <footer class="invoice-footer"><div><b>${company}</b>${settings.website ? `<span>${e(settings.website)}</span>` : ''}</div><div><span>${number}</span><span>Tagihan layanan internet</span></div></footer>
</main></body></html>`
}
