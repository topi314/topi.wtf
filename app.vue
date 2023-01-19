<template>
    <div>
        <nav-bar />
        <NuxtLayout>
            <NuxtPage />
        </NuxtLayout>
        <footer-bar />
    </div>
</template>
<script lang="ts" setup>
import { useHead } from '@vueuse/head'

const title = 'Topi • backend developer'
const description = '22 y/o backend developer from Germany'
const image = 'https://domain.com/my-image.jpg'
const url = 'https://topi.wtf'
const twitter = '@TopiSenpai'

useHead({
    title,
    meta: [
        {
            hid: 'description',
            property: 'description',
            content: description
        },
        {
            hid: 'og-type',
            property: 'og:type',
            content: 'website'
        },
        {
            hid: 'og-title',
            property: 'og:title',
            content: title
        },
        {
            hid: 'og-description',
            property: 'og:description',
            content: description
        },
        {
            hid: 'og-image',
            property: 'og:image',
            content: image
        },
        {
            hid: 'og-url',
            property: 'og:url',
            content: url
        },
        {
            hid: 't-type',
            name: 'twitter:card',
            content: 'summary_large_image'
        },
        {
            hid: 't-title',
            name: 'twitter:title',
            content: title
        },
        {
            hid: 't-description',
            name: 'twitter:description',
            content: description
        },
        {
            hid: 't-site',
            name: 'twitter:site',
            content: twitter
        }
    ]
})
</script>
<style lang="less">
#__nuxt {
    display: flex;
    flex-direction: column;
    width: 30rem;
    margin-left: auto;
    margin-right: auto;
}
</style>
