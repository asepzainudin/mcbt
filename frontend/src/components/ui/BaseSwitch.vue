<script setup lang="ts">
import { computed, useId } from 'vue'

interface Props {
  modelValue?: boolean
  label?: string
  description?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  label: '',
  description: '',
  disabled: false,
})

const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

const switchId = useId()
const checked = computed(() => props.modelValue)
</script>

<template>
  <div class="flex items-start justify-between gap-4 py-1">
    <div class="min-w-0">
      <label :for="switchId" class="block cursor-pointer text-sm font-medium text-foreground">
        {{ label }}
      </label>
      <p v-if="description" class="mt-0.5 text-xs leading-relaxed text-muted-foreground">
        {{ description }}
      </p>
    </div>
    <button
      :id="switchId"
      type="button"
      role="switch"
      :aria-checked="checked"
      :disabled="disabled"
      :class="[
        'relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
        'disabled:cursor-not-allowed disabled:opacity-50',
        checked ? 'bg-primary' : 'bg-gray-200 dark:bg-gray-700',
      ]"
      @click="emit('update:modelValue', !checked)"
    >
      <span
        :class="[
          'inline-block h-5 w-5 transform rounded-full bg-white shadow transition-transform',
          checked ? 'translate-x-[22px]' : 'translate-x-0.5',
        ]"
      />
    </button>
  </div>
</template>
