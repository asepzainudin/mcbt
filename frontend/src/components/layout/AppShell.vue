<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  GraduationCap,
  LayoutDashboard,
  ShieldCheck,
  LogOut,
  Moon,
  Sun,
  Menu,
  X,
  CalendarDays,
  DoorOpen,
  BookOpen,
  Users,
  FileQuestion,
  ClipboardList,
  FileEdit,
  Trophy,
  Flag,
  BarChart3,
} from 'lucide-vue-next'
import { ref } from 'vue'

import { useAuthStore } from '../../stores/auth'
import { useTheme } from '../../composables/useTheme'
import BaseBadge from '../ui/BaseBadge.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const { theme, toggle } = useTheme()

const sidebarOpen = ref(false)

const isAdmin = computed(() => auth.user?.roles.includes('admin') ?? false)
const isTeacher = computed(() => auth.user?.roles.includes('teacher') ?? false)
const isStaff = computed(() => isAdmin.value || isTeacher.value)

const navItems = computed(() => {
  const items = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  ]
  if (auth.user?.roles.includes('student')) {
    items.push({ to: '/candidate', label: 'Ujian Saya', icon: FileEdit })
    items.push({ to: '/my-results', label: 'Hasil Saya', icon: Trophy })
  }
  if (isStaff.value) {
    items.push(
      { to: '/exams', label: 'Ujian', icon: ClipboardList },
      { to: '/question-banks', label: 'Bank Soal', icon: FileQuestion },
      { to: '/laporan-ujian', label: 'Laporan Ujian', icon: BarChart3 },
      { to: '/question-reports', label: 'Laporan Soal', icon: Flag },
    )
  }
  if (isAdmin.value) {
    items.push({ to: '/roles', label: 'Roles', icon: ShieldCheck })
  }
  return items
})

const masterDataItems = computed(() => {
  const items: { to: string; label: string; icon: unknown }[] = []
  if (isAdmin.value) {
    items.push(
      { to: '/academic-years', label: 'Tahun Ajaran', icon: CalendarDays },
      { to: '/classes', label: 'Kelas', icon: DoorOpen },
      { to: '/subjects', label: 'Mata Pelajaran', icon: BookOpen },
      { to: '/teachers', label: 'Guru', icon: Users },
    )
  }
  if (isStaff.value) {
    items.push({ to: '/students', label: 'Siswa', icon: GraduationCap })
  }
  return items
})

const initials = computed(() => {
  const name = auth.user?.name ?? ''
  return name
    .split(' ')
    .map((w) => w[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
})

async function onLogout() {
  await auth.logout()
  router.push('/login')
}

function closeSidebar() {
  sidebarOpen.value = false
}
</script>

<template>
  <div class="flex min-h-screen bg-background text-foreground">
    <Transition
      enter-active-class="duration-200 ease-out"
      enter-from-class="-translate-x-full"
      leave-active-class="duration-150 ease-in"
      leave-to-class="-translate-x-full"
    >
      <div v-if="sidebarOpen" class="fixed inset-0 z-40 bg-black/60 md:hidden" @click="closeSidebar" />
    </Transition>

    <aside
      :class="[
        'fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-white/10 bg-sidebar-gradient text-slate-100 transition-transform duration-200',
        'md:static md:translate-x-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full',
      ]"
    >
      <div class="flex h-14 items-center justify-between border-b border-white/10 px-4">
        <router-link to="/" class="flex items-center gap-2.5" @click="closeSidebar">
          <span class="flex size-8 items-center justify-center rounded-lg bg-brand-gradient text-white shadow-sm">
            <GraduationCap class="size-4.5" />
          </span>
          <span class="text-lg font-bold tracking-tight text-white">MCBT</span>
        </router-link>
        <button class="text-slate-400 hover:text-white md:hidden" @click="closeSidebar">
          <X class="size-5" />
        </button>
      </div>

      <nav class="flex-1 space-y-1 overflow-y-auto p-3">
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          :class="[
            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors [&_svg]:size-4',
            route.path === item.to
                ? 'bg-brand-gradient text-white shadow-sm'
                : 'text-slate-300 hover:bg-white/10 hover:text-white',
          ]"
          @click="closeSidebar"
        >
          <component :is="item.icon" />
          {{ item.label }}
        </router-link>

        <template v-if="masterDataItems.length > 0">
          <p class="px-3 pb-1 pt-5 text-[11px] font-semibold uppercase tracking-wider text-slate-500">
            Master Data
          </p>
          <router-link
            v-for="item in masterDataItems"
            :key="item.to"
            :to="item.to"
            :class="[
              'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors [&_svg]:size-4',
              route.path === item.to
                ? 'bg-brand-gradient text-white shadow-sm'
                : 'text-slate-300 hover:bg-white/10 hover:text-white',
            ]"
            @click="closeSidebar"
          >
            <component :is="item.icon" />
            {{ item.label }}
          </router-link>
        </template>
      </nav>

      <div class="border-t border-white/10 p-3">
        <button
          class="mb-2 flex w-full items-center gap-3 rounded-lg bg-white/10 p-3 text-left transition-colors hover:bg-white/20"
          title="Profil Saya"
          @click="router.push('/profile'); closeSidebar()"
        >
          <span
            class="flex size-9 shrink-0 items-center justify-center rounded-full bg-brand-gradient text-xs font-bold text-white"
          >
            {{ initials }}
          </span>
          <span class="min-w-0 flex-1">
            <span class="block truncate text-sm font-medium text-white">{{ auth.user?.name }}</span>
            <span class="block truncate text-xs text-slate-400">@{{ auth.user?.username }}</span>
          </span>
        </button>
        <button
          class="flex w-full items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition-colors hover:bg-white/10 hover:text-white"
          @click="onLogout"
        >
          <LogOut class="size-4" />
          Keluar
        </button>
      </div>
    </aside>

    <div class="flex min-w-0 flex-1 flex-col">
      <header
        class="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur md:px-6"
      >
        <div class="flex items-center gap-3">
          <button
            class="rounded-lg p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground md:hidden"
            aria-label="Buka menu"
            @click="sidebarOpen = true"
          >
            <Menu class="size-5" />
          </button>
          <h2 class="text-sm font-semibold capitalize text-muted-foreground">
            {{ route.name === 'dashboard' ? '' : String(route.name) }}
          </h2>
        </div>

        <div class="flex items-center gap-2">
          <BaseBadge v-for="role in auth.user?.roles ?? []" :key="role" tone="primary">
            {{ role }}
          </BaseBadge>
          <button
            class="flex items-center gap-2 rounded-full border border-border bg-card py-1 pl-1 pr-3 transition-colors hover:border-primary/40 hover:bg-accent"
            title="Profil Saya"
            @click="router.push('/profile')"
          >
            <span
              class="flex size-7 shrink-0 items-center justify-center rounded-full bg-brand-gradient text-[11px] font-bold text-white"
            >
              {{ initials }}
            </span>
            <span class="hidden max-w-36 truncate text-sm font-medium sm:block">
              {{ auth.user?.name }}
            </span>
          </button>
          <button
            class="rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            :aria-label="theme === 'dark' ? 'Mode terang' : 'Mode gelap'"
            @click="toggle()"
          >
            <Sun v-if="theme === 'dark'" class="size-4" />
            <Moon v-else class="size-4" />
          </button>
        </div>
      </header>

      <main class="flex-1 p-4 md:p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
