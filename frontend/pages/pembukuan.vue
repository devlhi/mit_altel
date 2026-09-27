<script setup lang="ts">
import { exportCsv } from '~/lib/exportCsv'
const from = ref(localISODate().slice(0, 7) + '-01'), to = ref(localISODate())
const data = ref<any>(null), loading = ref(true), loadError = ref(''), busy = ref(false)
const tab = ref('neraca'), pops = ref<any[]>([])
const tabs = [{ id: 'neraca', name: 'Neraca' }, { id: 'profit', name: 'Laba-rugi' }, { id: 'cash', name: 'Arus kas' }, { id: 'journals', name: 'Jurnal' }, { id: 'audit', name: 'Riwayat' }]
const journalPages = usePagination(() => data.value?.journals || [], { resetOn: [from, to] })
const auditPages = usePagination(() => data.value?.audit || [])
const totals = computed(() => data.value?.totals || {})
const ready = computed(() => !!data.value?.start_date)
let request = 0
async function load() {
  const id = ++request; loading.value = true; loadError.value = ''
  try { const [report, sites] = await Promise.all([api(`/api/pembukuan?from=${tab.value === 'neraca' ? to.value.slice(0, 4) + '-01-01' : from.value}&to=${to.value}`), api('/api/pop')]); if (id === request) { data.value = report; pops.value = sites } }
  catch (e) { if (id === request) loadError.value = errMsg(e) }
  finally { if (id === request) loading.value = false }
}
onMounted(load)
function group(kind: string) { return (data.value?.accounts || []).filter((a: any) => a.kind === kind && (tab.value === 'profit' ? a.period : a.balance) !== 0) }
const opening = ref(false), journalOpen = ref(false), detail = ref<any>(null), auditDetail = ref<any>(null)
const openingDate = ref(localISODate()), zero = ref(false)
const initialLines = () => [{ account: '1102', debit: 0, credit: 0 }, { account: '3101', debit: 0, credit: 0 }]
const openingLines = ref(initialLines())
const form = ref<any>({}), preset = ref('expense')
const presets: Record<string, any> = {
  expense: { debit: '5199', credit: '1102', flow: 'operating' },
  asset: { debit: '1201', credit: '1102', flow: 'investing' },
  capital: { debit: '1102', credit: '3101', flow: 'financing' },
  transfer: { debit: '1102', credit: '1101', flow: 'operating' },
  partner: { debit: '5106', credit: '1102', flow: 'operating' },
  bhp: { debit: '5107', credit: '2102', flow: 'operating' },
  uso: { debit: '5108', credit: '2103', flow: 'operating' },
  manual: { debit: '', credit: '', flow: 'operating' }
}
function applyPreset() { const p = presets[preset.value]; form.value.flow = p.flow; form.value.lines = [{ account: p.debit, debit: 0, credit: 0 }, { account: p.credit, debit: 0, credit: 0 }] }
function openJournal() { form.value = { date: localISODate(), description: '', reference: '', pop_id: 0 }; preset.value = 'expense'; applyPreset(); journalOpen.value = true }
async function saveOpening() {
  busy.value = true
  try { await api('/api/pembukuan/opening', { method: 'POST', body: { date: openingDate.value, lines: openingLines.value, zero: zero.value } }); opening.value = false; from.value = openingDate.value; if (to.value < from.value) to.value = from.value; toast('Saldo awal tersimpan'); await load() }
  catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
async function saveJournal() {
  busy.value = true
  try { await api('/api/pembukuan/journal', { method: 'POST', body: form.value }); journalOpen.value = false; tab.value = 'journals'; toast('Jurnal dicatat'); await load() }
  catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
async function sync() {
  busy.value = true
  try { const result = await api('/api/pembukuan/sync', { method: 'POST' }); toast(`${result.changes} perubahan tagihan dibukukan`); await load() }
  catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
const voidReason = ref('')
async function cancelJournal() {
  busy.value = true
  try { await api(`/api/pembukuan/journal/${detail.value.id}/void`, { method: 'POST', body: { reason: voidReason.value } }); detail.value = null; toast('Jurnal dibatalkan; riwayat tetap tersimpan'); await load() }
  catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
function accountName(code: string) { return data.value?.accounts.find((a: any) => a.code === code)?.name || code }
const flowName: Record<string, string> = { operating: 'Operasional', investing: 'Investasi', financing: 'Pendanaan', opening: 'Saldo awal' }
function download() {
  const rows: (string | number)[][] = [['Pembukuan internal', data.value.from, data.value.to], ['Mulai pembukuan', data.value.start_date], ['Perubahan tagihan belum disinkronkan', data.value.pending], ['Akun', 'Nama', 'Kelompok', 'Debit kumulatif', 'Kredit kumulatif', 'Saldo akhir', 'Mutasi periode']]
  for (const a of data.value.accounts) rows.push([a.code, a.name, a.kind, a.debit, a.credit, a.balance, a.period])
  rows.push([], ['Ringkasan', 'Nilai']); for (const [key, val] of Object.entries(totals.value)) rows.push([key, Number(val)])
  rows.push([], ['Tanggal', 'ID jurnal', 'Uraian', 'Referensi', 'Akun', 'Debit', 'Kredit', 'Sumber', 'Status', 'POP'])
  for (const j of data.value.journals) for (const l of j.lines) rows.push([j.date, j.id, j.description, j.reference, accountName(l.account), l.debit, l.credit, j.source || 'Manual', j.void ? 'Dibatalkan' : 'Aktif', pops.value.find(p => p.id === j.pop_id)?.nama || ''])
  exportCsv(`pembukuan-${data.value.from}-${data.value.to}.csv`, rows)
}
</script>
<template>
  <div class="card op-hero"><div><h2>Laporan keuangan</h2><p>Posisi keuangan per tanggal, dengan rincian setiap akun.</p></div><div v-if="ready" class="op-actions"><button class="btn btn-sec" :disabled="busy || loading || !!loadError" @click="download">Ekspor CSV</button><button class="btn btn-primary" :disabled="busy" @click="openJournal">Tambah jurnal</button></div></div>
  <form class="op-filters" @submit.prevent="load"><label v-if="tab !== 'neraca'">Dari tanggal<input v-model="from" type="date" class="ctrl" required /></label><label>{{ tab === 'neraca' ? 'Tanggal posisi neraca' : 'Sampai tanggal' }}<input v-model="to" type="date" class="ctrl" required /></label><button class="btn btn-sec" :disabled="loading">Tampilkan</button><div class="spacer"></div><button v-if="ready" class="btn btn-sec" type="button" :disabled="busy" @click="sync">{{ busy ? 'Memproses…' : 'Sinkronkan tagihan' }}</button></form>
  <div v-if="loading" class="empty">Memuat pembukuan…</div>
  <div v-else-if="loadError" class="op-note op-warning" role="alert">{{ loadError }} <button class="btn btn-sec" @click="load">Coba lagi</button></div>
  <template v-else-if="data">
    <div v-if="!ready" class="card op-empty" style="margin-top:20px"><span class="op-eyebrow">Mulai dengan saldo sebenarnya</span><h3>Pembukuan Anda belum dimulai.</h3><p class="op-text">Isi kas/bank, piutang, aset, utang, dan modal pada tanggal mulai. Belum ada saldo atau transaksi contoh yang dimasukkan.</p><button class="btn btn-primary" @click="opening = true">Isi saldo awal</button><div class="op-steps"><div><b>01 · Tetapkan saldo awal</b><span>Jumlah debit harus sama dengan kredit. Saldo awal hanya diisi satu kali.</span></div><div><b>02 · Sinkronkan tagihan</b><span>Tagihan menjadi piutang dan pendapatan. Pembayaran tercatat masuk ke kas atau bank.</span></div><div><b>03 · Catat biaya & aset</b><span>Masukkan beban, pembelian peralatan, utang, serta modal melalui jurnal.</span></div></div></div>
    <template v-if="data">
      <div v-if="data.pending || data.warning" class="op-note op-warning" role="status"><strong>{{ data.warning || `${data.pending} perubahan tagihan belum dibukukan.` }}</strong>Laporan mencerminkan jurnal yang sudah tercatat. Sinkronkan untuk memasukkan perubahan tagihan terbaru.</div>
      <div v-if="(tab === 'neraca' ? data.to : data.from) < data.start_date" class="op-note op-warning">Pembukuan dimulai {{ data.start_date }}. Laporan sebelum tanggal tersebut tidak lengkap.</div>
      <div v-if="tab !== 'neraca'" class="op-metrics"><div class="card op-metric"><span>Total aset</span><strong>{{ fmtRp(totals.asset) }}</strong><small>Posisi {{ data.to }}</small></div><div class="card op-metric"><span>Total kewajiban</span><strong>{{ fmtRp(totals.liability) }}</strong><small>Utang yang tercatat</small></div><div class="card op-metric"><span>Ekuitas</span><strong>{{ fmtRp(totals.equity) }}</strong><small>Termasuk hasil usaha kumulatif</small></div><div class="card op-metric"><span>Laba / rugi periode</span><strong>{{ fmtRp(totals.profit) }}</strong><small>{{ data.from }} — {{ data.to }}</small></div></div>
      <div class="op-tabs" role="group" aria-label="Jenis laporan"><button v-for="t in tabs" :key="t.id" :class="{ active: tab === t.id }" :aria-pressed="tab === t.id" @click="tab = t.id; load()">{{ t.name }}</button></div>
      <BalanceSheet v-if="tab === 'neraca'" :report="data" />
      <div v-if="tab === 'profit'" class="op-split"><div class="card"><div class="op-panel-head"><h3>Pendapatan</h3></div><StatementTable :rows="group('income')" label="pendapatan" period /><div class="op-total"><span>Total pendapatan</span><span>{{ fmtRp(totals.income) }}</span></div></div><div class="card"><div class="op-panel-head"><h3>Beban usaha</h3></div><StatementTable :rows="group('expense')" label="beban" period /><div class="op-total"><span>Total beban</span><span>{{ fmtRp(totals.expense) }}</span></div></div><div class="op-note"><strong>Laba / rugi: {{ fmtRp(totals.profit) }}</strong>Pendapatan dicatat saat tagihan dibukukan. Biaya dan bagi hasil mitra perlu dicatat lewat jurnal.</div></div>
      <div v-if="tab === 'cash'" class="card"><div class="op-panel-head"><h3>Pergerakan kas & bank</h3><span class="cell-sub">{{ data.from }} — {{ data.to }}</span></div><div class="op-body"><div class="op-kv"><span>Saldo sebelum periode</span><b>{{ fmtRp(totals.cash_open) }}</b></div><div v-if="totals.opening_cash" class="op-kv"><span>Saldo awal pembukuan dalam periode</span><b>{{ fmtRp(totals.opening_cash) }}</b></div><div v-for="key in ['operating', 'investing', 'financing']" :key="key" class="op-kv"><span>{{ flowName[key] }}</span><b>{{ fmtRp(totals[key]) }}</b></div><p class="op-text">Arus kas berdasarkan mutasi akun Kas dan Bank serta kelompok jurnal. Transfer antar-kas/bank saling menghapus pada total.</p></div><div class="op-total"><span>Saldo akhir kas & bank</span><span>{{ fmtRp(totals.cash_close) }}</span></div></div>
      <div v-if="tab === 'journals'" class="card"><div class="op-panel-head"><h3>Jurnal transaksi</h3><span class="cell-sub">{{ journalPages.total }} catatan</span></div><div class="table-wrap"><table class="tbl"><thead><tr><th>Tanggal / ID</th><th>Uraian</th><th>Sumber</th><th class="num">Debit</th><th>Status</th><th>Detail</th></tr></thead><tbody><tr v-for="j in journalPages.rows" :key="j.id"><td>{{ j.date }}<div class="cell-sub">JRN-{{ j.id }}</div></td><td class="cell-main">{{ j.description }}<div class="cell-sub">{{ j.reference }}</div></td><td>{{ j.source === 'opening' ? 'Saldo awal' : j.source ? 'Tagihan' : 'Manual' }}</td><td class="num">{{ fmtRp(j.lines.reduce((n: number, l: any) => n + l.debit, 0)) }}</td><td><span class="op-status" :class="{ off: j.void }">{{ j.void ? 'Dibatalkan' : 'Aktif' }}</span></td><td><button class="btn btn-sec btn-sm" @click="detail = j; voidReason = ''">Lihat</button></td></tr><tr v-if="!journalPages.total"><td colspan="6" class="empty">Belum ada jurnal dalam periode ini.</td></tr></tbody></table></div><TablePagination v-model:page="journalPages.page" v-model:page-size="journalPages.pageSize" :total="journalPages.total" label="jurnal" /></div>
      <div v-if="tab === 'audit'" class="card"><div class="op-panel-head"><h3>Riwayat perubahan</h3><span class="cell-sub">Seluruh periode</span></div><div class="table-wrap"><table class="tbl"><thead><tr><th>Waktu</th><th>Aktivitas</th><th>Rincian</th></tr></thead><tbody><tr v-for="a in auditPages.rows" :key="a.id"><td>{{ new Date(a.at).toLocaleString('id-ID') }}</td><td>{{ a.action }}</td><td><button class="btn btn-sec btn-sm" @click="auditDetail = a">Lihat perubahan</button></td></tr><tr v-if="!auditPages.total"><td colspan="3" class="empty">Belum ada riwayat.</td></tr></tbody></table></div><TablePagination v-model:page="auditPages.page" v-model:page-size="auditPages.pageSize" :total="auditPages.total" label="riwayat pembukuan" /></div>
      <p v-if="ready" class="op-text">Mulai pembukuan: {{ data.start_date }}. Tagihan dibukukan pada tanggal 1 periode tagihan; penerimaan mengikuti tanggal pembayaran. Transfer masuk ke akun Bank, pembayaran tunai ke Kas. Pembayaran tagihan lama mengurangi piutang saldo awal.</p>
    </template>
  </template>
  <UiModal :open="opening" title="Tetapkan saldo awal" wide :busy="busy" @close="opening = false"><form id="opening-form" @submit.prevent="saveOpening"><div class="field"><label for="opening-date">Tanggal mulai pembukuan</label><input id="opening-date" v-model="openingDate" type="date" class="ctrl" required /></div><div class="op-note">Isi saldo tepat sebelum transaksi pada tanggal mulai. Aset biasanya debit; utang dan modal kredit. Akumulasi penyusutan dicatat kredit. Piutang harus mencakup tagihan lama yang belum dibayar.<strong>Koreksi setelah disimpan dilakukan melalui jurnal penyesuaian.</strong></div><JournalEditor v-model="openingLines" :accounts="(data?.accounts || []).filter((a: any) => ['asset', 'liability', 'equity'].includes(a.kind))" /><label class="op-text"><input v-model="zero" type="checkbox" /> Perusahaan mulai tanpa saldo; saya mengonfirmasi saldo awal nol jika seluruh nilai kosong.</label></form><template #footer><button class="btn btn-sec" :disabled="busy" @click="opening = false">Batal</button><button class="btn btn-primary" type="submit" form="opening-form" :disabled="busy">Simpan saldo awal</button></template></UiModal>
  <UiModal :open="journalOpen" title="Catat jurnal transaksi" wide :busy="busy" @close="journalOpen = false"><form id="journal-form" @submit.prevent="saveJournal"><div class="grid-form"><div class="field"><label for="journal-preset">Jenis transaksi</label><select id="journal-preset" v-model="preset" class="ctrl" @change="applyPreset"><option value="expense">Biaya operasional</option><option value="asset">Pembelian aset</option><option value="capital">Setoran modal</option><option value="transfer">Transfer kas / bank</option><option value="partner">Pembayaran bagi hasil mitra</option><option value="bhp">Pencadangan BHP</option><option value="uso">Pencadangan USO</option><option value="manual">Jurnal penyesuaian / lainnya</option></select></div><div class="field"><label for="journal-date">Tanggal transaksi</label><input id="journal-date" v-model="form.date" type="date" :min="data?.start_date" class="ctrl" required /></div><div class="field full"><label for="journal-desc">Uraian</label><input id="journal-desc" v-model="form.description" class="ctrl" required maxlength="500" placeholder="Contoh: Pembayaran bandwidth September" /></div><div class="field"><label for="journal-ref">Referensi bukti</label><input id="journal-ref" v-model="form.reference" class="ctrl" maxlength="200" /></div><div class="field"><label for="journal-pop">Alokasi POP</label><select id="journal-pop" v-model.number="form.pop_id" class="ctrl"><option :value="0">Umum / kantor</option><option v-for="p in pops" :key="p.id" :value="p.id">{{ p.kode }} · {{ p.nama }}</option></select></div><div class="field full"><label for="journal-flow">Kelompok arus kas</label><select id="journal-flow" v-model="form.flow" class="ctrl"><option value="operating">Operasional</option><option value="investing">Investasi</option><option value="financing">Pendanaan</option></select></div></div><div class="op-note">Isi nilai debit dan kredit yang sama. Pencadangan BHP/USO mencatat beban serta utang; pembayaran utang dicatat dengan debit akun utang dan kredit Bank.</div><JournalEditor v-model="form.lines" :accounts="data?.accounts || []" /></form><template #footer><button class="btn btn-sec" :disabled="busy" @click="journalOpen = false">Batal</button><button class="btn btn-primary" :disabled="busy" form="journal-form" type="submit">Catat jurnal</button></template></UiModal>
  <UiModal :open="!!detail" :title="`Jurnal JRN-${detail?.id || ''}`" :busy="busy" @close="detail = null"><template v-if="detail"><p class="op-text">{{ detail.date }} · {{ detail.description }}<br>{{ detail.reference }} · {{ flowName[detail.flow] }}</p><div v-for="(l, i) in detail.lines" :key="i" class="op-kv"><span>{{ accountName(l.account) }}</span><b>{{ l.debit ? `D ${fmtRp(l.debit)}` : `K ${fmtRp(l.credit)}` }}</b></div><div v-if="!detail.source && !detail.void" class="op-note op-warning">Pembatalan mengeluarkan jurnal dari laporan dan menyimpan alasannya pada riwayat.<form @submit.prevent="cancelJournal"><label for="void-reason">Alasan pembatalan</label><input id="void-reason" v-model="voidReason" class="ctrl" minlength="5" maxlength="500" required /><button class="btn btn-sec" style="margin-top:10px" :disabled="busy">Batalkan jurnal ini</button></form></div></template></UiModal>
  <UiModal :open="!!auditDetail" title="Rincian perubahan" @close="auditDetail = null"><p class="op-text">{{ auditDetail?.action }}</p><pre class="op-audit">{{ auditDetail ? JSON.stringify(JSON.parse(auditDetail.detail), null, 2) : '' }}</pre></UiModal>
</template>
