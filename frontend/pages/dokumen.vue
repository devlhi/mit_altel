<script setup lang="ts">
const list = ref<any[]>([])
const mitras = ref<any[]>([])
const loading = ref(true)
const loadError = ref('')
const q = ref('')
const fJenis = ref('')
const fPeriode = ref('')
const viewMode = ref<'grid' | 'list'>('grid')
const { data: settings } = useAppSettings()
const createOpen = ref(false)
const saving = ref(false)
const createForm = ref({ jenis: 'PKS', mitra_id: '', tanggal: localISODate(), periode: localISODate().slice(0, 7), masa_bulan: 12 })
const finance = ref<any>(null)
const financeLoading = ref(false)
const financeError = ref('')
const preview = ref<any>(null)
const previewHtml = ref('')
const previewLoading = ref(false)
const previewError = ref('')
const frameReady = ref(false)
const docFrame = ref<HTMLIFrameElement | null>(null)
const waTarget = ref<any>(null), waHistory = ref<any[]>([])
const waStatus = (d: any) => (({ accepted: 'Diterima MPWA', failed: 'Ditolak MPWA', sending: 'Diproses MPWA', unknown: 'Periksa MPWA' } as Record<string, string>)[waHistory.value.find(m => m.kind === 'document' && m.source_id === d.id)?.status as string] || (d.kirim?.length ? 'Pengiriman manual tercatat' : 'Belum dikirim'))
const share = ref<any>(null)
const shareVia = ref('wa')
const shareOpened = ref(false)
const recording = ref(false)
let financeRequest = 0
let previewRequest = 0

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const [docs, partners, messages] = await Promise.all([api('/api/dokumen'), api('/api/mitra'), api('/api/whatsapp/history')])
    waHistory.value = messages
    list.value = docs; mitras.value = partners
  } catch (e) { loadError.value = errMsg(e) }
  finally { loading.value = false }
}
onMounted(async () => {
  await load()
  viewMode.value = localStorage.getItem('isp_document_view') === 'list' ? 'list' : 'grid'
  const previewIntent = useState<any>('docPreviewIntent', () => null)
  if (previewIntent.value) {
    const found = list.value.find(d => d.id === previewIntent.value)
    if (found) await bukaPreview(found)
    previewIntent.value = null
  }
  const intent = useState<any>('docIntent', () => null)
  if (intent.value) {
    openCreate(intent.value.jenis, intent.value.mitra_id, intent.value.periode)
    intent.value = null
  }
})
const filtered = computed(() => list.value.filter(d => {
  const term = q.value.trim().toLowerCase()
  return (!term || `${d.nomor} ${d.mitra_nama}`.toLowerCase().includes(term)) &&
    (!fJenis.value || d.jenis === fJenis.value) && (!fPeriode.value || (d.periode || d.tanggal.slice(0, 7)) === fPeriode.value)
}))
const pages = usePagination(filtered, { pageSize: 9, resetOn: [q, fJenis, fPeriode] })
const visibleDocs = computed(() => pages.rows)
const pksCount = computed(() => list.value.filter(d => d.jenis === 'PKS').length)
const fabCount = computed(() => list.value.filter(d => d.jenis === 'FAB').length)
const selectedMitra = computed(() => mitras.value.find(m => String(m.id) === String(createForm.value.mitra_id)))
const fabSummary = computed(() => finance.value?.rekap?.find((r: any) => String(r.mitra_id) === String(createForm.value.mitra_id)))
const duplicateFAB = computed(() => createForm.value.jenis === 'FAB' && list.value.some(d => d.jenis === 'FAB' && String(d.mitra_id) === String(createForm.value.mitra_id) && d.periode === createForm.value.periode))
const canIssue = computed(() => !!selectedMitra.value && !!createForm.value.tanggal && !saving.value && !duplicateFAB.value &&
  (createForm.value.jenis === 'PKS' || (!financeLoading.value && !financeError.value && !!fabSummary.value)))
watch(viewMode, mode => { if (import.meta.client) localStorage.setItem('isp_document_view', mode) })
function resetFilters() { q.value = ''; fJenis.value = ''; fPeriode.value = '' }
function openCreate(jenis = 'PKS', mitraId: any = '', periode = '') {
  createForm.value = { jenis, mitra_id: mitraId ? String(mitraId) : '', tanggal: localISODate(), periode: periode || localISODate().slice(0, 7), masa_bulan: settings.value?.masa_pks_bulan || 12 }
  createOpen.value = true
}
async function loadFinance() {
  const request = ++financeRequest
  finance.value = null; financeError.value = ''; financeLoading.value = false
  if (!createOpen.value || createForm.value.jenis !== 'FAB' || !createForm.value.periode) return
  financeLoading.value = true
  try {
    const result = await api(`/api/keuangan?periode=${encodeURIComponent(createForm.value.periode)}`)
    if (request === financeRequest) finance.value = result
  } catch (e) { if (request === financeRequest) financeError.value = errMsg(e) }
  finally { if (request === financeRequest) financeLoading.value = false }
}
watch([createOpen, () => createForm.value.jenis, () => createForm.value.periode], loadFinance)
async function terbitkan() {
  if (!canIssue.value) return
  saving.value = true
  try {
    const d = await api<any>('/api/dokumen', { method: 'POST', body: { ...createForm.value, mitra_id: Number(createForm.value.mitra_id) } })
    toast(`${d.jenis} ${d.nomor} berhasil diterbitkan`)
    createOpen.value = false
    await load()
    await bukaPreview(d)
  } catch (e) { toast(errMsg(e), 'error') }
  finally { saving.value = false }
}
function clearPreviewHtml() {
  previewHtml.value = ''
}
async function bukaPreview(d: any) {
  const request = ++previewRequest
  preview.value = d; previewLoading.value = true; previewError.value = ''; frameReady.value = false
  clearPreviewHtml()
  try {
    const html = await $fetch<string>(`/api/dokumen/${d.id}/html?mode=preview`, { headers: { Authorization: `Bearer ${getToken() || ''}` }, responseType: 'text' })
    if (request === previewRequest) previewHtml.value = html
  } catch (e: any) {
    if (request === previewRequest) previewError.value = e?.status === 401 ? 'Sesi berakhir. Masuk kembali untuk membuka dokumen.' : 'Dokumen gagal dimuat. Coba buka kembali pratinjau.'
  } finally { if (request === previewRequest) previewLoading.value = false }
}
function closePreview() { ++previewRequest; preview.value = null; clearPreviewHtml() }
onBeforeUnmount(() => { ++previewRequest; ++financeRequest; clearPreviewHtml() })
function cetak() {
  try { docFrame.value?.contentWindow?.focus(); docFrame.value?.contentWindow?.print() }
  catch { toast('Dialog cetak tidak dapat dibuka. Muat ulang pratinjau dan coba lagi.', 'error') }
}
function mitraOf(d: any) { return mitras.value.find(m => m.id === d.mitra_id) || {} }
function openShare(d: any) { share.value = d; shareOpened.value = false; shareVia.value = mitraOf(d).telepon ? 'wa' : 'email' }
watch(shareVia, () => { shareOpened.value = false })
const shareMessage = computed(() => {
  if (!share.value) return ''
  const d = share.value, m = mitraOf(d)
  const summary = d.jenis === 'FAB' ? `\nPeriode: ${periodeLabel(d.periode)}\nBagi hasil: ${fmtRp(d.nilai)}` : ''
  return `Yth. ${m.pemilik || m.nama || d.mitra_nama},\n\nBerikut ${d.jenis === 'PKS' ? 'Perjanjian Kerja Sama' : 'Faktur Bisnis'} dari ${settings.value?.nama_isp || 'ISP'}.\nNomor: ${d.nomor}\nTanggal: ${fmtDate(d.tanggal)}${summary}\n\nMohon periksa dokumen PDF yang disertakan dan konfirmasi penerimaannya.\n\nTerima kasih.\n${settings.value?.nama_isp || 'ISP'}`
})
function openDraft() {
  const m = mitraOf(share.value)
  if (shareVia.value === 'wa') {
    const num = waNumber(m.telepon)
    if (!num) { toast('Nomor WhatsApp belum diisi. Lengkapi kontak di halaman Mitra.', 'error'); return }
    window.open(`https://wa.me/${num}?text=${encodeURIComponent(shareMessage.value)}`, '_blank', 'noopener,noreferrer')
  } else {
    if (!m.email) { toast('Email belum diisi. Lengkapi kontak di halaman Mitra.', 'error'); return }
    window.location.href = `mailto:${encodeURIComponent(m.email)}?subject=${encodeURIComponent(`[${share.value.jenis}] ${share.value.nomor}`)}&body=${encodeURIComponent(shareMessage.value)}`
  }
  shareOpened.value = true
}
async function recordSent() {
  if (!shareOpened.value || recording.value) return
  recording.value = true
  try {
    const res = await api<any>(`/api/dokumen/${share.value.id}/kirim`, { method: 'POST', body: { via: shareVia.value } })
    share.value.kirim = res.kirim
    if (preview.value?.id === share.value.id) preview.value.kirim = res.kirim
    share.value = null
    toast('Pengiriman dicatat berdasarkan konfirmasi Anda')
    await load()
  } catch (e) { toast(errMsg(e), 'error') }
  finally { recording.value = false }
}
async function hapus(d: any) {
  if (!await useConfirm().ask(`Hapus arsip ${d.nomor} untuk ${d.mitra_nama}? Dokumen yang sudah dikirim tetap berada pada penerima.`)) return
  try {
    await api(`/api/dokumen/${d.id}`, { method: 'DELETE' })
    if (preview.value?.id === d.id) closePreview()
    toast('Dokumen dihapus'); await load()
  } catch (e) { toast(errMsg(e), 'error') }
}
</script>

<template>
  <div class="page-head document-heading"><div><h1>Dokumen mitra</h1><p class="desc">Dari kesepakatan kerja sama sampai rekap bagi hasil.</p></div><button class="btn btn-primary" :disabled="loading || !!loadError" @click="openCreate('PKS')"><Ico name="plus" :size="18" />Buat dokumen</button></div>
  <div class="document-formats">
    <section class="document-format format-pks"><div class="format-copy"><span class="format-label"><Ico name="file" :size="15" />Perjanjian kerja sama</span><h2>Awali kerja sama<br>dengan PKS.</h2><p>Kesepakatan layanan, masa berlaku,<br>dan pembagian hasil dengan mitra.</p><button class="btn format-button" :disabled="loading || !!loadError" @click="openCreate('PKS')">Buat PKS<Ico name="plus" :size="16" /></button></div><div class="format-paper"><DocumentPaper kind="PKS" :company="settings?.nama_isp" /></div></section>
    <section class="document-format format-fab"><div class="format-copy"><span class="format-label"><Ico name="wallet" :size="15" />Faktur bisnis</span><h2>Rekap bagi hasil,<br>siap dibagikan.</h2><p>Rincian tagihan dan perhitungan<br>pendapatan mitra setiap bulan.</p><button class="btn format-button" :disabled="loading || !!loadError" @click="openCreate('FAB')">Buat FAB<Ico name="plus" :size="16" /></button></div><div class="format-paper"><DocumentPaper kind="FAB" :company="settings?.nama_isp" /></div></section>
  </div>
  <NuxtLink to="/keuangan" class="invoice-library-link"><span class="invoice-link-icon"><Ico name="money" :size="22" /></span><span><b>Invoice pelanggan</b><small>Lihat tagihan, buka preview invoice, dan simpan PDF dari halaman Keuangan.</small></span><span class="invoice-link-action">Buka invoice<Ico name="chevron" :size="16" /></span></NuxtLink>
  <NuxtLink to="/surat-pop" class="invoice-library-link"><span class="invoice-link-icon"><Ico name="file" :size="22" /></span><span><b>Surat POP Komdigi</b><small>KBLI, lokasi POP, kop perusahaan, dan tanda tangan / stempel dalam surat siap cetak.</small></span><span class="invoice-link-action">Buka surat POP<Ico name="chevron" :size="16" /></span></NuxtLink>
  <section class="document-library" aria-label="Arsip dokumen" :aria-busy="loading">
    <div class="library-heading"><div><h2>Arsip dokumen<span class="heading-count">{{ list.length }}</span></h2><p>Semua dokumen yang telah Anda terbitkan.</p></div><div class="view-controls"><div class="view-toggle" role="group" aria-label="Tampilan arsip"><button :class="{active:viewMode === 'grid'}" :aria-pressed="viewMode === 'grid'" aria-label="Tampilan kartu" title="Tampilan kartu" @click="viewMode='grid'"><Ico name="grid" :size="17" /></button><button :class="{active:viewMode === 'list'}" :aria-pressed="viewMode === 'list'" aria-label="Tampilan daftar" title="Tampilan daftar" @click="viewMode='list'"><Ico name="list" :size="18" /></button></div><button class="icon-btn" aria-label="Muat ulang dokumen" :disabled="loading" @click="load"><Ico name="refresh" :size="17" /></button></div></div>
    <div class="library-toolbar"><div class="document-tabs" role="group" aria-label="Jenis dokumen"><button v-for="kind in ['', 'PKS', 'FAB']" :key="kind" :class="{on:fJenis === kind}" :aria-pressed="fJenis === kind" @click="fJenis=kind">{{ kind || 'Semua dokumen' }}<span>{{ kind === 'PKS' ? pksCount : kind === 'FAB' ? fabCount : list.length }}</span></button></div><div class="library-filters"><div class="search-box"><Ico name="search" :size="17" /><input v-model="q" aria-label="Cari dokumen" type="search" placeholder="Cari nomor atau mitra…" /></div><label class="month-filter"><span class="sr-only">Filter bulan dokumen</span><input v-model="fPeriode" class="ctrl" type="month" aria-label="Filter bulan dokumen" /></label></div></div>
    <div v-if="q || fPeriode" class="active-filters"><span>{{ filtered.length }} dokumen ditemukan</span><button @click="resetFilters">Hapus filter<Ico name="x" :size="13" /></button></div>
    <div v-if="loading" class="empty card" role="status"><p>Memuat arsip dokumen…</p></div>
    <div v-else-if="loadError" class="empty card" role="alert"><div class="e-ico"><Ico name="alert" /></div><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
    <div v-else-if="!filtered.length" class="empty card"><div class="e-ico"><Ico name="file" :size="25" /></div><h3>{{ list.length ? 'Dokumen tidak ditemukan' : 'Dokumen pertama Anda dimulai di sini' }}</h3><p style="margin-top:8px">{{ list.length ? 'Ubah pencarian atau hapus filter untuk melihat dokumen lainnya.' : 'Siapkan PKS untuk mitra baru atau FAB untuk rekap bulanan.' }}</p><button v-if="list.length" class="btn btn-sec" @click="resetFilters">Hapus filter</button><button v-else class="btn btn-primary" @click="openCreate('PKS')">Buat PKS pertama</button></div>
    <div v-else-if="viewMode === 'grid'" class="document-gallery">
      <article v-for="d in visibleDocs" :key="d.id" class="document-tile">
        <button class="tile-preview" :class="d.jenis === 'FAB' ? 'tile-fab' : 'tile-pks'" :aria-label="`Pratinjau ${d.nomor}`" @click="bukaPreview(d)"><span class="tile-type">{{ d.jenis }}</span><DocumentPaper :kind="d.jenis" :company="settings?.nama_isp" /><span class="tile-open"><Ico name="eye" :size="17" />Pratinjau dokumen</span></button>
        <div class="tile-info"><button class="document-number" @click="bukaPreview(d)">{{ d.nomor }}</button><p>{{ d.mitra_nama }}</p><div class="tile-detail"><span>{{ d.jenis === 'FAB' ? periodeLabel(d.periode) : fmtDate(d.tanggal) }}</span><b v-if="d.jenis === 'FAB'">{{ fmtRp(d.nilai) }}</b><span v-else>Perjanjian kerja sama</span></div></div>
        <div class="tile-footer"><span class="badge" :class="d.kirim?.length ? 'b-success' : 'b-muted'">{{ waStatus(d) }}</span><div class="row-actions"><button class="icon-btn" :aria-label="`Kirim WhatsApp ${d.nomor}`" title="Kirim WhatsApp" @click="waTarget = d"><Ico name="send" :size="15" /></button><button class="icon-btn danger" :aria-label="`Hapus ${d.nomor}`" title="Hapus dokumen" @click="hapus(d)"><Ico name="trash" :size="15" /></button></div></div>
      </article>
    </div>
    <div v-else class="card table-wrap document-table" tabindex="0" aria-label="Daftar dokumen, geser untuk melihat semua kolom"><table class="tbl"><thead><tr><th scope="col">Nama dokumen</th><th scope="col">Mitra</th><th scope="col">Tanggal terbit</th><th scope="col" class="num">Bagi hasil</th><th scope="col">Pengiriman</th><th scope="col" style="text-align:right">Tindakan</th></tr></thead><tbody><tr v-for="d in visibleDocs" :key="d.id"><td><div class="document-cell"><span class="doc-symbol" :class="{fab:d.jenis === 'FAB'}"><Ico name="file" :size="19" /></span><div><button class="document-number" @click="bukaPreview(d)">{{ d.nomor }}</button><div class="cell-sub">{{ d.jenis === 'FAB' ? periodeLabel(d.periode) : 'Perjanjian kerja sama' }}</div></div></div></td><td class="cell-main">{{ d.mitra_nama }}</td><td class="cell-sub">{{ fmtDate(d.tanggal) }}</td><td class="num">{{ d.jenis === 'FAB' ? fmtRp(d.nilai) : '—' }}</td><td><span class="badge" :class="d.kirim?.length ? 'b-success' : 'b-muted'">{{ d.kirim?.length ? 'Tercatat' : 'Belum dicatat' }}</span></td><td><div class="row-actions"><button class="btn btn-sec btn-sm" :aria-label="`Lihat ${d.nomor}`" @click="bukaPreview(d)"><Ico name="eye" :size="15" />Lihat</button><button class="icon-btn" :aria-label="`Kirim WhatsApp ${d.nomor}`" title="Kirim WhatsApp" @click="waTarget = d"><Ico name="send" :size="15" /></button><button class="icon-btn danger" :aria-label="`Hapus ${d.nomor}`" title="Hapus dokumen" @click="hapus(d)"><Ico name="trash" :size="15" /></button></div></td></tr></tbody></table></div>
    <TablePagination v-if="!loading && !loadError" v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" :page-sizes="[9, 18, 36]" label="dokumen" style="margin-top:20px" />
  </section>
  <UiModal :open="createOpen" :title="`Buat ${createForm.jenis}`" wide :busy="saving" @close="createOpen = false">
    <div v-if="!mitras.length" class="empty"><p>Tambahkan mitra terlebih dahulu untuk menerbitkan dokumen.</p><NuxtLink class="btn btn-primary" to="/mitra">Buka halaman Mitra</NuxtLink></div>
    <form v-else id="issue-document" @submit.prevent="terbitkan">
      <div class="radio-cards"><label v-for="kind in ['PKS', 'FAB']" :key="kind" class="radio-card" :class="{ sel: createForm.jenis === kind }"><input v-model="createForm.jenis" type="radio" :value="kind" :disabled="saving" /><div><b>{{ kind }}</b><span>{{ kind === 'PKS' ? 'Perjanjian kerja sama' : 'Rekap bagi hasil bulanan' }}</span></div></label></div>
      <div class="issue-layout"><div>
        <div class="field"><label for="doc-mitra">Mitra kerja sama <span aria-hidden="true">*</span></label><select id="doc-mitra" v-model="createForm.mitra_id" class="ctrl" required :disabled="saving"><option value="" disabled>Pilih mitra penerima</option><option v-for="m in mitras" :key="m.id" :value="String(m.id)">{{ m.nama }}{{ m.status !== 'Aktif' ? ` (${m.status})` : '' }}</option></select></div>
        <div class="field"><label for="doc-date">Tanggal terbit <span aria-hidden="true">*</span></label><input id="doc-date" v-model="createForm.tanggal" class="ctrl" type="date" required :disabled="saving" /></div>
        <div v-if="createForm.jenis === 'PKS'" class="field"><label for="doc-duration">Masa kerja sama (bulan)</label><input id="doc-duration" v-model.number="createForm.masa_bulan" class="ctrl" type="number" min="1" max="120" step="1" required :disabled="saving" /><span class="hint">Terhitung sejak tanggal terbit dokumen.</span></div>
        <div v-else class="field"><label for="doc-period">Periode tagihan</label><input id="doc-period" v-model="createForm.periode" class="ctrl" type="month" required :disabled="saving" /><span class="hint">Bagi hasil dihitung dari tagihan berstatus lunas.</span></div>
      </div><aside class="issue-summary" aria-live="polite"><h4>Ringkasan penerbitan</h4>
        <p v-if="!selectedMitra" class="cell-sub">Pilih mitra untuk melihat ringkasan dokumen.</p>
        <template v-else><strong>{{ selectedMitra.nama }}</strong><p class="cell-sub">{{ selectedMitra.pemilik || 'Pemilik belum diisi' }}<br>{{ selectedMitra.wilayah || 'Wilayah belum diisi' }}</p>
          <dl class="kv"><dt>Tanggal terbit</dt><dd>{{ fmtDate(createForm.tanggal) }}</dd><dt>Bagi hasil mitra / ISP</dt><dd>{{ selectedMitra.share_mitra }}% / {{ 100 - selectedMitra.share_mitra }}%</dd><template v-if="createForm.jenis === 'PKS'"><dt>Masa kerja sama</dt><dd>{{ createForm.masa_bulan }} bulan</dd></template></dl>
          <template v-if="createForm.jenis === 'FAB'"><p v-if="financeLoading" class="cell-sub">Menghitung rekap…</p><div v-else-if="financeError" class="notice error"><div>{{ financeError }}<br><button type="button" class="btn btn-sec btn-sm" @click="loadFinance">Coba lagi</button></div></div><template v-else-if="fabSummary"><dl class="kv"><dt>Total tagihan</dt><dd>{{ fmtRp(fabSummary.total_tagihan) }}</dd><dt>Tagihan lunas</dt><dd>{{ fmtRp(fabSummary.total_lunas) }}</dd></dl><div class="issue-total"><span>Bagi hasil untuk mitra</span><strong>{{ fmtRp(fabSummary.nilai) }}</strong></div><p v-if="!fabSummary.total_lunas" class="hint" style="margin-top:10px">Belum ada tagihan lunas. FAB akan bernilai Rp0.</p></template></template>
        </template>
      </aside></div>
      <div v-if="selectedMitra && createForm.jenis === 'FAB'" class="op-note" :class="{ 'op-warning': !selectedMitra.rekening }"><strong>Rekening penerima bagi hasil</strong><span style="white-space:pre-line">{{ selectedMitra.rekening || 'Belum diisi. Lengkapi rekening pada Edit Mitra sebelum dokumen dipakai untuk pembayaran.' }}</span></div>
      <p class="op-text">Dokumen baru menyimpan identitas perusahaan, data mitra, rekening, dan template saat diterbitkan. Perubahan Pengaturan setelahnya tidak mengubah isi dokumen baru yang sudah tersimpan.</p>
      <div v-if="duplicateFAB" class="notice warning"><Ico name="info" :size="17" /><p>FAB untuk mitra dan periode ini sudah ada. Buka dokumen tersebut untuk menghindari rekap ganda.</p></div>
      <p class="hint">Nomor dokumen dibuat otomatis saat diterbitkan. Setelah itu, periksa pratinjau sebelum menyimpan PDF atau menyiapkan pengiriman.</p>
    </form>
    <template #footer><button class="btn btn-sec" :disabled="saving" @click="createOpen = false">Batal</button><button v-if="mitras.length" class="btn btn-primary" type="submit" form="issue-document" :disabled="!canIssue"><Ico name="file" :size="16" />{{ saving ? 'Menerbitkan…' : `Terbitkan ${createForm.jenis}` }}</button></template>
  </UiModal>

  <UiModal :open="!!preview" title="Pratinjau dokumen" hg @close="closePreview">
    <template v-if="preview"><div class="doc-preview-meta"><div><b>{{ preview.nomor }}</b><p class="cell-sub">{{ preview.mitra_nama }} · {{ fmtDate(preview.tanggal) }}</p></div><span class="chip">{{ preview.jenis }} · Format A4</span></div>
      <div v-if="previewLoading" class="empty" role="status">Menyiapkan dokumen…</div><div v-else-if="previewError" class="empty" role="alert"><p>{{ previewError }}</p><button class="btn btn-sec" @click="bukaPreview(preview)">Coba lagi</button></div>
      <iframe v-else-if="previewHtml" ref="docFrame" :srcdoc="previewHtml" class="doc-frame" title="Isi dokumen siap cetak" sandbox="allow-same-origin allow-modals" @load="frameReady = true"></iframe>
      <p class="preview-hint">Pilih “Simpan PDF / cetak”, lalu pilih tujuan “Save as PDF” pada dialog cetak. PDF dapat dilampirkan melalui tombol Kirim WhatsApp.</p>
    </template>
    <template #footer><button class="btn btn-ghost" @click="closePreview">Tutup</button><button class="btn btn-ghost" @click="openShare(preview)">Draf manual / email</button><button class="btn btn-sec" @click="waTarget = preview"><Ico name="message" :size="15" /> Kirim WhatsApp</button><button class="btn btn-primary" :disabled="!frameReady || !!previewError" @click="cetak"><Ico name="printer" :size="16" /> Simpan PDF / cetak</button></template>
  </UiModal>

  <WhatsAppComposer :open="!!waTarget" kind="document" :source-id="waTarget?.id || 0" @close="waTarget = null" @sent="load" />
  <UiModal :open="!!share" title="Kirim WhatsApp" :busy="recording" @close="share = null">
    <template v-if="share"><p class="cell-main">{{ share.nomor }}</p><p class="cell-sub" style="margin-bottom:18px">Penerima: {{ share.mitra_nama }}</p>
      <div class="notice"><Ico name="info" :size="18" /><p>Simpan PDF dari pratinjau, buka draf pesan, lalu lampirkan PDF dan kirim melalui aplikasi pilihan Anda.</p></div>
      <div class="field"><label for="share-channel">Kirim melalui</label><select id="share-channel" v-model="shareVia" class="ctrl"><option value="wa">WhatsApp</option><option value="email">Email</option></select><span class="hint">{{ shareVia === 'wa' ? (mitraOf(share).telepon || 'Nomor WhatsApp belum diisi') : (mitraOf(share).email || 'Email belum diisi') }}</span></div>
      <div class="field"><label for="share-message">Draf pesan</label><textarea id="share-message" :value="shareMessage" readonly rows="8" class="ctrl"></textarea></div>
      <button class="btn btn-sec" @click="openDraft"><Ico :name="shareVia === 'wa' ? 'message' : 'mail'" :size="16" /> Buka draf {{ shareVia === 'wa' ? 'WhatsApp' : 'email' }}</button>
      <p v-if="shareOpened" class="hint" style="margin-top:15px">Aplikasi tidak dapat memverifikasi pengiriman. Setelah Anda benar-benar mengirim pesan dan PDF, klik “Saya sudah mengirim”.</p>
    </template>
    <template #footer><button class="btn btn-sec" :disabled="recording" @click="share = null">Tutup</button><button class="btn btn-primary" :disabled="!shareOpened || recording" @click="recordSent">{{ recording ? 'Mencatat…' : 'Saya sudah mengirim' }}</button></template>
  </UiModal>
</template>




