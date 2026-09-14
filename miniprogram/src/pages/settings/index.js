const store = require('../../store/index')

Page({
  data: {
    isLoggedIn: false,
    version: '1.0.0',
    cacheSize: '0 KB',
    settings: {
      notification: true,
      sound: true,
      vibration: true
    }
  },

  onLoad() {
    this.loadSettings()
    this.calculateCacheSize()
    this.checkLoginStatus()
  },

  onShow() {
    this.checkLoginStatus()
  },

  checkLoginStatus() {
    this.setData({ isLoggedIn: store.isLoggedIn() })
  },

  loadSettings() {
    const settings = wx.getStorageSync('settings') || {}
    this.setData({ settings: { ...this.data.settings, ...settings } })
  },

  saveSettings() {
    wx.setStorageSync('settings', this.data.settings)
  },

  calculateCacheSize() {
    try {
      const info = wx.getStorageInfoSync()
      const sizeInKB = info.currentSize
      let cacheSize = ''
      
      if (sizeInKB < 1024) {
        cacheSize = `${sizeInKB} KB`
      } else {
        cacheSize = `${(sizeInKB / 1024).toFixed(2)} MB`
      }
      
      this.setData({ cacheSize })
    } catch (e) {
      this.setData({ cacheSize: '未知' })
    }
  },

  // 通知设置
  onNotificationChange(e) {
    this.setData({
      'settings.notification': e.detail.value
    })
    this.saveSettings()
    wx.showToast({
      title: e.detail.value ? '已开启' : '已关闭',
      icon: 'none'
    })
  },

  toggleNotification() {
    this.setData({
      'settings.notification': !this.data.settings.notification
    })
    this.saveSettings()
  },

  // 声音设置
  onSoundChange(e) {
    this.setData({
      'settings.sound': e.detail.value
    })
    this.saveSettings()
  },

  toggleSound() {
    this.setData({
      'settings.sound': !this.data.settings.sound
    })
    this.saveSettings()
  },

  // 震动设置
  onVibrationChange(e) {
    this.setData({
      'settings.vibration': e.detail.value
    })
    this.saveSettings()
  },

  toggleVibration() {
    this.setData({
      'settings.vibration': !this.data.settings.vibration
    })
    this.saveSettings()
  },

  // 清除缓存
  clearCache() {
    wx.showModal({
      title: '清除缓存',
      content: '确定要清除所有缓存数据吗？',
      success: (res) => {
        if (res.confirm) {
          wx.showLoading({ title: '清除中...' })
          
          setTimeout(() => {
            try {
              wx.clearStorageSync()
              // 重新设置登录信息
              const token = wx.getStorageSync('token')
              const userInfo = wx.getStorageSync('userInfo')
              
              this.calculateCacheSize()
              wx.hideLoading()
              
              wx.showToast({
                title: '清除成功',
                icon: 'success'
              })
            } catch (e) {
              wx.hideLoading()
              wx.showToast({
                title: '清除失败',
                icon: 'none'
              })
            }
          }, 500)
        }
      }
    })
  },

  // 导出数据
  exportData() {
    wx.showModal({
      title: '导出数据',
      content: '数据导出功能开发中...',
      showCancel: false
    })
  },

  // 隐私政策
  viewPrivacy() {
    wx.showModal({
      title: '隐私政策',
      content: '我们严格保护您的个人隐私...\n\n1. 信息收集\n2. 信息使用\n3. 信息存储\n4. 信息共享\n5. Cookie政策',
      showCancel: false
    })
  },

  // 用户协议
  viewTerms() {
    wx.showModal({
      title: '用户协议',
      content: '欢迎使用面试模拟系统\n\n1. 服务条款\n2. 用户权利\n3. 用户义务\n4. 免责声明',
      showCancel: false
    })
  },

  // 检查更新
  checkUpdate() {
    wx.showLoading({ title: '检查中...' })
    
    setTimeout(() => {
      wx.hideLoading()
      wx.showModal({
        title: '检查更新',
        content: '当前已是最新版本',
        showCancel: false
      })
    }, 1000)
  },

  // 关于我们
  viewAbout() {
    wx.showModal({
      title: '关于我们',
      content: `面试模拟系统 v${this.data.version}\n\nAI 驱动的面试练习平台\n\n帮助用户更好地准备面试\n\n© 2024 All Rights Reserved`,
      showCancel: false
    })
  },

  // 退出登录
  logout() {
    wx.showModal({
      title: '退出登录',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          const app = getApp()
          app.logout()
          store.clearAuth()
          
          wx.showToast({
            title: '已退出',
            icon: 'success'
          })
          
          setTimeout(() => {
            wx.reLaunch({ url: '/pages/index/index' })
          }, 1500)
        }
      }
    })
  }
})
