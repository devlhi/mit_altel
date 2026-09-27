export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: '2025-01-01',
  css: ['~/assets/css/main.css', '~/assets/css/workspace.css', '~/assets/css/operations.css'],
  nitro: {
    devProxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true }
    }
  },
  app: {
    head: {
      title: 'ALTELMATARA NETWORK · PT Ali Telematika Nusantara',
      htmlAttrs: { lang: 'id' },
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'PT Ali Telematika Nusantara — manajemen layanan internet ALTELMATARA NETWORK, mitra, invoice, dan dokumen PKS/FAB.' }
      ],
      link: [{ rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }]
    }
  }
})
