// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    pages: true,
    css: ['~/assets/less/global.less'],
    nitro: {
        preset: 'node-server'
    }
})
