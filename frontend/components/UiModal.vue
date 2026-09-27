<script setup lang="ts">
const props = defineProps({
  open: { type: Boolean, default: false }, title: { type: String, default: '' },
  wide: { type: Boolean, default: false }, hg: { type: Boolean, default: false },
  busy: { type: Boolean, default: false }
})
const emit = defineEmits(['close'])
const panel = ref<HTMLElement | null>(null)
const titleId = useId()
let returnFocus: HTMLElement | null = null
let releaseScroll: (() => void) | undefined
let background: { el: HTMLElement, inert: boolean }[] = []
function close() { if (!props.busy) emit('close') }
function cleanup() {
  releaseScroll?.(); releaseScroll = undefined
  background.forEach(({ el, inert }) => { if (el.isConnected) el.inert = inert })
  background = []
  if (returnFocus?.isConnected) returnFocus.focus()
  returnFocus = null
}
watch(() => props.open, async open => {
  if (!import.meta.client) return
  if (!open) { cleanup(); return }
  returnFocus = document.activeElement as HTMLElement
  releaseScroll = lockPageScroll()
  await nextTick()
  if (!props.open || !panel.value) return
  background = [...document.querySelectorAll<HTMLElement>('#__nuxt, .overlay')]
    .filter(el => !el.contains(panel.value))
    .map(el => ({ el, inert: el.inert }))
  background.forEach(({ el }) => { el.inert = true })
  const initialFocus = panel.value.querySelector<HTMLElement>('[autofocus]') || panel.value
  initialFocus.focus()
}, { immediate: true })
onBeforeUnmount(cleanup)
function onKey(event: KeyboardEvent) {
  if (event.key === 'Escape') { event.stopPropagation(); close() }
  if (event.key === 'Tab' && panel.value) trapFocus(event, panel.value)
}
</script>
<template>
  <Teleport to="body">
    <div v-if="open" class="overlay" @mousedown.self="close" @keydown="onKey">
      <div ref="panel" class="modal" :class="{ wide, hg }" role="dialog" aria-modal="true" :aria-labelledby="titleId" :aria-busy="busy" tabindex="-1">
        <div class="modal-head"><h3 :id="titleId">{{ title }}</h3><button class="icon-btn" type="button" aria-label="Tutup dialog" :disabled="busy" @click="close"><Ico name="x" :size="17" /></button></div>
        <div class="modal-body"><slot /></div>
        <div v-if="$slots.footer" class="modal-foot"><slot name="footer" /></div>
      </div>
    </div>
  </Teleport>
</template>
