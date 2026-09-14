const api = require('../../api/index')
const store = require('../../store/index')

Page({
  data: {
    isLoggedIn: false,
    userInfo: null,
    stats: null,
    recentInterviews: []
  },

  onLoad() {
    this.checkLoginStatus()
  },

  onShow() {
    this.checkLoginStatus()
    if (store.isLoggedIn()) {
      this.loadData()
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

  loadData() {
    // 加载统计数据
    api.profile.getStats().then(res => {
      this.setData({ stats: res.data })
    }).catch(err => {
      console.log('加载统计数据失败', err)
    })

    // 加载最近面试
    api.interview.list().then(res => {
      const list = (res.data || []).slice(0, 5)
      this.setData({ recentInterviews: list })
    }).catch(err => {
      console.log('加载面试记录失败', err)
    })
  },

  goLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    })
  },

  goToInterview() {
    wx.navigateTo({
      url: '/pages/interview/create'
    })
  },

  goToPractice() {
    // 跳转到题库页面
    wx.navigateTo({
      url: '/pages/questions/index'
    })
  },

  goToReport() {
    // 跳转到面试列表查看报告
    wx.navigateTo({
      url: '/pages/interview/list'
    })
  },

  goToInterviewDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/interview/detail?id=${id}`
    })
  }
})
