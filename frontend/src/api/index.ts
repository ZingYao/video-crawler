// 导入Wails绑定
import { GetServerPort } from '../../wailsjs/go/main/App'
import { isWails } from '../utils/wails'

// API 基础配置
const getApiBaseUrl = async (): Promise<string> => {
  // 检查是否在Wails环境中
  let isWailsEnv = isWails()
  
  // 如果第一次检测失败，尝试延迟检测
  if (!isWailsEnv) {
    // 等待一段时间后再次检测
    await new Promise(resolve => setTimeout(resolve, 500))
    isWailsEnv = isWails()
  }
  
  // console.log('getApiBaseUrl - 环境检测:', {
  //   isWails: isWailsEnv
  // })
  
  if (isWailsEnv) {
    // 在Wails环境中，动态获取后端服务端口
    try {
      const port = await GetServerPort()
      console.log('getApiBaseUrl - Wails端口获取:', { port })
      
      if (port && port > 0) {
        const result = `http://localhost:${port}`
        console.log('getApiBaseUrl - 返回Wails URL:', result)
        return result
      }
    } catch (error) {
      console.warn('无法获取Wails服务端口，使用默认端口:', error)
    }
    // 回退到默认端口
    const fallback = 'http://localhost:8080'
    console.log('getApiBaseUrl - 回退到默认端口:', fallback)
    return fallback
  }
  
  // 在浏览器环境中，使用当前域名
  const browserUrl = window.location.origin
  console.log('getApiBaseUrl - 返回浏览器 URL:', browserUrl)
  return browserUrl
}

import router from '@/router'
import { useAuthStore } from '@/stores/auth'
import { message } from 'ant-design-vue'

// 业务码处理
function handleBusinessCode(result: any) {
  if (result && typeof result === 'object' && 'code' in result && result.code === 6) {
    // 登录过期
    const DURATION = 1.5 // 秒
    message.error('登录已过期，请重新登录', DURATION)
    const auth = useAuthStore()
    const currentPath = router.currentRoute.value.fullPath
    auth.logout()
    // 等待提示展示后再跳转
    setTimeout(() => {
      router.push({ path: '/login', query: { redirect: currentPath } })
    }, DURATION * 1000)
  }
}

// 通用请求方法
async function request<T = any>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const apiBaseUrl = await getApiBaseUrl()
  const url = `${apiBaseUrl}${path}`
  
  // 调试信息（生产环境构建时会自动移除）
  console.log('API请求:', {
    path,
    apiBaseUrl,
    fullUrl: url,
    isWails: !!(window as any).go,
    currentDomain: window.location.origin,
    currentProtocol: window.location.protocol
  })
  
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }
  // 统一拦截响应头中的角色信息并落地到本地缓存
  try {
    const isAdmin = response.headers.get('X-User-Is-Admin')
    const isSiteAdmin = response.headers.get('X-User-Is-Site-Admin')
    if (isAdmin !== null || isSiteAdmin !== null) {
      const userRaw = localStorage.getItem('user')
      if (userRaw) {
        const user = JSON.parse(userRaw)
        if (isAdmin !== null) user.isAdmin = isAdmin === 'true'
        if (isSiteAdmin !== null) user.isSiteAdmin = isSiteAdmin === 'true'
        localStorage.setItem('user', JSON.stringify(user))
      }
    }
  } catch {}

  const json = await response.json()
  handleBusinessCode(json)
  return json as T
}

// 带认证的请求方法
async function authenticatedRequest<T = any>(
  path: string,
  token: string,
  options: RequestInit = {}
): Promise<T> {
  return request<T>(path, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })
}

// 配置相关API
export const configAPI = {
  // 获取系统配置
  getConfig: () => request('/api/config'),
}

// 用户相关API
export const userAPI = {
  // 用户登录
  login: (data: { username: string; password: string }) =>
    request('/api/user/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // 用户注册
  register: (data: { username: string; nickname: string; password: string }) =>
    request('/api/user/register', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // 获取用户详情
  getUserDetail: (token: string, userId?: string) =>
    authenticatedRequest(`/api/user/detail${userId ? `?user_id=${userId}` : ''}`, token),

  // 保存用户信息
  saveUser: (token: string, data: {
    user_id: string
    username: string
    nickname: string
    password?: string
    is_admin?: boolean
    is_site_admin?: boolean
    allow_login?: boolean
  }) =>
    authenticatedRequest('/api/user/save', token, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // 获取用户列表
  getUserList: (token: string) =>
    authenticatedRequest('/api/user/list', token),

  // 删除用户
  deleteUser: (token: string, userId: string) =>
    authenticatedRequest('/api/user/delete', token, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    }),

  // 修改用户登录状态
  changeLoginStatus: (token: string, userId: string, allowLogin: boolean) =>
    authenticatedRequest('/api/user/allow-login-status-change', token, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, allow_login: allowLogin }),
    }),

  // 管理员伪登录（不记录登录历史）
  adminImpersonateLogin: (token: string, userId: string) =>
    authenticatedRequest('/api/user/admin-impersonate-login', token, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId })
    }),
}

// 系统相关API
export const systemAPI = {
  // 健康检查
  health: () => request('/health'),

  // API信息
  apiInfo: (token: string) => authenticatedRequest('/api', token),
}

// 视频源相关API
export const videoSourceAPI = {
  // 获取视频源列表
  getVideoSourceList: (token: string) =>
    authenticatedRequest('/api/video-source/list', token),

  // 获取视频源详情
  getVideoSourceDetail: (token: string, id: string) =>
    authenticatedRequest(`/api/video-source/detail?id=${id}`, token),

  // 保存视频源（创建或更新）
  saveVideoSource: (token: string, data: any) =>
    authenticatedRequest('/api/video-source/save', token, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // 删除视频源
  deleteVideoSource: (token: string, id: string) =>
    authenticatedRequest('/api/video-source/delete', token, {
      method: 'POST',
      body: JSON.stringify({ id }),
    }),

  // 检查视频源资源状态（单个）
  checkStatus: (token: string, id: string) =>
    authenticatedRequest(`/api/video-source/check-status?id=${encodeURIComponent(id)}`, token),

  // 设置视频源状态
  setStatus: (token: string, id: string, status: number) =>
    authenticatedRequest('/api/video-source/set-status', token, {
      method: 'POST',
      body: JSON.stringify({ id, status }),
    }),

  // 导入视频源配置
  importVideoSources: (token: string, data: any[]) =>
    authenticatedRequest('/api/video-source/import', token, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
}

// 历史相关API
export const historyAPI = {
  // 获取观看历史
  getVideoHistory: (token: string, userId: string) =>
    authenticatedRequest(`/api/history/video?user_id=${encodeURIComponent(userId)}`, token),

  // 获取搜索历史（备用）
  getSearchHistory: (token: string, userId: string) =>
    authenticatedRequest(`/api/history/search?user_id=${encodeURIComponent(userId)}`, token),

  // 获取登录历史（备用）
  getLoginHistory: (token: string, userId: string) =>
    authenticatedRequest(`/api/history/login?user_id=${encodeURIComponent(userId)}`, token),
}

// 视频搜索/详情/播放相关API
export const videoAPI = {
  // 搜索（按站点）
  search: (token: string, sourceId: string, keyword: string) =>
    authenticatedRequest(`/api/video/search?source_id=${encodeURIComponent(sourceId)}&keyword=${encodeURIComponent(keyword)}`, token),

  // 详情
  detail: (token: string, sourceId: string, url: string) =>
    authenticatedRequest(`/api/video/detail?source_id=${encodeURIComponent(sourceId)}&url=${encodeURIComponent(url)}`, token),

  // 播放地址
  playUrl: (token: string, sourceId: string, url: string) =>
    authenticatedRequest(`/api/video/url?source_id=${encodeURIComponent(sourceId)}&url=${encodeURIComponent(url)}`, token),
}

// 导出基础请求方法
export { request, authenticatedRequest }
