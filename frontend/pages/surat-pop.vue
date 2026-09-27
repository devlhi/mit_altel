<script setup lang="ts">
const route = useRoute()
const options = ref<any>(null), records = ref<any[]>([])
const loading = ref(true), loadError = ref(''), busy = ref(false), tab = ref('compose'), search = ref('')
const fresh = () => ({ mode: 'template', service_id: '', kbli: '', manual_number: '', locations: [{ pop_id: 0, name: '', address: '' }] })
const form = ref<any>(fresh())
const filtered = computed(() => records.value.filter(r => `${r.number} ${r.data.kbli} ${r.data.locations.map((p: any) => `${p.name} ${p.address}`).join(' ')}`.toLowerCase().includes(search.value.toLowerCase())))
const pages = usePagination(filtered, { resetOn: search })
const chosen = computed(() => options.value?.services.find((s: any) => s.id === form.value.service_id))
const preview = ref(false), previewHTML = ref(''), previewRecord = ref<any>(null), previewInput = ref<any>(null), frameReady = ref(false)
const previewCanIssue = ref(false)
const frame = ref<HTMLIFrameElement | null>(null), viewport = ref<HTMLElement | null>(null)
const sheetHeight = ref(1123), scale = ref(1), fitMode = ref('width')
const cancelRecord = ref<any>(null), cancelReason = ref('')
let observer: ResizeObserver | undefined
function resize() {
  const widthScale = ((viewport.value?.clientWidth || 818) - 24) / 794
  const heightScale = Math.max(220, window.innerHeight - 320) / sheetHeight.value
  scale.value = Math.min(1, Math.max(.15, fitMode.value === 'page' ? Math.min(widthScale, heightScale) : widthScale))
}
watch(fitMode, resize)
watch(viewport, el => { observer?.disconnect(); if (el) { observer = new ResizeObserver(resize); observer.observe(el); resize() } })
onMounted(() => window.addEventListener('resize', resize))
onBeforeUnmount(() => { observer?.disconnect(); window.removeEventListener('resize', resize) })
async function ready() {
  const doc = frame.value?.contentDocument
  if (!doc) return
  await Promise.allSettled([...doc.images].map(img => img.decode()))
  sheetHeight.value = Math.max(1123, doc.documentElement.scrollHeight, doc.body.scrollHeight)
  frameReady.value = true; resize()
}
async function load() {
  loading.value = true; loadError.value = ''
  try {
    const [o, a] = await Promise.all([api('/api/surat-pop/options'), api('/api/surat-pop')])
    options.value = o; records.value = a
    if (route.query.pop) {
      const sites = await api('/api/pop')
      const site = sites.find((p: any) => String(p.id) === route.query.pop)
      if (site) form.value.locations = [{ pop_id: site.id, name: '', address: site.alamat }]
    }
  } catch (e) { loadError.value = errMsg(e) } finally { loading.value = false }
}
onMounted(load)
function changeService() { form.value.kbli = chosen.value?.kbli || '' }
async function review() {
  if (busy.value) return
  busy.value = true
  try {
    const data = JSON.parse(JSON.stringify(form.value)); data.request_key = Array.from(crypto.getRandomValues(new Uint8Array(24)), b => b.toString(16).padStart(2, '0')).join('')
    const res = await api('/api/surat-pop/preview', { method: 'POST', body: data })
    previewCanIssue.value = res.can_issue === true
    previewInput.value = res.data; previewRecord.value = null; showPreview(res.html)
  } catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
function showPreview(html: string) { fitMode.value = 'width'; sheetHeight.value = 1123; frameReady.value = false; previewHTML.value = html; preview.value = true }
async function readHTML(id: number) { return await $fetch<string>(`/api/surat-pop/${id}/html`, { headers: { Authorization: `Bearer ${getToken() || ''}` }, responseType: 'text' }) }
async function viewRecord(record: any) {
  busy.value = true
  try { const html = await readHTML(record.id); previewRecord.value = record; previewInput.value = null; showPreview(html) }
  catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
async function issue() {
  if (!previewInput.value || !previewCanIssue.value || busy.value) return
  busy.value = true
  try {
    const record = await api('/api/surat-pop', { method: 'POST', body: previewInput.value })
    // Switch to the saved record before reading its HTML, so a failed refresh cannot issue it again.
    previewRecord.value = record; previewInput.value = null; previewHTML.value = ''; frameReady.value = false
    records.value = [record, ...records.value.filter(r => r.id !== record.id)]
    form.value = fresh(); tab.value = 'archive'
    toast(`Surat ${record.number} diterbitkan`)
    showPreview(await readHTML(record.id))
  } catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
function printLetter() { try { frame.value?.contentWindow?.focus(); frame.value?.contentWindow?.print() } catch { toast('Dialog cetak tidak dapat dibuka. Buka kembali surat.', 'error') } }
function copyRecord(r: any) {
  form.value = { ...fresh(), service_id: r.data.service_id, kbli: r.data.kbli, locations: r.data.locations.map((p: any) => ({ pop_id: p.pop_id || 0, name: '', address: p.address })) }
  tab.value = 'compose'; toast('Alamat dan KBLI disalin. Isi nomor surat baru untuk melanjutkan.')
}
async function voidLetter() {
  busy.value = true
  try { await api(`/api/surat-pop/${cancelRecord.value.id}/void`, { method: 'POST', body: { reason: cancelReason.value } }); records.value = await api('/api/surat-pop'); cancelRecord.value = null; toast('Surat dibatalkan. Arsip dan nomor tetap tersimpan.') }
  catch (e) { toast(errMsg(e), 'error') } finally { busy.value = false }
}
</script>

<template>
  <div class="pop-letter-page">
    <div class="letter-tabs" role="group" aria-label="Halaman surat POP">
      <button :class="{ active: tab === 'compose' }" :aria-pressed="tab === 'compose'" @click="tab = 'compose'"><Ico name="pencil" :size="16" /> Buat surat</button>
      <button :class="{ active: tab === 'archive' }" :aria-pressed="tab === 'archive'" @click="tab = 'archive'"><Ico name="file" :size="16" /> Arsip surat <span>{{ records.length }}</span></button>
      <NuxtLink to="/pop">Data POP <Ico name="chevron" :size="14" /></NuxtLink>
    </div>
    <div v-if="loading" class="card empty" role="status">Memuat surat POP…</div>
    <div v-else-if="loadError" class="card empty" role="alert">{{ loadError }} <button class="btn btn-sec" @click="load">Coba lagi</button></div>
    <form v-else-if="tab === 'compose'" id="pop-letter-form" class="card letter-compose" @submit.prevent="review">
      <header class="letter-compose-head"><h2>Surat POP ke Komdigi</h2><p>Isi alamat POP, nomor surat, dan pilih KBLI. Surat lengkap disiapkan otomatis.</p></header>
      <div class="letter-compose-grid">
        <section class="letter-fields" aria-label="Isian surat">
          <div class="field"><label for="letter-number">Nomor surat</label><input id="letter-number" v-model="form.manual_number" class="ctrl" required maxlength="120" placeholder="Contoh: 001/POP/ATN-KOMDIGI/IX/2026" /><small class="letter-help">Nomor surat yang akan diterbitkan.</small></div>
          <div class="field"><label for="letter-service">KBLI</label><select id="letter-service" v-model="form.service_id" class="ctrl" required @change="changeService"><option value="" disabled>Pilih KBLI</option><option v-for="s in options.services" :key="s.id" :value="s.id">{{ s.kbli }} — {{ s.description }}{{ s.scope ? ` (${s.scope})` : '' }}</option><option value="manual">KBLI lainnya (isi manual)</option></select><small v-if="chosen" class="letter-help">Nama KBLI: {{ chosen.description }}.<br>Lingkup izin: {{ chosen.name }}.<br><a v-if="chosen.source_url" :href="chosen.source_url" target="_blank" rel="noopener noreferrer">Lihat rujukan OSS</a></small><div v-if="form.service_id === 'manual'" class="manual-kbli"><label for="letter-kbli">Kode KBLI manual</label><input id="letter-kbli" v-model="form.kbli" class="ctrl" inputmode="numeric" pattern="[0-9]{5}" maxlength="5" required placeholder="5 digit KBLI" /></div></div>
          <div v-for="(p, i) in form.locations" :key="i" class="field letter-address"><div class="address-label"><label :for="`letter-pop-address-${i}`">{{ form.locations.length === 1 ? 'Alamat POP' : `Alamat POP ${i + 1}` }}</label><button v-if="form.locations.length > 1" class="btn btn-ghost btn-sm" type="button" :aria-label="`Hapus alamat POP ${i + 1}`" @click="form.locations.splice(i, 1)">Hapus</button></div><textarea :id="`letter-pop-address-${i}`" v-model="p.address" class="ctrl" rows="4" required maxlength="700" placeholder="Jalan / dusun, desa / kelurahan, kecamatan, kabupaten / kota, provinsi"></textarea></div>
          <button type="button" class="btn btn-sec btn-sm" :disabled="form.locations.length >= 30" @click="form.locations.push({ pop_id: 0, name: '', address: '' })"><Ico name="plus" :size="15" /> Tambah alamat POP</button>
          <div class="letter-compose-submit"><button class="btn btn-primary" type="submit" :disabled="busy"><Ico name="eye" :size="17" /> {{ busy ? 'Menyiapkan surat…' : 'Pratinjau surat' }}</button><span>Periksa surat sebelum diterbitkan.</span></div>
        </section>
        <aside class="letter-automatic" aria-label="Isi otomatis surat">
          <img :src="options.letterhead" alt="Kop PT Ali Telematika Nusantara" />
          <h3>Selebihnya sudah otomatis</h3>
          <dl><div><dt>Tanggal surat</dt><dd>{{ fmtDate(localISODate()) }}</dd></div><div><dt>Tujuan</dt><dd>{{ options.profile.recipient }}</dd></div><div><dt>Isi surat</dt><dd>Redaksi lanjutan dan verifikasi OSS mengikuti contoh Anda.</dd></div><div><dt>Penandatangan</dt><dd>{{ options.profile.signer }}<span>{{ options.profile.job }}</span></dd></div></dl>
          <p class="signature-info"><Ico name="file" :size="17" /><span>Kop, tanda tangan, dan stempel perusahaan sudah terpasang.</span></p>
        </aside>
      </div>
    </form>
    <section v-else class="card letter-archive"><div class="archive-toolbar"><div><h2>Arsip surat POP</h2><p class="letter-help">Surat yang telah diterbitkan beserta lokasi dan tanda tangannya.</p></div><div class="search-box"><Ico name="search" :size="16" /><input v-model="search" aria-label="Cari surat POP" placeholder="Nomor, KBLI, lokasi…" /></div></div><div class="table-wrap"><table class="tbl"><thead><tr><th>Nomor / tanggal</th><th>KBLI / lokasi</th><th>Status</th><th>Aksi</th></tr></thead><tbody><tr v-for="r in pages.rows" :key="r.id"><td class="cell-main">{{ r.number }}<div class="cell-sub">{{ fmtDate(r.data.date) }} · {{ r.data.mode === 'initial' ? 'Permohonan awal' : 'Lanjutan' }}</div></td><td>KBLI {{ r.data.kbli }} ({{ r.data.kbli_version }})<div class="cell-sub">{{ r.data.locations.map((p: any) => p.name || p.address).join('; ') }}</div></td><td><span class="op-status" :class="{ off: r.void_reason }">{{ r.void_reason ? 'Dibatalkan' : 'Terbit' }}</span><div v-if="r.void_reason" class="cell-sub">{{ r.void_reason }}</div></td><td><div class="op-actions"><button class="btn btn-sec btn-sm" :disabled="busy" @click="viewRecord(r)">Buka surat</button><button class="btn btn-ghost btn-sm" :disabled="busy" @click="copyRecord(r)">Salin isi</button><button v-if="!r.void_reason" class="btn btn-ghost btn-sm" :disabled="busy" @click="cancelRecord = r; cancelReason = ''">Batalkan</button></div></td></tr><tr v-if="!pages.total"><td colspan="4" class="empty">{{ search ? 'Surat tidak ditemukan.' : 'Belum ada surat terbit. Isi formulir pada tab Buat surat.' }}</td></tr></tbody></table></div><TablePagination v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" label="surat POP" /></section>
    <UiModal :open="preview" :title="previewRecord ? previewRecord.number : 'Pratinjau surat POP'" hg :busy="busy" @close="preview = false">
      <div class="letter-preview-status"><span>{{ previewRecord ? (previewRecord.void_reason ? 'Surat dibatalkan' : 'Surat terbit') : 'Draf siap diperiksa' }}</span><div class="letter-zoom" role="group" aria-label="Ukuran pratinjau"><span>A4 · {{ Math.round(scale * 100) }}%</span><button :class="{ active: fitMode === 'page' }" :aria-pressed="fitMode === 'page'" @click="fitMode = 'page'">Pas halaman</button><button :class="{ active: fitMode === 'width' }" :aria-pressed="fitMode === 'width'" @click="fitMode = 'width'">Pas lebar</button></div></div>
      <div v-if="previewHTML" ref="viewport" class="letter-viewport"><div class="letter-sheet-size" :style="{ width: `${794 * scale}px`, height: `${sheetHeight * scale}px` }"><iframe ref="frame" :srcdoc="previewHTML" title="Isi surat POP Komdigi" sandbox="allow-same-origin allow-modals" :style="{ height: `${sheetHeight}px`, transform: `scale(${scale})` }" @load="ready"></iframe></div></div><div v-else class="empty"><p>Surat tersimpan. Buka ulang untuk memuat tampilannya.</p><button v-if="previewRecord" class="btn btn-sec" :disabled="busy" @click="viewRecord(previewRecord)">Muat surat</button></div>
      <p class="letter-help">{{ previewRecord ? 'Untuk PDF, pilih “Save as PDF” pada dialog cetak dan matikan header/footer browser.' : 'Nomor, KBLI, dan alamat mengikuti isian Anda. Isi surat, kop, TTD, dan stempel sudah disiapkan otomatis.' }}</p>
      <template #footer><button class="btn btn-sec" :disabled="busy" @click="preview = false">{{ previewRecord ? 'Tutup' : 'Kembali ke formulir' }}</button><button v-if="!previewRecord" class="btn btn-primary" :disabled="busy || !frameReady || !previewCanIssue" @click="issue">{{ busy ? 'Menerbitkan…' : 'Terbitkan surat' }}</button><button v-else class="btn btn-primary" :disabled="busy || !frameReady" @click="printLetter"><Ico name="printer" :size="16" /> Simpan PDF / cetak</button></template>
    </UiModal>
    <UiModal :open="!!cancelRecord" title="Batalkan surat POP" :busy="busy" @close="cancelRecord = null"><p>{{ cancelRecord?.number }}</p><p class="letter-help">Nomor dan arsip tetap disimpan. Surat diberi penanda dibatalkan.</p><form id="void-pop-letter" @submit.prevent="voidLetter"><div class="field"><label for="void-letter-reason">Alasan pembatalan *</label><textarea id="void-letter-reason" v-model="cancelReason" class="ctrl" required maxlength="500" rows="3"></textarea></div></form><template #footer><button class="btn btn-sec" :disabled="busy" @click="cancelRecord = null">Kembali</button><button class="btn btn-primary" :disabled="busy" type="submit" form="void-pop-letter">Batalkan surat</button></template></UiModal>
  </div>
</template>

<style scoped>
:global(.modal-body:has(.letter-viewport)){scrollbar-gutter:stable}
.letter-tabs{display:flex;align-items:center;gap:4px;border-bottom:1px solid #dce3ed;margin-bottom:22px}.letter-tabs button{display:flex;align-items:center;gap:8px;border:0;background:transparent;padding:14px 18px;color:#5e6e84;font-weight:600;border-bottom:2px solid transparent}.letter-tabs button.active{color:#2064b6;border-color:#2064b6}.letter-tabs span{font-size:11px;background:#eaf0f7;border-radius:10px;padding:2px 7px}.letter-tabs a{display:flex;align-items:center;margin-left:auto;color:#2064b6;font-size:12px}.letter-tabs button:focus-visible{outline:2px solid #2064b6;outline-offset:3px}
.letter-compose{overflow:hidden}.letter-compose-head{padding:26px 30px;border-bottom:1px solid #e5eaf1}.letter-compose-head h2{font-size:22px;letter-spacing:-.5px;margin:0 0 8px}.letter-compose-head p{font-size:13px;line-height:1.7;color:#65748a;margin:0}.letter-compose-grid{display:grid;grid-template-columns:minmax(0,1fr) 310px}.letter-fields{padding:28px 30px}.letter-fields .field{margin-bottom:22px}.letter-fields label{font-size:13px}.letter-fields select{max-width:100%;text-overflow:ellipsis}.letter-help{display:block;color:#65748a;font-size:12px;line-height:1.7;margin:8px 0 0}.letter-help a{color:#2064b6;text-decoration:underline}.manual-kbli{margin-top:14px}.address-label{display:flex;align-items:center;justify-content:space-between;gap:12px}.letter-compose-submit{display:flex;align-items:center;gap:16px;flex-wrap:wrap;margin-top:28px;padding-top:22px;border-top:1px solid #e5eaf1}.letter-compose-submit span{font-size:12px;color:#65748a}.letter-automatic{padding:26px 24px;background:#f8fafc;border-left:1px solid #e5eaf1}.letter-automatic img{width:100%;height:auto;background:white;border:1px solid #e5eaf1;padding:8px}.letter-automatic h3{font-size:15px;margin:24px 0 20px}.letter-automatic dl{margin:0;display:grid;gap:19px}.letter-automatic dt{font-size:11px;color:#65748a;margin-bottom:6px}.letter-automatic dd{margin:0;font-size:12px;color:#283e58;line-height:1.6;white-space:pre-line;overflow-wrap:anywhere}.letter-automatic dd span{display:block;color:#65748a}.signature-info{display:flex;gap:8px;align-items:flex-start;color:#22695c;font-size:12px;line-height:1.6;border-top:1px solid #dce8e5;padding-top:18px;margin:24px 0 0}.signature-info svg{flex-shrink:0;margin-top:2px}
.archive-toolbar{display:flex;justify-content:space-between;gap:16px;padding:24px;align-items:center}.archive-toolbar h2{font-size:18px;margin:0}.letter-archive td{max-width:350px;overflow-wrap:anywhere}.letter-archive .op-actions{flex-wrap:wrap}.letter-preview-status{display:flex;align-items:center;flex-wrap:wrap;justify-content:space-between;gap:14px;font-size:12px;color:#5e6e84;padding:0 0 14px}.letter-zoom{display:flex;align-items:center;gap:5px;flex-wrap:wrap}.letter-zoom button{font:inherit;padding:6px 9px;background:white;border:1px solid #dce3ed;border-radius:5px;color:#536780}.letter-zoom button.active{background:#eaf2fb;color:#2064b6;border-color:#b2cbe6}.letter-viewport{max-height:calc(100dvh - 290px);overflow:auto;background:#e9eef4;padding:12px;border-radius:6px}.letter-sheet-size{position:relative;margin:auto;box-shadow:0 3px 16px #23364b18}.letter-sheet-size iframe{border:0;width:794px;position:absolute;left:0;top:0;transform-origin:top left;background:#fff}
@media(max-width:1050px){.letter-compose-grid{grid-template-columns:minmax(0,1fr) 260px}.letter-fields,.letter-compose-head{padding:22px}.letter-automatic{padding:22px 18px}}@media(max-width:760px){.letter-compose-grid{grid-template-columns:1fr}.letter-automatic{border-left:0;border-top:1px solid #e5eaf1}.letter-automatic img{max-width:350px}.letter-automatic dl{grid-template-columns:1fr 1fr;gap:18px}.letter-fields,.letter-compose-head{padding:20px}.letter-compose-head h2{font-size:20px}.letter-compose-submit{align-items:stretch;flex-direction:column;gap:10px}.letter-compose-submit button{justify-content:center}.letter-tabs button{padding:12px 10px;font-size:12px}.letter-tabs a{font-size:11px}.archive-toolbar{align-items:stretch;flex-direction:column;padding:18px}.archive-toolbar .search-box{width:100%}}@media(max-width:400px){.letter-automatic dl{grid-template-columns:1fr}.letter-preview-status{font-size:11px}}
</style>
