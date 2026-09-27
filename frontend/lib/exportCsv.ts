export function exportCsv(name: string, rows: (string | number)[][]) {
  const text = rows.map(row => row.map(value => {
    let s = String(value ?? '')
    if (typeof value === 'string' && /^[=+\-@\t\r]/.test(s)) s = `'${s}`
    return `"${s.replace(/"/g, '""')}"`
  }).join(';')).join('\r\n')
  const url = URL.createObjectURL(new Blob(['\uFEFF', text], { type: 'text/csv;charset=utf-8;' }))
  const a = document.createElement('a'); a.href = url; a.download = name; a.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
