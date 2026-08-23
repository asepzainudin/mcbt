<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RefreshCw } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import {
  api,
  apiErrorMessage,
} from '../lib/axios'
import type { ApiResponse, PaginationMeta, RoleItem } from '../types/api'
import {
  BaseBadge,
  BaseButton,
  BasePagination,
  BaseSelect,
  BaseTable,
  EmptyState,
  LoadingState,
} from '../components/ui'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()

const roles = ref<RoleItem[]>([])
const meta = ref<PaginationMeta | null>(null)
const loading = ref(false)
const page = ref(1)
const limit = ref(10)

const limitOptions = [
  { value: '5', label: '5 / halaman' },
  { value: '10', label: '10 / halaman' },
  { value: '25', label: '25 / halaman' },
]

async function fetchRoles() {
  loading.value = true
  try {
    const res = await api.get<ApiResponse<RoleItem[]>>('/roles', {
      params: { page: page.value, limit: limit.value },
    })
    roles.value = res.data.data
    meta.value = res.data.meta ?? null
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat daftar role.'))
  } finally {
    loading.value = false
  }
}

watch(limit, () => {
  page.value = 1
})

onMounted(fetchRoles)

const rangeText = computed(() => {
  if (!meta.value) return ''
  const start = (meta.value.page - 1) * meta.value.limit + 1
  const end = Math.min(meta.value.page * meta.value.limit, meta.value.total_items)
  return `Menampilkan ${start}–${end} dari ${meta.value.total_items} role`
})
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Daftar Role</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Kelola role untuk kontrol akses RBAC — hanya admin.
          </p>
        </div>
        <div class="flex items-end gap-2">
          <BaseSelect v-model="limit" :options="limitOptions" />
          <BaseButton variant="outline" size="icon" aria-label="Muat ulang" @click="fetchRoles()">
            <RefreshCw :class="loading && 'animate-spin'" />
          </BaseButton>
        </div>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <EmptyState
          v-if="roles.length === 0"
          title="Belum ada role"
          message="Role bawaan dibuat oleh migration seeder."
        >
          <BaseButton variant="outline" size="sm" class="mt-3" @click="fetchRoles()">
            Muat ulang
          </BaseButton>
        </EmptyState>

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                #
              </th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                ID
              </th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Code
              </th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Status
              </th>
            </template>

            <tr v-for="(role, i) in roles" :key="role.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3 text-muted-foreground">
                {{ ((meta?.page ?? 1) - 1) * (meta?.limit ?? limit) + i + 1 }}
              </td>
              <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ role.id }}</td>
              <td class="px-4 py-3">
                <span class="font-medium">{{ role.code }}</span>
                <BaseBadge :tone="role.code === 'admin' ? 'primary' : role.code === 'teacher' ? 'info' : 'success'" class="ml-2">
                  {{ role.code }}
                </BaseBadge>
              </td>
              <td class="px-4 py-3 text-right">
                <BaseBadge tone="outline">aktif</BaseBadge>
              </td>
            </tr>
          </BaseTable>

          <div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
            <p class="text-sm text-muted-foreground">{{ rangeText }}</p>
            <BasePagination
              v-if="meta"
              :page="meta.page"
              :total-pages="meta.total_pages"
              @change="
                (p) => {
                  page = p
                  fetchRoles()
                }
              "
            />
          </div>
        </template>
      </template>
    </div>
  </AppShell>
</template>
