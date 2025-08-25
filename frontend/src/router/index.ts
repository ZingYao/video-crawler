import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import UserManagementView from '../views/UserManagementView.vue'
import UserEditView from '../views/UserEditView.vue'
import VideoSourceManagementView from '../views/VideoSourceManagementView.vue'
const VideoSourceEditView = () => import('../views/VideoSourceEditView.vue')
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
      meta: { requiresAuth: true, title: '首页' }
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { requiresAuth: false, title: '登录' }
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView,
      meta: { requiresAuth: false, title: '注册' }
    },
    {
      path: '/user-management',
      name: 'user-management',
      component: UserManagementView,
      meta: { requiresAuth: true, requiresAdmin: true, title: '用户管理' }
    },
    {
      path: '/user/edit/:userId?',
      name: 'user-edit',
      component: UserEditView,
      meta: { requiresAuth: true, title: '编辑用户' }
    },
    {
      path: '/profile',
      name: 'profile',
      component: UserEditView,
      meta: { requiresAuth: true, title: '个人中心' }
    },
    {
      path: '/video-source-management',
      name: 'video-source-management',
      component: VideoSourceManagementView,
      meta: { requiresAuth: true, requiresAdmin: true, title: '视频源管理' }
    },
    {
      path: '/video-source-edit/:id?',
      name: 'video-source-edit',
      component: VideoSourceEditView,
      meta: { requiresAuth: true, requiresAdmin: true, title: '编辑视频源' }
    },
    {
      path: '/history/watch/:userId?',
      name: 'watch-history',
      component: WatchHistoryView,
      meta: { requiresAuth: true, title: '观看历史' }
    },
    {
      path: '/movie',
      name: 'movie',
      component: MovieView,
      meta: { requiresAuth: true, title: '观影' }
    },
    {
      path: '/watch/:sourceId',
      name: 'watch',
      component: WatchView,
      meta: { requiresAuth: true, title: '观看' }
    },
    {
      path: '/404',
      name: 'not-found-404',
      component: NotFoundView,
      meta: { requiresAuth: false, title: '页面未找到' }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: NotFoundView,
      meta: { requiresAuth: false, title: '页面未找到' }
    },
  ]
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  const configStore = useConfigStore()
  
  // 如果配置未加载，先加载配置
  if (!configStore.isLoaded) {
    await configStore.loadConfig()
  }
  
  // 初始化认证状态
  authStore.initAuth()
  
  // 如果系统配置为不需要登录，登录注册相关页面展示404
  if (!configStore.needsLogin()) {
    if (to.path === '/login' || to.path === '/register' || to.path === '/user-management' || to.path === '/profile' || to.path.startsWith('/user/edit')) {
      next('/404')
      return
    }
    // 不需要登录时，其他页面都允许访问
    next()
    return
  }
  
  // 需要登录的情况下的原有逻辑
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，重定向到登录页，并带回跳转
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (to.meta.requiresAdmin && !(authStore.user?.isAdmin || authStore.user?.isSiteAdmin)) {
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

// 全局后置钩子：设置浏览器标题
router.afterEach((to) => {
  const baseTitle = 'Video Crawler'
  const t = typeof to.meta.title === 'function' ? (to.meta.title as any)(to) : (to.meta.title as string)
  const dynamic =
    to.name === 'watch' && to.query?.title
      ? `播放 - ${String(to.query.title)}`
      : to.name === 'movie'
      ? (to.query?.q ? `观影 - ${String(to.query.q)}` : '观影')
      : undefined
  const finalTitle = dynamic ?? t
  document.title = [finalTitle, baseTitle].filter(Boolean).join(' | ')
})
