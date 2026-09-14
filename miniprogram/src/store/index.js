// 简单状态管理
const NOT_FOUND = '__NOT_FOUND__'

class Store {
  constructor() {
    this.state = {
      user: wx.getStorageSync('userInfo') || null,
      token: wx.getStorageSync('token') || null,
      interviewHistory: []
    }
    this.listeners = []
  }

  // 获取状态
  getState() {
    return this.state
  }

  // 更新状态
  setState(newState) {
    this.state = { ...this.state, ...newState }
    this.notify()
  }

  // 订阅状态变化
  subscribe(listener) {
    this.listeners.push(listener)
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener)
    }
  }

  // 通知所有监听器
  notify() {
    this.listeners.forEach(listener => listener(this.state))
  }

  // 设置用户信息
  setUser(user) {
    this.state.user = user
    if (user) {
      wx.setStorageSync('userInfo', user)
    } else {
      wx.removeStorageSync('userInfo')
    }
    this.notify()
  }

  // 设置 Token
  setToken(token) {
    this.state.token = token
    if (token) {
      wx.setStorageSync('token', token)
    } else {
      wx.removeStorageSync('token')
    }
    this.notify()
  }

  // 清除登录状态
  clearAuth() {
    this.state.user = null
    this.state.token = null
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('token')
    this.notify()
  }

  // 设置面试历史
  setInterviewHistory(list) {
    this.state.interviewHistory = list
    this.notify()
  }

  // 添加面试记录
  addInterview(interview) {
    this.state.interviewHistory.unshift(interview)
    this.notify()
  }

  // 检查是否已登录
  isLoggedIn() {
    return !!this.state.token
  }
}

const store = new Store()

module.exports = store
