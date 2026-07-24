import axios from 'axios'
import { Message } from 'element-ui'
import router from '@/router'

const request = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL || '/api/v1',
  timeout: 30000
})

// 请求拦截器：自动携带 token
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器：统一处理错误
request.interceptors.response.use(
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
      if (status === 401) {
        localStorage.removeItem('token')
        const redirect = router.currentRoute.fullPath
        router.push(`/login?redirect=${encodeURIComponent(redirect)}`)
        Message.warning('登录已过期，请重新登录')
      } else if (status === 429) {
        Message.warning(data.message || '操作过于频繁，请稍后再试')
      } else {
        Message.error(data.message || '服务器错误，请稍后重试')
      }
    } else {
      Message.error('网络连接失败')
    }
    return Promise.reject(error)
  }
)

export default request
