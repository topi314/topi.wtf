<template>
    <nav>
        <NuxtLink href="/">index</NuxtLink>
        <NuxtLink href="/blog">blog</NuxtLink>
        <NuxtLink href="/projects">projects</NuxtLink>
        <select v-model="$colorMode.preference">
            <option value="system">System</option>
            <option value="light">Light</option>
            <option value="dark">Dark</option>
            <option value="sepia">Sepia</option>
        </select>
    </nav>
</template>
<style lang="less" scoped>
nav {
    padding: 2rem;
    display: flex;
    align-items: center;
    * {
        margin: 0.5rem;
        padding: 0.5rem;
    }
}
</style>
