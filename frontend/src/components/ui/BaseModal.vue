<script setup lang="ts">
import { X } from 'lucide-vue-next'

interface Props {
  open: boolean
  title?: string
}

const props = withDefaults(defineProps<Props>(), { title: '' })
const emit = defineEmits<{ (e: 'close'): void }>()

function onBackdrop() {
  if (props.open) emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="duration-200 ease-out"
      enter-from-class="opacity-0"
      leave-active-class="duration-150 ease-in"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="onBackdrop" />
        <div
          class="relative z-10 w-full max-w-md rounded-xl border border-border bg-popover p-6 text-popover-foreground shadow-2xl"
          role="dialog"
          aria-modal="true"
        >
          <div class="mb-3 flex items-start justify-between gap-4">
            <h2 class="text-lg font-semibold">{{ title }}</h2>
            <button
              class="rounded-md p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
              aria-label="Tutup"
              @click="emit('close')"
            >
              <X class="size-4" />
            </button>
          </div>
          <slot />
          <template v-if="$slots.footer">
            <div class="mt-5 flex justify-end gap-2">
              <slot name="footer" />
            </div>
          </template>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
