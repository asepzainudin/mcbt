<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { ChevronDown, Search, X } from 'lucide-vue-next'

export interface SearchSelectOption {
  value: string
  label: string
  hint?: string
  disabled?: boolean
}

interface Props {
  modelValue?: string
  options: SearchSelectOption[]
  label?: string
  placeholder?: string
  searchPlaceholder?: string
  error?: string
  required?: boolean
  emptyText?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  label: '',
  placeholder: 'Pilih…',
  searchPlaceholder: 'Cari…',
  error: '',
  required: false,
  emptyText: 'Tidak ada data yang cocok',
})

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const inputId = useId()
const query = ref('')
const open = ref(false)
const wrapperRef = ref<HTMLDivElement | null>(null)

const selectedOption = computed(() =>
  props.options.find((o) => o.value === props.modelValue),
)

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.options
  return props.options.filter(
    (o) =>
      o.label.toLowerCase().includes(q) ||
      (o.hint ?? '').toLowerCase().includes(q),
  )
})

function displayText(): string {
  if (selectedOption.value) return selectedOption.value.label
  return ''
}

function onInput(event: Event) {
  query.value = (event.target as HTMLInputElement).value
  open.value = true
}

function onFocus() {
  query.value = ''
  open.value = true
}

function select(option: SearchSelectOption) {
  if (option.disabled) return
  emit('update:modelValue', option.value)
  query.value = ''
  open.value = false
}

function clear() {
  emit('update:modelValue', '')
  query.value = ''
  open.value = false
  wrapperRef.value?.querySelector('input')?.focus()
}

function onBlur(event: FocusEvent) {
  const next = event.relatedTarget as Node | null
  if (next && wrapperRef.value?.contains(next)) return
  open.value = false
  query.value = ''
}
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" :for="inputId" class="text-sm font-medium leading-none text-foreground">
      {{ label }}
      <span v-if="required" class="text-destructive" title="Wajib diisi" aria-hidden="true">*</span>
    </label>

    <div ref="wrapperRef" class="relative" @focusout="onBlur">
      <div class="relative">
        <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          :id="inputId"
          :value="open ? query : displayText()"
          :placeholder="selectedOption ? selectedOption.label : placeholder"
          :aria-invalid="error !== ''"
          :aria-required="required"
          autocomplete="off"
          role="combobox"
          :aria-expanded="open"
          :class="[
            'flex h-9 w-full rounded-lg border border-input bg-transparent pl-9 pr-16 py-1 text-sm shadow-sm transition-colors',
            'placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            error !== '' && 'border-destructive focus-visible:ring-destructive',
          ]"
          @input="onInput"
          @focus="onFocus"
        />
        <span class="absolute inset-y-0 right-0 flex items-center pr-2">
          <button
            v-if="modelValue"
            type="button"
            class="rounded-md p-1 text-muted-foreground hover:text-foreground"
            aria-label="Hapus pilihan"
            @mousedown.prevent="clear"
          >
            <X class="size-4" />
          </button>
          <ChevronDown class="size-4 text-muted-foreground" />
        </span>
      </div>

      <ul
        v-if="open"
        class="absolute z-40 mt-1 max-h-60 w-full overflow-y-auto rounded-lg border border-border bg-popover py-1 text-popover-foreground shadow-lg"
        role="listbox"
      >
        <li v-if="filtered.length === 0" class="px-3 py-2 text-sm text-muted-foreground">
          {{ emptyText }}
        </li>
        <li v-for="opt in filtered" :key="opt.value">
          <button
            type="button"
            role="option"
            :aria-selected="opt.value === modelValue"
            :disabled="opt.disabled"
            :class="[
              'flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm',
              opt.value === modelValue ? 'bg-primary/10 font-medium text-primary' : 'hover:bg-accent hover:text-accent-foreground',
              opt.disabled && 'cursor-not-allowed opacity-40',
            ]"
            @mousedown.prevent="select(opt)"
          >
            <span class="truncate">{{ opt.label }}</span>
            <span v-if="opt.hint" class="ml-2 shrink-0 text-xs text-muted-foreground">{{ opt.hint }}</span>
          </button>
        </li>
      </ul>
    </div>

    <p v-if="error" class="text-xs font-medium text-destructive">{{ error }}</p>
  </div>
</template>
