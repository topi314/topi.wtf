<template>
    <nav>
        <NuxtLink href="/">About Me</NuxtLink>
        <NuxtLink href="/blog">Blog</NuxtLink>
        <NuxtLink href="/projects">Projects</NuxtLink>
        <select v-model="$colorMode.preference">
            <option value="system">System</option>
            <option value="dark">Dark</option>
            <option value="light">Light</option>
            <option value="sepia">Sepia</option>
        </select>
    </nav>
</template>
<style lang="less" scoped>
@import '@/assets/less/global.less';

nav {
    padding: 2rem;
    display: flex;
    align-items: center;
    a {
        margin: 0.5rem;
        padding: 0.5rem;
        text-decoration: none;
        background-color: @bg-secondary;
        border-radius: 4px;
        :hover {
            background-color: darken(@bg-secondary, 10%);
        }
    }
}
</style>
