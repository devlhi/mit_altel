export const TOKEN_KEY = 'isp_token'

export function getToken(): string | null {
  if (import.meta.client) return localStorage.getItem(TOKEN_KEY)
  return null
}

export function setToken(t: string) {
  if (import.meta.client) localStorage.setItem(TOKEN_KEY, t)
}

export function clearToken() {
  if (import.meta.client) {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem('isp_nama')
  }
}

export function errMsg(e: any): string {
  return e?.data?.error || e?.message || 'Terjadi kesalahan, silakan coba lagi'
}

interface ApiOpts {
  method?: string
  body?: any
}

export async function api<T = any>(url: string, opts: ApiOpts = {}): Promise<T> {
  try {
    return await $fetch<T>(url, {
      method: (opts.method || 'GET') as any,
      body: opts.body,
      headers: getToken() ? { Authorization: `Bearer ${getToken()}` } : undefined
    })
  } catch (e: any) {
    const status = e?.status || e?.response?.status
    if (status === 401 && import.meta.client && !location.pathname.startsWith('/login')) {
      clearToken()
      location.href = '/login'
    }
    throw e
  }
}
