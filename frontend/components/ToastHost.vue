<template>
  <div class="toasts" role="status" aria-live="polite" aria-atomic="false">
    <div v-for="t in list" :key="t.id" class="toast" :class="'t-' + t.type">
      <Ico :name="t.type === 'error' ? 'alert' : t.type === 'info' ? 'info' : 'check'" :size="15" />
      <span>{{ t.msg }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
const list = useToastList()
</script>
