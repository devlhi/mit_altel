<script setup lang="ts">
const route = useRoute()
const { data: settings } = useAppSettings()
const navOpen = ref(false)
const navPanel = ref<HTMLElement | null>(null)
const navButton = ref<HTMLElement | null>(null)
const searchOpen = ref(false)
const searchQuery = ref('')
const nav = [
  { label: 'Dashboard', to: '/', icon: 'grid', hint: 'Ringkasan operasional dan pendapatan' },
  { label: 'Mitra', to: '/mitra', icon: 'users', hint: 'Kerja sama, kontak, dan bagi hasil' },
  { label: 'Pelanggan', to: '/pelanggan', icon: 'wifi', hint: 'Pelanggan dan layanan internet' },
  { label: 'Data POP', to: '/pop', icon: 'building', hint: 'Lokasi jaringan, perangkat, dan pelanggan' },
  { label: 'Paket layanan', to: '/paket', icon: 'box', hint: 'Kecepatan dan harga langganan' },
  { label: 'Tiket gangguan', to: '/tiket', icon: 'wrench', hint: 'Laporan dan penanganan gangguan' }
]
const nav2 = [
  { label: 'Keuangan', to: '/keuangan', icon: 'wallet', hint: 'Tagihan dan rekap pembayaran' },
  { label: 'Pembukuan', to: '/pembukuan', icon: 'grid', hint: 'Saldo awal, jurnal, neraca, laba-rugi, arus kas' },
  { label: 'BHP & USO', to: '/bhp-uso', icon: 'sliders', hint: 'Estimasi kewajiban telekomunikasi tahunan' },
  { label: 'Dokumen', to: '/dokumen', icon: 'file', hint: 'Perjanjian PKS dan faktur FAB' },
  { label: 'Surat POP', to: '/surat-pop', icon: 'file', hint: 'Surat Komdigi, KBLI, lokasi POP, tanda tangan dan stempel' },
  { label: 'WhatsApp', to: '/whatsapp', icon: 'message', hint: 'Gateway MPWA dan riwayat pengiriman' }
]
const allPages = [...nav, ...nav2, { label: 'Pengaturan', to: '/pengaturan', icon: 'sliders', hint: 'Profil perusahaan, template, dan akun' }]
const searchResults = computed(() => allPages.filter(p => `${p.label} ${p.hint}`.toLowerCase().includes(searchQuery.value.toLowerCase().trim())))
const titles: Record<string, [string, string]> = {
  '/': ['Dashboard', 'Ringkasan aktivitas layanan Anda.'],
  '/mitra': ['Mitra kerja sama', 'Bangun hubungan mitra dan kelola pembagian hasil.'],
  '/pelanggan': ['Pelanggan', 'Semua pelanggan dan layanan mereka, dalam satu tempat.'],
  '/paket': ['Paket layanan', 'Pilihan kecepatan dan harga untuk kebutuhan pelanggan.'],
  '/tiket': ['Tiket gangguan', 'Pantau laporan masuk dan progres penanganannya.'],
  '/keuangan': ['Keuangan', 'Kelola penagihan dan pantau pendapatan setiap bulan.'],
  '/pembukuan': ['Pembukuan', 'Posisi keuangan dan transaksi perusahaan yang tertata.'],
  '/bhp-uso': ['BHP & USO', 'Hitung estimasi kewajiban dari pendapatan yang tercatat.'],
  '/pop': ['Data POP', 'Pusatkan data lokasi dan infrastruktur jaringan Anda.'],
  '/dokumen': ['Dokumen mitra', 'Perjanjian dan rekap bagi hasil yang tertata.'],
  '/surat-pop': ['Surat POP Komdigi', 'Penerbitan dan arsip surat lokasi jaringan perusahaan.'],
  '/whatsapp': ['WhatsApp', 'Kirim pesan dan dokumen melalui MPWA.'],
  '/pengaturan': ['Pengaturan', 'Atur identitas perusahaan, dokumen, dan akun Anda.']
}
const t = computed(() => titles[route.path] || ['Manajemen ISP', ''])
const nama = ref('Administrator')
const tanggal = ref('')
const initials = computed(() => nama.value.split(' ').slice(0, 2).map(n => n[0]).join('').toUpperCase())
const customHeading = computed(() => ['/', '/dokumen'].includes(route.path))
useHead({ title: computed(() => `${t.value[0]} · ${settings.value?.nama_isp || 'Ali Telematika'}`) })
onMounted(() => {
  nama.value = localStorage.getItem('isp_nama') || 'Administrator'
  tanggal.value = new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
  window.addEventListener('keydown', searchShortcut)
})
watch(() => route.path, () => { closeNav(); searchOpen.value = false })
let releaseScroll: (() => void) | undefined
watch(navOpen, async open => {
  if (open) {
    releaseScroll = lockPageScroll()
    await nextTick()
    navPanel.value?.querySelector<HTMLElement>('button, a')?.focus()
  } else {
    releaseScroll?.(); releaseScroll = undefined
    await nextTick(); navButton.value?.focus()
  }
})
watch(searchOpen, open => { if (open) searchQuery.value = '' })
onBeforeUnmount(() => { releaseScroll?.(); window.removeEventListener('keydown', searchShortcut) })
function closeNav() { navOpen.value = false }
function onNavKey(event: KeyboardEvent) {
  if (!navOpen.value) return
  if (event.key === 'Escape') closeNav()
  if (event.key === 'Tab' && navPanel.value) trapFocus(event, navPanel.value)
}
function searchShortcut(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
    if (document.querySelector('.overlay') && !searchOpen.value) return
    event.preventDefault(); searchOpen.value = !searchOpen.value
  }
}
function openPage(path: string) { searchOpen.value = false; navigateTo(path) }
function logout() { clearToken(); clearNuxtData('app_settings'); navigateTo('/login') }
function isActive(to: string) { return to === '/' ? route.path === '/' : route.path.startsWith(to) }
</script>

<template>
  <div class="layout" :class="{ 'nav-open': navOpen }">
    <a class="skip-link" href="#main-content">Langsung ke konten</a>
    <aside id="navigation" ref="navPanel" class="sidebar" :class="{ 'mobile-closed': !navOpen }" aria-label="Navigasi utama" @keydown="onNavKey">
      <NuxtLink to="/" class="brand official-brand" aria-label="PT Ali Telematika Nusantara, dashboard">
        <BrandLogo :width="200" />
        <span class="brand-sub">Manajemen layanan internet</span>
      </NuxtLink>
      <button class="icon-btn nav-close" aria-label="Tutup navigasi" @click="closeNav"><Ico name="x" :size="18" /></button>
      <div class="workspace-switch"><div class="workspace-icon"><Ico name="building" :size="17" /></div><div><b>Ruang kerja utama</b><span>Operasional perusahaan</span></div></div>
      <nav class="menu">
        <div class="menu-label">Ruang kerja</div>
        <NuxtLink v-for="m in nav" :key="m.to" :to="m.to" class="menu-item" :class="{ active: isActive(m.to) }" :aria-current="isActive(m.to) ? 'page' : undefined" @click="closeNav"><Ico :name="m.icon" :size="19" /><span>{{ m.label }}</span><span v-if="isActive(m.to)" class="nav-active-dot"></span></NuxtLink>
        <div class="menu-label">Bisnis</div>
        <NuxtLink v-for="m in nav2" :key="m.to" :to="m.to" class="menu-item" :class="{ active: isActive(m.to) }" :aria-current="isActive(m.to) ? 'page' : undefined" @click="closeNav"><Ico :name="m.icon" :size="19" /><span>{{ m.label }}</span><span v-if="m.to === '/dokumen'" class="nav-tag">PKS / FAB</span></NuxtLink>
      </nav>
      <div class="sidebar-bottom">
        <NuxtLink to="/pengaturan" class="menu-item" :class="{active:isActive('/pengaturan')}" @click="closeNav"><Ico name="sliders" :size="19" />Pengaturan</NuxtLink>
      </div>
    </aside>
    <button v-if="navOpen" class="scrim" tabindex="-1" aria-label="Tutup panel navigasi" @click="closeNav"></button>
    <div class="main" :inert="navOpen || undefined">
      <header class="topbar">
        <div class="topbar-leading"><button ref="navButton" class="icon-btn menu-btn" aria-label="Buka navigasi" :aria-expanded="navOpen" aria-controls="navigation" @click="navOpen = true"><Ico name="menu" :size="19" /></button><div class="breadcrumb"><span>Ruang kerja</span><Ico name="chevron" :size="13" /><b>{{ t[0] }}</b></div></div>
        <div class="topbar-tools">
          <button class="global-search" aria-label="Buka pencarian halaman" title="Cari halaman (Ctrl+K)" @click="searchOpen = true"><Ico name="search" :size="17" /><span>Cari halaman…</span><kbd>Ctrl K</kbd></button>
          <span class="topbar-date"><Ico name="calendar" :size="16" />{{ tanggal }}</span>
          <div class="topbar-profile" role="group" aria-label="Akun pengguna">
            <NuxtLink to="/pengaturan" class="account-link" :aria-label="`Buka pengaturan akun ${nama}`" :title="nama">
              <span class="avatar" aria-hidden="true">{{ initials }}</span>
              <span class="account-copy"><b>{{ nama }}</b><span>Administrator ISP</span></span>
            </NuxtLink>
            <button class="icon-btn account-logout" aria-label="Keluar dari akun" title="Keluar dari akun" @click="logout"><Ico name="logout" :size="17" /></button>
          </div>
        </div>
      </header>
      <main id="main-content" class="content" :class="{ 'custom-heading': customHeading }" tabindex="-1">
        <div v-if="!customHeading" class="page-head"><div><h1>{{ t[0] }}</h1><p class="desc">{{ t[1] }}</p></div></div>
        <slot />
        <footer class="workspace-footer"><span>{{ settings?.nama_isp || 'Ali Telematika' }}</span><span>Manajemen layanan internet</span></footer>
      </main>
    </div>
    <UiModal :open="searchOpen" title="Buka halaman" @close="searchOpen = false">
      <div class="search-box command-input"><Ico name="search" :size="20" /><input v-model="searchQuery" autofocus aria-label="Cari halaman" placeholder="Mitra, tagihan, dokumen…" @keydown.enter.prevent="searchResults.length && openPage(searchResults[0].to)" /></div>
      <div class="command-results"><button v-for="item in searchResults" :key="item.to" class="command-result" @click="openPage(item.to)"><span class="command-icon"><Ico :name="item.icon" :size="20" /></span><span><b>{{ item.label }}</b><small>{{ item.hint }}</small></span><Ico name="arrow" :size="16" /></button><p v-if="!searchResults.length" class="empty">Halaman tidak ditemukan. Coba kata pencarian lain.</p></div>
    </UiModal>
    <ToastHost /><ConfirmHost />
  </div>
</template>
<style scoped>
@media (max-width:900px) { .sidebar.mobile-closed { visibility: hidden; } }
</style>

