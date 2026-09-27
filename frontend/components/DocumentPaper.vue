<script setup lang="ts">
defineProps({ kind: { type: String, default: 'PKS' }, company: { type: String, default: 'Ali Telematika' } })
</script>

<template>
  <div class="paper-miniature" :class="kind === 'FAB' ? 'paper-fab' : 'paper-pks'" aria-hidden="true">
    <div class="paper-letterhead"><BrandLogo :width="95" /><span>{{ company }}</span></div>
    <div class="paper-rule"></div>
    <strong class="paper-title">{{ kind === 'PKS' ? 'Perjanjian kerja sama' : 'Faktur bisnis' }}</strong>
    <span class="paper-caption">{{ kind === 'PKS' ? 'Layanan internet' : 'Rekapitulasi bagi hasil' }}</span>
    <div class="paper-lines"><i></i><i></i><i></i></div>
    <div v-if="kind === 'FAB'" class="paper-table"><div></div><i></i><i></i><i></i><b></b></div>
    <template v-else><b class="paper-section">Ruang lingkup kerja sama</b><div class="paper-lines"><i></i><i></i><i></i><i></i></div></template>
    <div class="paper-signatures"><span></span><span></span></div>
  </div>
</template>
