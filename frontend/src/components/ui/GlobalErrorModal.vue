<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'

import { useUiStore } from '../../stores/ui'
import BaseModal from './BaseModal.vue'
import BaseButton from './BaseButton.vue'

const ui = useUiStore()

function reload() {
  window.location.reload()
}
</script>

<template>
  <BaseModal open title="Terjadi Kesalahan" @close="ui.closeErrorModal()">
    <div class="flex items-start gap-3">
      <div
        class="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive/15 text-destructive"
      >
        <AlertTriangle class="size-5" />
      </div>
      <p class="text-sm leading-relaxed text-muted-foreground">
        {{ ui.errorModalMessage }}
      </p>
    </div>
    <template #footer>
      <BaseButton variant="outline" @click="ui.closeErrorModal()">Tutup</BaseButton>
      <BaseButton @click="reload">Muat Ulang Halaman</BaseButton>
    </template>
  </BaseModal>
</template>
