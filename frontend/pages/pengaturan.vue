<script setup lang="ts">
const form = ref<any>(null)
const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const tplTab = ref<'pks' | 'fab'>('pks')
const passOpen = ref(false)
const passForm = ref({ lama: '', baru: '', ulang: '' })
const passSaving = ref(false)

const placeholderList = [
  ['{{nomor}}', 'Nomor dokumen, mis. PKS/2026/003'],
  ['{{hari}}, {{tanggal}}', 'Hari & tanggal terbit (Jumat, 26 September 2026)'],
  ['{{nama_isp}}, {{alamat_isp}}, {{kota}}', 'Identitas ISP'],
  ['{{penandatangan}}, {{jabatan}}', 'Penanda tangan ISP'],
  ['{{nama_mitra}}, {{pemilik_mitra}}, {{alamat_mitra}}', 'Data mitra'],
  ['{{share_mitra}}, {{share_isp}}', 'Persentase bagi hasil'],
  ['{{masa_bulan}}, {{tanggal_akhir}}', 'Masa berlaku PKS'],
  ['{{periode_label}}, {{rincian_rows}}', 'Periode & baris rincian tagihan (FAB)'],
  ['{{total_lunas}}, {{nilai_bagihasil}}, {{nilai_terbilang}}', 'Nilai FAB'],
  ['{{rekening}}', 'Rekening perusahaan untuk invoice'],
  ['{{rekening_mitra}}', 'Rekening penerima bagi hasil mitra (FAB)']
]

const placeholderPages = usePagination(placeholderList)

async function load() {
  loadError.value = ''
  loading.value = true
  try {
    form.value = await api('/api/settings')
  } catch (e) {
    loadError.value = errMsg(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    form.value = await api('/api/settings', { method: 'PUT', body: form.value })
    await refreshNuxtData('app_settings')
    toast('Pengaturan tersimpan')
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    saving.value = false
  }
}

async function pulihkanTemplate() {
  const jenis = tplTab.value === 'pks' ? 'PKS' : 'FAB'
  const ok = await useConfirm().ask(`Kembalikan template ${jenis} ke bawaan? Perubahan Anda pada template ini akan hilang.`)
  if (!ok) return
  try {
    const def = await api<any>('/api/settings/defaults')
    if (tplTab.value === 'pks') form.value.tpl_pks = def.tpl_pks
    else form.value.tpl_fab = def.tpl_fab
    toast(`Template ${jenis} dikembalikan ke bawaan. Klik Simpan untuk menerapkan.`)
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}

async function gantiPassword() {
  const f = passForm.value
  if (!f.lama || !f.baru) {
    toast('Isi password lama dan baru', 'error')
    return
  }
  if (f.baru !== f.ulang) {
    toast('Konfirmasi password baru tidak sama', 'error')
    return
  }
  passSaving.value = true
  try {
    await api('/api/password', { method: 'PUT', body: { lama: f.lama, baru: f.baru } })
    toast('Password berhasil diganti')
    passOpen.value = false
    passForm.value = { lama: '', baru: '', ulang: '' }
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    passSaving.value = false
  }
}

async function muatContoh() {
  const ok = await useConfirm().ask(
    'Muat ulang data contoh? SEMUA data mitra, pelanggan, tiket, tagihan, dan dokumen akan diganti dengan data contoh. Akun login tidak berubah.'
  )
  if (!ok) return
  try {
    await api('/api/seed', { method: 'POST', body: {} })
    toast('Data contoh berhasil dimuat ulang')
  } catch (e) {
    toast(errMsg(e), 'error')
  }
}
</script>

<template>
  <div v-if="loading" class="empty"><p>Memuat pengaturan…</p></div>

  <div v-else-if="loadError" class="empty" role="alert"><p>{{ loadError }}</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
  <template v-else-if="form">
    <div class="toolbar">
      <div class="spacer"></div>
      <button class="btn btn-primary" :disabled="saving" type="submit" form="settings-profile">{{ saving ? 'Menyimpan…' : 'Simpan Pengaturan' }}</button>
    </div>

    <div class="grid-2" style="align-items:start">
      <div class="card">
        <div class="card-head">
          <h3>Profil ISP</h3>
          <span class="cell-sub">tampil di kop dokumen</span>
        </div>
        <div class="card-pad">
          <div class="settings-brand"><BrandLogo :width="285" /></div>
          <form id="settings-profile" class="grid-form" @submit.prevent="save">
            <div class="field">
              <label for="pengaturan-form-nama-isp">Nama ISP *</label>
              <input id="pengaturan-form-nama-isp" v-model="form.nama_isp" required class="ctrl" />
            </div>
            <div class="field">
              <label for="pengaturan-form-kota">Kota</label>
              <input id="pengaturan-form-kota" v-model="form.kota" class="ctrl" />
            </div>
            <div class="field full">
              <label for="pengaturan-form-alamat">Alamat</label>
              <input id="pengaturan-form-alamat" v-model="form.alamat" class="ctrl" />
            </div>
            <div class="field">
              <label for="pengaturan-form-telepon">Telepon</label>
              <input id="pengaturan-form-telepon" v-model="form.telepon" class="ctrl" />
            </div>
            <div class="field">
              <label for="pengaturan-form-email">Email</label>
              <input id="pengaturan-form-email" v-model="form.email" class="ctrl" type="email" />
            </div>
            <div class="field">
              <label for="pengaturan-form-website">Website</label>
              <input id="pengaturan-form-website" v-model="form.website" class="ctrl" />
            </div>
            <div class="field">
              <label for="pengaturan-form-npwp">NPWP</label>
              <input id="pengaturan-form-npwp" v-model="form.npwp" class="ctrl" />
            </div>
            <div class="field">
              <label for="pengaturan-form-penandatangan">Penanda Tangan</label>
              <input id="pengaturan-form-penandatangan" v-model="form.penandatangan" class="ctrl" placeholder="Nama direktur" />
            </div>
            <div class="field">
              <label for="pengaturan-form-jabatan">Jabatan</label>
              <input id="pengaturan-form-jabatan" v-model="form.jabatan" class="ctrl" />
            </div>
            <div class="field full">
              <label for="pengaturan-form-rekening">Rekening perusahaan (pembayaran invoice)</label>
              <textarea id="pengaturan-form-rekening" v-model="form.rekening" class="ctrl" rows="2" placeholder="Nama bank dan nomor rekening&#10;Nama pemilik rekening"></textarea>
            </div>
            <div class="field">
              <label for="pengaturan-form-masa-pks-bulan">Masa PKS Default (bulan)</label>
              <input id="pengaturan-form-masa-pks-bulan" v-model.number="form.masa_pks_bulan" class="ctrl" type="number" min="1" max="120" required />
            </div>
          </form>
        </div>
      </div>

      <div style="display:flex;flex-direction:column;gap:16px">
        <div class="card">
          <div class="card-head">
            <h3>Template Dokumen</h3>
            <div class="tabs">
              <button type="button" :class="{ on: tplTab === 'pks' }" @click="tplTab = 'pks'">PKS</button>
              <button type="button" :class="{ on: tplTab === 'fab' }" @click="tplTab = 'fab'">FAB</button>
            </div>
          </div>
          <div class="card-pad">
            <textarea
              v-if="tplTab === 'pks'"
              v-model="form.tpl_pks"
              aria-label="Template PKS"
              class="ctrl mono"
              rows="14"
              spellcheck="false"
            ></textarea>
            <textarea v-else v-model="form.tpl_fab" aria-label="Template FAB" class="ctrl mono" rows="14" spellcheck="false"></textarea>
            <div class="toolbar" style="margin:12px 0 0">
              <button class="btn btn-sec btn-sm" @click="pulihkanTemplate"><Ico name="refresh" :size="14" /> Kembalikan Default</button>
              <span class="cell-sub">HTML sederhana · placeholder</span> <span class="cell-sub mono" v-text="'{{ nama_placeholder }}'"></span>
            </div>
            <details style="margin-top:12px">
              <summary style="font-size:13px;cursor:pointer;color:var(--muted)">Daftar placeholder yang tersedia</summary>
              <div class="table-wrap" style="margin-top:8px"><table class="tbl">
                <tbody>
                  <tr v-for="p in placeholderPages.rows" :key="p[0]">
                    <td class="mono" style="white-space:nowrap">{{ p[0] }}</td>
                    <td class="cell-sub">{{ p[1] }}</td>
                  </tr>
                </tbody>
              </table></div>
              <TablePagination v-model:page="placeholderPages.page" v-model:page-size="placeholderPages.pageSize" :total="placeholderPages.total" label="placeholder" />
            </details>
          </div>
        </div>

        <div class="card">
          <div class="card-head"><h3>Keamanan</h3></div>
          <div class="card-pad">
            <p class="cell-sub" style="margin-bottom:12px">Ganti password login aplikasi secara berkala.</p>
            <button class="btn btn-sec" @click="passOpen = true"><Ico name="key" :size="15" /> Ganti Password</button>
          </div>
        </div>

        <div class="card">
          <div class="card-head"><h3>Data</h3></div>
          <div class="card-pad">
            <p class="cell-sub" style="margin-bottom:12px">
              Memuat data contoh akan mengganti seluruh data operasional. Pastikan administrator telah mencadangkan data sebelum melanjutkan.
            </p>
            <button class="btn btn-sec" style="color:var(--error)" @click="muatContoh"><Ico name="refresh" :size="15" /> Muat Ulang Data Contoh</button>
          </div>
        </div>
      </div>
    </div>

    <MPWASettings />

    <UiModal :open="passOpen" title="Ganti Password" @close="passOpen = false">
      <div class="field">
        <label for="pengaturan-passForm-lama">Password Lama</label>
        <input id="pengaturan-passForm-lama" v-model="passForm.lama" class="ctrl" type="password" />
      </div>
      <div class="field">
        <label for="pengaturan-passForm-baru">Password Baru</label>
        <input id="pengaturan-passForm-baru" v-model="passForm.baru" class="ctrl" type="password" />
        <span class="hint">Minimal 6 karakter</span>
      </div>
      <div class="field">
        <label for="pengaturan-passForm-ulang">Ulangi Password Baru</label>
        <input id="pengaturan-passForm-ulang" v-model="passForm.ulang" class="ctrl" type="password" />
      </div>
      <template #footer>
        <button class="btn btn-sec" @click="passOpen = false">Batal</button>
        <button class="btn btn-primary" :disabled="passSaving" @click="gantiPassword">{{ passSaving ? 'Menyimpan…' : 'Ganti Password' }}</button>
      </template>
    </UiModal>
  </template>
</template>



