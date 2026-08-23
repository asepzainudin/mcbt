<script setup lang="ts">
import { computed, useId } from 'vue'
import { ChevronDown } from 'lucide-vue-next'

import { cn } from '../../lib/utils'

export interface SelectOption {
  value: string | number
  label: string
}

interface Props {
  modelValue?: string | number
  label?: string
  options: SelectOption[]
  error?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  label: '',
  error: '',
})

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const selectId = useId()
const hasError = computed(() => props.error !== '')
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" :for="selectId" class="text-sm font-medium leading-none text-foreground">
      {{ label }}
    </label>
    <div class="relative">
      <select
        :id="selectId"
        :value="modelValue"
        :aria-invalid="hasError"
        :class="
          cn(
            'flex h-9 w-full appearance-none rounded-lg border border-input bg-transparent px-3 py-1 pr-9 text-sm shadow-sm',
            'transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            'disabled:cursor-not-allowed disabled:opacity-50',
            hasError && 'border-destructive focus-visible:ring-destructive',
          )
        "
        @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
      >
        <option v-for="opt in options" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
      <ChevronDown
        class="pointer-events-none absolute inset-y-0 right-3 my-auto size-4 text-muted-foreground"
      />
    </div>
    <p v-if="error" class="text-xs font-medium text-destructive">{{ error }}</p>
  </div>
</template>
