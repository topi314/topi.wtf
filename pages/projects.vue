<template>
    <div>projects</div>
</template>
<script lang="ts" setup>
import { useHead } from '@vueuse/head'

const title = "Topi's • Projects"
const description = 'Projects'
const url = 'https://topi.wtf/blog'

useHead({
    title,
    meta: [
        {
            hid: 'description',
            property: 'description',
            content: description
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
            hid: 'og-url',
            property: 'og:url',
            content: url
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
        }
    ]
})
</script>
