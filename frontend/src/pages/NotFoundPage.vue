<script setup lang="ts">
import { Compass } from 'lucide-vue-next'
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center gap-4 bg-background px-4 text-center">
    <div class="flex size-16 items-center justify-center rounded-2xl bg-primary/10">
      <Compass class="size-8 text-primary" />
    </div>
    <p class="text-6xl font-bold tracking-tight">404</p>
    <p class="text-muted-foreground">Halaman yang kamu cari tidak ditemukan.</p>
    <router-link
      to="/"
      class="inline-flex h-9 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
    >
      Kembali ke Dashboard
    </router-link>
  </div>
</template>
