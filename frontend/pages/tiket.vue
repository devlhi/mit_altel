<script setup lang="ts">
const list = ref<any[]>([])
const pelanggans = ref<any[]>([])
const loading = ref(true)
const loadError = ref('')
const q = ref('')
const fStatus = ref('')

const formOpen = ref(false)
const editing = ref<any>(null)
const saving = ref(false)
const blank = () => ({
  pelanggan_id: '', jenis: 'No Internet', prioritas: 'Sedang',
  status: 'Baru', teknisi: '', deskripsi: ''
})
const form = ref<any>(blank())

const jenisOpsi = ['No Internet', 'Mati Total', 'Lambat', 'Putus-putus', 'Perangkat', 'Pindah Alamat', 'Administrasi', 'Lainnya']
const prioOpsi = ['Rendah', 'Sedang', 'Tinggi', 'Urgent']
const statusOpsi = ['Baru', 'Diproses', 'Selesai']

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    ;[list.value, pelanggans.value] = await Promise.all([api('/api/tiket'), api('/api/pelanggan')])
  } catch (e) {
    loadError.value = errMsg(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

const filtered = computed(() =>
  list.value.filter((t) => {
    const s = q.value.toLowerCase()
    const okQ = !s || [t.nomor, t.pelanggan_nama].join(' ').toLowerCase().includes(s)
    const okS = !fStatus.value || t.status === fStatus.value
    return okQ && okS
  })
)

const pages = usePagination(filtered, { resetOn: [q, fStatus] })

function openCreate() {
  editing.value = null
  form.value = blank()
  if (pelanggans.value.length) form.value.pelanggan_id = pelanggans.value[0].id
  formOpen.value = true
}
function openEdit(t: any) {
  editing.value = t
  form.value = { ...t }
  formOpen.value = true
}
async function save() {
  if (saving.value) return
  if (!form.value.pelanggan_id) {
    toast('Pilih pelanggan yang melapor', 'error')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await api(`/api/tiket/${editing.value.id}`, { method: 'PUT', body: form.value })
      toast('Tiket diperbarui')
    } else {
      await api('/api/tiket', { method: 'POST', body: form.value })
      toast('Tiket dibuat')
    }
    formOpen.value = false
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    saving.value = false
  }
}
async function hapus(t: any) {
  const ok = await useConfirm().ask(`Hapus tiket ${t.nomor}?`)
  if (!ok) return
  try {
    await api(`/api/tiket/${t.id}`, { method: 'DELETE' })
    toast('Tiket dihapus')
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
async function majuStatus(t: any) {
  const urut = ['Baru', 'Diproses', 'Selesai']
  const idx = urut.indexOf(t.status)
  const next = urut[Math.min(idx + 1, urut.length - 1)]
  try {
    await api(`/api/tiket/${t.id}`, {
      method: 'PUT',
      body: { ...t, status: next }
    })
    toast(`Tiket ${t.nomor} → ${next}`)
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
function prioBadge(p: string) {
  return p === 'Urgent' || p === 'Tinggi' ? 'b-error' : p === 'Sedang' ? 'b-warn' : 'b-muted'
}
function statBadge(s: string) {
  return s === 'Baru' ? 'b-info' : s === 'Diproses' ? 'b-warn' : 'b-success'
}
</script>

<template>
  <div class="toolbar">
    <div class="search-box">
      <Ico name="search" :size="15" />
      <input aria-label="Cari data" v-model="q" type="text" placeholder="Cari nomor tiket / pelanggan…" />
    </div>
    <select aria-label="Filter status" v-model="fStatus" class="ctrl" style="width:auto">
      <option value="">Semua Status</option>
      <option v-for="s in statusOpsi" :key="s" :value="s">{{ s }}</option>
    </select>
    <div class="spacer"></div>
    <button class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tiket Baru</button>
  </div>

  <div class="card">
    <div v-if="loading" class="empty"><p>Memuat tiket…</p></div>
    <div v-else-if="loadError" class="empty" role="alert"><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
    <div v-else-if="!filtered.length" class="empty">
      <div class="e-ico"><Ico name="wrench" /></div>
      <p>{{ list.length ? 'Tidak ada tiket yang cocok dengan filter.' : 'Belum ada tiket gangguan.' }}</p>
      <button v-if="!list.length" class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tiket Baru</button>
    </div>
    <div v-else class="table-wrap">
      <table class="tbl">
        <thead>
          <tr>
            <th>Nomor</th>
            <th>Pelanggan</th>
            <th>Gangguan</th>
            <th>Prioritas</th>
            <th>Status</th>
            <th>Teknisi</th>
            <th>Dibuat</th>
            <th style="text-align:right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in pages.rows" :key="t.id">
            <td class="mono" style="font-size:12.5px">{{ t.nomor }}</td>
            <td>
              <div class="cell-main" style="font-size:13.5px">{{ t.pelanggan_nama }}</div>
              <div class="cell-sub">{{ t.mitra_nama }}</div>
            </td>
            <td style="font-size:13px">{{ t.jenis }}</td>
            <td><span class="badge" :class="prioBadge(t.prioritas)">{{ t.prioritas }}</span></td>
            <td><span class="badge" :class="statBadge(t.status)">{{ t.status }}</span></td>
            <td style="font-size:13px">{{ t.teknisi || '-' }}</td>
            <td class="cell-sub">{{ fmtDate(t.dibuat) }}</td>
            <td>
              <div class="row-actions">
                <button v-if="t.status !== 'Selesai'" class="icon-btn accent" :aria-label="'Majukan ke ' + (t.status === 'Baru' ? 'Diproses' : 'Selesai')" title="'Majukan ke ' + (t.status === 'Baru' ? 'Diproses' : 'Selesai')" @click="majuStatus(t)">
                  <Ico name="check" :size="15" />
                </button>
                <button class="icon-btn" aria-label="Edit" title="Edit" @click="openEdit(t)"><Ico name="pencil" :size="15" /></button>
                <button class="icon-btn danger" aria-label="Hapus" title="Hapus" @click="hapus(t)"><Ico name="trash" :size="15" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <TablePagination v-if="!loading && !loadError" v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" label="tiket" />
  </div>

  <UiModal :open="formOpen" :busy="saving" :title="editing ? 'Edit Tiket ' + editing.nomor : 'Tiket Gangguan Baru'" @close="formOpen = false">
    <form id="form-tiket" @submit.prevent="save">
      <div class="field">
        <label for="tiket-form-pelanggan-id">Pelanggan *</label>
        <select id="tiket-form-pelanggan-id" v-model="form.pelanggan_id" class="ctrl" required>
          <option v-for="p in pelanggans" :key="p.id" :value="p.id">
            {{ p.nama }} — {{ p.mitra_nama }}
          </option>
        </select>
      </div>
      <div class="grid-form">
        <div class="field">
          <label for="tiket-form-jenis">Jenis Gangguan</label>
          <select id="tiket-form-jenis" v-model="form.jenis" class="ctrl">
            <option v-for="j in jenisOpsi" :key="j" :value="j">{{ j }}</option>
          </select>
        </div>
        <div class="field">
          <label for="tiket-form-prioritas">Prioritas</label>
          <select id="tiket-form-prioritas" v-model="form.prioritas" class="ctrl">
            <option v-for="p in prioOpsi" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div class="field">
          <label for="tiket-form-status">Status</label>
          <select id="tiket-form-status" v-model="form.status" class="ctrl">
            <option v-for="s in statusOpsi" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
        <div class="field">
          <label for="tiket-form-teknisi">Teknisi</label>
          <input id="tiket-form-teknisi" v-model="form.teknisi" class="ctrl" placeholder="Nama teknisi" />
        </div>
        <div class="field full">
          <label for="tiket-form-deskripsi">Deskripsi</label>
          <textarea id="tiket-form-deskripsi" v-model="form.deskripsi" class="ctrl" rows="3" placeholder="Kronologi / detail laporan…"></textarea>
        </div>
      </div>
    </form>
    <template #footer>
      <button class="btn btn-sec" :disabled="saving" @click="formOpen = false">Batal</button>
      <button class="btn btn-primary" :disabled="saving" type="submit" form="form-tiket">{{ saving ? 'Menyimpan…' : 'Simpan' }}</button>
    </template>
  </UiModal>
</template>




