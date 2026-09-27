let scrollLocks = 0
let previousOverflow = ''
export function lockPageScroll() {
  if (scrollLocks++ === 0) {
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  }
  let released = false
  return () => {
    if (released) return
    released = true
    if (--scrollLocks === 0) document.body.style.overflow = previousOverflow
  }
}
export function trapFocus(event: KeyboardEvent, container: HTMLElement) {
  const elements = [...container.querySelectorAll<HTMLElement>('a[href], button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled), iframe, [tabindex="0"]')].filter(el => el.getClientRects().length > 0)
  const first = elements[0]
  const last = elements[elements.length - 1]
  if (!first) { event.preventDefault(); container.focus(); return }
  if (event.shiftKey && (document.activeElement === first || document.activeElement === container || !container.contains(document.activeElement))) {
    event.preventDefault(); last.focus()
  } else if (!event.shiftKey && (document.activeElement === last || !container.contains(document.activeElement))) {
    event.preventDefault(); first.focus()
  }
}
