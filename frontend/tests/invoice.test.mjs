import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { invoiceNumber, renderInvoice } from '../lib/invoice.ts'

const css = readFileSync(new URL('../assets/css/invoice.css', import.meta.url), 'utf8')
const bill = { id:42, pelanggan_id:9, pelanggan_nama:'Rina', paket_nama:'Home 20', mitra_nama:'Mitra Utama', periode:'2026-09', jumlah:123456, status:'Belum' }
const profile = { nama_isp:'Ali Telematika', alamat:'Jl. Mawar', kota:'Bandung', rekening:'Bank BRI 123\na.n. ISP', telepon:'022 1234', email:'billing@example.test' }
const render = (b=bill, customer={alamat:'Jl. Melati'}, settings=profile) => renderInvoice(b, customer, settings, css).replaceAll('\u00a0',' ')

test('invoice identifies the stored bill and uses its amount, not the current package price', () => {
  const customer = {alamat:'Jl. Melati',harga:999999}
  const before = JSON.stringify({bill,customer,profile})
  const html = render(bill,customer)
  assert.match(html, /INV\/202609\/000042/)
  assert.match(html, /September 2026/)
  assert.match(html, /<dt>Sisa pembayaran<\/dt><dd>Rp 123\.456<\/dd>/)
  assert.match(html, /<dt>Sudah dibayar<\/dt><dd>Rp 0<\/dd>/)
  assert.ok(!html.includes('999.999'))
  assert.ok(!html.includes('Jatuh tempo'))
  assert.ok(!html.includes('PPN'))
  assert.equal(JSON.stringify({bill,customer,profile}),before)
})

test('paid invoice has zero balance and displays the recorded payment', () => {
  const html = render({...bill,status:'Lunas',tanggal_bayar:'2026-09-26',metode:'Transfer'})
  assert.match(html, /Tagihan ini sudah lunas/)
  assert.match(html, /<dt>Sisa pembayaran<\/dt><dd>Rp 0<\/dd>/)
  assert.match(html, /<dt>Sudah dibayar<\/dt><dd>Rp 123\.456<\/dd>/)
  assert.match(html, /26 September 2026/)
  assert.match(html, /<dt>Metode<\/dt><dd>Transfer<\/dd>/)
  assert.ok(!html.includes('Cantumkan nomor invoice'))
  assert.ok(!html.includes(profile.rekening))
})

test('reopened invoice does not show obsolete payment details', () => {
  const html = render({...bill,status:'Belum',tanggal_bayar:'2026-09-26',metode:'Transfer'})
  assert.match(html, /Belum dibayar/)
  assert.ok(!html.includes('26 September 2026'))
  assert.ok(!html.includes('<dd>Transfer</dd>'))
})

test('all customer, service, payment and company values are escaped', () => {
  const attack = '<img src=x onerror="alert(1)"><script>alert(2)</script>&'
  const html = render({...bill,pelanggan_nama:attack,paket_nama:attack,mitra_nama:attack},{alamat:attack},{nama_isp:attack,alamat:attack,kota:attack,email:attack,telepon:attack,rekening:attack,website:attack})
  assert.ok(html.includes('&lt;img src=x onerror=&quot;alert(1)&quot;&gt;'))
  assert.ok(html.includes('&lt;script&gt;alert(2)&lt;/script&gt;&amp;'))
  assert.ok(!html.includes('<img src=x'))
  assert.equal((html.match(/<img /g) || []).length, 1)
  assert.ok(!html.includes('<script'))
  const paid = render({...bill,status:'Lunas',metode:attack})
  assert.ok(!paid.includes('<img src=x'))
})

test('missing optional information does not invent payment dates or bank details', () => {
  const html = render(bill,{}, {nama_isp:'ISP'})
  assert.match(html,/Hubungi ISP untuk informasi rekening/)
  assert.ok(!html.includes('undefined'))
  assert.ok(!html.includes('null'))
  const paid = render({...bill,status:'Lunas'}, {}, {nama_isp:'ISP'})
  assert.ok(!paid.includes('<dt>Tanggal</dt>'))
  assert.ok(!paid.includes('<dt>Metode</dt>'))
  assert.ok(!paid.includes('Invalid Date'))
})

test('invoice numbers are stable and distinguish separate bills', () => {
  assert.equal(invoiceNumber(bill),'INV/202609/000042')
  assert.equal(invoiceNumber({...bill,status:'Lunas'}),invoiceNumber(bill))
  assert.notEqual(invoiceNumber({...bill,id:43}),invoiceNumber(bill))
  assert.notEqual(invoiceNumber({...bill,periode:'2026-10'}),invoiceNumber(bill))
})

test('a complete company address does not repeat its city', () => {
  const html = render(bill, {}, {nama_isp:'PT Ali Telematika Nusantara',alamat:'Tuladenggi, Kab. Gorontalo 96137',kota:'Gorontalo'})
  assert.match(html, /<address>Tuladenggi, Kab\. Gorontalo 96137<\/address>/)
  const partial = render(bill, {}, {nama_isp:'ISP',alamat:'Jl. Mawar',kota:'Bandung'})
  assert.match(partial, /<address>Jl\. Mawar<br>Bandung<\/address>/)
})
