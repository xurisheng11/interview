// app.js
const api = require('./api/index')
const store = require('./store/index')

App({
  globalData: {
    userInfo: null,
    token: null,
    apiBaseUrl: 'http://localhost:8080/api/v1'
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
        wx.setStorageSync('token', res.data.token)
        wx.setStorageSync('userInfo', res.data.user)
        return res.data
      }
      throw new Error(res.message || '登录失败')
    })
  },

  logout() {
    this.globalData.token = null
    this.globalData.userInfo = null
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
  },

  isLoggedIn() {
    return !!this.globalData.token
  }
})
