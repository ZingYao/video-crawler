import { ref } from 'vue'
import { defineStore } from 'pinia'
import MD5 from 'crypto-js/md5'
import { userAPI } from '@/api'

export interface User {
  id: string
  username: string
  nickname?: string
  isAdmin?: boolean
  allowLogin?: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  nickname: string
  password: string
}

export interface LoginResponse {
  id: string
  nickname:string
  token: string
  is_admin?: boolean | null
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const isAuthenticated = ref(!!token.value)

  // 真实登录API调用
  const login = async (username: string, password: string): Promise<void> => {
    try {
      // 前端对密码进行MD5加密，避免在请求日志中暴露明文密码
      const encryptedPassword = MD5(password).toString()
      
      const data = await userAPI.login({ username, password: encryptedPassword })
      
      if (data.code !== 0) {
        throw new Error(data.message || '登录失败')
      }

      const loginResponse: LoginResponse = data.data
      
      // 获取用户信息（这里简化处理，实际应该从token中解析或调用用户详情接口）
      const userInfo: User = {
        id: loginResponse.id, // 实际应该从token解析或API获取
        username: username,
        nickname: loginResponse.nickname,
        isAdmin: loginResponse.is_admin || false, // 当 isAdmin 不存在时视为 false
        allowLogin: true
      }
      
      user.value = userInfo
      token.value = loginResponse.token
      isAuthenticated.value = true
      
      // 保存到本地存储
      localStorage.setItem('token', loginResponse.token)
      localStorage.setItem('user', JSON.stringify(userInfo))
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('网络错误，请检查服务器连接')
    }
  }

  // 注册API调用
  const register = async (registerData: RegisterRequest): Promise<void> => {
    try {
      // 前端对密码进行MD5加密，避免在请求日志中暴露明文密码
      const encryptedPassword = MD5(registerData.password).toString()
      
      const data = await userAPI.register({
        username: registerData.username,
        nickname: registerData.nickname,
        password: encryptedPassword
      })
      
      if (data.code !== 0) {
        throw new Error(data.message || '注册失败')
      }
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('网络错误，请检查服务器连接')
    }
  }

  const logout = () => {
    user.value = null
    token.value = null
    isAuthenticated.value = false
    
    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const initAuth = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    
    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
        isAuthenticated.value = true
      } catch (error) {
        console.error('Failed to restore auth state:', error)
        logout()
      }
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    login,
    register,
    logout,
    initAuth
  }
})
