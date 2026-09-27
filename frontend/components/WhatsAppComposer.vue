<script setup lang="ts">
const props = withDefaults(defineProps<{ open: boolean; kind: string; sourceId?: number }>(), { sourceId: 0 })
const emit = defineEmits<{ close: []; sent: [] }>()
const data = ref<any>(null), loading = ref(false), error = ref(''), busy = ref(false), result = ref<any>(null)
const recipient = ref(''), message = ref(''), mode = ref('text'), file = ref<File | null>(null), requestKey = ref(''), attachment = ref(''), uncertain = ref(false), retryPayload = ref<any>(null)
let request = 0
const statusLabels: Record<string, string> = { accepted: 'Diterima MPWA', failed: 'Ditolak gateway', unknown: 'Hasil belum pasti', sending: 'Sedang diproses' }
watch(() => [props.open, props.kind, props.sourceId], async () => { if (!props.open) { ++request; return }; await load() }, { immediate: true })
async function load() {
  const id = ++request; loading.value = true; error.value = ''; result.value = null; uncertain.value = false; retryPayload.value = null; file.value = null; attachment.value = ''; data.value = null
  const bytes = crypto.getRandomValues(new Uint8Array(20)); requestKey.value = [...bytes].map(b => b.toString(16).padStart(2, '0')).join('')
  try { const d = await api(`/api/whatsapp/preview?kind=${props.kind}&id=${props.sourceId}`); if (id !== request) return; data.value = d; recipient.value = d.source.recipient || ''; message.value = d.source.message || ''; mode.value = d.can_link && props.kind !== 'custom' ? 'link' : 'text' }
  catch (e) { if (id === request) error.value = errMsg(e) } finally { if (id === request) loading.value = false }
}
function selectFile(event: Event) { file.value = (event.target as HTMLInputElement).files?.[0] || null; attachment.value = '' }
async function send() {
  if (busy.value || result.value) return; busy.value = true; error.value = ''
  try {
    if (!retryPayload.value) {
      if (mode.value === 'pdf') {
        if (!file.value || file.value.size > 8 * 1024 * 1024) throw new Error('Pilih file PDF maksimal 8 MB.')
        if (!attachment.value) { const form = new FormData(); form.append('file', file.value); const uploaded = await api('/api/whatsapp/attachment', { method: 'POST', body: form }); attachment.value = uploaded.attachment_id }
      }
      retryPayload.value = { request_key: requestKey.value, revision: data.value.source.revision, kind: props.kind, source_id: props.sourceId, recipient: recipient.value, message: message.value, include_link: mode.value === 'link', attachment_id: mode.value === 'pdf' ? attachment.value : '' }
    }
    uncertain.value = true
    result.value = await api('/api/whatsapp/send', { method: 'POST', body: retryPayload.value }); uncertain.value = false; emit('sent')
  } catch (e: any) {
    error.value = errMsg(e)
    if (e?.status >= 400 && e?.status < 500) { uncertain.value = false; retryPayload.value = null }
  } finally { busy.value = false }
}
const valid = computed(() => data.value?.enabled && recipient.value.trim() && message.value.trim() && (mode.value !== 'pdf' || file.value) && !busy.value)
</script>
<template>
  <UiModal :open="open" title="Kirim melalui WhatsApp" wide :busy="busy" @close="emit('close')">
    <div v-if="loading" class="empty">Menyiapkan pesan…</div>
    <p v-if="error" class="op-note op-warning" role="alert">{{ error }}</p>
    <div v-if="result" class="wa-result" role="status"><span class="op-status" :class="{ off: result.status !== 'accepted' }">{{ statusLabels[result.status] }}</span><h3>{{ result.reference }}</h3><p>Nomor penerima: {{ result.recipient }}</p><p>{{ result.status === 'accepted' ? 'MPWA menerima permintaan pengiriman. Status pesan sampai atau dibaca perlu diperiksa di gateway.' : result.error || 'Permintaan sedang diproses. Periksa riwayat pengiriman.' }}</p><NuxtLink to="/whatsapp" class="btn btn-sec" @click="emit('close')">Lihat riwayat</NuxtLink></div>
    <form v-else-if="data" id="wa-send-form" @submit.prevent="send"><p class="cell-main">{{ data.source.reference }}</p><p class="cell-sub" style="margin:5px 0 20px">{{ data.source.name }}<span v-if="data.sender"> · Dikirim dari {{ data.sender }}</span></p>
      <div v-if="!data.enabled" class="op-note op-warning">MPWA belum aktif. Isi URL gateway, API key, dan sender di <NuxtLink to="/pengaturan#mpwa" class="op-link" @click="emit('close')">Pengaturan MPWA</NuxtLink>.</div>
      <fieldset :disabled="busy || uncertain" class="wa-fieldset"><div class="wa-compose-grid"><div>
        <div class="field"><label for="wa-recipient">Nomor WhatsApp penerima</label><input id="wa-recipient" v-model="recipient" class="ctrl" type="tel" required maxlength="24" placeholder="08xxxxxxxxxx atau 628xxxxxxxxxx"><span class="hint">Nomor 08 otomatis diubah menjadi 628. Perubahan di sini berlaku untuk pengiriman ini.</span></div>
        <div class="field"><label for="wa-mode">Isi pengiriman</label><select id="wa-mode" v-model="mode" class="ctrl"><option value="text">Ringkasan teks</option><option v-if="kind !== 'custom'" value="link" :disabled="!data.can_link">Teks + tautan dokumen (7 hari)</option><option value="pdf" :disabled="!data.can_link">Teks + lampiran PDF</option></select><span v-if="!data.can_link" class="hint">Tambahkan URL publik aplikasi di Pengaturan untuk mengirim tautan atau PDF.</span></div>
        <div v-if="mode === 'pdf'" class="field"><label for="wa-pdf">PDF yang akan dikirim</label><input id="wa-pdf" class="ctrl" type="file" accept="application/pdf,.pdf" required @change="selectFile"><span class="hint">Simpan PDF dari preview invoice/dokumen, lalu pilih file tersebut. Maksimal 8 MB. Tautan pengambilan file untuk MPWA berlaku 24 jam.</span></div>
        <div class="field"><label for="wa-message">Pesan</label><textarea id="wa-message" v-model="message" class="ctrl" rows="10" maxlength="10000" required></textarea></div>
      </div><aside class="wa-preview"><div class="wa-preview-head"><Ico name="message" :size="18" /><b>Preview pesan</b></div><div class="wa-bubble"><p>{{ message || 'Isi pesan untuk melihat preview.' }}</p><div v-if="mode === 'link'" class="wa-attachment">Lihat {{ data.source.reference }}<br><span>Tautan dokumen pribadi ditambahkan saat dikirim. Berlaku 7 hari.</span></div><div v-if="mode === 'pdf'" class="wa-attachment"><Ico name="file" :size="16" /> {{ file?.name || 'Pilih lampiran PDF' }}</div></div><p class="hint">Periksa nomor penerima, nominal, rekening, dan lampiran sebelum mengirim.</p></aside></div></fieldset>
      <div v-if="uncertain" class="op-note op-warning">Koneksi terputus saat meminta pengiriman. “Periksa hasil” memakai identitas permintaan yang sama agar pesan tidak dikirim dua kali.</div>
    </form>
    <template #footer><button class="btn btn-sec" :disabled="busy" @click="emit('close')">{{ result ? 'Tutup' : 'Batal' }}</button><button v-if="data && !result" class="btn btn-primary" type="submit" form="wa-send-form" :disabled="!valid"><Ico name="send" :size="16" />{{ busy ? 'Memproses…' : uncertain ? 'Periksa hasil' : 'Kirim sekarang' }}</button></template>
  </UiModal>
</template>
<style scoped>
.wa-fieldset{border:0;padding:0;min-width:0}.wa-compose-grid{display:grid;grid-template-columns:1fr .9fr;gap:28px}.wa-preview{background:#edf3f8;border:1px solid #dee7ef;border-radius:12px;padding:18px;align-self:start}.wa-preview-head{display:flex;gap:9px;align-items:center;font-size:12px;color:#395970;margin-bottom:16px}.wa-bubble{background:white;padding:17px;border-radius:0 12px 12px 12px;box-shadow:0 2px 4px #223d5610;overflow-wrap:anywhere}.wa-bubble p{white-space:pre-wrap;font:12px/1.8 var(--font,Arial);margin:0;color:#244256}.wa-preview>.hint{margin-top:15px}.wa-attachment{border-top:1px solid #e5edf1;padding-top:15px;margin-top:15px;font-size:12px;color:#1762a7;line-height:1.7;overflow-wrap:anywhere}.wa-attachment span{font-size:11px;color:#687e93}.wa-result{padding:22px 0}.wa-result h3{margin-top:18px}.wa-result p{font-size:13px;line-height:1.8;color:#586b83;margin:12px 0}.wa-result .btn{margin-top:12px}@media(max-width:700px){.wa-compose-grid{grid-template-columns:1fr;gap:16px}}
</style>
