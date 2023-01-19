// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    pages: true,
    css: ['~/assets/less/global.less'],
    modules: ['@nuxtjs/color-mode'],
    colorMode: {
        preference: 'system',
        fallback: 'dark',
        storageKey: 'theme'
    },
    nitro: {
        preset: 'node-server'
    },
    app: {
        head: {
            charset: 'utf-8',
        }
    }
})
