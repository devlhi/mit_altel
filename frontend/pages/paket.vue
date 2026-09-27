<script setup lang="ts">
const list = ref<any[]>([])
const loading = ref(true)
const loadError = ref('')
const pages = usePagination(list)

const formOpen = ref(false)
const editing = ref<any>(null)
const saving = ref(false)
const blank = () => ({ nama: '', kecepatan: 10, harga: 150000, keterangan: '' })
const form = ref<any>(blank())

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    list.value = await api('/api/paket')
  } catch (e) {
    loadError.value = errMsg(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

function openCreate() {
  editing.value = null
  form.value = blank()
  formOpen.value = true
}
function openEdit(k: any) {
  editing.value = k
  form.value = { ...k }
  formOpen.value = true
}
async function save() {
  if (saving.value) return
  if (!form.value.nama.trim()) {
    toast('Nama paket wajib diisi', 'error')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await api(`/api/paket/${editing.value.id}`, { method: 'PUT', body: form.value })
      toast('Paket diperbarui')
    } else {
      await api('/api/paket', { method: 'POST', body: form.value })
      toast('Paket baru ditambahkan')
    }
    formOpen.value = false
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    saving.value = false
  }
}
async function hapus(k: any) {
  const ok = await useConfirm().ask(`Hapus paket "${k.nama}"?`)
  if (!ok) return
  try {
    await api(`/api/paket/${k.id}`, { method: 'DELETE' })
    toast('Paket dihapus')
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
</script>

<template>
  <div class="toolbar">
    <div class="spacer"></div>
    <button class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tambah Paket</button>
  </div>

  <div class="card">
    <div v-if="loading" class="empty"><p>Memuat paket…</p></div>
    <div v-else-if="loadError" class="empty" role="alert"><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
    <div v-else-if="!list.length" class="empty">
      <div class="e-ico"><Ico name="box" /></div>
      <p>Belum ada paket layanan. Buat paket pertama Anda.</p>
      <button class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tambah Paket</button>
    </div>
    <div v-else class="table-wrap">
      <table class="tbl">
        <thead>
          <tr>
            <th>Paket</th>
            <th class="num">Kecepatan</th>
            <th class="num">Harga / Bulan</th>
            <th class="num">Dipakai</th>
            <th>Keterangan</th>
            <th style="text-align:right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="k in pages.rows" :key="k.id">
            <td class="cell-main">{{ k.nama }}</td>
            <td class="num">{{ k.kecepatan }} Mbps</td>
            <td class="num cell-main">{{ fmtRp(k.harga) }}</td>
            <td class="num">{{ k.dipakai }} pelanggan</td>
            <td class="cell-sub">{{ k.keterangan || '-' }}</td>
            <td>
              <div class="row-actions">
                <button class="icon-btn" aria-label="Edit" title="Edit" @click="openEdit(k)"><Ico name="pencil" :size="15" /></button>
                <button class="icon-btn danger" aria-label="Hapus" title="Hapus" @click="hapus(k)"><Ico name="trash" :size="15" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <TablePagination v-if="!loading && !loadError" v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" label="paket" />
  </div>

  <UiModal :open="formOpen" :busy="saving" :title="editing ? 'Edit Paket' : 'Tambah Paket'" @close="formOpen = false">
    <form id="form-paket" @submit.prevent="save">
      <div class="grid-form">
        <div class="field">
          <label for="paket-form-nama">Nama Paket *</label>
          <input id="paket-form-nama" v-model="form.nama" class="ctrl" placeholder="Home 20" required />
        </div>
        <div class="field">
          <label for="paket-form-kecepatan">Kecepatan (Mbps)</label>
          <input id="paket-form-kecepatan" v-model.number="form.kecepatan" class="ctrl" type="number" min="1" />
        </div>
        <div class="field full">
          <label for="paket-form-harga">Harga per Bulan (Rp)</label>
          <input id="paket-form-harga" v-model.number="form.harga" class="ctrl" type="number" min="0" step="1" />
        </div>
        <div class="field full">
          <label for="paket-form-keterangan">Keterangan</label>
          <textarea id="paket-form-keterangan" v-model="form.keterangan" class="ctrl" rows="2" placeholder="Contoh: untuk keluarga, unlimited"></textarea>
        </div>
      </div>
    </form>
    <template #footer>
      <button class="btn btn-sec" :disabled="saving" @click="formOpen = false">Batal</button>
      <button class="btn btn-primary" :disabled="saving" type="submit" form="form-paket">{{ saving ? 'Menyimpan…' : 'Simpan' }}</button>
    </template>
  </UiModal>
</template>




