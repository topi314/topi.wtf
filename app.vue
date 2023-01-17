<template>
  <div>
    <div><NuxtLink href="/">index</NuxtLink> <NuxtLink href="/blog">blog</NuxtLink> <NuxtLink href="/projects">projects</NuxtLink></div>
    <NuxtLayout>
      <NuxtLoadingIndicator />
      <NuxtPage/>
    </NuxtLayout>
  </div>
</template>
