<script setup lang="ts">
import { invoiceNumber } from '~/lib/invoice'
const periode = ref(localISODate().slice(0, 7))
const data = ref<any>(null)
const loading = ref(true)
const generating = ref(false)
const loadError = ref('')
const q = ref('')
const statusFilter = ref('')
const paying = ref<number | null>(null)
const invoice = ref<any>(null)
const paymentBill = ref<any>(null)
const paymentDate = ref(localISODate())
const paymentMethod = ref('Transfer')
let loadRequest = 0
const filteredBills = computed(() => (data.value?.tagihan || []).filter((t: any) =>
  (!statusFilter.value || t.status === statusFilter.value) &&
  `${t.pelanggan_nama} ${t.mitra_nama} ${invoiceNumber(t)}`.toLowerCase().includes(q.value.trim().toLowerCase())
))

const billPages = usePagination<any>(filteredBills, { resetOn: [q, statusFilter, periode] })
const partnerPages = usePagination<any>(() => data.value?.rekap || [], { resetOn: periode })

async function load() {
  const request = ++loadRequest
  data.value = null
  loadError.value = ''
  if (!periode.value) { loading.value = false; return }
  loading.value = true
  try {
    const result = await api(`/api/keuangan?periode=${periode.value}`)
    if (request === loadRequest) data.value = result
  } catch (e) {
    if (request === loadRequest) loadError.value = errMsg(e)
  } finally {
    if (request === loadRequest) loading.value = false
  }
}
onMounted(load)
watch(periode, load)

async function generate() {
  if (generating.value || !periode.value) return
  generating.value = true
  try {
    const res = await api<any>('/api/keuangan/generate', { method: 'POST', body: { periode: periode.value } })
    toast(res.dibuat ? `${res.dibuat} tagihan dibuat untuk ${periodeLabel(periode.value)}` : 'Semua pelanggan aktif sudah punya tagihan periode ini', res.dibuat ? 'success' : 'info')
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    generating.value = false
  }
}

async function toggleBayar(t: any) {
  if (paying.value !== null) return
  if (t.status !== 'Lunas') { paymentBill.value = t; paymentDate.value = localISODate(); paymentMethod.value = 'Transfer'; return }
  const baru = t.status === 'Lunas' ? 'Belum' : 'Lunas'
  if (!await useConfirm().ask(baru === 'Lunas' ? `Catat pembayaran ${fmtRp(t.jumlah)} dari ${t.pelanggan_nama} sebagai lunas?` : `Batalkan status lunas tagihan ${t.pelanggan_nama}? Tagihan akan kembali masuk ke daftar belum dibayar.`)) return
  paying.value = t.id
  try {
    await api(`/api/tagihan/${t.id}`, { method: 'PUT', body: { status: baru } })
    toast(baru === 'Lunas' ? `Tagihan ${t.pelanggan_nama} ditandai lunas` : `Tagihan ${t.pelanggan_nama} dibuka kembali`)
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    paying.value = null
  }
}

async function recordPayment() {
  if (!paymentBill.value || paying.value !== null) return
  paying.value = paymentBill.value.id
  try { await api(`/api/tagihan/${paymentBill.value.id}`, { method: 'PUT', body: { status: 'Lunas', metode: paymentMethod.value, tanggal_bayar: paymentDate.value } }); paymentBill.value = null; toast('Pembayaran dicatat. Sinkronkan tagihan di Pembukuan untuk memperbarui jurnal.'); await load() }
  catch (e) { toast(errMsg(e), 'error') } finally { paying.value = null }
}

function buatFAB(r: any) {
  const st = useState<any>('docIntent', () => null)
  st.value = { jenis: 'FAB', mitra_id: r.mitra_id, periode: periode.value }
  navigateTo('/dokumen')
}
</script>

<template>
  <div class="toolbar">
    <label style="font-size:13px;font-weight:600;display:flex;align-items:center;gap:8px">
      Periode
      <input v-model="periode" type="month" class="ctrl" style="width:auto" />
    </label>
    <button class="btn btn-primary" :disabled="generating || loading || !periode" @click="generate">
      <Ico name="plus" :size="15" /> {{ generating ? 'Memproses…' : 'Buat tagihan bulanan' }}
    </button>
    <span class="cell-sub">Membuat tagihan untuk semua pelanggan aktif yang belum ditagih pada periode ini.</span>
  </div>

  <div v-if="loading && !data" class="empty"><p>Memuat data keuangan…</p></div>
  <div v-else-if="loadError" class="empty" role="alert"><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
  <div v-else-if="!periode" class="empty"><p>Pilih bulan untuk melihat tagihan dan bagi hasil mitra.</p></div>

  <template v-else-if="data">
    <div class="stat-grid">
      <div class="card stat">
        <div class="stat-ico ico-blue"><Ico name="money" :size="20" /></div>
        <div>
          <div class="val" style="font-size:17px">{{ fmtRp(data.summary.total) }}</div>
          <div class="lbl">Total Tagihan · {{ periodeLabel(periode) }}</div>
        </div>
      </div>
      <div class="card stat">
        <div class="stat-ico ico-green"><Ico name="check" :size="20" /></div>
        <div>
          <div class="val" style="font-size:17px">{{ fmtRp(data.summary.lunas) }}</div>
          <div class="lbl">Sudah Lunas ({{ data.summary.pct }}%)</div>
        </div>
      </div>
      <div class="card stat">
        <div class="stat-ico ico-amber"><Ico name="clock" :size="20" /></div>
        <div>
          <div class="val" style="font-size:17px">{{ fmtRp(data.summary.belum) }}</div>
          <div class="lbl">Belum Dibayar</div>
        </div>
      </div>
      <div class="card stat">
        <div class="stat-ico ico-blue"><Ico name="file" :size="20" /></div>
        <div>
          <div class="val">{{ data.summary.jumlah_tagihan }}</div>
          <div class="lbl">Jumlah Tagihan</div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom:20px">
      <div class="card-head">
        <div><h3>Invoice pelanggan</h3><p class="cell-sub">{{ periodeLabel(periode) }}. Buka preview untuk melihat invoice dan menyimpan PDF.</p></div>
        <span class="chip">{{ data.tagihan.length }} invoice</span>
      </div>
      <div v-if="!data.tagihan.length" class="empty">
        <div class="e-ico"><Ico name="money" /></div>
        <p>Belum ada tagihan pada periode ini.</p>
        <button class="btn btn-primary" :disabled="generating" @click="generate"><Ico name="plus" :size="15" /> {{ generating ? 'Memproses…' : 'Buat tagihan bulanan' }}</button>
      </div>
      <div v-else>
        <div class="archive-toolbar"><div class="search-box"><Ico name="search" :size="16" /><input v-model="q" aria-label="Cari tagihan" placeholder="Cari pelanggan atau nomor invoice…" /></div><select v-model="statusFilter" class="ctrl" style="width:auto" aria-label="Status pembayaran"><option value="">Semua pembayaran</option><option value="Belum">Belum dibayar</option><option value="Lunas">Lunas</option></select></div>
        <div class="table-wrap">
        <table class="tbl">
          <thead>
            <tr>
              <th>Pelanggan / invoice</th>
              <th>Paket</th>
              <th>Mitra</th>
              <th class="num">Jumlah</th>
              <th>Status</th>
              <th>Tgl Bayar</th>
              <th style="text-align:right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in billPages.rows" :key="t.id">
              <td class="cell-main" style="font-size:13.5px">{{ t.pelanggan_nama }}<button class="invoice-number" :aria-label="`Preview invoice ${invoiceNumber(t)} untuk ${t.pelanggan_nama}`" @click="invoice=t">{{ invoiceNumber(t) }}</button></td>
              <td style="font-size:13px">{{ t.paket_nama }}</td>
              <td style="font-size:13px">{{ t.mitra_nama }}</td>
              <td class="num">{{ fmtRp(t.jumlah) }}</td>
              <td>
                <span class="badge" :class="t.status === 'Lunas' ? 'b-success' : 'b-warn'">{{ t.status }}</span>
                <div v-if="t.metode" class="cell-sub">{{ t.metode }}</div>
              </td>
              <td class="cell-sub">{{ t.tanggal_bayar ? fmtDate(t.tanggal_bayar) : '-' }}</td>
              <td>
                <div class="row-actions">
                  <button class="btn btn-sec btn-sm" :aria-label="`Preview invoice untuk ${t.pelanggan_nama}`" @click="invoice=t"><Ico name="eye" :size="15" />Preview invoice</button>
                  <button class="btn btn-sm" :disabled="paying !== null" :class="t.status === 'Lunas' ? 'btn-sec' : 'btn-primary'" @click="toggleBayar(t)">
                    {{ paying === t.id ? 'Menyimpan…' : t.status === 'Lunas' ? 'Batalkan lunas' : 'Tandai lunas' }}
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filteredBills.length"><td colspan="7" class="empty">Tidak ada tagihan yang cocok dengan pencarian.</td></tr>
          </tbody>
        </table>
        </div>
      </div>
      <TablePagination v-model:page="billPages.page" v-model:page-size="billPages.pageSize" :total="billPages.total" label="tagihan" :disabled="loading" />
    </div>

    <div class="card">
      <div class="card-head">
        <h3>Rekap per Mitra — dasar penerbitan FAB</h3>
      </div>
      <div class="table-wrap">
        <table class="tbl">
          <thead>
            <tr>
              <th>Mitra</th>
              <th class="num">Pelanggan Aktif</th>
              <th class="num">Total Tagihan</th>
              <th class="num">Lunas</th>
              <th class="num">Share</th>
              <th class="num">Bagi Hasil Mitra</th>
              <th style="text-align:right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in partnerPages.rows" :key="r.mitra_id">
              <td class="cell-main">{{ r.mitra_nama }}</td>
              <td class="num">{{ r.pelanggan_aktif }}</td>
              <td class="num">{{ fmtRp(r.total_tagihan) }}</td>
              <td class="num">{{ fmtRp(r.total_lunas) }}</td>
              <td class="num">{{ r.share_pct }}%</td>
              <td class="num cell-main">{{ fmtRp(r.nilai) }}</td>
              <td>
                <div class="row-actions">
                  <button class="btn btn-sm btn-sec" @click="buatFAB(r)"><Ico name="file" :size="13" /> Terbitkan FAB</button>
                </div>
              </td>
            </tr>
            <tr v-if="!data.rekap.length">
              <td colspan="7" style="text-align:center;color:var(--muted)">Belum ada mitra.</td>
            </tr>
          </tbody>
        </table>
      </div>
      <TablePagination v-model:page="partnerPages.page" v-model:page-size="partnerPages.pageSize" :total="partnerPages.total" label="rekap mitra" :disabled="loading" />
    </div>
  </template>
  <InvoicePreview :bill="invoice" @close="invoice=null" />
  <UiModal :open="!!paymentBill" title="Catat pembayaran tagihan" :busy="paying !== null" @close="paymentBill = null"><form id="bill-payment" @submit.prevent="recordPayment"><p class="op-text">{{ paymentBill?.pelanggan_nama }} · {{ fmtRp(paymentBill?.jumlah || 0) }}</p><div class="grid-form"><div class="field"><label for="payment-date">Tanggal pembayaran</label><input id="payment-date" v-model="paymentDate" type="date" :max="localISODate()" class="ctrl" required /></div><div class="field"><label for="payment-method">Metode pembayaran</label><select id="payment-method" v-model="paymentMethod" class="ctrl"><option>Transfer</option><option>Tunai</option></select></div></div><p class="op-text">Transfer masuk ke akun Bank, tunai ke Kas ketika tagihan disinkronkan ke pembukuan.</p></form><template #footer><button class="btn btn-sec" :disabled="paying !== null" @click="paymentBill = null">Batal</button><button class="btn btn-primary" :disabled="paying !== null" type="submit" form="bill-payment">Simpan pembayaran</button></template></UiModal>
</template>

<style scoped>
.invoice-number{display:block;padding:3px 0 0;background:none;border:0;color:var(--accent);font-size:10px;text-align:left;font-weight:400;white-space:nowrap}.invoice-number:hover{text-decoration:underline}.row-actions{align-items:center}.card-head .chip{white-space:nowrap}@media(max-width:600px){.card-head{align-items:flex-start}.card-head .cell-sub{font-size:11px}.card-head .chip{font-size:10px}.row-actions .btn{min-height:38px}}
</style>


