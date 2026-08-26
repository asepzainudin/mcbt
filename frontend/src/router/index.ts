import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useUiStore } from '../stores/ui'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../pages/LoginPage.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../pages/DashboardPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../pages/ProfilePage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/roles',
      name: 'roles',
      component: () => import('../pages/RolesPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/academic-years',
      name: 'academic-years',
      component: () => import('../pages/AcademicYearsPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/classes',
      name: 'classes',
      component: () => import('../pages/ClassesPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/subjects',
      name: 'subjects',
      component: () => import('../pages/SubjectsPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/teachers',
      name: 'teachers',
      component: () => import('../pages/TeachersPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/students',
      name: 'students',
      component: () => import('../pages/StudentsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/exams',
      name: 'exams',
      component: () => import('../pages/ExamsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/exams/:id/sections',
      name: 'exam-sections',
      component: () => import('../pages/ExamSectionsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/exams/:id/review',
      name: 'exam-review',
      component: () => import('../pages/exams/ExamReviewPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/candidate',
      name: 'candidate-exams',
      component: () => import('../pages/CandidateExamsPage.vue'),
      meta: { requiresAuth: true, requiresRole: 'student' },
    },
    {
      path: '/candidate/attempts/:id',
      name: 'candidate-attempt',
      component: () => import('../pages/CandidateAttemptPage.vue'),
      meta: { requiresAuth: true, requiresRole: 'student' },
    },
    {
      path: '/exams/:id/results',
      name: 'exam-results',
      component: () => import('../pages/ExamResultsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/candidate/attempts/:id/discussion',
      name: 'candidate-discussion',
      component: () => import('../pages/CandidateDiscussionPage.vue'),
      meta: { requiresAuth: true, requiresRole: 'student' },
    },
    {
      path: '/question-reports',
      name: 'question-reports',
      component: () => import('../pages/QuestionReportsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/my-results',
      name: 'my-results',
      component: () => import('../pages/StudentResultsPage.vue'),
      meta: { requiresAuth: true, requiresRole: 'student' },
    },
    {
      path: '/exams/:id/results',
      name: 'exam-results',
      component: () => import('../pages/ExamResultsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/exams/:id/answers',
      name: 'exam-answers',
      component: () => import('../pages/ExamAnswersPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/exams/:id/grading',
      name: 'exam-grading',
      component: () => import('../pages/GradingPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/exams/:id/schedule',
      name: 'exam-schedule',
      component: () => import('../pages/ExamSchedulePage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/question-banks',
      name: 'question-banks',
      component: () => import('../pages/QuestionBanksPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/question-banks/:id',
      name: 'question-bank-detail',
      component: () => import('../pages/BankQuestionsPage.vue'),
      meta: { requiresAuth: true, staff: true },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../pages/NotFoundPage.vue'),
      meta: { public: true },
    },
  ],
})

let bootstrapped = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!bootstrapped && !to.meta.public) {
    await auth.bootstrap()
    bootstrapped = true
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }

  if (to.meta.requiresAdmin && !auth.user?.roles.includes('admin')) {
    const ui = useUiStore()
    ui.toastError('Halaman ini hanya untuk admin.')
    return { name: 'dashboard' }
  }

  if (to.meta.staff && !(auth.user?.roles.includes('admin') || auth.user?.roles.includes('teacher'))) {
    const ui = useUiStore()
    ui.toastError('Halaman ini hanya untuk admin dan guru.')
    return { name: 'dashboard' }
  }

  if (to.meta.requiresRole && !auth.user?.roles.includes(to.meta.requiresRole as string)) {
    const ui = useUiStore()
    ui.toastError('Anda tidak memiliki akses ke halaman ini.')
    return { name: 'dashboard' }
  }

  return true
})

export default router
