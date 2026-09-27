<template>
  <UiModal :open="!!st" title="Konfirmasi" @close="resolve(false)">
    <div style="display:flex;gap:14px;align-items:flex-start">
      <div class="stat-ico" :class="destructive ? 'ico-red' : 'ico-blue'"><Ico :name="destructive ? 'alert' : 'info'" :size="20" /></div>
      <p style="font-size:14px">{{ st?.msg }}</p>
    </div>
    <template #footer>
      <button class="btn btn-sec" type="button" @click="resolve(false)">Batal</button>
      <button class="btn" :class="destructive ? 'btn-danger' : 'btn-primary'" type="button" @click="resolve(true)">{{ st?.msg.startsWith('Hapus') ? 'Ya, hapus' : 'Konfirmasi' }}</button>
    </template>
  </UiModal>
</template>

<script setup lang="ts">
const st = useConfirmState()
const destructive = computed(() => /hapus|data contoh|kembalikan/i.test(st.value?.msg || ''))
function resolve(v: boolean) {
  if (st.value) st.value.resolve(v)
  st.value = null
}
</script>
