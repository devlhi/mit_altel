<script setup lang="ts">
const props = defineProps<{ rows: any[], label: string, period?: boolean }>()
const pages = usePagination(() => props.rows)
</script>
<template>
  <div class="table-wrap"><table class="tbl"><thead><tr><th>Akun</th><th class="num">{{ period ? 'Nilai periode' : 'Saldo akhir' }}</th></tr></thead><tbody><tr v-for="a in pages.rows" :key="a.code"><td><span class="cell-sub mono">{{ a.code }}</span> {{ a.name }}</td><td class="num">{{ fmtRp(period ? a.period : a.balance) }}</td></tr><tr v-if="!pages.total"><td colspan="2" class="empty">Belum ada saldo.</td></tr></tbody></table></div>
  <TablePagination v-model:page="pages.page" v-model:page-size="pages.pageSize" :total="pages.total" :label="label" />
</template>
