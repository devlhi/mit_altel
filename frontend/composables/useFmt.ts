export function localISODate(date = new Date()): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function fmtRp(n: number | undefined | null): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0
  }).format(n || 0)
}

export function fmtShort(n: number | undefined | null): string {
  const v = n || 0
  if (v >= 1e9) return (v / 1e9).toFixed(1).replace('.', ',') + ' M'
  if (v >= 1e6) return (v / 1e6).toFixed(1).replace('.', ',') + ' jt'
  if (v >= 1e3) return Math.round(v / 1e3) + ' rb'
  return String(v)
}

export function fmtDate(iso: string | undefined | null): string {
  if (!iso) return '-'
  const d = new Date(iso.length === 10 ? iso + 'T00:00:00' : iso)
  if (isNaN(d.getTime())) return iso
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

export function periodeLabel(p: string | undefined | null): string {
  if (!p) return '-'
  const [y, m] = p.split('-')
  const bulan = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  const idx = parseInt(m, 10) - 1
  if (isNaN(idx) || idx < 0 || idx > 11) return p
  return `${bulan[idx]} ${y}`
}

export function waNumber(t: string | undefined | null): string {
  const d = (t || '').replace(/\D/g, '')
  if (!d) return ''
  if (d.startsWith('62')) return d
  if (d.startsWith('0')) return '62' + d.slice(1)
  if (d.startsWith('8')) return '62' + d
  return d
}

// useAppSettings: profil ISP untuk header dokumen & pesan kirim.
export function useAppSettings() {
  return useAsyncData('app_settings', () => api<any>('/api/settings'), {
    default: () => ({ nama_isp: 'ISP', penandatangan: '', telepon: '', email: '' })
  })
}
