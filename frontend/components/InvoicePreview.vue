<script setup lang="ts">
import { renderInvoice } from '~/lib/invoice'
import invoiceCSS from '~/assets/css/invoice.css?raw'
const props = defineProps<{ bill: any | null }>()
const emit = defineEmits<{ close: [] }>()
const waOpen = ref(false)
const html = ref('')
const loading = ref(false)
const error = ref('')
const sessionExpired = ref(false)
const ready = ref(false)
const zoom = ref<'page' | 'width' | 'actual'>('page')
const frame = ref<HTMLIFrameElement | null>(null)
const canvas = ref<HTMLElement | null>(null)
const canvasWidth = ref(794)
const canvasHeight = ref(600)
const paperHeight = ref(1123)
const scale = computed(() => zoom.value === 'actual' ? 1 : Math.max(0.1, Math.min(1, (canvasWidth.value - 24) / 794, zoom.value === 'page' ? (canvasHeight.value - 24) / paperHeight.value : 1)))
let requestId = 0
let resizeObserver: ResizeObserver | undefined

watch(canvas, el => {
  resizeObserver?.disconnect()
  if (!el) return
  canvasWidth.value = el.clientWidth
  canvasHeight.value = el.clientHeight
  resizeObserver = new ResizeObserver(() => { canvasWidth.value = el.clientWidth; canvasHeight.value = el.clientHeight })
  resizeObserver.observe(el)
}, { flush: 'post' })

watch(() => props.bill?.id, id => {
  ++requestId
  html.value = ''; ready.value = false; error.value = ''; loading.value = false
  zoom.value = import.meta.client && window.matchMedia('(min-width: 801px)').matches ? 'width' : 'page'; paperHeight.value = 1123
  if (id) load()
}, { immediate: true })

async function load() {
  const id = props.bill?.id
  if (!id) return
  const request = ++requestId
  loading.value = true; ready.value = false; html.value = ''; error.value = ''; sessionExpired.value = false
  try {
    const [finance, customers, settings] = await Promise.all([
      api(`/api/keuangan?periode=${encodeURIComponent(props.bill.periode)}`), api('/api/pelanggan'), api('/api/settings')
    ])
    if (request !== requestId) return
    const current = finance.tagihan.find((item: any) => item.id === id)
    if (!current) { error.value = 'Tagihan ini tidak ditemukan. Tutup preview dan muat ulang daftar tagihan.'; return }
    html.value = renderInvoice(current, customers.find((item: any) => item.id === current.pelanggan_id) || {}, settings, invoiceCSS)
  } catch (e: any) {
    if (request !== requestId) return
    const status = e?.status || e?.response?.status
    sessionExpired.value = status === 401
    error.value = status === 401 ? 'Sesi berakhir. Masuk kembali untuk melihat invoice.' : status === 404 ? 'Tagihan ini tidak ditemukan. Tutup preview dan muat ulang daftar tagihan.' : 'Preview invoice belum berhasil dimuat. Periksa koneksi lalu coba lagi.'
  } finally { if (request === requestId) loading.value = false }
}
function onReady() {
  if (!html.value) return
  paperHeight.value = Math.max(1123, frame.value?.contentDocument?.documentElement.scrollHeight || 0)
  ready.value = true
}
async function setZoom(value: 'page' | 'width' | 'actual') {
  zoom.value = value
  await nextTick()
  canvas.value?.scrollTo({ top: 0, left: 0 })
}
function printInvoice() {
  if (!ready.value || !frame.value?.contentWindow) return
  try { frame.value.contentWindow.focus(); frame.value.contentWindow.print() }
  catch { toast('Dialog cetak belum dapat dibuka. Coba muat ulang preview.', 'error') }
}
function close() { waOpen.value = false; ++requestId; emit('close') }
function signIn() { close(); clearToken(); navigateTo('/login') }
onBeforeUnmount(() => { ++requestId; resizeObserver?.disconnect() })
</script>

<template>
  <UiModal :open="!!bill" title="Preview invoice" hg @close="close">
    <div v-if="bill" class="invoice-preview-layout">
      <section class="invoice-viewer" aria-label="Pratinjau invoice A4">
        <div class="invoice-viewer-toolbar"><span><Ico name="file" :size="16" />A4 <small>210 × 297 mm</small></span><div class="invoice-zoom" role="group" aria-label="Ukuran pratinjau"><button :class="{active:zoom === 'page'}" :aria-pressed="zoom === 'page'" @click="setZoom('page')">Pas layar</button><button :class="{active:zoom === 'width'}" :aria-pressed="zoom === 'width'" @click="setZoom('width')">Pas lebar</button><button :class="{active:zoom === 'actual'}" :aria-pressed="zoom === 'actual'" @click="setZoom('actual')">100%</button></div></div>
        <div v-if="loading" class="invoice-viewer-state" role="status"><Ico name="file" :size="29" /><p>Menyiapkan invoice…</p></div>
        <div v-else-if="error" class="invoice-viewer-state" role="alert"><Ico name="alert" :size="27" /><p>{{ error }}</p><button v-if="sessionExpired" class="btn btn-sec" @click="signIn">Masuk kembali</button><button v-else class="btn btn-sec" @click="load">Coba lagi</button></div>
        <div v-else-if="html" ref="canvas" class="invoice-canvas" tabindex="0" aria-label="Lembar invoice, geser untuk melihat seluruh halaman">
          <div class="invoice-paper-scale" :style="{width:`${794*scale}px`,height:`${paperHeight*scale}px`}"><iframe ref="frame" :srcdoc="html" title="Invoice pelanggan siap cetak" sandbox="allow-same-origin allow-modals" :style="{height:`${paperHeight}px`,transform:`scale(${scale})`}" @load="onReady"></iframe></div>
        </div>
      </section>
    </div>
    <template #footer><span class="invoice-print-tip">Pilih Save as PDF pada dialog cetak untuk mengunduh.</span><button class="btn btn-sec" @click="close">Tutup preview</button><button class="btn btn-sec" :disabled="!ready || loading || !!error" @click="waOpen = true"><Ico name="message" :size="16" /> Kirim WhatsApp</button><button class="btn btn-primary" :disabled="!ready || loading || !!error" @click="printInvoice"><Ico name="printer" :size="16" />Cetak / simpan PDF</button></template>
  </UiModal>
  <WhatsAppComposer :open="waOpen" kind="invoice" :source-id="bill?.id || 0" @close="waOpen = false" />
</template>

<style scoped>
.invoice-preview-layout{min-width:0;height:100%;min-height:0}
.invoice-viewer{min-width:0;height:100%;min-height:0;display:flex;flex-direction:column;background:#e8edf5}
.invoice-viewer-toolbar{display:flex;flex:none;align-items:center;justify-content:space-between;gap:12px;padding:10px 20px;background:#f8faff;border-bottom:1px solid var(--border);font-size:11px}
.invoice-viewer-toolbar>span{display:flex;align-items:center;gap:8px;color:#586c84}.invoice-viewer-toolbar small{font-size:10px;color:#7b8aa0}
.invoice-zoom{display:flex;gap:3px;padding:3px;border:1px solid #dce4ef;border-radius:7px;background:#fff}.invoice-zoom button{border:0;border-radius:4px;background:none;padding:6px 10px;font-size:11px;color:#62728a;min-height:28px}.invoice-zoom button.active{background:#e7f0ff;color:#165dcc;font-weight:600}
.invoice-canvas{flex:1;min-height:0;overflow:auto;overflow-anchor:none;overscroll-behavior:contain;padding:12px;background:#e8edf5}
.invoice-paper-scale{position:relative;margin:0 auto;box-shadow:0 3px 18px #25436912}.invoice-paper-scale iframe{position:absolute;top:0;left:0;width:794px;border:0;transform-origin:top left;background:#fff;display:block}
.invoice-viewer-state{flex:1;min-height:0;overflow:auto;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:30px;text-align:center;color:var(--muted);gap:17px;font-size:13px}.invoice-viewer-state p{max-width:35ch}
.invoice-print-tip{margin-right:auto;align-self:center;max-width:40ch;font-size:11px;color:var(--muted)}
@media(max-width:800px){.invoice-viewer-toolbar{padding:9px 14px}.invoice-zoom button{min-height:32px;padding:6px 8px}.invoice-viewer-toolbar small,.invoice-print-tip{display:none}}
</style>

<style>
.modal.hg:has(.invoice-preview-layout){max-width:1040px;height:min(960px,calc(100dvh - 48px));border-radius:14px}
.modal:has(.invoice-preview-layout)>.modal-body{flex:1;padding:0;overflow:hidden;min-height:0}
.modal:has(.invoice-preview-layout)>.modal-head{padding:16px 20px}
.modal:has(.invoice-preview-layout)>.modal-foot{padding:15px 20px;background:#fff}
@media(max-width:800px){.modal.hg:has(.invoice-preview-layout){height:calc(100dvh - 24px)}.modal:has(.invoice-preview-layout)>.modal-head,.modal:has(.invoice-preview-layout)>.modal-foot{padding:14px}}
</style>
