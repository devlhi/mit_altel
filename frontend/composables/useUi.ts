import type { Ref } from 'vue'

export interface ToastItem {
  id: number
  msg: string
  type: 'success' | 'error' | 'info'
}

export function useToastList(): Ref<ToastItem[]> {
  return useState<ToastItem[]>('toasts', () => [])
}

export function toast(msg: string, type: ToastItem['type'] = 'success') {
  const list = useToastList()
  const id = Date.now() + Math.random()
  list.value.push({ id, msg, type })
  setTimeout(() => {
    list.value = list.value.filter((t) => t.id !== id)
  }, 3800)
}

interface ConfirmState {
  msg: string
  resolve: (v: boolean) => void
}

export function useConfirmState(): Ref<ConfirmState | null> {
  return useState<ConfirmState | null>('confirm', () => null)
}

export function useConfirm() {
  const st = useConfirmState()
  return {
    ask(msg: string): Promise<boolean> {
      return new Promise((resolve) => {
        st.value = { msg, resolve }
      })
    }
  }
}
