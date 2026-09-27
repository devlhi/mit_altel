<script setup lang="ts">
const props = defineProps<{ modelValue: any[], accounts: any[] }>()
const emit = defineEmits(['update:modelValue'])
function update(i: number, key: string, value: any) { emit('update:modelValue', props.modelValue.map((l, n) => n === i ? { ...l, [key]: value } : l)) }
const debit = computed(() => props.modelValue.reduce((s, l) => s + Number(l.debit || 0), 0))
const credit = computed(() => props.modelValue.reduce((s, l) => s + Number(l.credit || 0), 0))
</script>
<template>
  <div class="journal-editor">
    <div v-for="(line, i) in modelValue" :key="i" class="journal-line">
      <label>Akun {{ i + 1 }}<select class="ctrl" :value="line.account" required @change="update(i, 'account', ($event.target as HTMLSelectElement).value)"><option value="">Pilih akun</option><option v-for="a in accounts" :key="a.code" :value="a.code">{{ a.code }} · {{ a.name }}</option></select></label>
      <label>Debit (Rp)<input class="ctrl" type="number" min="0" max="1000000000000" step="1" :value="line.debit" @input="update(i, 'debit', Number(($event.target as HTMLInputElement).value))" /></label>
      <label>Kredit (Rp)<input class="ctrl" type="number" min="0" max="1000000000000" step="1" :value="line.credit" @input="update(i, 'credit', Number(($event.target as HTMLInputElement).value))" /></label>
      <button type="button" class="icon-btn danger" :aria-label="`Hapus baris ${i + 1}`" @click="emit('update:modelValue', modelValue.filter((_, n) => n !== i))"><Ico name="trash" :size="16" /></button>
    </div>
    <button type="button" class="btn btn-sec" @click="emit('update:modelValue', [...modelValue, { account: '', debit: 0, credit: 0 }])">Tambah baris</button>
    <div class="op-note" :class="{ 'op-warning': debit !== credit }">Debit {{ fmtRp(debit) }} · Kredit {{ fmtRp(credit) }}<strong>{{ debit === credit ? 'Seimbang' : `Selisih ${fmtRp(debit - credit)}` }}</strong></div>
  </div>
</template>
