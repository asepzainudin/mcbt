<script setup lang="ts">
import { CheckCircle2, Info, TriangleAlert, X, XCircle } from 'lucide-vue-next'
import type { Component } from 'vue'

import { useUiStore, type ToastType } from '../../stores/ui'

const ui = useUiStore()

const icons: Record<ToastType, Component> = {
  success: CheckCircle2,
  error: XCircle,
  warning: TriangleAlert,
  info: Info,
}

const styles: Record<ToastType, string> = {
  success: 'border-success/30 bg-card text-card-foreground',
  error: 'border-destructive/40 bg-card text-card-foreground',
  warning: 'border-warning/40 bg-card text-card-foreground',
  info: 'border-primary/40 bg-card text-card-foreground',
}

const iconStyles: Record<ToastType, string> = {
  success: 'bg-success/15 text-success',
  error: 'bg-destructive/15 text-destructive',
  warning: 'bg-warning/20 text-warning',
  info: 'bg-primary/15 text-primary',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed right-4 top-4 z-[60] flex w-80 flex-col gap-2">
      <TransitionGroup
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-x-6 opacity-0"
        leave-active-class="transition duration-150 ease-in"
        leave-to-class="opacity-0"
      >
        <div
          v-for="toast in ui.toasts"
          :key="toast.id"
          :class="[
            'flex items-start gap-3 rounded-xl border p-3.5 shadow-lg backdrop-blur',
            styles[toast.type],
          ]"
          role="alert"
        >
          <span
            :class="[
              'mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full [&_svg]:size-3.5',
              iconStyles[toast.type],
            ]"
          >
            <component :is="icons[toast.type]" />
          </span>
          <p class="flex-1 pt-0.5 text-sm leading-snug">{{ toast.message }}</p>
          <button
            class="pt-1 text-muted-foreground transition-colors hover:text-foreground"
            @click="ui.dismiss(toast.id)"
          >
            <X class="size-4" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
