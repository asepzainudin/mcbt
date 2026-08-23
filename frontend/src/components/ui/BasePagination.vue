<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight, MoreHorizontal } from 'lucide-vue-next'

import { cn } from '../../lib/utils'
import BaseButton from './BaseButton.vue'

interface Props {
  page: number
  totalPages: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'change', page: number): void }>()

const pages = computed<number[]>(() => {
  const total = props.totalPages
  const current = props.page
  const window = 2
  const list: number[] = []
  for (let p = Math.max(1, current - window); p <= Math.min(total, current + window); p++) {
    list.push(p)
  }
  return list
})

function go(p: number) {
  if (p >= 1 && p <= props.totalPages && p !== props.page) emit('change', p)
}
</script>

<template>
  <nav v-if="totalPages > 1" class="flex items-center justify-center gap-1" aria-label="Pagination">
    <BaseButton variant="outline" size="icon" :disabled="page <= 1" @click="go(page - 1)">
      <ChevronLeft />
    </BaseButton>

    <button
      v-if="pages[0] !== undefined && pages[0] > 1"
      :class="cn('h-9 rounded-lg px-3 text-sm hover:bg-accent hover:text-accent-foreground')"
      @click="go(1)"
    >
      1
    </button>
    <span v-if="pages[0] !== undefined && pages[0] > 2" class="px-1 text-muted-foreground">
      <MoreHorizontal class="size-4" />
    </span>

    <button
      v-for="p in pages"
      :key="p"
      :class="
        cn(
          'min-w-[36px] rounded-lg px-3 py-1.5 text-sm transition-colors',
          p === page
            ? 'bg-primary font-semibold text-primary-foreground shadow-sm'
            : 'hover:bg-accent hover:text-accent-foreground',
        )
      "
      :aria-current="p === page ? 'page' : undefined"
      @click="go(p)"
    >
      {{ p }}
    </button>

    <span
      v-if="pages.length && pages[pages.length - 1]! < totalPages - 1"
      class="px-1 text-muted-foreground"
    >
      <MoreHorizontal class="size-4" />
    </span>
    <button
      v-if="pages.length && pages[pages.length - 1]! < totalPages"
      class="h-9 rounded-lg px-3 text-sm hover:bg-accent hover:text-accent-foreground"
      @click="go(totalPages)"
    >
      {{ totalPages }}
    </button>

    <BaseButton variant="outline" size="icon" :disabled="page >= totalPages" @click="go(page + 1)">
      <ChevronRight />
    </BaseButton>
  </nav>
</template>
