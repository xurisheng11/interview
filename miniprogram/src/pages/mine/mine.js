const app = getApp()
const store = require('../../store/index')
const api = require('../../api/index')

Page({
  data: {
    isLoggedIn: false,
    userInfo: null,
    stats: {
      totalInterviews: 0,
      totalQuestions: 0,
      collections: 0,
      maxScore: 0
    }
  },

  onShow() {
    this.checkLoginStatus()
    if (store.isLoggedIn()) {
      this.loadStats()
    }
  },

  checkLoginStatus() {
    const userInfo = wx.getStorageSync('userInfo')
    const token = wx.getStorageSync('token')
    
    this.setData({
      isLoggedIn: !!token,
      userInfo: userInfo
    })
  },

  loadStats() {
    api.profile.getStats().then(res => {
      this.setData({ stats: res.data || {} })
    }).catch(err => {
      console.log('加载统计数据失败', err)
    })
  },

  goLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    })
  },

  editProfile() {
    // 跳转到编辑资料页面
    wx.navigateTo({
      url: '/pages/profile/index'
    })
  },

  viewHistory() {
    // 面试列表是 tabBar 页面，必须用 switchTab
    wx.switchTab({
      url: '/pages/interview/list'
    })
  },

  viewInterviews() {
    this.viewHistory()
  },

  viewQuestions() {
    // 跳转到题库
    wx.navigateTo({
      url: '/pages/questions/index'
    })
  },

  viewCollections() {
    wx.navigateTo({
      url: '/pages/collections/index'
    })
  },

  viewContributions() {
    wx.navigateTo({
      url: '/pages/contribution/my'
    })
  },

  viewResume() {
    wx.navigateTo({
      url: '/pages/resume/list'
    })
  },

  viewLearn() {
    wx.navigateTo({
      url: '/pages/learn/index'
    })
  },

  viewCommunity() {
    wx.navigateTo({
      url: '/pages/community/index'
    })
  },

  viewCompanyIntel() {
    wx.navigateTo({
      url: '/pages/company/intel'
    })
  },

  viewCompanyQuestions() {
    wx.navigateTo({
      url: '/pages/contribution/company'
    })
  },

  viewSettings() {
    // 跳转到设置页面
    wx.navigateTo({
      url: '/pages/settings/index'
    })
  },

  viewHelp() {
    // 跳转到帮助页面
    wx.navigateTo({
      url: '/pages/help/index'
    })
  },

  viewAbout() {
    wx.showModal({
      title: '关于我们',
      content: '面试模拟系统 v1.0.0\n\nAI 驱动的面试练习平台\n\n帮助用户更好地准备面试\n\n© 2024 All Rights Reserved',
      showCancel: false
    })
  },

  logout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          app.logout()
          store.clearAuth()
          this.setData({
            isLoggedIn: false,
            userInfo: null,
            stats: {
              totalInterviews: 0,
              totalQuestions: 0,
              collections: 0,
              maxScore: 0
            }
          })
          wx.showToast({
            title: '已退出登录',
            icon: 'success'
          })
        }
      }
    })
  }
})
