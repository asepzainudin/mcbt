<script setup lang="ts">
import { Inbox } from 'lucide-vue-next'

interface Props {
  title?: string
  message?: string
}

withDefaults(defineProps<Props>(), {
  title: 'Tidak ada data',
  message: 'Data belum tersedia saat ini.',
})
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-1 py-14 text-center">
    <div class="mb-3 flex size-12 items-center justify-center rounded-full bg-muted">
      <Inbox class="size-6 text-muted-foreground" />
    </div>
    <p class="font-medium text-foreground">{{ title }}</p>
    <p class="text-sm text-muted-foreground">{{ message }}</p>
    <slot />
  </div>
</template>
