// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@vueuse/nuxt'
  ],

  runtimeConfig: {
    // Private keys are only available on the server (Nitro)
    apiBase: process.env.NUXT_API_BASE || 'http://localhost:12000',
    public: {
      // Public keys are exposed to the client (Browser)
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:12000',
      discordAppId: process.env.DISCORD_APP_ID || '1505920684256264293'
    }
  },

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  routeRules: {
    '/api/**': {
      cors: true
    }
  },

  compatibilityDate: '2024-07-11',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  }
})
