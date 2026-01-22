import axios from 'axios'
import router from '../router'
import { ElMessage } from 'element-plus'

// 自动获取基础地址：
// 1. 优先使用环境变量
// 2. 开发环境 (DEV): 优先使用 localhost 避免代理拦截，否则使用探测到的主机名
const devBase = `http://localhost:8080/api/v1`
const baseURL = import.meta.env.VITE_API_BASE_URL || (import.meta.env.DEV ? devBase : '/api/v1')

console.log('[API] Base URL:', baseURL)

const service = axios.create({
  baseURL: baseURL,
  timeout: 30000, // 延长超时时间到 30s，防止备份过慢
  withCredentials: false // 避免跨域携带 Cookie 导致的复杂性 (JWT 在 Header 里)
})

// 请求拦截器
service.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
      ElMessage.error('会话已过期，请重新登录')
    }
    return Promise.reject(error)
  }
)

export default service