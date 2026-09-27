import { computed, reactive, ref, toValue, watch } from 'vue'
import type { MaybeRefOrGetter, WatchSource } from 'vue'

export function usePagination<T>(source: MaybeRefOrGetter<readonly T[]>, options: {
  pageSize?: number
  resetOn?: WatchSource | WatchSource[]
} = {}) {
  const page = ref(1)
  const pageSize = ref(options.pageSize || 10)
  const items = computed(() => toValue(source))
  const total = computed(() => items.value.length)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const rows = computed(() => items.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))

  // Deleting the last row of a page must never leave the user on an empty page.
  watch(totalPages, count => { page.value = Math.min(page.value, count) }, { flush: 'sync' })
  watch(pageSize, () => { page.value = 1 }, { flush: 'sync' })
  if (options.resetOn) watch(options.resetOn, () => { page.value = 1 }, { flush: 'sync' })

  return reactive({ page, pageSize, total, totalPages, rows })
}
