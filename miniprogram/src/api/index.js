// API 请求模块

const request = (options) => {
  return new Promise((resolve, reject) => {
    const app = getApp()
    const token = wx.getStorageSync('token')
    
    wx.request({
      url: app.globalData.apiBaseUrl + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      timeout: options.timeout || 30000,
      success: (res) => {
        if (res.statusCode === 200) {
          if (res.data.code === 200) {
            resolve(res.data)
          } else if (res.data.code === 401) {
            // Token 过期，跳转登录
            wx.removeStorageSync('token')
            wx.removeStorageSync('userInfo')
            wx.redirectTo({ url: '/pages/login/login' })
            reject(new Error(res.data.message || '登录已过期'))
          } else {
            wx.showToast({
              title: res.data.message || '请求失败',
              icon: 'none'
            })
            reject(new Error(res.data.message))
          }
        } else {
          wx.showToast({
            title: '服务器错误',
            icon: 'none'
          })
          reject(new Error('服务器错误'))
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '网络连接失败',
          icon: 'none'
        })
        reject(err)
      }
    })
  })
}

// 认证相关 API
export const auth = {
  // 微信登录
  wxLogin: (code) => request({
    url: '/auth/wxlogin',
    method: 'POST',
    data: { code }
  }),
  
  // 用户名密码登录
  login: (account, password) => request({
    url: '/auth/login',
    method: 'POST',
    data: { account, password }
  }),
  
  // 注册
  register: (data) => request({
    url: '/auth/register',
    method: 'POST',
    data
  }),
  
  // 获取当前用户信息
  getMe: () => request({
    url: '/auth/me'
  })
}

// 面试相关 API
export const interview = {
  // 创建面试
  create: (data) => request({
    url: '/interviews',
    method: 'POST',
    data
  }),
  
  // 获取面试列表
  list: () => request({
    url: '/interviews'
  }),
  
  // 获取面试详情
  get: (id) => request({
    url: `/interviews/${id}`
  }),
  
  // 提交回答
  submitAnswer: (id, data) => request({
    url: `/interviews/${id}/answers`,
    method: 'POST',
    data
  }),
  
  // 完成面试
  complete: (id) => request({
    url: `/interviews/${id}/complete`,
    method: 'PUT'
  })
}

// 题库相关 API
export const question = {
  // 获取题库列表
  list: () => request({
    url: '/questions'
  }),
  
  // 获取题目详情
  get: (id) => request({
    url: `/questions/${id}`
  })
}

// 报告相关 API
export const report = {
  // 获取面试报告
  get: (interviewId) => request({
    url: `/reports/${interviewId}`
  }),
  
  // 获取分享报告
  getShare: (token) => request({
    url: `/reports/share/${token}`
  })
}

// 用户资料相关 API
export const profile = {
  get: () => request({
    url: '/profile'
  }),
  
  update: (data) => request({
    url: '/profile',
    method: 'PUT',
    data
  }),
  
  getStats: () => request({
    url: '/profile/stats'
  }),
  
  getTrend: () => request({
    url: '/profile/trend'
  })
}

export default {
  wxLogin: auth.wxLogin,
  login: auth.login,
  register: auth.register,
  getMe: auth.getMe,
  interview,
  question,
  report,
  profile
}
