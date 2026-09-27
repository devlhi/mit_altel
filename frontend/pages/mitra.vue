<script setup lang="ts">
const list = ref<any[]>([])
const loading = ref(true)
const loadError = ref('')
const q = ref('')
const fStatus = ref('')

const formOpen = ref(false)
const editing = ref<any>(null)
const saving = ref(false)
const blank = () => ({
  nama: '', kode: '', jenis: 'Reseller', pemilik: '', telepon: '', email: '',
  alamat: '', wilayah: '', npwp: '', status: 'Aktif', share_mitra: 30,
  tanggal_gabung: localISODate(), catatan: '', rekening: ''
})
const form = ref<any>(blank())

const detail = ref<any>(null)
const customerPages = usePagination<any>(() => detail.value?.pelanggan || [], { resetOn: () => detail.value?.mitra.id })
const documentPages = usePagination<any>(() => detail.value?.dokumen || [], { resetOn: () => detail.value?.mitra.id })

const jenisOpsi = ['Reseller', 'Agen', 'Warung Internet', 'Perumahan (RT/RW Net)', 'Lainnya']
const statusOpsi = ['Aktif', 'Pending', 'Nonaktif']

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    list.value = await api('/api/mitra')
  } catch (e) {
    loadError.value = errMsg(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

const filtered = computed(() =>
  list.value.filter((m) => {
    const s = q.value.toLowerCase()
    const okQ = !s || [m.nama, m.kode, m.wilayah, m.pemilik].join(' ').toLowerCase().includes(s)
    const okS = !fStatus.value || m.status === fStatus.value
    return okQ && okS
  })
)

const pages = usePagination(filtered, { resetOn: [q, fStatus] })

function openCreate() {
  editing.value = null
  form.value = blank()
  formOpen.value = true
}
function openEdit(m: any) {
  editing.value = m
  form.value = { ...m }
  formOpen.value = true
}
async function save() {
  if (saving.value) return
  if (!form.value.nama.trim()) {
    toast('Nama mitra wajib diisi', 'error')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await api(`/api/mitra/${editing.value.id}`, { method: 'PUT', body: form.value })
      toast('Data mitra diperbarui')
    } else {
      await api('/api/mitra', { method: 'POST', body: form.value })
      toast('Mitra baru ditambahkan')
    }
    formOpen.value = false
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    saving.value = false
  }
}
async function hapus(m: any) {
  const ok = await useConfirm().ask(`Hapus mitra "${m.nama}"? Dokumen terkait juga akan terhapus.`)
  if (!ok) return
  try {
    await api(`/api/mitra/${m.id}`, { method: 'DELETE' })
    toast('Mitra dihapus')
    await load()
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
async function bukaDetail(m: any) {
  try {
    detail.value = await api(`/api/mitra/${m.id}`)
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
function terbitkan(jenis: string) {
  const st = useState<any>('docIntent', () => null)
  st.value = { jenis, mitra_id: detail.value.mitra.id }
  detail.value = null
  navigateTo('/dokumen')
}
function statusBadge(s: string) {
  return s === 'Aktif' ? 'b-success' : s === 'Pending' ? 'b-warn' : 'b-muted'
}
</script>

<template>
  <div class="toolbar">
    <div class="search-box">
      <Ico name="search" :size="15" />
      <input aria-label="Cari data" v-model="q" type="text" placeholder="Cari nama, kode, wilayah…" />
    </div>
    <select aria-label="Filter status" v-model="fStatus" class="ctrl" style="width:auto">
      <option value="">Semua Status</option>
      <option v-for="s in statusOpsi" :key="s" :value="s">{{ s }}</option>
    </select>
    <div class="spacer"></div>
    <button class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tambah Mitra</button>
  </div>

  <div class="card">
    <div v-if="loading" class="empty"><p>Memuat data mitra…</p></div>
    <div v-else-if="loadError" class="empty" role="alert"><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
    <div v-else-if="!filtered.length" class="empty">
      <div class="e-ico"><Ico name="users" /></div>
      <p>{{ list.length ? 'Tidak ada mitra yang cocok dengan filter.' : 'Belum ada mitra. Tambahkan mitra pertama Anda.' }}</p>
      <button v-if="!list.length" class="btn btn-primary" @click="openCreate"><Ico name="plus" :size="15" /> Tambah Mitra</button>
    </div>
    <div v-else class="table-wrap">
      <table class="tbl">
        <thead>
          <tr>
            <th>Mitra</th>
            <th>Pemilik / Kontak</th>
            <th>Wilayah</th>
            <th class="num">Share</th>
            <th class="num">Pelanggan</th>
            <th>Status</th>
            <th style="text-align:right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in pages.rows" :key="m.id">
            <td>
              <div class="cell-main">{{ m.nama }}</div>
              <div class="cell-sub">{{ m.kode }} · {{ m.jenis }}</div>
            </td>
            <td>
              <div style="font-size:13px">{{ m.pemilik || '-' }}</div>
              <div class="cell-sub">{{ m.telepon || m.email || '' }}</div>
            </td>
            <td>{{ m.wilayah || '-' }}</td>
            <td class="num">{{ m.share_mitra }}%</td>
            <td class="num">{{ m.pelanggan_aktif }}<span class="cell-sub">/{{ m.total_pelanggan }}</span></td>
            <td><span class="badge" :class="statusBadge(m.status)">{{ m.status }}</span></td>
            <td>
              <div class="row-actions">
                <button class="icon-btn" aria-label="Detail" title="Detail" @click="bukaDetail(m)"><Ico name="eye" :size="15" /></button>
                <button class="icon-btn" aria-label="Edit" title="Edit" @click="openEdit(m)"><Ico name="pencil" :size="15" /></button>
                <button class="icon-btn danger" aria-label="Hapus" title="Hapus" @click="hapus(m)"><Ico name="trash" :size="15" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <TablePagination v-if="!loading && !loadError" v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" label="mitra" />
  </div>

  <UiModal :open="formOpen" :busy="saving" :title="editing ? 'Edit Mitra' : 'Tambah Mitra'" @close="formOpen = false">
    <form id="form-mitra" @submit.prevent="save">
      <div class="grid-form">
        <div class="field">
          <label for="mitra-form-nama">Nama Mitra *</label>
          <input id="mitra-form-nama" v-model="form.nama" class="ctrl" placeholder="CV Sinar Data" required />
        </div>
        <div class="field">
          <label for="mitra-form-kode">Kode</label>
          <input id="mitra-form-kode" v-model="form.kode" class="ctrl" placeholder="otomatis (MTR-001)" />
        </div>
        <div class="field">
          <label for="mitra-form-jenis">Jenis</label>
          <select id="mitra-form-jenis" v-model="form.jenis" class="ctrl">
            <option v-for="j in jenisOpsi" :key="j" :value="j">{{ j }}</option>
          </select>
        </div>
        <div class="field">
          <label for="mitra-form-pemilik">Nama Pemilik</label>
          <input id="mitra-form-pemilik" v-model="form.pemilik" class="ctrl" placeholder="Budi Santoso" />
        </div>
        <div class="field">
          <label for="mitra-form-telepon">Telepon / WhatsApp</label>
          <input id="mitra-form-telepon" v-model="form.telepon" class="ctrl" placeholder="0812-3456-7890" />
        </div>
        <div class="field">
          <label for="mitra-form-email">Email</label>
          <input id="mitra-form-email" v-model="form.email" class="ctrl" type="email" placeholder="mitra@email.com" />
        </div>
        <div class="field full">
          <label for="mitra-form-alamat">Alamat</label>
          <input id="mitra-form-alamat" v-model="form.alamat" class="ctrl" placeholder="Jl. Merdeka No. 45" />
        </div>
        <div class="field">
          <label for="mitra-form-wilayah">Wilayah / Kota</label>
          <input id="mitra-form-wilayah" v-model="form.wilayah" class="ctrl" placeholder="Bandung" />
        </div>
        <div class="field">
          <label for="mitra-form-npwp">NPWP</label>
          <input id="mitra-form-npwp" v-model="form.npwp" class="ctrl" placeholder="01.234.567.8-901.000" />
        </div>
        <div class="field">
          <label for="mitra-form-status">Status</label>
          <select id="mitra-form-status" v-model="form.status" class="ctrl">
            <option v-for="s in statusOpsi" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
        <div class="field">
          <label for="mitra-form-share-mitra">Bagi Hasil Mitra (%)</label>
          <input id="mitra-form-share-mitra" v-model.number="form.share_mitra" class="ctrl" type="number" min="0" max="100" step="0.1" />
          <span class="hint">Sisa {{ 100 - (form.share_mitra || 0) }}% untuk ISP</span>
        </div>
        <div class="field">
          <label for="mitra-form-tanggal-gabung">Tanggal Gabung</label>
          <input id="mitra-form-tanggal-gabung" v-model="form.tanggal_gabung" class="ctrl" type="date" />
        </div>
        <div class="field full"><label for="mitra-bank">Rekening penerima bagi hasil (FAB)</label><textarea id="mitra-bank" v-model="form.rekening" class="ctrl" rows="3" placeholder="Nama bank, nomor rekening, dan nama pemilik"></textarea><span class="hint">Disimpan pada FAB baru. Rekening perusahaan untuk invoice tetap di Pengaturan.</span></div>
        <div class="field full">
          <label for="mitra-form-catatan">Catatan</label>
          <textarea id="mitra-form-catatan" v-model="form.catatan" class="ctrl" rows="2" placeholder="Catatan tambahan…"></textarea>
        </div>
      </div>
    </form>
    <template #footer>
      <button class="btn btn-sec" :disabled="saving" @click="formOpen = false">Batal</button>
      <button class="btn btn-primary" :disabled="saving" type="submit" form="form-mitra">
        {{ saving ? 'Menyimpan…' : 'Simpan' }}
      </button>
    </template>
  </UiModal>

  <UiModal :open="!!detail" :title="detail?.mitra?.nama || 'Detail Mitra'" wide @close="detail = null">
    <template v-if="detail">
      <div class="stat-grid" style="grid-template-columns:repeat(4,1fr);margin-bottom:16px">
        <div class="card stat" style="border:0;padding:12px 14px">
          <div class="stat-ico ico-green" style="width:36px;height:36px"><Ico name="wifi" :size="17" /></div>
          <div><div class="val" style="font-size:17px">{{ detail.rekap.pelanggan_aktif }}</div><div class="lbl">Pelanggan Aktif</div></div>
        </div>
        <div class="card stat" style="border:0;padding:12px 14px">
          <div class="stat-ico ico-blue" style="width:36px;height:36px"><Ico name="money" :size="17" /></div>
          <div><div class="val" style="font-size:17px">{{ fmtShort(detail.rekap.total_lunas) }}</div><div class="lbl">Lunas Bulan Ini</div></div>
        </div>
        <div class="card stat" style="border:0;padding:12px 14px">
          <div class="stat-ico ico-amber" style="width:36px;height:36px"><Ico name="wallet" :size="17" /></div>
          <div><div class="val" style="font-size:17px">{{ fmtShort(detail.rekap.nilai) }}</div><div class="lbl">Bagi Hasil {{ detail.mitra.share_mitra }}%</div></div>
        </div>
        <div class="card stat" style="border:0;padding:12px 14px">
          <div class="stat-ico ico-blue" style="width:36px;height:36px"><Ico name="file" :size="17" /></div>
          <div><div class="val" style="font-size:17px">{{ detail.dokumen.length }}</div><div class="lbl">Dokumen</div></div>
        </div>
      </div>

      <dl class="kv" style="margin-bottom:18px">
        <dt>Pemilik</dt><dd>{{ detail.mitra.pemilik || '-' }}</dd>
        <dt>Kontak</dt><dd>{{ detail.mitra.telepon || '-' }} {{ detail.mitra.email ? '· ' + detail.mitra.email : '' }}</dd>
        <dt>Alamat</dt><dd>{{ detail.mitra.alamat || '-' }}, {{ detail.mitra.wilayah || '' }}</dd>
        <dt>Gabung</dt><dd>{{ fmtDate(detail.mitra.tanggal_gabung) }}</dd>
      </dl>

      <div class="toolbar" style="margin-bottom:12px">
        <button class="btn btn-primary btn-sm" @click="terbitkan('PKS')"><Ico name="file" :size="14" /> Terbitkan PKS</button>
        <button class="btn btn-sec btn-sm" @click="terbitkan('FAB')"><Ico name="file" :size="14" /> Terbitkan FAB</button>
      </div>

      <h4 style="font-size:13px;margin-bottom:6px">Pelanggan ({{ detail.pelanggan.length }})</h4>
      <div class="table-wrap">
        <table class="tbl">
          <thead><tr><th>Nama</th><th>Paket</th><th>Status</th></tr></thead>
          <tbody>
            <tr v-for="p in customerPages.rows" :key="p.id">
              <td><div class="cell-main">{{ p.nama }}</div><div class="cell-sub mono">{{ p.pppoe }}</div></td>
              <td>{{ p.paket }} <span class="cell-sub">({{ p.kecepatan }} Mbps)</span></td>
              <td><span class="badge" :class="p.status === 'Aktif' ? 'b-success' : p.status === 'Isolir' ? 'b-warn' : 'b-muted'">{{ p.status }}</span></td>
            </tr>
            <tr v-if="!detail.pelanggan.length"><td colspan="3" style="color:var(--muted)">Belum ada pelanggan.</td></tr>
          </tbody>
        </table>
      </div>

      <TablePagination v-model:page="customerPages.page" v-model:page-size="customerPages.pageSize" :total="customerPages.total" label="pelanggan mitra" style="margin-bottom:16px" />

      <h4 style="font-size:13px;margin-bottom:6px">Dokumen</h4>
      <div class="table-wrap">
        <table class="tbl">
          <thead><tr><th>Nomor</th><th>Tanggal</th><th>Nilai</th></tr></thead>
          <tbody>
            <tr v-for="d in documentPages.rows" :key="d.id">
              <td><span class="badge" :class="d.jenis === 'PKS' ? 'b-info' : 'b-success'">{{ d.jenis }}</span> <span class="mono">{{ d.nomor }}</span></td>
              <td>{{ fmtDate(d.tanggal) }}</td>
              <td class="num">{{ d.jenis === 'FAB' ? fmtRp(d.nilai) : '-' }}</td>
            </tr>
            <tr v-if="!detail.dokumen.length"><td colspan="3" style="color:var(--muted)">Belum ada dokumen.</td></tr>
          </tbody>
        </table>
      </div>
      <TablePagination v-model:page="documentPages.page" v-model:page-size="documentPages.pageSize" :total="documentPages.total" label="dokumen mitra" />
    </template>
    <template #footer>
      <button class="btn btn-sec" @click="detail = null">Tutup</button>
    </template>
  </UiModal>
</template>




