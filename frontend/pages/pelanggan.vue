<script setup lang="ts">
const list = ref<any[]>([])
const mitras = ref<any[]>([])
const pakets = ref<any[]>([])
const loading = ref(true)
const loadError = ref('')
const q = ref('')
const fMitra = ref('')
const fStatus = ref('')

const formOpen = ref(false)
const editing = ref<any>(null)
const saving = ref(false)
const blank = () => ({
  nama: '', alamat: '', telepon: '', mitra_id: '', paket_id: '', pppoe: '',
  status: 'Aktif', tanggal_pasang: localISODate(), catatan: ''
})
const form = ref<any>(blank())

const statusOpsi = ['Aktif', 'Isolir', 'Nonaktif']

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    ;[list.value, mitras.value, pakets.value] = await Promise.all([
      api('/api/pelanggan'),
      api('/api/mitra'),
      api('/api/paket')
    ])
  } catch (e) {
    loadError.value = errMsg(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

const filtered = computed(() =>
  list.value.filter((p) => {
    const s = q.value.toLowerCase()
    const okQ = !s || [p.nama, p.alamat, p.pppoe, p.telepon].join(' ').toLowerCase().includes(s)
    const okM = !fMitra.value || String(p.mitra_id) === String(fMitra.value)
    const okS = !fStatus.value || p.status === fStatus.value
    return okQ && okM && okS
  })
)

const pages = usePagination(filtered, { resetOn: [q, fMitra, fStatus] })

function openCreate() {
  editing.value = null
  form.value = blank()
  if (mitras.value.length) form.value.mitra_id = mitras.value[0].id
  if (pakets.value.length) form.value.paket_id = pakets.value[0].id
  formOpen.value = true
}
function openEdit(p: any) {
  editing.value = p
  form.value = { ...p }
  formOpen.value = true
}
async function save() {
  if (saving.value) return
  if (!form.value.nama.trim() || !form.value.mitra_id || !form.value.paket_id) {
    toast('Nama, mitra, dan paket wajib diisi', 'error')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await api(`/api/pelanggan/${editing.value.id}`, { method: 'PUT', body: form.value })
      toast('Data pelanggan diperbarui')
    } else {
      await api('/api/pelanggan', { method: 'POST', body: form.value })
      toast('Pelanggan baru ditambahkan')
    }
    formOpen.value = false
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    saving.value = false
  }
}
async function hapus(p: any) {
  const ok = await useConfirm().ask(
    `Hapus pelanggan "${p.nama}"? Tagihan dan tiket milik pelanggan ini juga akan terhapus.`
  )
  if (!ok) return
  try {
    await api(`/api/pelanggan/${p.id}`, { method: 'DELETE' })
    toast('Pelanggan dihapus')
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
</script>

<template>
  <div class="toolbar">
    <div class="search-box">
      <Ico name="search" :size="15" />
      <input aria-label="Cari data" v-model="q" type="text" placeholder="Cari nama, alamat, user PPPoE…" />
    </div>
    <select aria-label="Filter mitra" v-model="fMitra" class="ctrl" style="width:auto">
      <option value="">Semua Mitra</option>
      <option v-for="m in mitras" :key="m.id" :value="String(m.id)">{{ m.nama }}</option>
    </select>
    <select aria-label="Filter status" v-model="fStatus" class="ctrl" style="width:auto">
      <option value="">Semua Status</option>
      <option v-for="s in statusOpsi" :key="s" :value="s">{{ s }}</option>
    </select>
    <div class="spacer"></div>
    <button class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tambah Pelanggan</button>
  </div>

  <div class="card">
    <div v-if="loading" class="empty"><p>Memuat data pelanggan…</p></div>
    <div v-else-if="loadError" class="empty" role="alert"><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
    <div v-else-if="!filtered.length" class="empty">
      <div class="e-ico"><Ico name="wifi" /></div>
      <p>{{ list.length ? 'Tidak ada pelanggan yang cocok dengan filter.' : 'Belum ada pelanggan.' }}</p>
      <button v-if="!list.length" class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tambah Pelanggan</button>
    </div>
    <div v-else class="table-wrap">
      <table class="tbl">
        <thead>
          <tr>
            <th>Pelanggan</th>
            <th>Paket</th>
            <th>Mitra</th>
            <th>Alamat</th>
            <th class="num">Tagihan/Bln</th>
            <th>Status</th>
            <th style="text-align:right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in pages.rows" :key="p.id">
            <td>
              <div class="cell-main">{{ p.nama }}</div><div v-if="p.telepon" class="cell-sub">WA {{ p.telepon }}</div>
              <div class="cell-sub mono">{{ p.pppoe || '-' }}</div>
            </td>
            <td>
              <div style="font-size:13px">{{ p.paket_nama }}</div>
              <div class="cell-sub">{{ p.kecepatan }} Mbps</div>
            </td>
            <td style="font-size:13px">{{ p.mitra_nama }}</td>
            <td class="cell-sub" style="max-width:220px">{{ p.alamat }}</td>
            <td class="num">{{ fmtRp(p.harga) }}</td>
            <td>
              <span class="badge" :class="p.status === 'Aktif' ? 'b-success' : p.status === 'Isolir' ? 'b-warn' : 'b-muted'">{{ p.status }}</span>
            </td>
            <td>
              <div class="row-actions">
                <button class="icon-btn" aria-label="Edit" title="Edit" @click="openEdit(p)"><Ico name="pencil" :size="15" /></button>
                <button class="icon-btn danger" aria-label="Hapus" title="Hapus" @click="hapus(p)"><Ico name="trash" :size="15" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <TablePagination v-if="!loading && !loadError" v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" label="pelanggan" />
  </div>

  <UiModal :open="formOpen" :busy="saving" :title="editing ? 'Edit Pelanggan' : 'Tambah Pelanggan'" @close="formOpen = false">
    <form id="form-pelanggan" @submit.prevent="save">
      <div class="grid-form">
        <div class="field">
          <label for="pelanggan-form-nama">Nama Pelanggan *</label>
          <input id="pelanggan-form-nama" v-model="form.nama" class="ctrl" placeholder="Andi Pratama" required />
        </div>
        <div class="field">
          <label for="pelanggan-form-pppoe">User PPPoE / IP</label>
          <input id="pelanggan-form-pppoe" v-model="form.pppoe" class="ctrl" placeholder="andi.home" />
        </div>
        <div class="field">
          <label for="pelanggan-form-mitra-id">Mitra *</label>
          <select id="pelanggan-form-mitra-id" v-model="form.mitra_id" class="ctrl" required>
            <option v-for="m in mitras" :key="m.id" :value="m.id">{{ m.nama }}</option>
          </select>
        </div>
        <div class="field">
          <label for="pelanggan-form-paket-id">Paket *</label>
          <select id="pelanggan-form-paket-id" v-model="form.paket_id" class="ctrl" required>
            <option v-for="k in pakets" :key="k.id" :value="k.id">{{ k.nama }} — {{ fmtRp(k.harga) }}</option>
          </select>
        </div>
        <div class="field full"><label for="pelanggan-whatsapp">Nomor WhatsApp pelanggan</label><input id="pelanggan-whatsapp" v-model="form.telepon" class="ctrl" type="tel" maxlength="24" placeholder="08xxxxxxxxxx atau 628xxxxxxxxxx"><span class="hint">Penerima invoice dan pengingat pembayaran. Kosongkan jika belum tersedia.</span></div>
        <div class="field full">
          <label for="pelanggan-form-alamat">Alamat</label>
          <input id="pelanggan-form-alamat" v-model="form.alamat" class="ctrl" placeholder="Jl. Kenanga No. 3" />
        </div>
        <div class="field">
          <label for="pelanggan-form-status">Status</label>
          <select id="pelanggan-form-status" v-model="form.status" class="ctrl">
            <option v-for="s in statusOpsi" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
        <div class="field">
          <label for="pelanggan-form-tanggal-pasang">Tanggal Pasang</label>
          <input id="pelanggan-form-tanggal-pasang" v-model="form.tanggal_pasang" class="ctrl" type="date" />
        </div>
        <div class="field full">
          <label for="pelanggan-form-catatan">Catatan</label>
          <textarea id="pelanggan-form-catatan" v-model="form.catatan" class="ctrl" rows="2" placeholder="Catatan teknis / info pelanggan…"></textarea>
        </div>
      </div>
    </form>
    <template #footer>
      <button class="btn btn-sec" :disabled="saving" @click="formOpen = false">Batal</button>
      <button class="btn btn-primary" :disabled="saving" type="submit" form="form-pelanggan">{{ saving ? 'Menyimpan…' : 'Simpan' }}</button>
    </template>
  </UiModal>
</template>





