<script setup lang="ts">
const dash = ref<any>(null)
const loading = ref(true)
const selectedMonth = ref(-1)
const nama = ref('Administrator')
const greeting = ref('Selamat datang')
const thisMonth = periodeLabel(localISODate().slice(0, 7))
async function load() {
  loading.value = true
  try { dash.value = await api('/api/dashboard'); selectedMonth.value = dash.value.revenue_months.length - 1 }
  catch (e) { toast(errMsg(e), 'error') }
  finally { loading.value = false }
}
onMounted(() => {
  nama.value = localStorage.getItem('isp_nama') || 'Administrator'
  const hour = new Date().getHours()
  greeting.value = hour < 11 ? 'Selamat pagi' : hour < 15 ? 'Selamat siang' : hour < 18 ? 'Selamat sore' : 'Selamat malam'
  load()
})
function createDocument(jenis: string) {
  useState<any>('docIntent', () => null).value = { jenis }
  navigateTo('/dokumen')
}
function openDocument(d: any) {
  useState<any>('docPreviewIntent', () => null).value = d.id
  navigateTo('/dokumen')
}
const pct = computed(() => dash.value?.tagihan_bulan ? Math.min(100, Math.round(dash.value.pendapatan_bulan / dash.value.tagihan_bulan * 100)) : 0)
const outstanding = computed(() => Math.max(0, (dash.value?.tagihan_bulan || 0) - (dash.value?.pendapatan_bulan || 0)))
const months = computed<any[]>(() => dash.value?.revenue_months || [])
const chartMax = computed(() => Math.max(...months.value.map(m => m.value), 1000000))
const chartPoints = computed(() => months.value.map((m, i) => ({...m, x:48 + i * (592 / Math.max(1, months.value.length - 1)), y:185 - m.value / chartMax.value * 150})))
const linePath = computed(() => chartPoints.value.map((p, i) => `${i ? 'L' : 'M'} ${p.x} ${p.y}`).join(' '))
const areaPath = computed(() => chartPoints.value.length ? `${linePath.value} L ${chartPoints.value.at(-1)!.x} 185 L 48 185 Z` : '')
const selectedRevenue = computed(() => months.value[selectedMonth.value] || {label:'', value:0})
const documentCount = computed(() => dash.value?.dokumen_terbaru?.length || 0)
const partnerPages = usePagination<any>(() => dash.value?.rekap_mitra || [])
</script>

<template>
  <div class="page-head dashboard-heading">
    <div><h1>{{ greeting }}, {{ nama.split(' ')[0] }}<span class="greeting-dot">.</span></h1><p class="desc">Ini yang terjadi pada layanan Anda hari ini.</p></div>
    <NuxtLink to="/keuangan" class="btn btn-sec"><Ico name="calendar" :size="17" />{{ thisMonth }}</NuxtLink>
  </div>
  <div v-if="loading" class="dashboard-loading" role="status"><span class="loading-line"></span><p>Menyiapkan ringkasan operasional…</p></div>
  <div v-else-if="!dash" class="empty card"><div class="e-ico"><Ico name="alert" /></div><p>Ringkasan belum dapat dimuat.</p><button class="btn btn-sec" @click="load">Coba lagi</button></div>
  <template v-else>
    <section class="overview-metrics" aria-label="Ringkasan operasional">
      <NuxtLink to="/keuangan" class="overview-metric"><div class="metric-label"><span>Pendapatan bulan ini</span><span class="metric-icon metric-blue"><Ico name="wallet" :size="18" /></span></div><strong>{{ fmtRp(dash.pendapatan_bulan) }}</strong><span class="metric-note"><span class="status-dot green"></span>{{ pct }}% dari total tagihan</span></NuxtLink>
      <NuxtLink to="/pelanggan" class="overview-metric"><div class="metric-label"><span>Pelanggan aktif</span><span class="metric-icon metric-purple"><Ico name="wifi" :size="18" /></span></div><strong>{{ dash.pelanggan_aktif }}<small>pelanggan</small></strong><span class="metric-note">Layanan internet berjalan</span></NuxtLink>
      <NuxtLink to="/mitra" class="overview-metric"><div class="metric-label"><span>Mitra aktif</span><span class="metric-icon metric-green"><Ico name="users" :size="18" /></span></div><strong>{{ dash.mitra_aktif }}<small>mitra</small></strong><span class="metric-note">Bersama membangun layanan</span></NuxtLink>
      <NuxtLink to="/tiket" class="overview-metric"><div class="metric-label"><span>Tiket terbuka</span><span class="metric-icon metric-orange"><Ico name="wrench" :size="18" /></span></div><strong>{{ dash.tiket_terbuka }}<small>laporan</small></strong><span class="metric-note" :class="{'attention-note':dash.tiket_terbuka}"><span class="status-dot" :class="dash.tiket_terbuka ? 'orange' : 'green'"></span>{{ dash.tiket_terbuka ? 'Menunggu tindak lanjut' : 'Semua laporan selesai' }}</span></NuxtLink>
    </section>

    <div class="analytics-grid">
      <section class="card revenue-panel">
        <div class="panel-heading"><div><h2>Pendapatan layanan</h2><p>Rekap pembayaran lunas selama 6 bulan terakhir.</p></div><span class="chart-legend"><i></i>Pendapatan</span></div>
        <div class="revenue-readout" aria-live="polite"><strong>{{ fmtRp(selectedRevenue.value) }}</strong><span>{{ selectedRevenue.label }} · tagihan lunas</span></div>
        <div class="revenue-chart">
          <svg viewBox="0 0 680 214" role="img" aria-label="Grafik pendapatan 6 bulan; pilih bulan di bawah untuk melihat nominal">
            <g v-for="tick in [0,1,2,3]" :key="tick"><line x1="48" x2="660" :y1="185-tick*50" :y2="185-tick*50" class="chart-gridline" /><text x="0" :y="189-tick*50" class="chart-axis">{{ fmtShort(chartMax*tick/3) }}</text></g>
            <path :d="areaPath" class="chart-area"/><path :d="linePath" class="chart-line"/>
            <g v-for="(point,i) in chartPoints" :key="i"><line v-if="i === selectedMonth" :x1="point.x" :x2="point.x" y1="24" y2="185" class="chart-selection"/><circle :cx="point.x" :cy="point.y" :r="i === selectedMonth ? 6 : 4" class="chart-point" :class="{'is-selected':i === selectedMonth}"><title>{{ point.label }}: {{ fmtRp(point.value) }}</title></circle></g>
          </svg>
          <div class="chart-months"><button v-for="(month,i) in months" :key="i" :class="{selected:i === selectedMonth}" :aria-pressed="i === selectedMonth" :aria-label="`${month.label}, ${fmtRp(month.value)}`" @click="selectedMonth=i">{{ month.label }}</button></div>
        </div>
      </section>
      <section class="card collection-panel"><div class="panel-heading"><div><h2>Progres penagihan</h2><p>{{ thisMonth }}</p></div><Ico name="money" :size="19" /></div>
        <div class="collection-ring"><svg viewBox="0 0 120 120" aria-hidden="true"><circle cx="60" cy="60" r="48" class="ring-track"/><circle cx="60" cy="60" r="48" class="ring-value" :stroke-dasharray="`${pct*3.0159} 301.59`"/></svg><div><strong>{{ pct }}<span>%</span></strong><small>Terkumpul</small></div></div>
        <dl class="collection-values"><div><dt><i class="status-dot blue"></i>Sudah lunas</dt><dd>{{ fmtRp(dash.pendapatan_bulan) }}</dd></div><div><dt><i class="status-dot pale"></i>Belum dibayar</dt><dd>{{ fmtRp(outstanding) }}</dd></div></dl>
        <NuxtLink to="/keuangan" class="btn btn-sec collection-action">Kelola tagihan<Ico name="arrow" :size="16" /></NuxtLink>
      </section>
    </div>

    <div class="dashboard-work-grid">
      <section class="card"><div class="panel-heading"><div><h2>Perlu ditangani<span class="heading-count">{{ dash.tiket_terbuka }}</span></h2><p>Laporan gangguan yang masih terbuka.</p></div><NuxtLink to="/tiket" class="text-link">Lihat semua<Ico name="chevron" :size="14" /></NuxtLink></div>
        <div v-if="!dash.tiket_terbaru?.length" class="empty"><div class="e-ico"><Ico name="check" /></div><p>Tidak ada laporan yang menunggu penanganan.</p></div>
        <div v-else class="ticket-inbox"><NuxtLink v-for="ticket in dash.tiket_terbaru" :key="ticket.id" to="/tiket" class="ticket-inbox-item"><span class="ticket-avatar" :class="ticket.prioritas === 'Tinggi' || ticket.prioritas === 'Urgent' ? 'urgent' : ''"><Ico name="wrench" :size="17" /></span><div class="ticket-inbox-copy"><b>{{ ticket.pelanggan_nama }}</b><span>{{ ticket.jenis }}<i>·</i>{{ ticket.nomor }}</span></div><span class="badge" :class="ticket.status === 'Baru' ? 'b-warn' : 'b-info'">{{ ticket.status }}</span><Ico name="chevron" :size="14" /></NuxtLink></div>
      </section>
      <section class="card recent-documents"><div class="panel-heading"><div><h2>Dokumen terbaru</h2><p>Arsip kerja sama dan bagi hasil mitra.</p></div><NuxtLink to="/dokumen" class="text-link">Lihat semua<Ico name="chevron" :size="14" /></NuxtLink></div>
        <div class="recent-document-list"><button v-for="doc in dash.dokumen_terbaru" :key="doc.id" class="recent-document" @click="openDocument(doc)"><span class="doc-symbol" :class="{fab:doc.jenis === 'FAB'}"><Ico name="file" :size="19" /></span><span><b>{{ doc.nomor }}</b><small>{{ doc.mitra_nama }}</small></span><span class="recent-doc-date">{{ fmtDate(doc.tanggal) }}</span></button><p v-if="!documentCount" class="empty">Belum ada dokumen diterbitkan.</p></div>
        <div class="recent-doc-actions"><button @click="createDocument('PKS')"><Ico name="plus" :size="15" />Buat PKS</button><button @click="createDocument('FAB')"><Ico name="plus" :size="15" />Buat FAB</button></div>
      </section>
    </div>

    <section class="card partner-performance"><div class="panel-heading"><div><h2>Kontribusi mitra</h2><p>Rekap pelanggan dan bagi hasil {{ thisMonth.toLowerCase() }}.</p></div><NuxtLink to="/keuangan" class="text-link">Detail keuangan<Ico name="chevron" :size="14" /></NuxtLink></div><div class="table-wrap"><table class="tbl"><thead><tr><th>Mitra</th><th class="num">Pelanggan aktif</th><th class="num">Pembayaran lunas</th><th class="num">Persentase mitra</th><th class="num">Bagi hasil</th></tr></thead><tbody><tr v-for="r in partnerPages.rows" :key="r.mitra_id"><td><div class="partner-cell"><span class="partner-avatar">{{ r.mitra_nama.slice(0,2).toUpperCase() }}</span><b>{{ r.mitra_nama }}</b></div></td><td class="num">{{ r.pelanggan_aktif }}</td><td class="num">{{ fmtRp(r.total_lunas) }}</td><td class="num">{{ r.share_pct }}%</td><td class="num cell-main">{{ fmtRp(r.nilai) }}</td></tr><tr v-if="!dash.rekap_mitra?.length"><td colspan="5" class="empty">Tambahkan mitra untuk mulai melihat rekap.</td></tr></tbody></table></div><TablePagination v-model:page="partnerPages.page" v-model:page-size="partnerPages.pageSize" :total="partnerPages.total" label="kontribusi mitra" /></section>
  </template>
</template>
