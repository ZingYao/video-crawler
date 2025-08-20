import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import UserManagementView from '../views/UserManagementView.vue'
import UserEditView from '../views/UserEditView.vue'
import VideoSourceManagementView from '../views/VideoSourceManagementView.vue'
import VideoSourceEditView from '../views/VideoSourceEditView.vue'
// 懒加载观看历史页面（新增）
const WatchHistoryView = () => import('../views/WatchHistoryView.vue')
import MovieView from '../views/MovieView.vue'
const WatchView = () => import('../views/WatchView.vue')
const NotFoundView = () => import('../views/NotFoundView.vue')

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView,
      meta: { requiresAuth: false }
    },
    {
      path: '/user-management',
      name: 'user-management',
      component: UserManagementView,
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/user/edit/:userId?',
      name: 'user-edit',
      component: UserEditView,
      meta: { requiresAuth: true }
    },
    {
      path: '/profile',
      name: 'profile',
      component: UserEditView,
      meta: { requiresAuth: true }
    },
    {
      path: '/video-source-management',
      name: 'video-source-management',
      component: VideoSourceManagementView,
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/video-source-edit/:id?',
      name: 'video-source-edit',
      component: VideoSourceEditView,
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/history/watch/:userId?',
      name: 'watch-history',
      component: WatchHistoryView,
      meta: { requiresAuth: true }
    },
    {
      path: '/movie',
      name: 'movie',
      component: MovieView,
      meta: { requiresAuth: true }
    },
    {
      path: '/watch',
      name: 'watch',
      component: WatchView,
      meta: { requiresAuth: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: NotFoundView,
      meta: { requiresAuth: false }
    },
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  // 初始化认证状态
  authStore.initAuth()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，重定向到登录页，并带回跳转
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (to.meta.requiresAdmin && !authStore.user?.isAdmin) {
    // 需要管理员权限但用户不是管理员，重定向到首页
    next('/')
  } else if ((to.path === '/login' || to.path === '/register') && authStore.isAuthenticated) {
    // 已登录用户访问登录页或注册页，重定向到首页
    next('/')
  } else {
    next()
  }
})

export default router
