<script setup lang="ts">
definePageMeta({ layout: 'blank' })
const username = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const err = ref('')
async function submit() {
  if (loading.value) return
  loading.value = true; err.value = ''
  try {
    const res = await api<any>('/api/login', { method: 'POST', body: { username: username.value.trim(), password: password.value } })
    setToken(res.token)
    localStorage.setItem('isp_nama', res.nama || res.username || 'Administrator')
    await navigateTo('/')
  } catch (e) { err.value = errMsg(e) }
  finally { loading.value = false }
}
</script>
<template>
  <div class="login-wrap">
    <section class="login-story">
      <div class="brand official-brand"><div class="login-brand-plaque"><BrandLogo :width="280" /></div><span class="brand-sub">Ruang kerja PT Ali Telematika Nusantara</span></div>
      <div><h1>Terhubung lebih baik. Terkelola lebih rapi.</h1><p>Satu ruang kerja untuk pelanggan, mitra, dan setiap langkah operasional layanan internet Anda.</p><div class="network-motif" aria-hidden="true"><span><Ico name="building" :size="26" /></span><i></i><span><Ico name="wifi" :size="26" /></span><i></i><span><Ico name="users" :size="26" /></span></div></div>
      <footer>Mitra, operasional, dan dokumen dalam satu tempat.</footer>
    </section>
    <section class="login-panel"><div class="login-card"><h2>Selamat datang kembali</h2><p class="desc">Masuk untuk melanjutkan pengelolaan layanan.</p>
      <form @submit.prevent="submit"><div v-if="err" class="login-err" role="alert">{{ err }}</div><div class="field"><label for="f-user">Username</label><input id="f-user" v-model="username" class="ctrl" autocomplete="username" placeholder="Masukkan username" required autofocus :disabled="loading" /></div>
        <div class="field"><label for="f-pass">Password</label><div class="password-wrap"><input id="f-pass" v-model="password" class="ctrl" :type="showPassword ? 'text' : 'password'" autocomplete="current-password" placeholder="Masukkan password" required :disabled="loading" /><button type="button" class="icon-btn" :aria-label="showPassword ? 'Sembunyikan password' : 'Tampilkan password'" :aria-pressed="showPassword" @click="showPassword = !showPassword"><Ico name="eye" :size="17" /></button></div></div>
        <button class="btn btn-primary" type="submit" :disabled="loading">{{ loading ? 'Memeriksa akun…' : 'Masuk ke ruang kerja' }}<Ico name="arrow" :size="16" /></button>
      </form><p class="login-help">Butuh akses? Hubungi administrator ISP Anda.</p>
    </div></section>
  </div>
</template>

