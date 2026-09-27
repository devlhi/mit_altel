<script setup lang="ts">
const props = withDefaults(defineProps<{
  page: number
  pageSize: number
  total: number
  label: string
  pageSizes?: number[]
  disabled?: boolean
}>(), { pageSizes: () => [5, 10, 25, 50], disabled: false })
const emit = defineEmits<{ 'update:page': [value: number]; 'update:pageSize': [value: number] }>()
const pages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const start = computed(() => props.total ? (props.page - 1) * props.pageSize + 1 : 0)
const end = computed(() => Math.min(props.page * props.pageSize, props.total))
function resize(event: Event) {
  emit('update:pageSize', Number((event.target as HTMLSelectElement).value))
}
</script>

<template>
  <nav class="table-pagination" :aria-label="`Halaman ${label}`">
    <p class="pagination-summary" role="status" aria-live="polite">Menampilkan <b>{{ start }}–{{ end }}</b> dari <b>{{ total }}</b> {{ label }}</p>
    <div class="pagination-controls">
      <label class="pagination-size">Baris
        <select :value="pageSize" :disabled="disabled" :aria-label="`Baris per halaman ${label}`" @change="resize">
          <option v-for="size in pageSizes" :key="size" :value="size">{{ size }}</option>
        </select>
      </label>
      <div class="pagination-pages">
        <button type="button" class="btn btn-sec btn-sm" :disabled="disabled || page <= 1" :aria-label="`Halaman sebelumnya ${label}`" @click="emit('update:page', page - 1)">Sebelumnya</button>
        <span class="page-position" :aria-label="`Halaman ${page} dari ${pages}`">Hal. <b>{{ page }}</b> / {{ pages }}</span>
        <button type="button" class="btn btn-sec btn-sm" :disabled="disabled || page >= pages" :aria-label="`Halaman berikutnya ${label}`" @click="emit('update:page', page + 1)">Berikutnya</button>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.table-pagination{display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:12px 20px;padding:16px 20px;border-top:1px solid var(--border);font-size:11px;color:var(--muted);min-width:0}
.pagination-summary{margin:0;line-height:1.6;overflow-wrap:anywhere}.pagination-summary b{color:var(--text);font-weight:600}
.pagination-controls,.pagination-pages,.pagination-size{display:flex;align-items:center;gap:10px}
.pagination-controls{gap:18px;flex-wrap:wrap}.pagination-size{gap:7px;font-size:11px}
.pagination-size select{min-height:34px;border:1px solid var(--border-2);border-radius:6px;background:var(--surface);color:var(--text);padding:5px 8px;font:inherit}
.pagination-pages .btn{font-size:11px;min-height:34px;padding:7px 10px}.page-position{white-space:nowrap;font-variant-numeric:tabular-nums;font-size:10px}.page-position b{color:var(--text);font-weight:600}
@media(max-width:600px){.table-pagination{padding:14px 12px;gap:10px}.pagination-controls{width:100%;gap:10px}.pagination-pages{display:grid;grid-template-columns:minmax(0,1fr) auto minmax(0,1fr);gap:7px;width:100%}.pagination-pages .btn{padding:8px 5px;min-height:38px}.pagination-size{margin-left:auto}.pagination-size select{min-height:36px}}
</style>
