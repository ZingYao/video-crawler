// API 基础配置
const API_BASE_URL = window.location.origin

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
  const url = `${API_BASE_URL}${path}`
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
