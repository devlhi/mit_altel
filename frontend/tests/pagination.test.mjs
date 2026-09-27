import test from 'node:test'
import assert from 'node:assert/strict'
import { computed, effectScope, ref } from 'vue'
import { usePagination } from '../composables/usePagination.ts'

function setup(t, source, options) {
  const scope = effectScope()
  t.after(() => scope.stop())
  return scope.run(() => usePagination(source, options))
}
const records = n => Array.from({length:n}, (_,i) => ({id:i+1}))

test('walking every page preserves row order without skipping, repeating or mutating data', t => {
  const items = ref(records(23))
  const before = JSON.stringify(items.value)
  const pages = setup(t, items)
  const visited = []
  for (let n=1; n<=pages.totalPages; n++) {
    pages.page = n
    visited.push(...pages.rows.map(row => row.id))
    assert.ok(pages.rows.length <= pages.pageSize)
  }
  assert.deepEqual(visited, items.value.map(row => row.id))
  assert.equal(pages.rows.length,3)
  assert.equal(JSON.stringify(items.value),before)
  pages.page--
  assert.deepEqual(pages.rows.map(row => row.id),records(10).map(row => row.id+10))
})

test('changing a filter resets the page even when the new result has the same length', t => {
  const filter = ref('first')
  const items = computed(() => records(24).filter(row => filter.value === 'first' ? row.id<=12 : row.id>12))
  const pages = setup(t, items, {resetOn:filter})
  pages.page = 2
  filter.value = 'second'
  assert.equal(pages.page,1)
  assert.equal(pages.total,12)
  assert.equal(pages.rows[0].id,13)
})

test('deleting the final page clamps to the last available page', t => {
  const items = ref(records(21))
  const pages = setup(t,items)
  pages.page = 3
  items.value.pop()
  assert.equal(pages.page,2)
  assert.equal(pages.rows.length,10)
  items.value.push({id:22})
  assert.equal(pages.page,2)
})

test('changing page size resets to page one while keeping the full total', t => {
  const pages = setup(t,ref(records(32)))
  pages.page = 3
  pages.pageSize = 25
  assert.equal(pages.page,1)
  assert.equal(pages.total,32)
  assert.equal(pages.totalPages,2)
  assert.equal(pages.rows.length,25)
})

test('empty and single-page tables stay on page one', t => {
  const items = ref(records(12))
  const pages = setup(t,items)
  pages.page = 2
  items.value = []
  assert.equal(pages.page,1)
  assert.equal(pages.totalPages,1)
  assert.equal(pages.total,0)
  assert.deepEqual(pages.rows,[])
  items.value = records(4)
  assert.equal(pages.totalPages,1)
  assert.equal(pages.rows.length,4)
})

test('tables have independent state and changing the detail record resets its page', t => {
  const detail = ref({id:1,items:records(20)})
  const first = setup(t,()=>detail.value.items,{pageSize:5,resetOn:()=>detail.value.id})
  const second = setup(t,ref(records(20)))
  first.page = 3
  assert.equal(second.page,1)
  detail.value = {id:2,items:records(20)}
  assert.equal(first.page,1)
  assert.equal(first.pageSize,5)
})
