import axios from 'axios'
import { Message } from 'element-ui'
import router from '@/router'

// 管理后台专用 axios 实例，与主应用完全隔离
const adminRequest = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL || '/api/v1',
  timeout: 30000
})

// 请求拦截器：携带 admin_token
adminRequest.interceptors.request.use(
  config => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器：401 跳管理后台登录页
adminRequest.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code && res.code !== 200) {
      Message.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message))
    }
    return res
  },
  error => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401 || status === 403) {
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_user')
        router.push('/admin/login')
        Message.warning('管理员登录已过期，请重新登录')
      } else {
        Message.error(data.message || '服务器错误，请稍后重试')
      }
    } else {
      Message.error('网络连接失败')
    }
    return Promise.reject(error)
  }
)

export default adminRequest
