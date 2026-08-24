<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import {
  Bold,
  Code,
  Italic,
  List,
  ListOrdered,
  RemoveFormatting,
  Underline,
} from 'lucide-vue-next'

interface Props {
  modelValue?: string
  placeholder?: string
  minHeight?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: 'Tulis soal di sini…',
  minHeight: '120px',
})

const emit = defineEmits<{ (e: 'update:modelValue', html: string): void }>()

const editorRef = ref<HTMLDivElement | null>(null)
const showSource = ref(false)
const sourceText = ref('')

onMounted(() => {
  if (editorRef.value) {
    editorRef.value.innerHTML = props.modelValue ?? ''
  }
})

watch(
  () => props.modelValue,
  (val) => {
    if (editorRef.value && editorRef.value.innerHTML !== val) {
      editorRef.value.innerHTML = val ?? ''
    }
  },
)

function emitHtml() {
  if (editorRef.value) {
    emit('update:modelValue', editorRef.value.innerHTML)
  }
}

function exec(command: string, value?: string) {
  editorRef.value?.focus()
  document.execCommand(command, false, value)
  emitHtml()
}

function openSource() {
  if (!showSource.value && editorRef.value) {
    sourceText.value = editorRef.value.innerHTML
  }
  showSource.value = !showSource.value
}

function applySource() {
  if (editorRef.value) {
    editorRef.value.innerHTML = sourceText.value
    emitHtml()
  }
}

function onInput() {
  emitHtml()
}

const tools = [
  { icon: Bold, command: 'bold', title: 'Tebal (Ctrl+B)' },
  { icon: Italic, command: 'italic', title: 'Miring (Ctrl+I)' },
  { icon: Underline, command: 'underline', title: 'Garis bawah (Ctrl+U)' },
  { icon: List, command: 'insertUnorderedList', title: 'Daftar poin' },
  { icon: ListOrdered, command: 'insertOrderedList', title: 'Daftar bernomor' },
  { icon: RemoveFormatting, command: 'removeFormat', title: 'Hapus format' },
]
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-input bg-transparent shadow-sm focus-within:ring-2 focus-within:ring-ring">
    <div class="flex items-center gap-0.5 border-b border-border bg-muted/50 px-1.5 py-1">
      <button
        v-for="tool in tools"
        :key="tool.command"
        type="button"
        class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
        :title="tool.title"
        @click="exec(tool.command)"
      >
        <component :is="tool.icon" class="size-4" />
      </button>
      <span class="mx-1 h-4 w-px bg-border" />
      <button
        type="button"
        :class="[
          'flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium transition-colors',
          showSource ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
        ]"
        @click="openSource"
      >
        <Code class="size-3.5" />
        HTML
      </button>
    </div>

    <div
      v-show="!showSource"
      ref="editorRef"
      :contenteditable="true"
      :style="{ minHeight }"
      class="w-full px-3 py-2 text-sm outline-none prose-sm [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:list-disc [&_ul]:pl-5 [&_p]:my-1"
      :data-placeholder="placeholder"
      @input="onInput"
      @blur="emitHtml"
    />

    <textarea
      v-show="showSource"
      v-model="sourceText"
      :style="{ minHeight }"
      class="w-full resize-y px-3 py-2 font-mono text-xs outline-none"
      spellcheck="false"
      @change="applySource"
    />
  </div>
</template>

<style scoped>
[contenteditable][data-placeholder]:empty::before {
  content: attr(data-placeholder);
  color: var(--color-muted-foreground);
  pointer-events: none;
}
</style>
