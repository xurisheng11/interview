// app.js
const api = require('./api/index')
const store = require('./store/index')

App({
  globalData: {
    userInfo: null,
    token: null,
    // 后端经 cpolar 隧道暴露到公网（本地调试：http://localhost:8080/api/v1）
    apiBaseUrl: 'https://4f47a784.r2.cpolar.top/api/v1'
  },

  onLaunch() {
    // 检查登录状态
    this.checkLogin()
  },

  checkLogin() {
    const token = wx.getStorageSync('token')
    const userInfo = wx.getStorageSync('userInfo')
    
    if (token && userInfo) {
      this.globalData.token = token
      this.globalData.userInfo = userInfo
    }
  },

  login(code) {
    return api.wxLogin(code).then(res => {
      if (res.code === 200) {
        this.globalData.token = res.data.token
        this.globalData.userInfo = res.data.user
        // 同步 store 内存态，保持会话内登录状态一致
        store.setToken(res.data.token)
        store.setUser(res.data.user)
        return res.data
      }
      throw new Error(res.message || '登录失败')
    })
  },

  logout() {
    this.globalData.token = null
    this.globalData.userInfo = null
    // clearAuth 同时清内存与本地缓存
    store.clearAuth()
  },

  isLoggedIn() {
    return !!this.globalData.token
  }
})
