<script setup lang="ts">
import { AppToasts, GlobalErrorModal } from './components/ui'
import { useUiStore } from './stores/ui'

const ui = useUiStore()
</script>

<template>
  <router-view />
  <AppToasts />
  <GlobalErrorModal v-if="ui.errorModalOpen" />
</template>
