<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

interface Props {
  message?: string
}

withDefaults(defineProps<Props>(), { message: 'Memuat data…' })
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-3 py-14 text-muted-foreground">
    <Loader2 class="size-8 animate-spin text-primary" />
    <p class="text-sm">{{ message }}</p>
  </div>
</template>
