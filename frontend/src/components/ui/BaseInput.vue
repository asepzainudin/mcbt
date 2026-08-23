<script setup lang="ts">
import { computed, useId, ref } from 'vue'
import { Eye, EyeOff } from 'lucide-vue-next'

import { cn } from '../../lib/utils'

interface Props {
  modelValue?: string
  label?: string
  type?: string
  placeholder?: string
  error?: string
  autocomplete?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  label: '',
  type: 'text',
  placeholder: '',
  error: '',
  autocomplete: 'off',
})

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const inputId = useId()
const hasError = computed(() => props.error !== '')
const revealed = ref(false)

const effectiveType = computed(() =>
  props.type === 'password' && revealed.value ? 'text' : props.type,
)
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" :for="inputId" class="text-sm font-medium leading-none text-foreground">
      {{ label }}
    </label>
    <div class="relative">
      <input
        :id="inputId"
        :type="effectiveType"
        :value="modelValue"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :aria-invalid="hasError"
        :class="
          cn(
            'flex h-9 w-full rounded-lg border border-input bg-transparent px-3 py-1 text-sm shadow-sm',
            'placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2',
            'focus-visible:ring-ring transition-colors disabled:cursor-not-allowed disabled:opacity-50',
            type === 'password' && 'pr-10',
            hasError && 'border-destructive focus-visible:ring-destructive',
          )
        "
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <button
        v-if="type === 'password'"
        type="button"
        class="absolute inset-y-0 right-0 flex w-9 items-center justify-center rounded-e-lg text-muted-foreground hover:text-foreground"
        :aria-label="revealed ? 'Sembunyikan password' : 'Tampilkan password'"
        @click="revealed = !revealed"
      >
        <EyeOff v-if="revealed" class="size-4" />
        <Eye v-else class="size-4" />
      </button>
    </div>
    <p v-if="error" class="text-xs font-medium text-destructive">{{ error }}</p>
  </div>
</template>
